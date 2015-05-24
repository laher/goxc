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
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/platforms"
	"github.com/laher/goxc/source"
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
	TASK_ARCHIVE_ZIP    = "archive-zip"
	TASK_ARCHIVE_TAR_GZ = "archive-tar-gz"
	TASK_REMOVE_BIN     = "rmbin" //after zipping
	TASK_DOWNLOADS_PAGE = "downloads-page"
	TASK_DEB_GEN        = "deb"
	TASK_DEB_DEV        = "deb-dev"
	TASK_DEB_SOURCE     = "deb-source"
	TASK_PUBLISH_GITHUB = "publish-github"

	TASKALIAS_ALL        = "all"
	TASKALIAS_ARCHIVE    = "archive"
	TASKALIAS_CLEAN      = "clean"
	TASKALIAS_COMPILE    = "compile"
	TASKALIAS_DEBS       = "debs"
	TASKALIAS_DEFAULT    = "default"
	TASKALIAS_PACKAGE    = "package"
	TASKALIAS_PKG_BUILD  = "pkg-build"
	TASKALIAS_PKG_SOURCE = "pkg-source"
	TASKALIAS_VALIDATE   = "validate"
)

var (
	TASKS_ARCHIVE                     = []string{TASK_ARCHIVE_ZIP, TASK_ARCHIVE_TAR_GZ}
	TASKS_CLEAN                       = []string{TASK_GO_CLEAN, TASK_CLEAN_DESTINATION}
	TASKS_COMPILE                     = []string{TASK_GO_INSTALL, TASK_XC, TASK_CODESIGN, TASK_COPY_RESOURCES}
	TASKS_DEBS                        = []string{TASK_DEB_GEN, TASK_DEB_DEV, TASK_DEB_SOURCE}
	TASKS_PACKAGE                     = []string{TASK_ARCHIVE_ZIP, TASK_ARCHIVE_TAR_GZ, TASK_DEB_GEN, TASK_DEB_DEV, TASK_REMOVE_BIN, TASK_DOWNLOADS_PAGE}
	TASKS_PKG_BUILD                   = []string{TASK_DEB_GEN, TASK_DEB_DEV}
	TASKS_PKG_SOURCE                  = []string{TASK_DEB_SOURCE}
	TASKS_VALIDATE                    = []string{TASK_GO_VET, TASK_GO_TEST}
	TASKS_DEFAULT                     = append(append(append([]string{}, TASKS_VALIDATE...), TASKS_COMPILE...), TASKS_PACKAGE...)
	TASKS_OTHER                       = []string{TASK_BUILD_TOOLCHAIN, TASK_GO_FMT, TASK_PUBLISH_GITHUB}
	TASKS_ALL                         = append(append([]string{}, TASKS_OTHER...), TASKS_DEFAULT...)
	TASK_ALIASES_FOR_MERGING_SETTINGS = map[string][]string{TASKALIAS_PKG_BUILD: TASKS_PKG_BUILD, TASKALIAS_PKG_SOURCE: TASKS_PKG_SOURCE, TASKALIAS_DEBS: TASKS_DEBS}

	allTasks = make(map[string]Task)
	//Aliases are one or more tasks, in a specific order.
	Aliases = map[string][]string{
		TASKALIAS_ALL:        TASKS_ALL,
		TASKALIAS_ARCHIVE:    TASKS_ARCHIVE,
		TASKALIAS_CLEAN:      TASKS_CLEAN,
		TASKALIAS_COMPILE:    TASKS_COMPILE,
		TASKALIAS_DEFAULT:    TASKS_DEFAULT,
		TASKALIAS_PACKAGE:    TASKS_PACKAGE,
		TASKALIAS_PKG_BUILD:  TASKS_PKG_BUILD,
		TASKALIAS_PKG_SOURCE: TASKS_PKG_SOURCE,
		TASKALIAS_VALIDATE:   TASKS_VALIDATE,
	}
)

// Parameter object passed to a task.
type TaskParams struct {
	DestPlatforms                 []platforms.Platform
	AllPackageDirs                []string
	MainDirs                      []string
	AppName                       string
	WorkingDirectory, OutDestRoot string
	Settings                      *config.Settings
	MaxProcessors                 int
}

// A task is basically a user-defined function given a unique name, plus some 'default settings'
type Task struct {
	Name            string
	Description     string
	Run             func(TaskParams) error
	DefaultSettings map[string]interface{}
}

type ParallelizableTask struct {
	Name            string
	Description     string
	setUp           func(TaskParams) ([]platforms.Platform, error)
	perPlatform     func(TaskParams, platforms.Platform, chan error)
	tearDown        func(TaskParams) error
	DefaultSettings map[string]interface{}
}

// Register a task for use by goxc. Call from an 'init' function
func Register(task Task) {
	allTasks[task.Name] = task
}

func generateParallelizedRunFunc(pTask ParallelizableTask) func(TaskParams) error {
	fn := func(tp TaskParams) error {
		platforms, err := pTask.setUp(tp)
		if err != nil {
			return err
		}
		platCount := len(platforms)
		if platCount < 1 {
			return nil
		}
		numProcs := runtime.NumCPU()
		if !tp.Settings.IsQuiet() {
			log.Printf("Parallelizing %s for %d platforms, using max %d of %d processors", pTask.Name, platCount, tp.MaxProcessors, numProcs)
		}
		errchan := make(chan error)
		roundIdx := 0
		roundCount := tp.MaxProcessors
		totIdx := 0
		totCount := platCount
		errs := []error{}
		for roundIdx < roundCount && totIdx < totCount {
			pl := platforms[totIdx]
			go pTask.perPlatform(tp, pl, errchan)
			totIdx++
			roundIdx++
			if roundIdx >= roundCount {
				//block until all of this round is complete.
				i := 0
				for i < roundCount {
					err = <-errchan
					if err != nil {
						errs = append(errs, err)
					}
					i++
				}
				//reset roundIdx
				roundIdx = 0
			}
		}
		if roundIdx > 0 {
			//block until all remaining tasks are complete
			i := 0
			for i < roundIdx {
				err = <-errchan
				if err != nil {
					errs = append(errs, err)
				}
				i++
			}

		}
		//always tearDown incase you need to free resources
		if pTask.tearDown != nil {
			err = pTask.tearDown(tp)
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			log.Printf("Multiple errors (returning first one): %v", errs)
			return errs[0]
		}
		return nil
	}
	return fn
}

func RegisterParallelizable(pTask ParallelizableTask) {
	task := Task{
		Name:            pTask.Name,
		Description:     pTask.Description,
		Run:             generateParallelizedRunFunc(pTask),
		DefaultSettings: pTask.DefaultSettings}
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
func RunTasks(workingDirectory string, destPlatforms []platforms.Platform, settings *config.Settings, maxProcessors int) error {
	if settings.IsVerbose() {
		log.Printf("Using Go root: %s", settings.GoRoot)
		log.Printf("looping through each platform")
	}
	appName := core.GetAppName(settings.AppName, workingDirectory)

	outDestRoot, err := core.GetOutDestRoot(appName, workingDirectory, settings.ArtifactsDest)
	if err != nil {
		return err
	}
	defer log.SetPrefix("[goxc] ")
	exclusions := ResolveAliases(settings.TasksExclude)
	appends := ResolveAliases(settings.TasksAppend)
	mains := ResolveAliases(settings.Tasks)
	all := ResolveAliases(settings.TasksPrepend)
	//log.Printf("prepending %v", all)
	all = append(all, mains...)
	all = append(all, appends...)

	//exclude by resolved task names (not by aliases)
	tasksToRun := []string{}
	for _, taskName := range all {
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
			if settings.IsVerbose() {
				log.Printf("Task '%s' does NOT exist!", taskName)
			}
			return errors.New("Task '" + taskName + "' does not exist")
		}
	}
	mainDirs := []string{}
	allPackages := []string{}
	if len(tasksToRun) == 1 && tasksToRun[0] == "toolchain" {
		log.Printf("Toolchain task only - not searching for main dirs")
		//mainDirs = []string{workingDirectory}
	} else {
		var err error
		excludes := core.ParseCommaGlobs(settings.MainDirsExclude)
		excludesSource := core.ParseCommaGlobs(settings.SourceDirsExclude)
		excludesSource = append(excludesSource, excludes...)
		allPackages, err = source.FindSourceDirs(workingDirectory, "", excludesSource, settings.IsVerbose())
		if err != nil || len(allPackages) == 0 {
			log.Printf("Warning: could not establish list of source packages. Using working directory")
			allPackages = []string{workingDirectory}
		}
		mainDirs, err = source.FindMainDirs(workingDirectory, excludes, settings.IsVerbose())
		if err != nil || len(mainDirs) == 0 {
			log.Printf("Warning: could not find any main dirs: %v", err)
		} else {
			if settings.IsVerbose() {
				log.Printf("Found 'main package' dirs (len %d): %v", len(mainDirs), mainDirs)
			}
		}
	}
	if settings.IsVerbose() {
		log.Printf("Running tasks: %v", tasksToRun)
		log.Printf("All packages: %v", allPackages)
	}
	for _, taskName := range tasksToRun {
		log.SetPrefix("[goxc:" + taskName + "] ")
		if settings.IsVerbose() {
			log.Printf("Running task %s with settings: %v", taskName, settings.TaskSettings[taskName])
		}
		err := runTask(taskName, destPlatforms, allPackages, mainDirs, appName, workingDirectory, outDestRoot, settings, maxProcessors)
		if err != nil {
			// TODO: implement 'force' option.
			log.Printf("Stopping after '%s' failed with error '%v'", taskName, err)
			return err
		} else {
			if !settings.IsQuiet() {
				log.Printf("Task %s succeeded", taskName)
			}
		}
	}
	return nil
}

// run named task
func runTask(taskName string, destPlatforms []platforms.Platform, mainDirs []string, allPackages []string, appName, workingDirectory, outDestRoot string, settings *config.Settings, maxProcessors int) error {
	if taskV, keyExists := allTasks[taskName]; keyExists {
		tp := TaskParams{destPlatforms, mainDirs, allPackages, appName, workingDirectory, outDestRoot, settings, maxProcessors}
		return taskV.Run(tp)
	}
	log.Printf("Unrecognised task '%s'", taskName)
	return fmt.Errorf("Unrecognised task '%s'", taskName)
}

func FillTaskSettingsDefaults(settings *config.Settings) {
	if len(settings.Tasks) == 0 {
		settings.Tasks = Aliases[TASKALIAS_DEFAULT]
	}
	if settings.TaskSettings == nil {
		settings.TaskSettings = make(map[string]map[string]interface{})
	}
	//fill in per-task settings ...
	for _, t := range ListTasks() {
		if t.DefaultSettings != nil {
			if _, keyExists := settings.TaskSettings[t.Name]; !keyExists {
				settings.TaskSettings[t.Name] = t.DefaultSettings
			} else {
				//TODO go deeper still?
				for k, v := range t.DefaultSettings {
					taskSettings := settings.TaskSettings[t.Name]
					if _, keyExists = taskSettings[k]; !keyExists {
						taskSettings[k] = v
					}
				}
			}
		}
	}
}
