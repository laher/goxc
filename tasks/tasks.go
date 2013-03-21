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
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"log"
	"strings"
)

func RunTasks(workingDirectory string, settings config.Settings) {
	if settings.IsVerbose() {
		log.Printf("looping through each platform")
	}
	//0.5 add support for space delimiters (more like BuildConstraints)
	destOses := strings.FieldsFunc(settings.Os, func(r rune) bool { return r==',' || r==' '})
	destArchs := strings.FieldsFunc(settings.Arch, func(r rune) bool { return r==',' || r==' '})
	if len(destOses) == 0 { destOses = []string{""} }
	if len(destArchs) == 0 { destArchs = []string{""} }
	var destPlatforms [][]string
	for _, supportedPlatformArr := range core.PLATFORMS {
		supportedOs := supportedPlatformArr[0]
		supportedArch := supportedPlatformArr[1]
		for _, destOs := range destOses {
			if destOs == "" || supportedOs == destOs {
				for _, destArch := range destArchs {
					if destArch == "" || supportedArch == destArch {
						destPlatforms = append(destPlatforms, supportedPlatformArr)
					}
				}
			}
		}
	}
	appName := core.GetAppName(workingDirectory)
	outDestRoot := core.GetOutDestRoot(appName, settings.ArtifactsDest)
	for _, task := range settings.Tasks {
		err := runTask(task, destPlatforms, appName, workingDirectory, outDestRoot, settings)
		if err != nil {
			// TODO: implement 'force' option.
			return
		}
	}
}

func runTask(task string, destPlatforms [][]string, appName, workingDirectory, outDestRoot string, settings config.Settings) error {
	// 0.3.1 added clean, vet, test, install etc
	switch task {
	case config.TASK_GO_CLEAN:
		err := core.InvokeGo(workingDirectory, []string{"clean"}, settings)
		if err != nil {
			log.Printf("Clean failed! %s", err)
		}
		return err
	case config.TASK_GO_VET:
		err := core.InvokeGo(workingDirectory, []string{"vet"}, settings)
		if err != nil {
			log.Printf("Vet failed! %s", err)
		}
		return err
	case config.TASK_GO_TEST:
		dir := settings.GetTaskSetting(config.TASK_GO_TEST, "dir", "./...").(string)
		err := core.InvokeGo(workingDirectory, []string{"test", dir}, settings)
		if err != nil {
			log.Printf("Test failed! %s", err)
		}
		return err
	case config.TASK_GO_FMT:
		dir := settings.GetTaskSetting(config.TASK_GO_FMT, "dir", "./...").(string)
		err := core.InvokeGo(workingDirectory, []string{"fmt", dir}, settings)
		if err != nil {
			log.Printf("Fmt failed! %s", err)
		}
		return err
	case config.TASK_GO_INSTALL:
		err := core.InvokeGo(workingDirectory, []string{"install"}, settings)
		if err != nil {
			log.Printf("Install failed! %s", err)
		}
		return err
	case config.TASK_CODESIGN:
		return runTaskCodesign(destPlatforms, appName, outDestRoot, settings)
	case config.TASK_BUILD_TOOLCHAIN:
		return runTaskToolchain(destPlatforms, settings)
	case config.TASK_XC:
		return runTaskXC(destPlatforms, workingDirectory, settings)
	case config.TASK_ARCHIVE:
		return runTaskZip(destPlatforms, appName, workingDirectory, outDestRoot, settings)
	case config.TASK_REMOVE_BIN:
		return runTaskRmBin(destPlatforms, appName, outDestRoot, settings)
	case config.TASK_DOWNLOADS_PAGE:
		return runTaskDownloadsPage(destPlatforms, appName, workingDirectory, outDestRoot, settings)
	}
	// TODO: custom tasks
	log.Printf("Unrecognised task '%s'", task)
	return fmt.Errorf("Unrecognised task '%s'", task)
}
