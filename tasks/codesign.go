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
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/platforms"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
)

var codesignTask = Task{
	TASK_CODESIGN,
	"sign code for Mac. Only Mac hosts are supported for this task.",
	runTaskCodesign,
	map[string]interface{}{"id": ""}}

//runs automatically
func init() {
	register(codesignTask)
}

func runTaskCodesign(tp taskParams) error {
	//func runTaskCodesign(destPlatforms [][]string, outDestRoot string, appName string, settings config.Settings) error {
	for _, platformArr := range tp.destPlatforms {
		destOs := platformArr[0]
		destArch := platformArr[1]
		relativeBin := core.GetRelativeBin(destOs, destArch, tp.appName, false, tp.settings.GetFullVersionName())
		codesignPlat(destOs, destArch, tp.outDestRoot, relativeBin, tp.settings)
	}
	//TODO return error
	return nil
}

func codesignPlat(goos, arch string, outDestRoot string, relativeBin string, settings config.Settings) {
	// settings.codesign only works on OS X for binaries generated for OS X.
	id := settings.GetTaskSetting("codesign", "id")
	if id != "" && runtime.GOOS == platforms.DARWIN && goos == platforms.DARWIN {
		if err := signBinary(filepath.Join(outDestRoot, relativeBin), id.(string)); err != nil {
			log.Printf("codesign failed: %s", err)
		} else {
			log.Printf("Signed with ID: %q", id)
		}
	}
}

func signBinary(binPath string, id string) error {
	cmd := exec.Command("codesign")
	cmd.Args = append(cmd.Args, "-s", id, binPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
