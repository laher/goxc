package config

import (
	"github.com/laher/goxc/core"
	"log"
	"runtime"
)

func FillBuildSettingsDefaults(bs *BuildSettings) {
	bs.LdFlagsXVars = &map[string]interface{}{"TimeNow": "main.BUILD_DATE", "Version": "main.VERSION"}
}

//TODO fulfil all defaults
func FillSettingsDefaults(settings *Settings, workingDirectory string) {
	if settings.AppName == "" {
		settings.AppName = core.GetAppName(settings.AppName, workingDirectory)
	}
	if settings.ExecutablePathTemplate == "" {
		settings.ExecutablePathTemplate = core.OUTFILE_TEMPLATE_DEFAULT
	}
	if settings.ResourcesInclude == "" {
		settings.ResourcesInclude = core.RESOURCES_INCLUDE_DEFAULT
	}
	if settings.ResourcesExclude == "" {
		settings.ResourcesExclude = core.RESOURCES_EXCLUDE_DEFAULT
	}
	if settings.PackageVersion == "" {
		settings.PackageVersion = core.PACKAGE_VERSION_DEFAULT
	}
	if settings.BuildSettings == nil {
		bs := BuildSettings{}
		FillBuildSettingsDefaults(&bs)
		settings.BuildSettings = &bs
	}
	if settings.GoRoot == "" {
		if settings.IsVerbose() {
			log.Printf("Defaulting GoRoot to runtime.GOROOT (%s)", runtime.GOROOT())
		}
		settings.GoRoot = runtime.GOROOT()
	}
}
