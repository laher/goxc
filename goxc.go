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
	"github.com/laher/goxc/tasks"
)

const (
	MSG_HELP               = "Usage: goxc [options] <directory_name>\n"
	MSG_HELP_TOPICS        = "goxc -h <topic>\n"
	MSG_HELP_TOPICS_EG     = "Try:\n\tgoxc -h options\nor\n\tgoxc -h tasks\n"
	MSG_HELP_LINK          = "Please see https://github.com/laher/goxc/wiki for full details.\n"
	MSG_HELP_UNKNOWN_TOPIC = "Unknown topic '%s'. Try 'options' or 'tasks'\n"
	MSG_HELP_DESC          = "goxc cross-compiles go programs to multiple platforms at once."
)

// settings for this invocation of goxc
var (
	// VERSION is initialised by the linker during compilation if the appropriate flag is specified:
	// e.g. go build -ldflags "-X main.VERSION 0.1.2-abcd" goxc.go
	// thanks to minux for this advice
	// So, goxc does this automatically during 'go build'
	VERSION          = "0.5.x"
	settings         config.Settings
	configName       string
	isVersion        bool
	isHelp           bool
	isBuildToolchain bool
	tasksToRun       string
	tasksPlus        string
	tasksMinus       string
	isCliZipArchives string
	codesignId       string
	isWriteConfig    bool
	isVerbose        bool
)

func printHelp(flagSet *flag.FlagSet) {
	args := flagSet.Args()
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, MSG_HELP)
		fmt.Fprintf(os.Stderr, "Version '%s'\n", VERSION)
		fmt.Fprint(os.Stderr, MSG_HELP_DESC)
		fmt.Fprint(os.Stderr, MSG_HELP_TOPICS)
		fmt.Fprint(os.Stderr, MSG_HELP_TOPICS_EG)
	} else {
		printHelpTopic(flagSet, args[0])
	}
}

func printHelpTopic(flagSet *flag.FlagSet, topic string) {
	fmt.Fprint(os.Stderr, MSG_HELP)
	switch topic {
	case "options":
		printOptions(flagSet)
		return
	case "tasks":
		fmt.Fprint(os.Stderr, "Use the '-tasks=' option to specify tasks, and '-tasks-=' or '-tasks+=' to adjust them.\n\te.g. -tasks=default -tasks+=go-fmt -tasks-=rmbin\n")
		fmt.Fprint(os.Stderr, "\nAvailable tasks:\n")
		allTasks := tasks.ListTasks()
		for _, task := range allTasks {
			fmt.Fprintf(os.Stderr, " %s\t%s\n", task.Name, task.Description)
		}
		fmt.Fprint(os.Stderr, "\nTask aliases:\n")
		for alias, taskNames := range tasks.Aliases {
			fmt.Fprintf(os.Stderr, " %s\t%v\n", alias, taskNames)
		}
		return
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
//TODO honour build flags
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
		settings.Resources.Include = config.RESOURCES_INCLUDE_DEFAULT
	}
	if settings.Resources.Exclude == "" {
		settings.Resources.Exclude = config.RESOURCES_EXCLUDE_DEFAULT
	}
	if settings.PackageVersion == "" {
		settings.PackageVersion = config.PACKAGE_VERSION_DEFAULT
	}

	if len(settings.Tasks) == 0 {
		settings.Tasks = config.TASKS_DEFAULT
	}
	if settings.TaskSettings == nil {
		settings.TaskSettings = make(map[string]interface{})
	}

	//fill in per-task settings ...
	for _, t := range tasks.ListTasks() {
		if t.DefaultSettings != nil {
			if _, keyExists := settings.TaskSettings[t.Name]; !keyExists {
				settings.TaskSettings[t.Name] = t.DefaultSettings
			} else {
				//TODO go deeper still
				for k, v := range t.DefaultSettings {
					//log.Printf("k %s => v %s", k, v)
					taskSettings := settings.TaskSettings[t.Name].(map[string]interface{})
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
		err := config.WriteJsonConfig(workingDirectory, config.WrapJsonSettings(settings), configName, false)
		if err != nil {
			log.Printf("Could not write config file: %v", err)
		}
		// 0.2.5 writeConfig now just exits after writing config
	} else {
		//0.2.3 fillDefaults should only happen after writing config
		settings = fillDefaults(settings)
		//remove unwanted tasks here ...
		if tasksMinus != "" {
			removeTasks := tasks.ResolveAliases(strings.Split(tasksMinus, ","))
			for _, val := range removeTasks {
				settings.Tasks = remove(settings.Tasks, val)
			}
		}
		log.Printf("tasks: %v", settings.Tasks)

		if settings.IsVerbose() {
			log.Printf("Final settings %+v", settings)
		}
		//v2.0.0: Removed PKG_VERSION parsing
		destPlatforms := core.GetDestPlatforms(settings.Os, settings.Arch)
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
			settings.Verbosity = config.VERBOSITY_VERBOSE
		}
		if isBuildToolchain {
			tasksToRun = config.TASK_BUILD_TOOLCHAIN + "," + tasksToRun
		}
		if tasksPlus != "" {
			tasksToRun = tasksPlus + "," + tasksToRun
		}
		if tasksToRun != "" {
			settings.Tasks = tasks.ResolveAliases(strings.Split(tasksToRun, ","))
		}
		//0.2.3 NOTE this will be superceded soon
		//using string because that makes it overrideable
		//0.5.0 using Tasks instead of ArtifactTypes
		if isCliZipArchives == "true" || isCliZipArchives == "t" {
			//settings.ArtifactTypes = []string{config.ARTIFACT_TYPE_ZIP}
			settings.Tasks = append_if_missing(settings.Tasks, config.TASK_ARCHIVE)
		} else if isCliZipArchives == "false" || isCliZipArchives == "f" {
			settings.Tasks = remove(settings.Tasks, config.TASK_ARCHIVE)
		}
		//TODO use Setting
		if codesignId != "" {
			settings.SetTaskSetting(config.TASK_CODESIGN, "id", codesignId)
		}
	}
	//log.Printf("Settings: %s", settings)
	if isHelp {

		printHelp(flagSet)
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

	args := flagSet.Args()
	var workingDirectory string
	if len(args) < 1 {
		if isBuildToolchain {
			//default to HOME folder
			log.Printf("Building toolchain, so getting config from HOME directory. To use current folder's config, specify the folder (i.e. goxc -t .)")
			workingDirectory = userHomeDir()
		} else {
			if isVerbose {
				log.Printf("Using config from current folder")
			}
			//default to current folder
			workingDirectory = "."
		}
	} else {
		workingDirectory = args[0]
	}
	log.Printf("Working directory: '%s', Config name: %s", workingDirectory, configName)

	settings, err := mergeConfiguredSettings(workingDirectory, configName, !isWriteConfig)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Configuration file error. %s", err.Error())
			os.Exit(1)
		}
	}
	return workingDirectory, settings
}

func append_if_missing(arr []string, v string) []string {
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
	flagSet.StringVar(&configName, "c", config.CONFIG_NAME_DEFAULT, "config name (default='.goxc')")

	flagSet.StringVar(&settings.Os, "os", "", "Specify OS (default is all - \"linux darwin windows freebsd openbsd\")")
	flagSet.StringVar(&settings.Arch, "arch", "", "Specify Arch (default is all - \"386 amd64 arm\")")

	flagSet.StringVar(&settings.PackageVersion, "pv", "", "Package version (usually [major].[minor].[patch]. default='"+config.PACKAGE_VERSION_DEFAULT+"')")
	flagSet.StringVar(&settings.PackageVersion, "av", "", "DEPRECATED: Package version (deprecated option name)")
	flagSet.StringVar(&settings.PrereleaseInfo, "pr", "", "Prerelease info (usually 'alpha', 'snapshot' ...)")
	flagSet.StringVar(&settings.PrereleaseInfo, "pi", "", "DEPRECATED option name. Use -pr instead")
	flagSet.StringVar(&settings.BranchName, "br", "", "Branch name (use this if you've forked a repo)")
	flagSet.StringVar(&settings.BuildName, "bu", "", "Build name (use this for pre-release builds)")

	flagSet.StringVar(&settings.ArtifactsDest, "d", "", "Destination root directory (default=$GOBIN/(appname)-xc)")
	flagSet.StringVar(&codesignId, "codesign", "", "identity to sign darwin binaries with (only applied when host OS is 'darwin')")

	flagSet.StringVar(&settings.Resources.Include, "include", "", "Include resources in archives (default="+config.RESOURCES_INCLUDE_DEFAULT+")") //TODO: Add resources to non-zips & downloads.md

	//0.2.0 Not easy to 'merge' boolean config items. More flexible to translate them to string options anyway
	flagSet.BoolVar(&isHelp, "h", false, "Help")
	flagSet.BoolVar(&isHelp, "help", false, "Help")
	flagSet.BoolVar(&isVersion, "version", false, "Print version")

	flagSet.BoolVar(&isVerbose, "v", false, "Verbose")
	flagSet.StringVar(&isCliZipArchives, "z", "", "DEPRECATED (use archive & rmbin tasks instead): create ZIP archives instead of folders (true/false. default=true)")
	flagSet.StringVar(&tasksToRun, "tasks", "", "Tasks to run. Use `goxc -h tasks` for more details")
	flagSet.StringVar(&tasksPlus, "tasks+", "", "Additional tasks to run")
	flagSet.StringVar(&tasksMinus, "tasks-", "", "Tasks to exclude")
	flagSet.BoolVar(&isBuildToolchain, "t", false, "Build cross-compiler toolchain(s). Equivalent to -tasks=toolchain")
	flagSet.BoolVar(&isWriteConfig, "wc", false, "(over)write config. Overwrites are additive. Try goxc -wc to produce a starting point.")
	flagSet.Usage = func() {
		printHelpTopic(flagSet, "options")
	}
	return flagSet
}

func printOptions(flagSet *flag.FlagSet) {
	fmt.Print("Options:\n")
	taskOptions := []string{"t", "tasks", "tasks+", "tasks-"}
	versioningOptions := []string{"pv", "pi", "br", "bu"}
	deprecatedOptions := []string{"av", "z"}
	platformOptions := []string{"os", "arch"}
	cfOptions := []string{"wc", "c"}
	boolOptions := []string{"h", "v", "version", "t", "wc"}

	//help
	fmt.Printf("  -h             %s\n", flagSet.Lookup("h").Usage)
	fmt.Printf("  -h options     Show this information\n")
	fmt.Printf("  -h tasks       List all available tasks and task aliases\n")
	fmt.Printf("  -version       %s\n", flagSet.Lookup("version").Usage)
	fmt.Printf("  -v             %s\n", flagSet.Lookup("v").Usage)

	//tasks
	fmt.Printf("Tasks options:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if contains(taskOptions, flag.Name) {
			printFlag(flag, contains(boolOptions, flag.Name))
		}
	})

	fmt.Printf("Platform filtering:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if contains(platformOptions, flag.Name) {
			printFlag(flag, contains(boolOptions, flag.Name))
		}
	})

	fmt.Printf("Config files:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if contains(cfOptions, flag.Name) {
			printFlag(flag, contains(boolOptions, flag.Name))
		}
	})

	//versioning
	fmt.Printf("Package versioning:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if contains(versioningOptions, flag.Name) {
			printFlag(flag, contains(boolOptions, flag.Name))
		}
	})

	//most
	fmt.Printf("Other options:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if contains(taskOptions, flag.Name) ||
			contains(versioningOptions, flag.Name) ||
			contains(platformOptions, flag.Name) ||
			contains(cfOptions, flag.Name) ||
			contains(deprecatedOptions, flag.Name) ||
			contains([]string{"h", "help", "version", "v"}, flag.Name) {
			return
		}
		printFlag(flag, contains(boolOptions, flag.Name))
	})
	for _, _ = range []string{"h", "version"} {
	}
}

func contains(h []string, n string) bool {
	for _, e := range h {
		if e == n {
			return true
		}
	}
	return false
}

func printFlag(flag *flag.Flag, isBool bool) {
	padding := strings.Repeat(" ", 12-len(flag.Name))
	if isBool {
		format := "  -%s  %s%s\n"
		fmt.Printf(format, flag.Name, padding, flag.Usage)
		/*
			} else if strings.Contains(flag.DefValue, " ") {
				// put quotes on the value
				format := "  -%s=%q%s %s\n"
				fmt.Printf(format, flag.Name, flag.DefValue, padding, flag.Usage)
		*/
	} else {
		format := "  -%s= %s%s\n"
		fmt.Printf(format, flag.Name, padding, flag.Usage)
	}
}

//TODO user-level config file.
func userHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Printf("Could not get home folder: %s", err)
		return os.Getenv("HOME")
	}
	log.Printf("user dir: %s", usr.HomeDir)
	return usr.HomeDir
}

func main() {
	log.SetPrefix("[goxc] ")
	goXC(os.Args)
}
