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

	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/debber/debber-v0.3/deb"
	"github.com/debber/debber-v0.3/debgen"
	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/platforms"
	"github.com/laher/goxc/typeutils"
)

//runs automatically
func init() {
	Register(Task{
		TASK_DEB_DEV,
		"Build a '-dev.deb' sources package for use as a Debian/Ubuntu Linux dependency.",
		runTaskDebDev,
		map[string]interface{}{
			"metadata": map[string]interface{}{
				"maintainer":      "unknown",
				"maintainerEmail": "unknown@example.com",
			},
			"metadata-deb": map[string]interface{}{"Depends": "",
				"Build-Depends": "debhelper (>=4.0.0)",
			},
			"rmtemp":              true,
			"armarch":             "",
			"go-sources-dir":      ".",
			"other-mappped-files": map[string]interface{}{},
		},
	})
}

func runTaskDebDev(tp TaskParams) (err error) {
	build := false
	for _, dest := range tp.DestPlatforms {
		if dest.Os == platforms.LINUX {
			build = true
		}
	}
	if build {
		err := debDevBuild(tp)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	return
}

func debDevBuild(tp TaskParams) error {
	metadata := tp.Settings.GetTaskSettingMap(TASK_DEB_GEN, "metadata")
	//maintain support for old configs ...
	metadataDebX := tp.Settings.GetTaskSettingMap(TASK_DEB_GEN, "metadata-deb")
	otherMappedFilesFromSetting := tp.Settings.GetTaskSettingMap(TASK_DEB_GEN, "other-mapped-files")
	otherMappedFiles, err := calcOtherMappedFiles(otherMappedFilesFromSetting)
	if err != nil {
		return err
	}
	metadataDeb := map[string]string{}
	for k, v := range metadataDebX {
		val, ok := v.(string)
		if ok {
			metadataDeb[k] = val
		}
	}
	rmtemp := tp.Settings.GetTaskSettingBool(TASK_DEB_GEN, "rmtemp")
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
	goSourcesDir := tp.Settings.GetTaskSettingString(TASK_DEB_GEN, "go-sources-dir")
	mappedSources, err := debgen.GlobForGoSources(goSourcesDir, []string{build.DestDir, build.TmpDir})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	for k, v := range mappedSources {
		otherMappedFiles[k] = v
	}
	debArch := deb.ArchAll
	build.Arches = []deb.Architecture{debArch}
	build.Version = tp.Settings.GetFullVersionName()
	dgens, err := debgen.PrepareBasicDebGen(ctrl, build)
	if err != nil {
		return fmt.Errorf("Error preparing deb generator: %v", err)
	}
	//there should only be one for this platform.
	// Anyway this part maps all binaries.
	for _, dgen := range dgens {
		if strings.HasSuffix(dgen.DebWriter.Control.Get(deb.PackageFName), "-dev") {
			for k, v := range otherMappedFiles {
				dgen.DataFiles[k] = v
			}
			err = dgen.GenerateAllDefault()
			if err != nil {
				return fmt.Errorf("Error generating deb: %v", err)
			}
			if !tp.Settings.IsQuiet() {
				log.Printf("Wrote -dev deb to %s", filepath.Join(build.DestDir, dgen.DebWriter.Filename))
			}
		}
	}
	return err

}
