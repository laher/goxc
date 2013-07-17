package tasks

/*
   Copyright 2013 Am Laher

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"fmt"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/archive"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/platforms"
	"github.com/laher/goxc/typeutils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	//"strings"
)

//runs automatically
func init() {
	Register(Task{
		TASK_PKG_BUILD,
		"Build a binary package. Currently only supports .deb format for Debian/Ubuntu Linux.",
		runTaskPkgBuild,
		map[string]interface{}{"metadata": map[string]interface{}{"maintainer": "unknown"}, "metadata-deb": map[string]interface{}{"Depends": ""}, "rmtemp": true}})
}

func runTaskPkgBuild(tp TaskParams) (err error) {
	for _, dest := range tp.DestPlatforms {
		err := pkgBuildPlat(dest.Os, dest.Arch, tp)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	return
}

func pkgBuildPlat(destOs, destArch string, tp TaskParams) (err error) {
	if destOs == platforms.LINUX {
		//TODO rpm
		//TODO sdeb
		return debBuild(destOs, destArch, tp)
	}
	// TODO BSD ports?
	// TODO Mac pkgs?
	// TODO Windows - msi or something? Perhaps build an installer using 'https://github.com/jteeuwen/go-bindata' to pack the compressed executable
	return nil
}

func getDebControlFileContent(appName, maintainer, version, arch, description string, metadataDeb map[string]interface{}) []byte {
	control := fmt.Sprintf("Package: %s\nPriority: Extra\n", appName)
	if maintainer != "" {
		control = fmt.Sprintf("%sMaintainer: %s\n", control, maintainer)
	}
	//mandatory
	control = fmt.Sprintf("%sVersion: %s\n", control, version)

	control = fmt.Sprintf("%sArchitecture: %s\n", control, getDebArch(arch))
	for k, v := range metadataDeb {
		control = fmt.Sprintf("%s%s: %s\n", control, k, v)
	}
	control = fmt.Sprintf("%sDescription: %s\n", control, description)
	return []byte(control)
}

func getDebArch(destArch string) string {
	architecture := "all"
	switch destArch {
	case platforms.X86:
		architecture = "i386"
	case platforms.ARM:
		architecture = "armel"
	case platforms.AMD64:
		architecture = "amd64"
	}
	return architecture
}

func debBuild(destOs, destArch string, tp TaskParams) (err error) {
	metadata := tp.Settings.GetTaskSettingMap(TASK_PKG_BUILD, "metadata")
	metadataDeb := tp.Settings.GetTaskSettingMap(TASK_PKG_BUILD, "metadata-deb")
	rmtemp := tp.Settings.GetTaskSettingBool(TASK_PKG_BUILD, "rmtemp")
	debDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName()) //v0.8.1 dont use platform dir
	tmpDir := filepath.Join(debDir, ".goxc-temp")
	if rmtemp {
		defer os.RemoveAll(tmpDir)
	}
	os.MkdirAll(tmpDir, 0755)
	err = ioutil.WriteFile(filepath.Join(tmpDir, "debian-binary"), []byte("2.0\n"), 0644)
	if err != nil {
		return err
	}
	description := "?"
	if desc, keyExists := metadata["description"]; keyExists {
		description, err = typeutils.ToString(desc, "description")
		if err != nil {
			return err
		}
	}
	maintainer := "?"
	if maint, keyExists := metadata["maintainer"]; keyExists {
		maintainer, err = typeutils.ToString(maint, "maintainer")
		if err != nil {
			return err
		}
	}
	controlContent := getDebControlFileContent(tp.AppName, maintainer, tp.Settings.GetFullVersionName(), destArch, description, metadataDeb)
	if tp.Settings.IsVerbose() {
		log.Printf("Control file:\n%s", string(controlContent))
	}
	err = ioutil.WriteFile(filepath.Join(tmpDir, "control"), controlContent, 0644)
	if err != nil {
		return err
	}
	err = archive.TarGz(filepath.Join(tmpDir, "control.tar.gz"), []archive.ArchiveItem{archive.ArchiveItem{FileSystemPath: filepath.Join(tmpDir, "control"), ArchivePath: "control"}})
	if err != nil {
		return err
	}
	//build
	items := []archive.ArchiveItem{}

	for _, mainDir := range tp.MainDirs {
		exeName := filepath.Base(mainDir)
		relativeBin := core.GetRelativeBin(destOs, destArch, exeName, false, tp.Settings.GetFullVersionName())
		items = append(items, archive.ArchiveItem{FileSystemPath: filepath.Join(tp.OutDestRoot, relativeBin), ArchivePath: "/usr/bin/" + exeName})
	}
	//TODO add resources to /usr/share/appName/
	err = archive.TarGz(filepath.Join(tmpDir, "data.tar.gz"), items)
	if err != nil {
		return err
	}
	targetFile := filepath.Join(debDir, fmt.Sprintf("%s_%s_%s.deb", tp.AppName, tp.Settings.GetFullVersionName(), getDebArch(destArch))) //goxc_0.5.2_i386.deb")
	inputs := [][]string{
		[]string{filepath.Join(tmpDir, "debian-binary"), "debian-binary"},
		[]string{filepath.Join(tmpDir, "control.tar.gz"), "control.tar.gz"},
		[]string{filepath.Join(tmpDir, "data.tar.gz"), "data.tar.gz"}}
	err = archive.ArForDeb(targetFile, inputs)
	return
}
