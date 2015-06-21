// goxc is a build tool with a focus on cross-compilation, plus packaging and deployment features.
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
	"path/filepath"
	"runtime"
	"strings"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/platforms"
	"github.com/laher/goxc/tasks"
	_ "github.com/laher/goxc/tasks/github"
)

const (
	MSG_HELP               = "Usage: goxc [<option(s)>] [<task(s)>]\n"
	MSG_HELP_TOPICS        = "goxc -h <topic>\n"
	MSG_HELP_TOPICS_EG     = "More help:\n\tgoxc -h options\nor\n\tgoxc -h tasks\n"
	MSG_HELP_UNKNOWN_TOPIC = "Unknown topic '%s'. Try 'options' or 'tasks'\n"
)

var (
	// VERSION is initialised by the linker during compilation if the appropriate flag is specified:
	// e.g. go build -ldflags "-X main.VERSION 0.1.2-abcd" goxc.go
	// thanks to minux for this advice
	// So, goxc does this automatically during 'go build'
	VERSION     = "0.17.1"
	BUILD_DATE  = ""
	SOURCE_DATE = "2015-06-21T21:42:03+12:00"
	// settings for this invocation of goxc
	settings             config.Settings
	fBuildSettings       config.BuildSettings
	configName           string
	isVersion            bool
	isHelp               bool
	isHelpTasks          bool
	isBuildToolchain     bool
	tasksToRun           string
	tasksAppend          string
	tasksPrepend         string
	tasksMinus           string
	isCliZipArchives     string
	codesignId           string
	goRoot               string
	isWriteConfig        bool
	isWriteLocalConfig   bool
	isVerbose            bool
	isQuiet              bool
	workingDirectoryFlag string
	buildConstraints     string
	maxProcessors        int
	env                  config.Strslice
)

func printHelp(flagSet *flag.FlagSet) {
	args := flagSet.Args()
	if len(args) < 1 {
		printVersion(os.Stderr)
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
					out, err := json.MarshalIndent(map[string]map[string]interface{}{task.Name: task.DefaultSettings}, "", "\t")
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error displaying TaskSettings info %v\n", err)
					} else {
						fmt.Fprintf(os.Stderr, "Default settings for this task (JSON-formatted):\n \"TaskSettings\": %s \n", string(out))
					}
				} else {
					fmt.Fprintf(os.Stderr, "TaskSettings:\n No TaskSettings available for '%s'\n", task.Name)

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

func printVersion(output *os.File) {
	fmt.Fprintf(output, " goxc version: %s\n", VERSION)
	fmt.Fprintf(output, "  build date: %s\n", BUILD_DATE)
}

// goXC is the goxc startpoint
// In theory you could call this with a slice of flags
func goXC(call []string) error {
	interpretFlags(call)
	workingDirectory := getWorkingDir()
	mergeConfigIntoSettings(workingDirectory)
	if isWriteConfig || isWriteLocalConfig {
		err := config.WriteJsonConfig(workingDirectory, settings, configName, isWriteLocalConfig)
		if err != nil {
			log.Printf("Could not write config file: %v", err)
		}
		//0.2.5 writeConfig now just exits after writing config
		return err
	} else {
		//0.2.3 fillDefaults should only happen after writing config
		config.FillSettingsDefaults(&settings, workingDirectory)
		tasks.FillTaskSettingsDefaults(&settings)

		if settings.IsVerbose() {
			log.Printf("Final settings %+v", settings)
		}
		destPlatforms := platforms.GetDestPlatforms(settings.Os, settings.Arch)
		destPlatforms = platforms.ApplyBuildConstraints(settings.BuildConstraints, destPlatforms)
		err := tasks.RunTasks(workingDirectory, destPlatforms, &settings, maxProcessors)
		if err != nil {
			log.Printf("RunTasks error: %+v", err)
		}
		return err
	}
}

func flagVisitor(f *flag.Flag) {

	switch f.Name {
	case "build-processors":
		settings.BuildSettings.Processors = fBuildSettings.Processors
	case "build-race":
		settings.BuildSettings.Race = fBuildSettings.Race
	case "build-verbose":
		settings.BuildSettings.Verbose = fBuildSettings.Verbose
	case "build-print-commands":
		settings.BuildSettings.PrintCommands = fBuildSettings.PrintCommands
	case "build-ccflags":
		settings.BuildSettings.CcFlags = fBuildSettings.CcFlags
	case "build-compiler":
		settings.BuildSettings.Compiler = fBuildSettings.Compiler
	case "build-gccgoflags":
		settings.BuildSettings.GccGoFlags = fBuildSettings.GccGoFlags
	case "build-gcflags":
		settings.BuildSettings.GcFlags = fBuildSettings.GcFlags
	case "build-installsuffix":
		settings.BuildSettings.InstallSuffix = fBuildSettings.InstallSuffix
	case "build-ldflags":
		settings.BuildSettings.LdFlags = fBuildSettings.LdFlags
	case "build-ldflags-xvars":
		settings.BuildSettings.LdFlagsXVars = fBuildSettings.LdFlagsXVars
	case "build-tags":
		settings.BuildSettings.Tags = fBuildSettings.Tags

	case "env":
		env, ok := f.Value.(*config.Strslice)
		if !ok {
			log.Printf("Type error %v (%T)", f.Value, f.Value)
		}
		settings.Env = append(settings.Env, *env...)
	default:
		//log.Printf("Visiting flag %s", f.Name)
	}
}

func interpretSettings(call []string) string {
	interpretFlags(call)
	workingDirectory := getWorkingDir()
	mergeConfigIntoSettings(workingDirectory)
	return workingDirectory
}

func interpretFlags(call []string) {
	flagSet := setupFlags()
	if err := flagSet.Parse(call[1:]); err != nil {
		log.Printf("Error parsing arguments: %s", err)
		os.Exit(1)
	} else {
		/* TODO: normalize flags for merging with config settings
		specifiedFlags := make(map[string]interface{})
		flagSet.Visit(func(flg *flag.Flag) { specifiedFlags[flg.Name] = flg.Value  })
		log.Printf("Specified cli flags: %s", specifiedFlags)
		*/
		if isQuiet {
			settings.Verbosity = core.VerbosityQuiet
		}
		if isVerbose {
			settings.Verbosity = core.VerbosityVerbose
		}

		//0.6 use args. Parse into slice.
		//settings.Tasks = flagSet.Args()

		//0.10.x: per-task flags
		settings.Tasks, settings.TaskSettings, err = config.ParseCliTasksAndTaskSettings(flagSet.Args())
		if err != nil {
			log.Printf("Error parsing arguments: %s", err)
			os.Exit(1)
		}
		if settings.IsVerbose() {
			log.Printf("Tasks from CLI: %+v", settings.Tasks)
			log.Printf("Task settings from CLI: %+v", settings.TaskSettings)
		}
		settings.BuildSettings = &config.BuildSettings{}
		settings.Env = []string{}
		flagSet.Visit(flagVisitor)
		if settings.IsVerbose() {
			log.Printf("env from flags: %+v", settings.Env)
		}
		//the tasksToRun (-tasks=) flag is only kept incase people used it originally. To be taken out eventually
		if tasksToRun != "" {
			tasksToRunSlice := strings.FieldsFunc(tasksToRun, func(r rune) bool { return r == ',' || r == ' ' })
			settings.Tasks = append(settings.Tasks, tasksToRunSlice...)
		}
		if tasksPrepend != "" {
			settings.TasksPrepend = strings.FieldsFunc(tasksPrepend, func(r rune) bool { return r == ',' || r == ' ' })
		}
		if tasksAppend != "" {
			settings.TasksAppend = strings.FieldsFunc(tasksAppend, func(r rune) bool { return r == ',' || r == ' ' })
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
			settings.Tasks = remove(settings.TasksExclude, tasks.TASKALIAS_ARCHIVE)
			settings.Tasks = appendIfMissing(settings.Tasks, tasks.TASKALIAS_ARCHIVE)
		} else if isCliZipArchives == "false" || isCliZipArchives == "f" {
			settings.TasksExclude = appendIfMissing(settings.TasksExclude, tasks.TASKALIAS_ARCHIVE)
		}
		if codesignId != "" {
			settings.SetTaskSetting(tasks.TASK_CODESIGN, "id", codesignId)
		}
		if buildConstraints != "" {
			//strip single-quotes because these come from the shell and break ApplyBuildConstraints
			buildConstraints = strings.Replace(buildConstraints, "'", "", -1)
			settings.BuildConstraints = buildConstraints
		}
		maxProcessors = calcMaxProcessors(maxProcessors)
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
		printVersion(os.Stderr)
		os.Exit(0)
	}

	//only set it if non-default:
	if goRoot != runtime.GOROOT() && goRoot != "" {
		settings.GoRoot = goRoot
	}
	//sanity check
	if err := core.SanityCheck(goRoot); err != nil {
		log.Printf("Error: %s", err)
		log.Printf(core.MSG_INSTALL_GO_FROM_SOURCE)
		os.Exit(1)
	}
}
func calcMaxProcessors(max int) int {
	processors := runtime.NumCPU()
	if processors > 1 {
		if max > 0 {
			if max > processors {
				return processors
			} else {
				return max
			}
		} else if max < 0 {
			//ok -x
			result := processors + max
			if result < 1 {
				log.Printf("Requested less than one processor (%d %d = %d) - using 1 only!", processors, max, result)
				return 1
			} else {
				return result
			}
		} else {
			//default
			return processors - 1
		}
	} else {
		return 1
	}
}
func getWorkingDir() string {
	//0.6 do NOT use args[0]
	var workingDirectory string
	if workingDirectoryFlag != "" {
		workingDirectory = workingDirectoryFlag
	} else {
		if isBuildToolchain {
			//default to HOME dir
			log.Printf("Building toolchain, so getting config from HOME directory. To use current directory's config, use the wd option (i.e. goxc -t -wd=.)")
			workingDirectory = core.UserHomeDir()
		} else {
			if isVerbose {
				log.Printf("Using config from current directory")
			}
			//default to current directory
			workingDirectory = "."
		}
	}
	workingDirectoryAbs, err := filepath.Abs(workingDirectory)
	if err != nil {
		log.Printf("Could NOT resolve working directory %s", workingDirectory)
	} else {
		workingDirectory = workingDirectoryAbs
	}
	return workingDirectory
}

//merge configuration file
//maybe oneday: parse source
func mergeConfiguredSettings(dir string, configName string, isWriteMain, isWriteLocal bool) error {
	if settings.IsVerbose() {
		log.Printf("loading configured settings")
	}
	configuredSettings, err := config.LoadJsonConfigOverrideable(dir, configName, !isWriteMain && !isWriteLocal, isWriteLocal, settings.IsVerbose())
	if settings.IsVerbose() {
		log.Printf("Settings from config %s: %+v : %v", configName, configuredSettings, err)
	}
	//TODO: further error handling ?
	if err != nil {
		return err
	}
	// v0.14.x merge certain settings (particularly for pkg-build!)
	settings.MergeAliasedTaskSettings(tasks.TASK_ALIASES_FOR_MERGING_SETTINGS)
	configuredSettings.MergeAliasedTaskSettings(tasks.TASK_ALIASES_FOR_MERGING_SETTINGS)
	settings = config.Merge(settings, configuredSettings)
	return err
}

func mergeConfigIntoSettings(workingDirectory string) {
	//0.8 change default name due to new config inheritance rules
	if configName == "" {
		if isWriteConfig || isWriteLocalConfig {
			//for writing, default to the 'base' file.
			configName = core.GOXC_CONFIGNAME_BASE
		} else {
			//for reading, default to 'default'
			configName = core.GOXC_CONFIGNAME_DEFAULT
		}
	}

	err := mergeConfiguredSettings(workingDirectory, configName, isWriteConfig, isWriteLocalConfig)
	//log.Printf("TaskSettings: %+v", settings.TaskSettings)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Configuration file error. %s", err.Error())
			os.Exit(1)
		}
	}
	if settings.IsVerbose() {
		log.Printf("Working directory: '%s', Config name: '%s'", workingDirectory, configName)
	}
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
	flagSet.StringVar(&configName, "c", "", "config name")

	//TODO deprecate?
	flagSet.StringVar(&settings.Os, "os", "", "Specify OS (default is all - \"linux darwin windows freebsd openbsd solaris dragonfly\")")
	flagSet.StringVar(&settings.Arch, "arch", "", "Specify Arch (default is all - \"386 amd64 arm\")")

	//v0.6
	flagSet.StringVar(&buildConstraints, "bc", "", "Specify build constraints (e.g. 'linux,arm windows')")

	flagSet.StringVar(&workingDirectoryFlag, "wd", "", "Specify directory to work on")

	flagSet.StringVar(&settings.PackageVersion, "pv", "", "Package version (usually [major].[minor].[patch]. default='"+core.PACKAGE_VERSION_DEFAULT+"')")
	flagSet.StringVar(&settings.PackageVersion, "av", "", "DEPRECATED: Package version (deprecated option name)")
	flagSet.StringVar(&settings.PrereleaseInfo, "pr", "", "Prerelease info (usually 'alpha', 'snapshot' ...)")
	flagSet.StringVar(&settings.PrereleaseInfo, "pi", "", "DEPRECATED option name. Use -pr instead")
	flagSet.StringVar(&settings.BranchName, "br", "", "Branch name (use this if you've forked a repo)")
	flagSet.StringVar(&settings.BuildName, "bu", "", "Build name (use this for pre-release builds)")
	//	flagSet.StringVar(&settings.PreferredGoVersion, "goversion", "", "Preferred Go version")

	flagSet.StringVar(&settings.ArtifactsDest, "d", "", "Destination root directory (default=$GOBIN/(appname)-xc)")
	flagSet.StringVar(&settings.AppName, "n", "", "Application name. By default this is the directory name.")

	flagSet.StringVar(&settings.OutPath, "o", "", "Output file name for compilation (this string is a template, with default -o=\""+core.OUTFILE_TEMPLATE_DEFAULT+"\")")

	flagSet.StringVar(&codesignId, "codesign", "", "identity to sign darwin binaries with (only applied when host OS is 'darwin')")

	flagSet.StringVar(&settings.ResourcesInclude, "resources-include", "", "Include resources in archives (default="+core.RESOURCES_INCLUDE_DEFAULT+")")
	//deprecated
	flagSet.StringVar(&settings.ResourcesInclude, "include", "", "Include resources in archives (default="+core.RESOURCES_INCLUDE_DEFAULT+")")

	flagSet.StringVar(&settings.ResourcesExclude, "resources-exclude", "", "Include resources in archives (default="+core.RESOURCES_EXCLUDE_DEFAULT+")")
	flagSet.StringVar(&settings.MainDirsExclude, "main-dirs-exclude", "", "Exclude given comma-separated directories from 'main' packages (default="+core.MAIN_DIRS_EXCLUDE_DEFAULT+")")

	//0.2.0 Not easy to 'merge' boolean config items. More flexible to translate them to string options anyway
	flagSet.BoolVar(&isHelp, "h", false, "Help - options")
	flagSet.BoolVar(&isHelp, "help", false, "Help - options")
	flagSet.BoolVar(&isHelpTasks, "ht", false, "Help about tasks")
	flagSet.BoolVar(&isHelpTasks, "h-tasks", false, "Help about tasks")
	flagSet.BoolVar(&isHelpTasks, "help-tasks", false, "Help about tasks")
	flagSet.BoolVar(&isVersion, "version", false, "Print version")

	flagSet.BoolVar(&isVerbose, "v", false, "Verbose output")
	flagSet.BoolVar(&isQuiet, "q", false, "Quiet (no output except for errors)")
	flagSet.StringVar(&isCliZipArchives, "z", "", "DEPRECATED (use archive & rmbin tasks instead): create ZIP archives instead of directories (true/false. default=true)")
	flagSet.StringVar(&tasksToRun, "tasks", "", "Tasks to run. Use `goxc -ht` for more details")
	flagSet.StringVar(&tasksPrepend, "+tasks", "", "Additional tasks to run first. See '-help tasks' for tasks list")
	flagSet.StringVar(&tasksAppend, "tasks+", "", "Additional tasks to run last. See '-help tasks' for tasks list")
	flagSet.StringVar(&tasksMinus, "tasks-", "", "Tasks to exclude. See '-help tasks' for tasks list")
	flagSet.StringVar(&goRoot, "goroot", "", "Specify Go ROOT dir (useful when you have multiple Go installations)")
	flagSet.BoolVar(&isBuildToolchain, "t", false, "Build cross-compiler toolchain(s). Equivalent to -tasks=toolchain")
	flagSet.BoolVar(&isWriteConfig, "wc", false, "(over)write config. Overwrites are additive. Try goxc -wc to produce a starting point.")
	flagSet.BoolVar(&isWriteLocalConfig, "wlc", false, "write 'local' config")

	flagSet.IntVar(&maxProcessors, "max-processors", 0, "Max processors (for parallelizing tasks)")

	//var bs config.BuildSettings
	//bs.Processors = &processors
	//v0.10.x
	fBuildSettings = config.BuildSettings{}
	fBuildSettings.Processors = flagSet.Int("build-processors", 0, "Processors to use during build")
	fBuildSettings.Race = flagSet.Bool("build-race", false, "Build flag 'race'")
	fBuildSettings.Verbose = flagSet.Bool("build-verbose", false, "Build flag 'verbose'")
	fBuildSettings.PrintCommands = flagSet.Bool("build-print-commands", false, "Build flag 'print-commands'")
	fBuildSettings.CcFlags = flagSet.String("build-ccflags", "", "Build flag 'print-commands'")
	fBuildSettings.Compiler = flagSet.String("build-compiler", "", "Build flag 'compiler'")
	fBuildSettings.GccGoFlags = flagSet.String("build-gccgoflags", "", "Build flag")
	fBuildSettings.GcFlags = flagSet.String("build-gcflags", "", "Build flag")
	fBuildSettings.InstallSuffix = flagSet.String("build-installsuffix", "", "Build flag")
	fBuildSettings.LdFlags = flagSet.String("build-ldflags", "", "Build flag")
	fBuildSettings.Tags = flagSet.String("build-tags", "", "Build flag")

	env = config.Strslice{}
	flagSet.Var(&env, "env", "Use env variables")

	flagSet.Usage = func() {
		printHelpTopic(flagSet, "options")
	}
	return flagSet
}

func printOptions(flagSet *flag.FlagSet) {
	fmt.Print("Help Options:\n")
	taskOptions := []string{"t", "tasks+", "tasks-", "+tasks"}
	packageVersioningOptions := []string{"pv", "pr", "br", "bu"}
	deprecatedOptions := []string{"av", "z", "tasks", "h-tasks", "help-tasks", "ht"} //still work but not mentioned
	platformOptions := []string{"os", "arch", "bc"}
	cfOptions := []string{"wc", "c"}
	boolOptions := []string{"h", "v", "version", "t", "wc"}

	//help
	fmt.Printf("  -h <topic>     Help - default topic is 'options'. Also 'tasks', or any task or alias name.\n")
	fmt.Printf("  -ht            Help - show tasks (and task aliases)\n")
	fmt.Printf("  -version       %s\n", flagSet.Lookup("version").Usage)
	fmt.Printf("  -v             %s\n", flagSet.Lookup("v").Usage)

	fmt.Print("Help Topics:\n")
	fmt.Printf("  options 	    default)\n")
	fmt.Printf("  tasks         lists all tasks and aliases\n")
	fmt.Printf("  <task-name>   task description, task options, and default values\n")
	fmt.Printf("  <alias-name>  lists an alias's task(s)\n")

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

	//build
	fmt.Printf("Build:\n")
	flagSet.VisitAll(func(flag *flag.Flag) {
		if strings.HasPrefix(flag.Name, "build-") {
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
			core.ContainsString([]string{"h", "help", "h-options", "help-options", "version", "v"}, flag.Name) ||
			strings.HasPrefix(flag.Name, "build-") {
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

func main() {
	log.SetPrefix("[goxc] ")
	err := goXC(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
