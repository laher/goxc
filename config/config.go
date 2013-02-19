package config

type Settings struct {
	Verbose          bool
	//TODO: replace with list of artifact types
	ZipArchives      bool
	//TODO: replace Os/Arch with +build type arguments
	Os              string
	Arch             string

	PackageVersion   string
	ArtifactsDest    string
	Codesign         string
	IncludeResources string

	IsHelp           bool
	IsVersion        bool
	IsBuildToolchain bool
}

