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
	"github.com/debber/debber-v0.3/deb"
	"github.com/debber/debber-v0.3/debgen"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"fmt"
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/platforms"
	"github.com/laher/goxc/typeutils"
	"log"
	"os"
	"path/filepath"
)

//runs automatically
func init() {
	Register(Task{
		TASK_DEBGEN,
		"Build a .deb package for Debian/Ubuntu Linux.",
		runTaskDebGen,
		map[string]interface{}{"metadata": map[string]interface{}{"maintainer": "unknown"},
			"metadata-deb": map[string]interface{}{"Depends": ""},
			"rmtemp":       true, "armarch": "",
			"sources-glob":        "**.go",
			"sources-dir":         ".",
			"sources-relative-to": "",
		},
	})
}

func runTaskDebGen(tp TaskParams) (err error) {
	for _, dest := range tp.DestPlatforms {
		err := pkgDebPlat(dest, tp)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	return
}

func pkgDebPlat(dest platforms.Platform, tp TaskParams) (err error) {
	if dest.Os == platforms.LINUX {
		return debBuild(dest, tp)
	}
	return nil
}


// TODO fix armhf/armel distinction ...
func getDebArch(destArch string, armArchName string) deb.Architecture {
	var architecture deb.Architecture
	switch destArch {
	case platforms.X86:
		architecture = deb.ArchI386
	case platforms.ARM:
		architecture = deb.ArchArmhf
	case platforms.AMD64:
		architecture = deb.ArchAmd64
	}
	return architecture
}

func getArmArchName(settings *config.Settings) string {
	armArchName := settings.GetTaskSettingString(TASK_DEB, "armarch")
	if armArchName == "" {
		//derive it from GOARM version:
		goArm := settings.GetTaskSettingString(TASK_XC, "GOARM")
		if goArm == "5" {
			armArchName = "armel"
		} else {
			armArchName = "armhf"
		}
	}
	return armArchName
}

func debBuild(dest platforms.Platform, tp TaskParams) error {
	metadata := tp.Settings.GetTaskSettingMap(TASK_DEB, "metadata")
	armArchName := getArmArchName(tp.Settings)
	//maintain support for old configs ...
	metadataDebX := tp.Settings.GetTaskSettingMap(TASK_DEB, "metadata-deb")
	metadataDeb := map[string]string{}
	for k, v := range metadataDebX {
		val, ok := v.(string)
		if ok {
			metadataDeb[k] = val
		}
	}
	rmtemp := tp.Settings.GetTaskSettingBool(TASK_DEB, "rmtemp")
	debDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName()) //v0.8.1 dont use platform dir
	tmpDir := filepath.Join(debDir, ".goxc-temp")

	shortDescription := "?"
	if desc, keyExists := metadata["description"]; keyExists {
		var err error
		shortDescription, err = typeutils.ToString(desc, "description")
		if err != nil {
			return err
		}
	}
	longDescription := " "
	if ldesc, keyExists := metadata["long-description"]; keyExists {
		var err error
		longDescription, err = typeutils.ToString(ldesc, "long-description")
		if err != nil {
			return err
		}
	}
	maintainerName := "?"
	if maint, keyExists := metadata["maintainer"]; keyExists {
		var err error
		maintainerName, err = typeutils.ToString(maint, "maintainer")
		if err != nil {
			return err
		}
	}
	maintainerEmail := "example@example.org"
	if maintEmail, keyExists := metadata["maintainer-email"]; keyExists {
		var err error
		maintainerEmail, err = typeutils.ToString(maintEmail, "maintainer-email")
		if err != nil {
			return err
		}
	}
	//'dev' Package should be a separate task
	addDevPackage := false
	/*
		pkg := deb.NewPackage(tp.AppName, tp.Settings.GetFullVersionName(), maintainer, description)
		pkg.AdditionalControlData = metadataDeb*/

	build := debgen.NewBuildParams()
	build.DestDir = debDir
	build.TmpDir = tmpDir
	build.Init()
	build.IsRmtemp = rmtemp
	var ctrl *deb.Control
	//Read control data. If control file doesnt exist, use parameters ...
	fi, err := os.Open(filepath.Join(build.DebianDir, "control"))
	if os.IsNotExist(err) {
		log.Printf("WARNING - no debian 'control' file found. Use `debber` to generate proper debian metadata")
		ctrl = deb.NewControlDefault(tp.AppName, maintainerName, maintainerEmail, shortDescription, longDescription, addDevPackage)
	} else if err != nil {
		return fmt.Errorf("%v", err)
	} else {
		cfr := deb.NewControlFileReader(fi)
		ctrl, err = cfr.Parse()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	build.SourcesGlob = tp.Settings.GetTaskSettingString(TASK_DEB, "sources-glob")
	build.SourceDir = tp.Settings.GetTaskSettingString(TASK_DEB, "sources-dir")
	build.SourcesRelativeTo = tp.Settings.GetTaskSettingString(TASK_DEB, "sources-relative-to")
	debArch := getDebArch(dest.Arch, armArchName)
	build.Arches = []deb.Architecture{debArch}
	build.Version = tp.Settings.GetFullVersionName()
	dgens, err := debgen.PrepareBasicDebGen(ctrl, build)
	if err != nil {
		return fmt.Errorf("Error preparing deb generator: %v", err)
	}
	//there should only be one for this platform.
	// Anyway this part maps all binaries.
	for _, dgen := range dgens {
		for _, mainDir := range tp.MainDirs {
			var exeName string
			if len(tp.MainDirs) == 1 {
				exeName = tp.Settings.AppName
			} else {
				exeName = filepath.Base(mainDir)
			}
			binPath, err := core.GetAbsoluteBin(dest.Os, dest.Arch, tp.Settings.AppName, exeName, tp.WorkingDirectory, tp.Settings.GetFullVersionName(), tp.Settings.OutPath, tp.Settings.ArtifactsDest)
			if err != nil {
				return err
			}
			if dgen.DataFiles == nil {
				dgen.DataFiles = map[string]string{}
			}
			dgen.DataFiles["./usr/bin/"+exeName] = binPath
		}
		err = dgen.GenerateAllDefault()
		if err != nil {
			return fmt.Errorf("Error generating deb: %v", err)
		}
	}
	return err

	/*
		artifacts, err := deb.GetDebs(pkg)
		if err != nil {
			return err
		}
		log.Printf("Building arch: %s", debArch)
		artifact := artifacts[debArch]
		for _, mainDir := range tp.MainDirs {
			var exeName string
			if len(tp.MainDirs) == 1 {
				exeName = tp.Settings.AppName
			} else {
				exeName = filepath.Base(mainDir)

			}
			binPath, err := core.GetAbsoluteBin(dest.Os, dest.Arch, tp.Settings.AppName, exeName, tp.WorkingDirectory, tp.Settings.GetFullVersionName(), tp.Settings.OutPath, tp.Settings.ArtifactsDest)
			if err != nil {
				return err
			}
			if artifact.MappedFiles == nil {
				artifact.MappedFiles = map[string]string{}
			}
			artifact.MappedFiles["./usr/bin/"+exeName] = binPath
		}
		log.Printf("generating artifact '%v'", debArch)
		err = debgen.GenBinaryArtifact(artifact, build)
		if err != nil {
			return err
		}

		return err*/
	/*
		controlContent := getDebControlFileContent(tp.AppName, maintainer, tp.Settings.GetFullVersionName(), dest.Arch, armArchName, description, metadataDeb)
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
			var exeName string
			if len(tp.MainDirs) == 1 {
				exeName = tp.Settings.AppName
			} else {
				exeName = filepath.Base(mainDir)

			}
			binPath, err := core.GetAbsoluteBin(dest.Os, dest.Arch, tp.Settings.AppName, exeName, tp.WorkingDirectory, tp.Settings.GetFullVersionName(), tp.Settings.OutPath, tp.Settings.ArtifactsDest)
			if err != nil {
				return err
			}
			items = append(items, archive.ArchiveItem{FileSystemPath: binPath, ArchivePath: "/usr/bin/" + exeName})
		}
		//TODO add resources to /usr/share/appName/
		err = archive.TarGz(filepath.Join(tmpDir, "data.tar.gz"), items)
		if err != nil {
			return err
		}

		targetFile := filepath.Join(debDir, fmt.Sprintf("%s_%s_%s.deb", tp.AppName, tp.Settings.GetFullVersionName(), getDebArch(dest.Arch, armArchName))) //goxc_0.5.2_i386.deb")
		inputs := [][]string{
			[]string{filepath.Join(tmpDir, "debian-binary"), "debian-binary"},
			[]string{filepath.Join(tmpDir, "control.tar.gz"), "control.tar.gz"},
			[]string{filepath.Join(tmpDir, "data.tar.gz"), "data.tar.gz"}}
		err = ar.ArForDeb(targetFile, inputs)
		return
	*/
}
