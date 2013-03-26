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
)

type taskParams struct {
	destPlatforms                 [][]string
	appName                       string
	workingDirectory, outDestRoot string
	settings                      config.Settings
}

type Task struct {
	Name            string
	Description     string
	f               func(taskParams) error
	DefaultSettings map[string]interface{}
}

var (
	allTasks = make(map[string]Task)
	Aliases  = map[string][]string{
		core.TASKALIAS_CLEAN:    core.TASKS_CLEAN,
		core.TASKALIAS_VALIDATE: core.TASKS_VALIDATE,
		core.TASKALIAS_COMPILE:  core.TASKS_COMPILE,
		core.TASKALIAS_PACKAGE:  core.TASKS_PACKAGE,
		core.TASKALIAS_DEFAULT:  core.TASKS_DEFAULT,
		core.TASKALIAS_ALL:      core.TASKS_ALL}
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

func RunTasks(workingDirectory string, destPlatforms [][]string, settings config.Settings) {
	if settings.IsVerbose() {
		log.Printf("looping through each platform")
	}

	appName := core.GetAppName(workingDirectory)
	outDestRoot := core.GetOutDestRoot(appName, settings.ArtifactsDest, workingDirectory)
	defer log.SetPrefix("[goxc] ")
	exclusions := ResolveAliases(settings.TasksExclude)
	appends := ResolveAliases(settings.TasksAppend)
	mains := ResolveAliases(settings.Tasks)
	mains = append(mains, appends...)
	tasksToRun := []string{}
	for _, taskName := range mains {
		if !core.ContainsString(exclusions, taskName) {
			tasksToRun = append(tasksToRun, taskName)
		}
	}
	log.Printf("Running tasks: %v", tasksToRun)
	for _, taskName := range tasksToRun {
		if !core.ContainsString(exclusions, taskName) {
			log.SetPrefix("[goxc:" + taskName + "] ")
			err := runTask(taskName, destPlatforms, appName, workingDirectory, outDestRoot, settings)
			if err != nil {
				// TODO: implement 'force' option.
				log.Printf("Stopping after '%s' failed with error '%v'", taskName, err)
				return
			}
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
