package config

const (
	ARTIFACTS_DEST_DEFAULT    = ""
	BUILD_CONSTRAINTS_DEFAULT = ""
	CODESIGN_DEFAULT          = ""
	// Default resources to include. Comma-separated list of globs.
	RESOURCES_INCLUDE_DEFAULT       = "INSTALL*,README*,LICENSE*"
	RESOURCES_EXCLUDE_DEFAULT       = "*.go" //TODO
	OS_DEFAULT                      = ""
	ARCH_DEFAULT                    = ""
	PACKAGE_VERSION_DEFAULT         = "unknown"
	PACKAGE_PRERELEASE_INFO_DEFAULT = ""
	VERBOSE_DEFAULT                 = false
	ZIP_ARCHIVES_DEFAULT            = false
	IS_BUILDTOOLCHAIN_DEFAULT       = false
	BRANCH_ORIGINAL                 = "original"

	VERBOSITY_QUIET			= "q" //TODO
	VERBOSITY_DEFAULT		= "d"
	VERBOSITY_VERBOSE		= "v"

	TASK_BUILD_TOOLCHAIN		= "toolchain"
	TASK_CROSSCOMPILE		= "xc"
	TASK_CLEAN			= "clean" //TODO

	ARTIFACT_TYPE_ZIP		= "zip"
	ARTIFACT_TYPE_DEFAULT		= "bin"

	CONFIG_NAME_DEFAULT		= ".goxc"
)

type Resources struct {
	Include string
	Exclude string
}

type Settings struct {
	ArtifactsDest string
	//0.2.0 ArtifactTypes replaces ZipArchives bool
	ArtifactTypes []string //default = 'zip'. Also 'bin'
	Codesign      string //mac signing identity

	//0.2.0 Tasks replaces IsBuildToolChain bool
	Tasks            []string //TODO: clean,xc,toolchain

	//TODO: replace Os/Arch with BuildConstraints?
	Arch             string
	Os               string
	BuildConstraints string //TODO similar to build constraints used by Golang

	Resources     Resources

	//versioning
	PackageVersion string
	BranchName     string
	PrereleaseInfo string
	BuildName      string

	//0.2.0 Verbosity replaces Verbose bool
	Verbosity   string // none/debug/
}

func (s Settings) IsVerbose() bool {
	return s.Verbosity == VERBOSITY_VERBOSE
}

func (s Settings) IsZip() bool {
	for _, t := range s.ArtifactTypes { if t == ARTIFACT_TYPE_ZIP { return true } }
	return false
}

func (s Settings) IsBuildToolchain() bool {
	for _, t := range s.Tasks { if t == TASK_BUILD_TOOLCHAIN { return true } }
	return false
}

// Merge settings together with priority.
// TODO: is there a cleverer way to do this? Reflection, maybe. Maybe not.
func Merge(high Settings, low Settings) Settings {
	if high.ArtifactsDest == "" {
		high.ArtifactsDest = low.ArtifactsDest
	}
	if high.BuildConstraints == "" {
		high.BuildConstraints = low.BuildConstraints
	}

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
	//TODO: merging of booleans ??
	if high.Verbosity == "" {
		high.Verbosity = low.Verbosity
	}
	if high.Codesign == "" {
		high.Codesign = low.Codesign
	}
	if len(high.Tasks) == 0 {
		high.Tasks = low.Tasks
	}
	if len(high.ArtifactTypes) == 0 {
		high.ArtifactTypes = low.ArtifactTypes
	}

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
		settings.Tasks = []string{TASK_CROSSCOMPILE}
	}
	if len(settings.ArtifactTypes) == 0 {
		settings.ArtifactTypes = []string{ARTIFACT_TYPE_ZIP}
	}
	return settings
}
