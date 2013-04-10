// GOXC IS NOT READY FOR USE AS AN API - function names and packages will continue to change until version 1.0
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
	"github.com/laher/goxc/core"
	"log"
)

type Resources struct {
	Include string `json:",omitempty"`
	Exclude string `json:",omitempty"`
}

type Settings struct {
	ArtifactsDest string
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

	//0.5.0 adding extra tasks. TODO (maybe) - prepend
	TasksAppend []string `json:",omitempty"`

	//TODO: replace Os/Arch with BuildConstraints?
	Arch string `json:",omitempty"`
	Os   string `json:",omitempty"`

	//NEW 0.5.5 - implemented 0.5.7
	BuildConstraints string `json:",omitempty"`

	Resources Resources `json:",omitempty"`

	//versioning
	PackageVersion string `json:",omitempty"`
	BranchName     string `json:",omitempty"`
	PrereleaseInfo string `json:",omitempty"`
	BuildName      string `json:",omitempty"`

	//0.2.0 Verbosity replaces Verbose bool
	Verbosity string  `json:",omitempty"` // none/debug/

	//TaskSettings map[string]map[string]interface{}
	TaskSettings map[string]interface{} `json:",omitempty"`

	//for 0.6.0, to replace 'FormatVersion'
	GoxcConfigVersion string `json:"FormatVersion,omitempty"`
}

func (s Settings) IsVerbose() bool {
	return s.Verbosity == core.VERBOSITY_VERBOSE
}

func (s Settings) IsTask(taskName string) bool {
	for _, t := range s.Tasks {
		if t == taskName {
			return true
		}
	}
	return false
}

func (s Settings) SetTaskSetting(taskName, settingName string, value interface{}) {
	if s.TaskSettings == nil {
		s.TaskSettings = make(map[string]interface{})
	}
	if value, keyExists := s.TaskSettings[taskName]; keyExists {
		//ok
	} else {
		value = make(map[string]interface{})
		s.TaskSettings[taskName] = value
	}
	value.(map[string]interface{})[settingName] = value
}

func (s Settings) GetTaskSetting(taskName, settingName string) interface{} {
	if value, keyExists := s.TaskSettings[taskName]; keyExists {
		taskMap := value.(map[string]interface{})
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

func (settings Settings) GetFullVersionName() string {
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

// Merge settings together with priority.
// TODO: is there a cleverer way to do this? Reflection, maybe. Maybe not.
func Merge(high Settings, low Settings) Settings {
	if high.ArtifactsDest == "" {
		high.ArtifactsDest = low.ArtifactsDest
	}
	/* TODO - BuildConstraints
	if high.BuildConstraints == "" {
		high.BuildConstraints = low.BuildConstraints
	}
	*/
	if high.Resources.Exclude == "" {
		high.Resources.Exclude = low.Resources.Exclude
	}
	if high.Resources.Include == "" {
		high.Resources.Include = low.Resources.Include
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
	/* 0.5.0 codesign setting is replaced by task setting 'id'
	if high.Codesign == "" {
		high.Codesign = low.Codesign
	}
	*/
	if len(high.Tasks) == 0 {
		high.Tasks = low.Tasks
	}
	if len(high.TasksExclude) == 0 {
		high.TasksExclude = low.TasksExclude
	}
	/* 0.5.0 replaced.
	if len(high.ArtifactTypes) == 0 {
		high.ArtifactTypes = low.ArtifactTypes
	}
	*/
	if len(high.TaskSettings) == 0 {
		high.TaskSettings = low.TaskSettings
	} else {
		high.TaskSettings = mergeMaps(high.TaskSettings, low.TaskSettings)
	}
	return high
}

func mergeMaps(high, low map[string]interface{}) map[string]interface{} {
	//log.Printf("Merging %+v with %+v", high, low)
	if high == nil {
		return low
	}
	for key, lowVal := range low {
		//log.Printf("Merging key %s", key)
		if highVal, keyExists := high[key]; keyExists {
			// NOTE: go deeper for maps.
			// (Slices and other types should not go deeper)
			switch highValTyped := highVal.(type) {
			case map[string]interface{}:
				switch lowValTyped := lowVal.(type) {
				case map[string]interface{}:
					//log.Printf("Go deeper for key '%s'", key)
					high[key] = mergeMaps(highValTyped, lowValTyped)
				}
			}
		} else {
			high[key] = lowVal
		}
	}
	return high
}
