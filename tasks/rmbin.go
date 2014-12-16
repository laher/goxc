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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/platforms"
)

//runs automatically
func init() {
	Register(Task{
		"rmbin",
		"delete binary. Normally runs after 'archive' task to reduce size of output dir.",
		runTaskRmBin,
		nil})
}

func runTaskRmBin(tp TaskParams) error {
	for _, dest := range tp.DestPlatforms {
		for _, mainDir := range tp.MainDirs {
			var exeName string
			if len(tp.MainDirs) == 1 {
				exeName = tp.Settings.AppName
			} else {
				exeName = filepath.Base(mainDir)

			}
			err := rmBinPlat(dest, tp, exeName)
			if err != nil {
				//todo - add a force option?
				log.Printf("%v", err)
			}
		}
	}
	//TODO return error
	return nil
}

func rmBinPlat(dest platforms.Platform, tp TaskParams, exeName string) error {
	binPath, err := core.GetAbsoluteBin(dest.Os, dest.Arch, tp.Settings.AppName, exeName, tp.WorkingDirectory, tp.Settings.GetFullVersionName(), tp.Settings.OutPath, tp.Settings.ArtifactsDest)
	if err != nil {
		return err
	}
	err = os.Remove(binPath)
	if err != nil {
		return err
	}
	//if empty, remove dir
	binDir := filepath.Dir(binPath)
	files, err := ioutil.ReadDir(binDir)
	if err != nil {
		return err
	}
	if len(files) < 1 {
		err = os.Remove(binDir)
	}
	return err
}
