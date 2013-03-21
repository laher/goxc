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

import (
	"errors"
	"fmt"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	//"github.com/laher/goxc/archive"
	//"github.com/laher/goxc/config"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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

func GetMakeScriptPath(goroot string) string {
	gohostos := runtime.GOOS
	var scriptname string
	if gohostos == WINDOWS {
		scriptname = "make.bat"
	} else {
		scriptname = "make.bash"
	}
	return filepath.Join(goroot, "src", scriptname)
}

func SanityCheck(goroot string) error {
	if goroot == "" {
		return errors.New("GOROOT environment variable is NOT set.")
	}
	scriptpath := GetMakeScriptPath(goroot)
	_, err := os.Stat(scriptpath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New(fmt.Sprintf("Make script ('%s') does not exist!", scriptpath))
		} else {
			return errors.New(fmt.Sprintf("Error reading make script ('%s'): %v", scriptpath, err))
		}
	}
	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ParseIncludeResources(basedir string, includeResources string, isVerbose bool) []string {
	allMatches := []string{}
	if includeResources != "" {
		resourceGlobs := strings.Split(includeResources, ",")
		for _, resourceGlob := range resourceGlobs {
			matches, err := filepath.Glob(filepath.Join(basedir, resourceGlob))
			if err == nil {
				allMatches = append(allMatches, matches...)
			} else {
				log.Printf("GLOB error: %s: %s", resourceGlob, err)
			}
		}
	}
	if isVerbose {
		log.Printf("Resources to include: %v", allMatches)
	}
	return allMatches

}

func GetAppName(workingDirectory string) string {
	appDirname, err := filepath.Abs(workingDirectory)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	appName := filepath.Base(appDirname)
	return appName
}

func GetGoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Printf("GOPATH env variable not set! Using '.'")
		gopath = "."
	}
	return gopath
}

func GetOutDestRoot(appName string, artifactsDestSetting string) string {
	var outDestRoot string
	if artifactsDestSetting != "" {
		outDestRoot = artifactsDestSetting
	} else {
		gobin := os.Getenv("GOBIN")
		gopath := GetGoPath()
		if gobin == "" {
			// follow usual GO rules for making GOBIN
			gobin = filepath.Join(gopath, "bin")
		}
		outDestRoot = filepath.Join(gobin, appName+"-xc")
	}
	return outDestRoot
}

func GetRelativeBin(goos, arch string, appName string, isForMarkdown bool, fullVersionName string) string {
	var ending = ""
	if goos == WINDOWS {
		ending = ".exe"
	}
	if isForMarkdown {
		return filepath.Join(goos+"_"+arch, appName+ending)
	}
	relativeDir := filepath.Join(fullVersionName, goos+"_"+arch)
	return filepath.Join(relativeDir, appName+ending)
}
