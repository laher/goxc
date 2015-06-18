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
	"log"
	"os/exec"
	"path/filepath"
	"runtime"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/executils"
	"github.com/laher/goxc/platforms"
)

var codesignTask = Task{
	TASK_CODESIGN,
	"sign code for Mac. Only Mac hosts are supported for this task.",
	runTaskCodesign,
	map[string]interface{}{"id": ""}}

//runs automatically
func init() {
	Register(codesignTask)
}

func runTaskCodesign(tp TaskParams) (err error) {
	for _, dest := range tp.DestPlatforms {
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
			err = codesignPlat(dest.Os, dest.Arch, binPath, tp.Settings)
		}
	}
	//TODO return error
	return err
}

func codesignPlat(goos, arch string, binPath string, settings *config.Settings) error {
	// settings.codesign only works on OS X for binaries generated for OS X.
	id := settings.GetTaskSettingString("codesign", "id")
	if id != "" && runtime.GOOS == platforms.DARWIN && goos == platforms.DARWIN {
		if err := signBinary(binPath, id); err != nil {
			log.Printf("codesign failed: %s", err)
			return err
		} else {
			if !settings.IsQuiet() {
				log.Printf("Signed with ID: %q", id)
			}
			return nil
		}
	}
	return nil
}

func signBinary(binPath string, id string) error {
	cmd := exec.Command("codesign")
	cmd.Args = append(cmd.Args, "-s", id, binPath)

	return executils.StartAndWait(cmd)
}
