package config

import (
	//"log"
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
	TASK_CLEAN           = "clean"
	TASK_TEST            = "test"
	TASK_VET             = "vet"
	TASK_FMT             = "fmt"
	TASK_XC              = "xc"
	TASK_CODESIGN        = "codesign"
	TASK_INSTALL         = "install"
	TASK_ARCHIVE         = "archive" //zip
	TASK_REMOVE_BIN      = "rmbin"   //after zipping
	TASK_DOWNLOADS_PAGE  = "downloads-page"

	//ARTIFACT_TYPE_ZIP = "zip"
	//ARTIFACT_TYPE_BIN = "bin"

	CONFIG_NAME_DEFAULT = ".goxc"
)

var (
	TASKS_DEFAULT = []string{TASK_CLEAN, TASK_VET, TASK_TEST, TASK_INSTALL, TASK_CODESIGN, TASK_XC, TASK_ARCHIVE, TASK_REMOVE_BIN, TASK_DOWNLOADS_PAGE}
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

	TaskSettings map[string]map[string]interface{}
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
	/* 0.5.0 codesign setting is replaced by codesign task
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

	return high
}

//TODO fulfil all defaults
func FillDefaults(settings Settings) Settings {
	if settings.Resources.Include == "" {
		settings.Resources.Include = RESOURCES_INCLUDE_DEFAULT
	}
	if settings.Resources.Exclude == "" {
		settings.Resources.Exclude = RESOURCES_EXCLUDE_DEFAULT
	}
	if settings.PackageVersion == "" {
		settings.PackageVersion = PACKAGE_VERSION_DEFAULT
	}

	if len(settings.Tasks) == 0 {
		settings.Tasks = TASKS_DEFAULT
	}
	/* 0.5.0 see tasks
	if len(settings.ArtifactTypes) == 0 {
		settings.ArtifactTypes = []string{ARTIFACT_TYPE_ZIP}
	}
	*/
	return settings
}
