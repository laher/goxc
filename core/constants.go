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

const (
	AMD64   = "amd64"
	X86     = "386"
	ARM     = "arm"
	DARWIN  = "darwin"
	LINUX   = "linux"
	FREEBSD = "freebsd"
	NETBSD  = "netbsd"
	OPENBSD = "openbsd"
	WINDOWS = "windows"
	// Message to install go from source, incase it is missing
	MSG_INSTALL_GO_FROM_SOURCE = "goxc requires Go to be installed from Source. Please follow instructions at http://golang.org/doc/install/source"
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

	TASK_PKG_BUILD = "pkg-build"

	TASKALIAS_CLEAN = "clean"

	TASKALIAS_VALIDATE = "validate"
	TASKALIAS_COMPILE  = "compile"
	TASKALIAS_PACKAGE  = "package"

	TASKALIAS_DEFAULT = "default"
	TASKALIAS_ALL     = "all"

	//0.4 removed in favour of associated tasks
	//ARTIFACT_TYPE_ZIP = "zip"
	//ARTIFACT_TYPE_BIN = "bin"

	CONFIG_NAME_DEFAULT = ".goxc"
)

// Supported platforms
var PLATFORMS = [][]string{
	{DARWIN, X86},
	{DARWIN, AMD64},
	{LINUX, X86},
	{LINUX, AMD64},
	{LINUX, ARM},
	{FREEBSD, X86},
	{FREEBSD, AMD64},
	// {FREEBSD, ARM},
	// couldnt build toolchain for netbsd using a linux 386 host: 2013-02-19
	//	{NETBSD, X86},
	//	{NETBSD, AMD64},
	{OPENBSD, X86},
	{OPENBSD, AMD64},
	{WINDOWS, X86},
	{WINDOWS, AMD64},
}

var (
	TASKS_CLEAN    = []string{TASK_GO_CLEAN, TASK_CLEAN_DESTINATION}
	TASKS_VALIDATE = []string{TASK_GO_VET, TASK_GO_TEST}
	TASKS_COMPILE  = []string{TASK_GO_INSTALL, TASK_XC, TASK_CODESIGN}
	//TASKS_PACKAGE  = []string{TASK_ARCHIVE, TASK_PKG_BUILD, TASK_REMOVE_BIN, TASK_DOWNLOADS_PAGE}
	TASKS_PACKAGE  = []string{TASK_ARCHIVE, TASK_REMOVE_BIN, TASK_DOWNLOADS_PAGE}
	TASKS_DEFAULT  = append(append(append([]string{}, TASKS_VALIDATE...), TASKS_COMPILE...), TASKS_PACKAGE...)
	TASKS_OTHER    = []string{TASK_BUILD_TOOLCHAIN, TASK_GO_FMT}
	TASKS_ALL      = append(append([]string{}, TASKS_OTHER...), TASKS_DEFAULT...)
)
