// GOXC IS NOT READY FOR USE AS AN API - function names and packages will continue to change until version 1.0
package main

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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"runtime"
	"strings"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/platforms"
	"github.com/laher/goxc/tasks"
)

const (
	MSG_HELP               = "Usage: goxc [<option(s)>] [<task(s)>]\n"
	MSG_HELP_TOPICS        = "goxc -h <topic>\n"
	MSG_HELP_TOPICS_EG     = "More help:\n\tgoxc -h options\nor\n\tgoxc -h tasks\n"
	MSG_HELP_LINK          = "Please see https://github.com/laher/goxc/wiki for full details.\n"
	MSG_HELP_UNKNOWN_TOPIC = "Unknown topic '%s'. Try 'options' or 'tasks'\n"
	MSG_HELP_DESC          = "goxc cross-compiles go programs to multiple platforms at once.\n"
	MSG_HELP_TC            = "NOTE: You need to run `goxc -t` first. See https://github.com/laher/goxc/ for more details!\n"
)

// settings for this invocation of goxc
var (
	// VERSION is initialised by the linker during compilation if the appropriate flag is specified:
	// e.g. go build -ldflags "-X main.VERSION 0.1.2-abcd" goxc.go
	// thanks to minux for this advice
	// So, goxc does this automatically during 'go build'
	VERSION              = "0.6.x"
	settings             config.Settings
	configName           string
	isVersion            bool
	isHelp               bool
	isHelpTasks          bool
	isBuildToolchain     bool
	tasksToRun           string
	tasksPlus            string
	tasksMinus           string
	isCliZipArchives     string
	codesignId           string
	isWriteConfig        bool
	isVerbose            bool
	workingDirectoryFlag string
)

func printHelp(flagSet *flag.FlagSet) {
	args := flagSet.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "goxc version '%s'\n", VERSION)
		printHelpTopic(flagSet, "options")
	} else {
		printHelpTopic(flagSet, args[0])
	}
}

func printHelpTopic(flagSet *flag.FlagSet, topic string) {
	switch topic {
	case "options":
		fmt.Fprint(os.Stderr, MSG_HELP)
		printOptions(flagSet)
		return
	case "tasks":
		fmt.Fprint(os.Stderr, "Use commandline arguments to specify tasks, or '-tasks-=' or '-tasks+=' to adjust them.\n\ne.g. to run all the 'default' tasks skipping 'rmbin' and appending 'go-fmt':\n\t`goxc -tasks+=go-fmt -tasks-=rmbin default`\n")
		fmt.Fprint(os.Stderr, "\nAvailable tasks & aliases (specify aliases where possible):\n")
		allTasks := tasks.ListTasks()
		var padding string
		for _, task := range allTasks {
			if len(task.Name) < 14 {
				padding = strings.Repeat(" ", 14-len(task.Name))
			} else {
				padding = ""
			}
			fmt.Fprintf(os.Stderr, " %s  %s%s\n", task.Name, padding, task.Description)
		}
		for alias, taskNames := range tasks.Aliases {
			if len(alias) < 15 {
				padding = strings.Repeat(" ", 15-len(alias))
			} else {
				padding = ""
			}
			fmt.Fprintf(os.Stderr, " %s%s alias: %v\n", alias, padding, taskNames)
		}
		return
	default:
		//task help
		for _, task := range tasks.ListTasks() {
			if topic == task.Name {
				fmt.Fprintf(os.Stderr, "Task:\n '%s'\nDescription:\n  %s\n", task.Name, task.Description)
				if task.DefaultSettings != nil {
					out, err := json.MarshalIndent(map[string]map[string]interface{}{ task.Name : task.DefaultSettings }, "", "\t")
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error displaying TaskConfig info %v\n", err)
					} else {
						fmt.Fprintf(os.Stderr, "TaskConfig:\n\"TaskConfig\": %s \n", string(out))
					}
				} else {
					fmt.Fprintf(os.Stderr, "TaskConfig:\n No TaskConfig available for '%s'\n", task.Name)

				}
				return
			}
		}
		for alias, taskNames := range tasks.Aliases {
			if topic == alias {
				fmt.Fprintf(os.Stderr, "Alias '%s'\n'%s' runs the following tasks:\n  %s\n", alias, alias, taskNames)
				return
			}
		}

	}
	fmt.Fprintf(os.Stderr, MSG_HELP_UNKNOWN_TOPIC, topic)
	fmt.Fprint(os.Stderr, MSG_HELP_TOPICS)
	fmt.Fprint(os.Stderr, MSG_HELP_TOPICS_EG)
}

func printVersion(flagSet *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, " goxc version: %s\n", VERSION)
}

//merge configuration file
//maybe oneday: parse source
//TODO honour build flags ? (build flags are per-file rather than per-package, so not sure how this might work)
func mergeConfiguredSettings(dir string, configName string, useLocal bool) (config.Settings, error) {
	if settings.IsVerbose() {
		log.Printf("loading configured settings")
	}
	configuredSettings, err := config.LoadJsonConfigOverrideable(dir, configName, useLocal, settings.IsVerbose())
	if settings.IsVerbose() {
		log.Printf("Settings from config %s: %+v : %v", configName, configuredSettings, err)
	}
	//TODO: further error handling ?
	if err == nil {
		settings = config.Merge(settings, configuredSettings)
	}
	return settings, err
}

//TODO fulfil all defaults
func fillDefaults(settings config.Settings) config.Settings {
	if settings.Resources.Include == "" {
		settings.Resources.Include = core.RESOURCES_INCLUDE_DEFAULT
	}
	if settings.Resources.Exclude == "" {
		settings.Resources.Exclude = core.RESOURCES_EXCLUDE_DEFAULT
	}
	if settings.PackageVersion == "" {
		settings.PackageVersion = core.PACKAGE_VERSION_DEFAULT
	}
	if len(settings.Tasks) == 0 {
		settings.Tasks = tasks.TASKS_DEFAULT
	}
	if settings.TaskSettings == nil {
		settings.TaskSettings = make(map[string]map[string]interface{})
	}

	//fill in per-task settings ...
	for _, t := range tasks.ListTasks() {
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
	return settings
}

// goXC is the goxc startpoint
// In theory you could call this with a slice of flags
func goXC(call []string) {
	workingDirectory, settings := interpretSettings(call)
	if isWriteConfig {
		err := config.WriteJsonConfig(workingDirectory, settings, configName, false)
		if err != nil {
			log.Printf("Could not write config file: %v", err)
		}
		//0.2.5 writeConfig now just exits after writing config
	} else {
		//0.2.3 fillDefaults should only happen after writing config
		settings = fillDefaults(settings)

		if settings.IsVerbose() {
			log.Printf("Final settings %+v", settings)
		}
		//2.0.0: Removed PKG_VERSION parsing
		destPlatforms := platforms.GetDestPlatforms(settings.Os, settings.Arch)
		destPlatforms = platforms.ApplyBuildConstraints(settings.BuildConstraints, destPlatforms)
		tasks.RunTasks(workingDirectory, destPlatforms, settings)
	}
}

func interpretSettings(call []string) (string, config.Settings) {
	flagSet := setupFlags()
	if err := flagSet.Parse(call[1:]); err != nil {
		log.Printf("Error parsing arguments: %s", err)
		os.Exit(1)
	} else {
		if isVerbose {
			settings.Verbosity = core.VERBOSITY_VERBOSE
		}
		//0.6 use args. Parse into slice.
		settings.Tasks = flagSet.Args()

		//0.6 merge after tasks
		if tasksToRun != "" {
			tasksToRunSlice := strings.FieldsFunc(tasksToRun, func(r rune) bool { return r == ',' || r == ' ' })
			settings.Tasks = append(settings.Tasks, tasksToRunSlice...)
		}

		//TODO permit/recommend spaces
		if tasksPlus != "" {
			settings.TasksAppend = strings.FieldsFunc(tasksPlus, func(r rune) bool { return r == ',' || r == ' ' })
		}
		if tasksMinus != "" {
			settings.TasksExclude = strings.Split(tasksMinus, ",")
		}
		if isBuildToolchain {
			//0.6 prepend to settings.Tasks slice (instead of tasksToRun string)
			settings.Tasks = append([]string{tasks.TASK_BUILD_TOOLCHAIN}, settings.Tasks...)
		}
		//0.2.3 NOTE this will be superceded soon
		//using string because that makes it overrideable
		//0.5.0 using Tasks instead of ArtifactTypes
		if isCliZipArchives == "true" || isCliZipArchives == "t" {
			//settings.ArtifactTypes = []string{core.ARTIFACT_TYPE_ZIP}
			settings.Tasks = remove(settings.TasksExclude, tasks.TASK_ARCHIVE)
			settings.Tasks = appendIfMissing(settings.Tasks, tasks.TASK_ARCHIVE)
		} else if isCliZipArchives == "false" || isCliZipArchives == "f" {
			settings.TasksExclude = appendIfMissing(settings.TasksExclude, tasks.TASK_ARCHIVE)
		}
		//TODO use Setting
		if codesignId != "" {
			settings.SetTaskSetting(tasks.TASK_CODESIGN, "id", codesignId)
		}
	}
	if isHelp {
		printHelp(flagSet)
		os.Exit(0)
	}
	if isHelpTasks {
		printHelpTopic(flagSet, "tasks")
		os.Exit(0)
	}

	if isVersion {
		printVersion(flagSet)
		os.Exit(0)
	}
	//sanity check
	goroot := runtime.GOROOT()
	if err := core.SanityCheck(goroot); err != nil {
		log.Printf("Error: %s", err)
		log.Printf(core.MSG_INSTALL_GO_FROM_SOURCE)
		os.Exit(1)
	}
	//0.6 do NOT use args[0]
	var workingDirectory string
	if workingDirectoryFlag != "" {
		workingDirectory = workingDirectoryFlag
	} else {
		if isBuildToolchain {
			//default to HOME dir
			log.Printf("Building toolchain, so getting config from HOME directory. To use current directory's config, use the wd option (i.e. goxc -t -wd=.)")
			workingDirectory = userHomeDir()
		} else {
			if isVerbose {
				log.Printf("Using config from current directory")
			}
			//default to current directory
			workingDirectory = "."
		}
	}
	log.Printf("Working directory: '%s', Config name: '%s'", workingDirectory, configName)

	settings, err := mergeConfiguredSettings(workingDirectory, configName, !isWriteConfig)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Configuration file error. %s", err.Error())
			os.Exit(1)
		}
	}
	return workingDirectory, settings
}

func appendIfMissing(arr []string, v string) []string {
	ret := make([]string, len(arr))
	copy(ret, arr)
	for _, val := range arr {
		if val == v { //found. return.
			return ret
		}
	}
	return append(ret, v)
}

func remove(arr []string, v string) []string {
	ret := make([]string, len(arr))
	copy(ret, arr)
	for i, val := range arr {
		//fmt.Println(i, val)
		if val != v {
			continue
		}

		//fmt.Println(i, val, v)
		return append(ret[:i], ret[i+1:]...)
	}
	return ret
}

// Set up flags.
// Note use of empty strings as defaults, with 'actual' defaults .
// This is done to make merging options from configuration files easier.
func setupFlags() *flag.FlagSet {
	flagSet := flag.NewFlagSet("goxc", flag.ContinueOnError)
	flagSet.StringVar(&configName, "c", core.CONFIG_NAME_DEFAULT, "config name (default='')")

	//TODO deprecate?
	flagSet.StringVar(&settings.Os, "os", "", "Specify OS (default is all - \"linux darwin windows freebsd openbsd\")")
	flagSet.StringVar(&settings.Arch, "arch", "", "Specify Arch (default is all - \"386 amd64 arm\")")

	//TODO introduce and implement in time for 0.6
	flagSet.StringVar(&settings.BuildConstraints, "bc", "", "Specify build constraints (e.g. 'linux,arm windows')")

	flagSet.StringVar(&workingDirectoryFlag, "wd", "", "Specify directory to work on")

	flagSet.StringVar(&settings.PackageVersion, "pv", "", "Package version (usually [major].[minor].[patch]. default='"+core.PACKAGE_VERSION_DEFAULT+"')")
	flagSet.StringVar(&settings.PackageVersion, "av", "", "DEPRECATED: Package version (deprecated option name)")
	flagSet.StringVar(&settings.PrereleaseInfo, "pr", "", "Prerelease info (usually 'alpha', 'snapshot' ...)")
	flagSet.StringVar(&settings.PrereleaseInfo, "pi", "", "DEPRECATED option name. Use -pr instead")
	flagSet.StringVar(&settings.BranchName, "br", "", "Branch name (use this if you've forked a repo)")
	flagSet.StringVar(&settings.BuildName, "bu", "", "Build name (use this for pre-release builds)")

	flagSet.StringVar(&settings.ArtifactsDest, "d", "", "Destination root directory (default=$GOBIN/(appname)-xc)")
	flagSet.StringVar(&codesignId, "codesign", "", "identity to sign darwin binaries with (only applied when host OS is 'darwin')")

	flagSet.StringVar(&settings.Resources.Include, "include", "", "Include resources in archives (default="+core.RESOURCES_INCLUDE_DEFAULT+")") //TODO: Add resources to non-zips & downloads.md

	//0.2.0 Not easy to 'merge' boolean config items. More flexible to translate them to string options anyway
	flagSet.BoolVar(&isHelp, "h", false, "Help - options")
	flagSet.BoolVar(&isHelp, "help", false, "Help - options")
	flagSet.BoolVar(&isHelpTasks, "ht", false, "Help about tasks")
	flagSet.BoolVar(&isHelpTasks, "h-tasks", false, "Help about tasks")
	flagSet.BoolVar(&isHelpTasks, "help-tasks", false, "Help about tasks")
	flagSet.BoolVar(&isVersion, "version", false, "Print version")

	flagSet.BoolVar(&isVerbose, "v", false, "Verbose")
	flagSet.StringVar(&isCliZipArchives, "z", "", "DEPRECATED (use archive & rmbin tasks instead): create ZIP archives instead of directories (true/false. default=true)")
	flagSet.StringVar(&tasksToRun, "tasks", "", "Tasks to run. Use `goxc -ht` for more details")
	flagSet.StringVar(&tasksPlus, "tasks+", "", "Additional tasks to run. See -ht for tasks list")
	flagSet.StringVar(&tasksMinus, "tasks-", "", "Tasks to exclude. See -ht for tasks list")
	flagSet.BoolVar(&isBuildToolchain, "t", false, "Build cross-compiler toolchain(s). Equivalent to -tasks=toolchain")
	flagSet.BoolVar(&isWriteConfig, "wc", false, "(over)write config. Overwrites are additive. Try goxc -wc to produce a starting point.")
	flagSet.Usage = func() {
		printHelpTopic(flagSet, "options")
	}
	return flagSet
}

func printOptions(flagSet *flag.FlagSet) {
	fmt.Print("Options:\n")
	taskOptions := []string{"t", "tasks+", "tasks-"}
	packageVersioningOptions := []string{"pv", "pi", "br", "bu"}
	deprecatedOptions := []string{"av", "z", "tasks", "h-tasks", "help-tasks", "ht"} //still work but not mentioned
	platformOptions := []string{"os", "arch", "bc"}
	cfOptions := []string{"wc", "c"}
	boolOptions := []string{"h", "v", "version", "t", "wc"}

	//help
	fmt.Printf("  -h <topic>     Help - default topic is 'options'. Also 'tasks', or any task or alias name.\n")
	fmt.Printf("  -ht            Help - show tasks (and task aliases)\n")
	fmt.Printf("  -version       %s\n", flagSet.Lookup("version").Usage)
	fmt.Printf("  -v             %s\n", flagSet.Lookup("v").Usage)

	//tasks
	fmt.Printf("Tasks options:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if core.ContainsString(taskOptions, flag.Name) {
			printFlag(flag, core.ContainsString(boolOptions, flag.Name))
		}
	})

	fmt.Printf("Platform filtering:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if core.ContainsString(platformOptions, flag.Name) {
			printFlag(flag, core.ContainsString(boolOptions, flag.Name))
		}
	})

	fmt.Printf("Config files:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if core.ContainsString(cfOptions, flag.Name) {
			printFlag(flag, core.ContainsString(boolOptions, flag.Name))
		}
	})

	//versioning
	fmt.Printf("Package versioning:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if core.ContainsString(packageVersioningOptions, flag.Name) {
			printFlag(flag, core.ContainsString(boolOptions, flag.Name))
		}
	})

	//most
	fmt.Printf("Other options:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if core.ContainsString(taskOptions, flag.Name) ||
			core.ContainsString(packageVersioningOptions, flag.Name) ||
			core.ContainsString(platformOptions, flag.Name) ||
			core.ContainsString(cfOptions, flag.Name) ||
			core.ContainsString(deprecatedOptions, flag.Name) ||
			core.ContainsString([]string{"h", "help", "h-options", "help-options", "version", "v"}, flag.Name) {
			return
		}
		printFlag(flag, core.ContainsString(boolOptions, flag.Name))
	})
	for _, _ = range []string{"h", "version"} {
	}
}

func printFlag(flag *flag.Flag, isBool bool) {
	var padding string
	if len(flag.Name) < 12 {
		padding = strings.Repeat(" ", 12-len(flag.Name))
	} else {
		padding = ""
	}
	if isBool {
		format := "  -%s  %s%s\n"
		fmt.Printf(format, flag.Name, padding, flag.Usage)
	} else {
		format := "  -%s= %s%s\n"
		fmt.Printf(format, flag.Name, padding, flag.Usage)
	}
}

//TODO user-level config file.
func userHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Printf("Could not get home directory: %s", err)
		return os.Getenv("HOME")
	}
	log.Printf("user dir: %s", usr.HomeDir)
	return usr.HomeDir
}

func main() {
	log.SetPrefix("[goxc] ")
	goXC(os.Args)
}
