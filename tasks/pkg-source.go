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

/* INCOMPLETE!!
TODO: Complete sometime during 0.6.x
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

const (
	TASK_PKG_SOURCE = "pkg-source"
)

//runs automatically
func init() {
	register(Task{
		TASK_PKG_SOURCE,
		"Build a source package. Currently only supports 'source deb' format for Debian/Ubuntu Linux.",
		runTaskPkgSource,
		map[string]interface{}{"metadata": map[string]interface{}{"maintainer": "unknown"}, "metadata-deb": map[string]interface{}{"Depends": "", "Build-Depends" : "debhelper (>=4.0.0), golang-go, gcc"}, "rmtemp": true}})
}


func runTaskPkgSource(tp TaskParams) (err error) {
	var makeSourceDeb bool
	for _, platformArr := range tp.destPlatforms {
		destOs := platformArr[0]
		if destOs == platforms.LINUX {
			makeSourceDeb = true
		}
	}
	//TODO rpm
	if makeSourceDeb {
		err= debSourceBuild(tp)
	}
	//OK
	return
}


func debSourceBuild(tp TaskParams) (err error) {
	metadata := tp.settings.GetTaskSettingMap(TASK_PKG_SOURCE, "metadata")
	metadataDeb := tp.settings.GetTaskSettingMap(TASK_PKG_SOURCE, "metadata-deb")
	rmtemp := tp.settings.GetTaskSettingBool(TASK_PKG_SOURCE, "rmtemp")
	tmpDir := filepath.Join(tp.outDestRoot, ".goxc-temp")
	if rmtemp {
		defer os.RemoveAll(tmpDir)
	}
	description := ""
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
	version := tp.Settings.GetFullVersionName()
	arches := "any"
	buildDepends := "golang, debhelper"
	err := sdebPrepare(tp.WorkingDirectory, tp.appName, maintainer, version, arches, description, buildDepends, metadataDeb)

	if err != nil {
		return err
	}
	destDir := filepath.Join(tmpDir, "src")
	err = sdebCopySourceRecurse(tp.workingDirectory, destDir)
	if err != nil {
		return err
	}
//	controlContent := getDebControlFileContent(tp.appName, maintainer, tp.settings.GetFullVersionName(), destArch, description, metadataDeb)
	//if tp.settings.IsVerbose() {
		//log.Printf("Control file:\n%s", string(controlContent))
	//}

	//err = ioutil.WriteFile(filepath.Join(tmpDir, "control"), controlContent, 0644)
	if err != nil {
		return err
	}
	err = archive.TarGz(filepath.Join(tmpDir, "control.tar.gz"), [][]string{[]string{filepath.Join(tmpDir, "control"), "control"}})
	if err != nil {
		return err
	}

	//build
	//1. generate orig.tar.gz
	//memcached_1.2.5.orig.tar.gz


	//TODO add resources to /usr/share
	err = archive.TarGz(filepath.Join(tmpDir, appName + "_" + version + ".orig.tar.gz"), [][]string{[]string{appPath, "/usr/bin/" + tp.appName}})
	if err != nil {
		return err
	}

	//2. generate .diff.tar.gz (just containing debian/ directory)
	//generate debian/control
	//generate debian/rules
	//generate debian/changelog
	//generate debian/copyright
	//generate debian/README.Debian


	//3. generate .dsc file

	return
}

*/
