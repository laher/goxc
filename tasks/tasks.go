// tasks are units of work performed by goxc.
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
	"github.com/laher/goxc/platforms"
	"github.com/laher/goxc/source"
	"log"
	"strings"
)

const (
	TASK_BUILD_TOOLCHAIN = core.TASK_BUILD_TOOLCHAIN

	TASK_CLEAN_DESTINATION = "clean-destination"
	TASK_GO_CLEAN          = "go-clean"

	TASK_GO_VET  = "go-vet"
	TASK_GO_TEST = "go-test"
	TASK_GO_FMT  = "go-fmt"

	TASK_GO_INSTALL = "go-install"
	TASK_XC         = "xc"
	TASK_CODESIGN   = "codesign"

	TASK_COPY_RESOURCES = "copy-resources"
	TASK_ARCHIVE        = "archive" //zip
	TASK_REMOVE_BIN     = "rmbin"   //after zipping
	TASK_DOWNLOADS_PAGE = "downloads-page"

	TASK_PKG_BUILD = "pkg-build"

	TASKALIAS_CLEAN = "clean"

	TASKALIAS_VALIDATE = "validate"
	TASKALIAS_COMPILE  = "compile"
	TASKALIAS_PACKAGE  = "package"

	TASKALIAS_DEFAULT = "default"
	TASKALIAS_ALL     = "all"
)

var (
	TASKS_CLEAN    = []string{TASK_GO_CLEAN, TASK_CLEAN_DESTINATION}
	TASKS_VALIDATE = []string{TASK_GO_VET, TASK_GO_TEST}
	TASKS_COMPILE  = []string{TASK_GO_INSTALL, TASK_XC, TASK_CODESIGN, TASK_COPY_RESOURCES}
	TASKS_PACKAGE  = []string{TASK_ARCHIVE, TASK_PKG_BUILD, TASK_REMOVE_BIN, TASK_DOWNLOADS_PAGE}
	TASKS_DEFAULT  = append(append(append([]string{}, TASKS_VALIDATE...), TASKS_COMPILE...), TASKS_PACKAGE...)
	TASKS_OTHER    = []string{TASK_BUILD_TOOLCHAIN, TASK_GO_FMT}
	TASKS_ALL      = append(append([]string{}, TASKS_OTHER...), TASKS_DEFAULT...)
)

// Parameter object passed to a task.
type TaskParams struct {
	DestPlatforms                 []platforms.Platform
	MainDirs                      []string
	AppName                       string
	WorkingDirectory, OutDestRoot string
	Settings                      config.Settings
}

// A task is basically a user-defined function given a unique name, plus some 'default settings'
type Task struct {
	Name            string
	Description     string
	f               func(TaskParams) error
	DefaultSettings map[string]interface{}
}

var (
	allTasks = make(map[string]Task)
	//Aliases are one or more tasks, in a specific order.
	Aliases = map[string][]string{
		TASKALIAS_CLEAN:    TASKS_CLEAN,
		TASKALIAS_VALIDATE: TASKS_VALIDATE,
		TASKALIAS_COMPILE:  TASKS_COMPILE,
		TASKALIAS_PACKAGE:  TASKS_PACKAGE,
		TASKALIAS_DEFAULT:  TASKS_DEFAULT,
		TASKALIAS_ALL:      TASKS_ALL}
)

// Register a task for use by goxc. Call from an 'init' function
func Register(task Task) {
	allTasks[task.Name] = task
}

// resolve aliases into tasks
// TODO recurse. (currently aliases are only lists of tasks, not of aliases). Recursion would enable the extra flexibility
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

// list all available tasks
func ListTasks() []Task {
	tasks := []Task{}
	for _, t := range allTasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// run all given tasks
func RunTasks(workingDirectory string, destPlatforms []platforms.Platform, settings config.Settings) {
	if settings.IsVerbose() {
		log.Printf("looping through each platform")
	}
	appName := core.GetAppName(workingDirectory)
	mainDirs, err := source.FindMainDirs(workingDirectory)
	if err != nil {
		log.Printf("Warning: could not establish list of main dirs. Using working directory")
		mainDirs = []string{workingDirectory}
	}
	if len(mainDirs) > 1 {
		log.Printf("Multiple main dirs: %v", mainDirs)
	}
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
	//0.6 check all tasks are valid before continuing
	for _, taskName := range tasksToRun {
		if _, keyExists := allTasks[taskName]; !keyExists {
			if strings.HasPrefix(taskName, ".") {
				log.Printf("'%s' looks like a directory, not a task - specify 'working directory' with -wd option", taskName)
			}
			if e, _ := core.FileExists(taskName); e {
				log.Printf("'%s' looks like a directory, not a task - specify 'working directory' with -wd option", taskName)
			}
			log.Printf("Task %s does NOT exist!", taskName)
			return
		}
	}
	log.Printf("Running tasks: %v", tasksToRun)
	for _, taskName := range tasksToRun {
		log.SetPrefix("[goxc:" + taskName + "] ")
		err := runTask(taskName, destPlatforms, mainDirs, appName, workingDirectory, outDestRoot, settings)
		if err != nil {
			// TODO: implement 'force' option.
			log.Printf("Stopping after '%s' failed with error '%v'", taskName, err)
			return
		} else {
			log.Printf("Task %s succeeded", taskName)
		}
	}
}

// run named task
func runTask(taskName string, destPlatforms []platforms.Platform, mainDirs []string, appName, workingDirectory, outDestRoot string, settings config.Settings) error {
	if taskV, keyExists := allTasks[taskName]; keyExists {
		tp := TaskParams{destPlatforms, mainDirs, appName, workingDirectory, outDestRoot, settings}
		return taskV.f(tp)
	}
	log.Printf("Unrecognised task '%s'", taskName)
	return fmt.Errorf("Unrecognised task '%s'", taskName)
}
