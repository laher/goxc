// config package handles invocation settings for goxc, which can be set using a combination of cli flags plus json-based config files.
package config

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
	"log"

	"github.com/laher/goxc/core"
	"github.com/laher/goxc/typeutils"
)

/* 0.9: removed!
// Resources are files which a user might want to include or exclude from their package or archive
type Resources struct {
	Include string `json:",omitempty"`
	Exclude string `json:",omitempty"`
}
*/

// Invocation settings
type Settings struct {
	AppName       string `json:",omitempty"`
	ArtifactsDest string `json:",omitempty"`
	//0.13.x. If this starts with a FileSeparator then ignore top dir
	OutPath string `json:",omitempty"`

	//0.2.0 ArtifactTypes replaces ZipArchives bool
	//0.5.0 ArtifactTypes is replaced by tasks
	//ArtifactTypes []string //default = 'zip'. Also 'bin'
	//0.5.0 Codesign replaced by codesign task
	//Codesign      string   //mac signing identity

	//0.2.0 Tasks replaces IsBuildToolChain bool
	//0.5.0 Tasks is a much longer list.
	Tasks []string `json:",omitempty"`

	//0.5.0 adding exclusions. Easier for dealing with aliases. (e.g. Tasks=[default], TasksExclude=[rmbin] is easier than specifying individual tasks)
	TasksExclude []string `json:",omitempty"`

	//0.5.0 adding extra tasks.
	TasksAppend []string `json:",omitempty"`
	//0.9.9 adding 'prepend'
	TasksPrepend []string `json:",omitempty"`

	//0.6 complement Os/Arch with BuildConstraints
	Arch string `json:",omitempty"`
	Os   string `json:",omitempty"`
	//NEW 0.5.5 - implemented 0.5.7
	BuildConstraints string `json:",omitempty"`

	//0.7.x experimental option - only 0.7.3. Removed 0.7.4
	//PrependCurrentEnv string `json:",omitempty"`

	//0.9 changed from struct to ResourcesInclude & ResourcesExclude
	ResourcesInclude string `json:",omitempty"`
	ResourcesExclude string `json:",omitempty"`

	//0.10.x source exclusion
	MainDirsExclude string `json:",omitempty"`
	//0.13.x source exclusion (source dirs)
	SourceDirsExclude string `json:",omitempty"`

	//versioning
	PackageVersion string `json:",omitempty"`
	BranchName     string `json:",omitempty"`
	PrereleaseInfo string `json:",omitempty"`
	BuildName      string `json:",omitempty"`

	//0.2.0 Verbosity replaces Verbose bool
	Verbosity string `json:",omitempty"` // none/debug/

	//TaskSettings map[string]map[string]interface{}
	TaskSettings map[string]map[string]interface{} `json:",omitempty"`

	//DEPRECATED (since v0.9. See GoxcConfigVersion)
	FormatVersion string `json:",omitempty"`

	//v0.9, to replace 'FormatVersion'
	GoxcConfigVersion string `json:"ConfigVersion,omitempty"`

	BuildSettings *BuildSettings `json:",omitempty"`

	GoRoot string `json:"-"` //only settable by a flag

	//TODO?
	//PreferredGoVersion string `json:",omitempty"` //try to use a go version...

	//v0.10.x
	Env []string `json:",omitempty"`
}

func (s *Settings) IsVerbose() bool {
	return s.Verbosity == core.VerbosityVerbose
}

func (s *Settings) IsQuiet() bool {
	return s.Verbosity == core.VerbosityQuiet
}

func (s *Settings) IsTask(taskName string) bool {
	for _, t := range s.Tasks {
		if t == taskName {
			return true
		}
	}
	return false
}

func (s *Settings) SetTaskSetting(taskName, settingName string, value interface{}) {
	if s.TaskSettings == nil {
		s.TaskSettings = make(map[string]map[string]interface{})
	}
	if value, keyExists := s.TaskSettings[taskName]; keyExists {
		//ok
	} else {
		value = make(map[string]interface{})
		s.TaskSettings[taskName] = value
	}
	value.(map[string]interface{})[settingName] = value
}

// this helps whenever a 'task' gets refactored to become an 'alias' (e.g. 'pkg-build' task renamed to 'deb', and 'pkg-build' became a task alias)
// note that the task settings take precedence over the alias' settings.
func (s *Settings) MergeAliasedTaskSettings(aliases map[string][]string) {
	//for each per-task setting name/value pair:
	for taskSettingTaskName, taskSettingValue := range s.TaskSettings {
		//loop through aliases, getting alias,tasks:
		for staticSettingAlias, aliasedTasks := range aliases {
			// if the specified TaskSetting == this alias:
			if staticSettingAlias == taskSettingTaskName {
				//merge into results ...
				for _, aliasedTask := range aliasedTasks {
					aliasedTaskValue, exists := s.TaskSettings[aliasedTask]
					//if settings already exists for the actual task - merge
					if exists {
						if s.IsVerbose() {
							log.Printf("Task settings specified for %s. Merging %v with %v", aliasedTask, aliasedTaskValue, taskSettingValue)
						}
						// merge
						s.TaskSettings[aliasedTask] = typeutils.MergeMaps(aliasedTaskValue, taskSettingValue)
					} else {
						if s.IsVerbose() {
							log.Printf("alias didnt exist. Setting %s to %v", aliasedTask, taskSettingValue)
						}
						//add ...
						s.TaskSettings[aliasedTask] = taskSettingValue
					}
				}
			}
		}
	}
}

func (s *Settings) GetTaskSetting(taskName, settingName string) interface{} {
	if value, keyExists := s.TaskSettings[taskName]; keyExists {
		taskMap := value //.(map[string]interface{})
		if settingValue, keyExists := taskMap[settingName]; keyExists {
			return settingValue
		}
	} else {
		if s.IsVerbose() {
			log.Printf("No settings for task '%s'", taskName)
			log.Printf("All task settings: %+v", s.TaskSettings)
		}
	}
	return nil
}
func (s *Settings) GetTaskSettingMap(taskName, settingName string) map[string]interface{} {
	retUntyped := s.GetTaskSetting(taskName, settingName)
	if retUntyped == nil {
		return nil
	}
	mp, err := typeutils.ToMap(retUntyped, taskName+"."+settingName)
	if err != nil {
		//already logged
	}
	return mp

}

func (s *Settings) GetTaskSettingStringSlice(taskName, settingName string) []string {
	retUntyped := s.GetTaskSetting(taskName, settingName)
	if retUntyped == nil {
		return []string{}
	}
	strSlice, err := typeutils.ToStringSlice(retUntyped, taskName+"."+settingName)
	if err != nil {
		//already logged
	}
	return strSlice
}

func (s *Settings) GetTaskSettingString(taskName, settingName string) string {
	retUntyped := s.GetTaskSetting(taskName, settingName)
	if retUntyped == nil {
		return ""
	}
	str, err := typeutils.ToString(retUntyped, taskName+"."+settingName)
	if err != nil {
		//already logged
	}
	return str
}

func (s *Settings) GetTaskSettingBool(taskName, settingName string) bool {
	retUntyped := s.GetTaskSetting(taskName, settingName)
	if s.IsVerbose() {
		log.Printf("Setting %s.%s resolves to %v", taskName, settingName, retUntyped)
	}
	if retUntyped == nil {
		return false
	}
	ret, err := typeutils.ToBool(retUntyped, taskName+"."+settingName)
	if err != nil {
		//already logged
	}
	return ret
}

func (s *Settings) GetTaskSettingInt(taskName, settingName string, defaultValue int) int {
	retUntyped := s.GetTaskSetting(taskName, settingName)
	if retUntyped == nil {
		return defaultValue
	}
	ret, err := typeutils.ToInt(retUntyped, taskName+"."+settingName)
	if err != nil {
		//already logged
	}
	return ret
}

//Builds version name from PackageVersion, BranchName, PrereleaseInfo, BuildName
//This breakdown is mainly based on 'semantic versioning' See http://semver.org/
//The difference being that you can specify a branch name (which becomes part of the 'prerelease info' as named by semver)
func (settings *Settings) GetFullVersionName() string {
	versionName := settings.PackageVersion
	if settings.BranchName != "" {
		versionName += "-" + settings.BranchName
	}
	if settings.PrereleaseInfo != "" {
		if settings.BranchName == "" {
			versionName += "-"
		} else {
			versionName += "."
		}
		versionName += settings.PrereleaseInfo
	}
	if settings.BuildName != "" {
		versionName += "+b" + settings.BuildName
	}
	return versionName
}

// DEPRECATED!! Merge settings together with priority.
// TODO: deprecate in favour of map merge.
func Merge(high Settings, low Settings) Settings {
	if high.ArtifactsDest == "" {
		high.ArtifactsDest = low.ArtifactsDest
	}
	if high.OutPath == "" {
		high.OutPath = low.OutPath
	}
	//0.6 Adding BuildConstraints
	if high.BuildConstraints == "" {
		high.BuildConstraints = low.BuildConstraints
	}

	if high.ResourcesInclude == "" {
		high.ResourcesInclude = low.ResourcesInclude
	}
	if high.ResourcesExclude == "" {
		high.ResourcesExclude = low.ResourcesExclude
	}
	if high.MainDirsExclude == "" {
		high.MainDirsExclude = low.MainDirsExclude
	}
	if high.AppName == "" {
		high.AppName = low.AppName
	}
	if high.Arch == "" {
		high.Arch = low.Arch
	}

	if high.Os == "" {
		high.Os = low.Os
	}
	if high.BuildConstraints == "" {
		high.BuildConstraints = low.BuildConstraints
	}
	if high.PackageVersion == "" {
		high.PackageVersion = low.PackageVersion
	}
	if high.BranchName == "" {
		high.BranchName = low.BranchName
	}
	if high.PrereleaseInfo == "" {
		high.PrereleaseInfo = low.PrereleaseInfo
	}
	if high.BuildName == "" {
		high.BuildName = low.BuildName
	}
	//TODO: merging of booleans ??
	if high.Verbosity == "" {
		high.Verbosity = low.Verbosity
	}
	//0.5.0 codesign setting is replaced by task setting 'id'
	if len(high.Tasks) == 0 {
		high.Tasks = low.Tasks
	}
	//0.6 fixed missing 'merge'
	if len(high.TasksAppend) == 0 {
		high.TasksAppend = low.TasksAppend
	}
	//0.10 adding 'prepend'
	if len(high.TasksPrepend) == 0 {
		high.TasksPrepend = low.TasksPrepend
	}
	if len(high.TasksExclude) == 0 {
		high.TasksExclude = low.TasksExclude
	}
	//0.5.0 replaced ArtifactTypes
	if len(high.TaskSettings) == 0 {
		high.TaskSettings = low.TaskSettings
	} else {
		high.TaskSettings = typeutils.MergeMapsStringMapStringInterface(high.TaskSettings, low.TaskSettings)
	}
	//0.9 BuildSettings
	if high.BuildSettings == nil || high.BuildSettings.IsEmpty() {
		high.BuildSettings = low.BuildSettings
	} else if low.BuildSettings != nil {
		if high.BuildSettings.Processors == nil {
			high.BuildSettings.Processors = low.BuildSettings.Processors
		}
		if high.BuildSettings.Race == nil {
			high.BuildSettings.Race = low.BuildSettings.Race
		}
		if high.BuildSettings.Verbose == nil {
			high.BuildSettings.Verbose = low.BuildSettings.Verbose
		}
		if high.BuildSettings.PrintCommands == nil {
			high.BuildSettings.PrintCommands = low.BuildSettings.PrintCommands
		}
		if high.BuildSettings.CcFlags == nil {
			high.BuildSettings.CcFlags = low.BuildSettings.CcFlags
		}
		if high.BuildSettings.Compiler == nil {
			high.BuildSettings.Compiler = low.BuildSettings.Compiler
		}
		if high.BuildSettings.GccGoFlags == nil {
			high.BuildSettings.GccGoFlags = low.BuildSettings.GccGoFlags
		}
		if high.BuildSettings.GcFlags == nil {
			high.BuildSettings.GcFlags = low.BuildSettings.GcFlags
		}
		if high.BuildSettings.InstallSuffix == nil {
			high.BuildSettings.InstallSuffix = low.BuildSettings.InstallSuffix
		}
		if high.BuildSettings.LdFlags == nil {
			high.BuildSettings.LdFlags = low.BuildSettings.LdFlags
		}
		if high.BuildSettings.LdFlagsXVars == nil {
			high.BuildSettings.LdFlagsXVars = low.BuildSettings.LdFlagsXVars
		}
		if high.BuildSettings.Tags == nil {
			high.BuildSettings.Tags = low.BuildSettings.Tags
		}
	}
	if len(high.Env) == 0 {
		high.Env = low.Env
	}
	return high
}
