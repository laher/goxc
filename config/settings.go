package config

import (
	"log"
)

const (
	ARTIFACTS_DEST_DEFAULT    = ""
	BUILD_CONSTRAINTS_DEFAULT = ""
	CODESIGN_DEFAULT          = ""
	// Default resources to include. Comma-separated list of globs.
	RESOURCES_INCLUDE_DEFAULT = "INSTALL*,README*,LICENSE*"
	RESOURCES_EXCLUDE_DEFAULT = "*.go" //TODO
	OS_DEFAULT                = ""
	ARCH_DEFAULT              = ""
	PACKAGE_VERSION_DEFAULT   = "unknown"
	PRERELEASE_INFO_DEFAULT   = "SNAPSHOT"
	VERBOSE_DEFAULT           = false
	ZIP_ARCHIVES_DEFAULT      = false
	IS_BUILDTOOLCHAIN_DEFAULT = false
	BRANCH_ORIGINAL           = "original"

	VERBOSITY_QUIET   = "q" //TODO
	VERBOSITY_DEFAULT = "d"
	VERBOSITY_VERBOSE = "v"

	TASK_BUILD_TOOLCHAIN = "toolchain"

	TASK_CLEAN_DESTINATION = "clean-destination"
	TASK_GO_CLEAN          = "go-clean"

	TASK_GO_VET  = "go-vet"
	TASK_GO_TEST = "go-test"
	TASK_GO_FMT  = "go-fmt"

	TASK_GO_INSTALL = "go-install"
	TASK_XC         = "xc"
	TASK_CODESIGN   = "codesign"

	TASK_ARCHIVE        = "archive" //zip
	TASK_REMOVE_BIN     = "rmbin"   //after zipping
	TASK_DOWNLOADS_PAGE = "downloads-page"

	TASKALIAS_CLEAN    = "clean"
	TASKALIAS_DEFAULT  = "default"
	TASKALIAS_VALIDATE = "validate"
	TASKALIAS_PACKAGE  = "package"
	TASKALIAS_ALL      = "all"

	//0.4 removed in favour of associated tasks
	//ARTIFACT_TYPE_ZIP = "zip"
	//ARTIFACT_TYPE_BIN = "bin"

	CONFIG_NAME_DEFAULT = ".goxc"
)

var (
	TASKS_CLEAN    = []string{TASK_GO_CLEAN, TASK_CLEAN_DESTINATION}
	TASKS_VALIDATE = []string{TASK_GO_VET, TASK_GO_TEST, TASK_GO_INSTALL}
	TASKS_COMPILE  = []string{TASK_GO_INSTALL, TASK_XC, TASK_CODESIGN}
	TASKS_PACKAGE  = []string{TASK_ARCHIVE, TASK_REMOVE_BIN, TASK_DOWNLOADS_PAGE}
	TASKS_DEFAULT  = append(append(append([]string{}, TASKS_VALIDATE...), TASKS_COMPILE...), TASKS_PACKAGE...)
	TASKS_OTHER    = []string{TASK_BUILD_TOOLCHAIN, TASK_GO_FMT}
	TASKS_ALL      = append(append([]string{}, TASKS_OTHER...), TASKS_DEFAULT...)
)

type Resources struct {
	Include string
	Exclude string
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
	Tasks []string

	//TODO: replace Os/Arch with BuildConstraints?
	Arch string
	Os   string

	//TODO: similar to build constraints used by Golang
	// BuildConstraints []string

	Resources Resources

	//versioning
	PackageVersion string
	BranchName     string
	PrereleaseInfo string
	BuildName      string

	//0.2.0 Verbosity replaces Verbose bool
	Verbosity string // none/debug/

	//TaskSettings map[string]map[string]interface{}
	TaskSettings map[string]interface{}
}

func (s Settings) IsVerbose() bool {
	return s.Verbosity == VERBOSITY_VERBOSE
}

func (s Settings) IsBuildToolchain() bool {
	return s.IsTask(TASK_BUILD_TOOLCHAIN)
}

func (s Settings) IsXC() bool {
	return s.IsTask(TASK_XC)
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
