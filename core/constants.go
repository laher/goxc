package core

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

//messages ... int8n?
const (
	// Message to install go from source, incase it is missing
	MSG_INSTALL_GO_FROM_SOURCE = "goxc requires Go to be installed from Source. Please follow instructions at http://golang.org/doc/install/source"
)

//defaults ...
const (
	ARTIFACTS_DEST_TEMPLATE_DEFAULT = "{{.GoBin}}{{.PS}}{{.AppName}}-xc"
	OUTFILE_TEMPLATE_DEFAULT        = "{{.Dest}}{{.PS}}{{.Version}}{{.PS}}{{.Os}}_{{.Arch}}{{.PS}}{{.ExeName}}{{.Ext}}"
	OUTFILE_TEMPLATE_FORMARKDOWN    = "{{.Dest}}{{.PS}}{{.Os}}_{{.Arch}}{{.PS}}{{.ExeName}}{{.Ext}}"
	BUILD_CONSTRAINTS_DEFAULT       = ""
	CODESIGN_DEFAULT                = ""

	// Default resources to include. Comma-separated list of globs.
	RESOURCES_INCLUDE_DEFAULT = "INSTALL*,README*,LICENSE*"
	RESOURCES_EXCLUDE_DEFAULT = "*.go" //TODO
	// Main dirs to exclude by default (Godeps!)
	MAIN_DIRS_EXCLUDE_DEFAULT = "Godeps,testdata"

	OS_DEFAULT              = ""
	ARCH_DEFAULT            = ""
	PACKAGE_VERSION_DEFAULT = "snapshot"
	PRERELEASE_INFO_DEFAULT = "SNAPSHOT"
	BRANCH_ORIGINAL         = "original"

	VerbosityDefault = "d"
	VerbosityQuiet   = "q" //TODO
	VerbosityVerbose = "v"
	//0.4 removed in favour of associated tasks
	//ARTIFACT_TYPE_ZIP = "zip"
	//ARTIFACT_TYPE_BIN = "bin"

	GOXC_FILE_EXT           = ".goxc.json"
	GOXC_LOCAL_FILE_EXT     = ".goxc.local.json"
	GOXC_CONFIGNAME_DEFAULT = "default"
	GOXC_CONFIGNAME_BASE    = ""

	//taskname required by config/json
	TASK_BUILD_TOOLCHAIN = "toolchain"
	//windows required inside core methods
	WINDOWS = "windows"
)
