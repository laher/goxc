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
)

type Resources struct {
	Include string
	Exclude string
}

type Settings struct {
	//build variables
	ArtifactsDest string
	Codesign      string //mac signing identity
	Resources     Resources

	//TODO: deprecate: replace with artifact-types
	ZipArchives   bool
	ArtifactTypes []string //default = 'zip'. Also 'bin'

	//TODO: deprecate: replace with tasks
	IsBuildToolchain bool
	Tasks            []string //TODO: clean,xc,toolchain

	//TODO: replace Os/Arch with BuildConstraints?
	Os               string
	Arch             string
	BuildConstraints string //TODO similar to build constraints used by Golang

	//versioning
	PackageVersion string
	BranchName     string
	PrereleaseInfo string
	BuildName      string

	//invocation settings
	Verbose   bool
	IsHelp    bool
	IsVersion bool
}

// Merge settings together with priority.
// TODO: is there a cleverer way to do this? Reflection, maybe. Maybe not.
func Merge(high Settings, low Settings) Settings {
	if high.ArtifactsDest == "" {
		high.ArtifactsDest = low.ArtifactsDest
	}
	if high.BuildConstraints == BUILD_CONSTRAINTS_DEFAULT {
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
	if high.Verbose == VERBOSE_DEFAULT {
		high.Verbose = low.Verbose
	}
	if high.Codesign == CODESIGN_DEFAULT {
		high.Codesign = low.Codesign
	}

	if high.IsBuildToolchain == IS_BUILDTOOLCHAIN_DEFAULT {
		high.IsBuildToolchain = low.IsBuildToolchain
	}
	if high.ZipArchives == ZIP_ARCHIVES_DEFAULT {
		high.ZipArchives = low.ZipArchives
	}

	return high
}

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
	return settings
}
