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

type taskParams struct {
	destPlatforms                 [][]string
	appName                       string
	workingDirectory, outDestRoot string
	settings                      config.Settings
}

type Task struct {
	Name        string
	Description string
	f           func(taskParams) error
}

var (
	allTasks = make(map[string]Task)
	Aliases  = map[string][]string{
		config.TASKALIAS_DEFAULT:  config.TASKS_DEFAULT,
		config.TASKALIAS_PACKAGE:  config.TASKS_PACKAGE,
		config.TASKALIAS_VALIDATE: config.TASKS_VALIDATE,
		config.TASKALIAS_ALL:      config.TASKS_ALL}
)

func register(task Task) {
	allTasks[task.Name] = task
}

func ResolveAliases(tasks []string) []string {
	ret := []string{}
	for _, taskName := range tasks {
		if aliasTasks, keyExists := Aliases[taskName]; keyExists {
			ret = append(ret, aliasTasks...)
		} else {
			ret = append(ret, taskName)
		}
	}
	return ret
}

func ListTasks() []Task {
	tasks := []Task{}
	for _, t := range allTasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func RunTasks(workingDirectory string, settings config.Settings) {
	if settings.IsVerbose() {
		log.Printf("looping through each platform")
	}
	//0.5 add support for space delimiters (more like BuildConstraints)
	destOses := strings.FieldsFunc(settings.Os, func(r rune) bool { return r == ',' || r == ' ' })
	destArchs := strings.FieldsFunc(settings.Arch, func(r rune) bool { return r == ',' || r == ' ' })
	if len(destOses) == 0 {
		destOses = []string{""}
	}
	if len(destArchs) == 0 {
		destArchs = []string{""}
	}
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
	outDestRoot := core.GetOutDestRoot(appName, settings.ArtifactsDest, workingDirectory)
	for _, taskName := range settings.Tasks {
		err := runTask(taskName, destPlatforms, appName, workingDirectory, outDestRoot, settings)
		if err != nil {
			// TODO: implement 'force' option.
			log.Printf("Stopping after '%s' failed.", taskName)
			return
		}
	}
}

func runTask(taskName string, destPlatforms [][]string, appName, workingDirectory, outDestRoot string, settings config.Settings) error {
	if taskV, keyExists := allTasks[taskName]; keyExists {
		tp := taskParams{destPlatforms, appName, workingDirectory, outDestRoot, settings}
		return taskV.f(tp)
	}

	// TODO: custom tasks
	log.Printf("Unrecognised task '%s'", taskName)
	return fmt.Errorf("Unrecognised task '%s'", taskName)
}
