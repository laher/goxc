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

func GetGoPathElement(workingDirectory string) string {
	//build.Import(path, srcDir string, mode ImportMode) (*Package, error)
	var gopath string
	gopathVar := os.Getenv("GOPATH")
	if gopathVar == "" {
		log.Printf("GOPATH env variable not set! Using '.'")
		gopath = "."
	} else {
		gopaths := filepath.SplitList(gopathVar)
		validGopaths := []string{}
		workingDirectoryAbs, err := filepath.Abs(workingDirectory)
		//log.Printf("workingDirectory %s, (abs) %s", workingDirectory, workingDirectoryAbs)
		if err != nil {
			//strange. TODO: investigate
			workingDirectoryAbs = workingDirectory
		}
		//see if you can match the workingDirectory
		for _, gopathi := range gopaths {
			//if empty or GOROOT, continue
			//logic taken from http://tip.golang.org/src/pkg/go/build/build.go
			if gopathi == "" || gopathi == runtime.GOROOT() || strings.HasPrefix(gopathi, "~") {
				continue
			} else {
				validGopaths = append(validGopaths, gopathi)
			}
			gopathAbs, err := filepath.Abs(gopathi)
			if err != nil {
				//strange. TODO: investigate
				gopathAbs = gopathi
			}
			//log.Printf("gopath element %s, (abs) %s", gopathi, gopathAbs)
			//working directory is inside this path element. Use it!
			if strings.HasPrefix(workingDirectoryAbs, gopathAbs) {
				return gopathi
			}
		}
		if len(validGopaths) > 0 {
			gopath = validGopaths[0]

		} else {
			log.Printf("GOPATH env variable not valid! Using '.'")
			gopath = "."
		}
	}
	return gopath
}

//0.5 add support for space delimiters (similar to BuildConstraints)
//0.5 add support for different oses/services
func GetDestPlatforms(specifiedOses string, specifiedArches string) [][]string {
	destOses := strings.FieldsFunc(specifiedOses, func(r rune) bool { return r == ',' || r == ' ' })
	destArchs := strings.FieldsFunc(specifiedArches, func(r rune) bool { return r == ',' || r == ' ' })

	for _, o := range destOses {
		supported := false
		for _, supportedPlatformArr := range PLATFORMS {
			supportedOs := supportedPlatformArr[0]
			if o == supportedOs {
				supported = true
			}
		}
		if !supported {
			log.Printf("WARNING: Operating System '%s' is unsupported", o)
		}
	}
	for _, o := range destArchs {
		supported := false
		for _, supportedPlatformArr := range PLATFORMS {
			supportedArch := supportedPlatformArr[1]
			if o == supportedArch {
				supported = true
			}
		}
		if !supported {
			log.Printf("WARNING: Architecture '%s' is unsupported", o)
		}
	}
	if len(destOses) == 0 {
		destOses = []string{""}
	}
	if len(destArchs) == 0 {
		destArchs = []string{""}
	}
	var destPlatforms [][]string
	for _, supportedPlatformArr := range PLATFORMS {
		supportedOs := supportedPlatformArr[0]
		supportedArch := supportedPlatformArr[1]
		for _, destOs := range destOses {
			if destOs == "" || supportedOs == destOs {
				for _, destArch := range destArchs {
					if destArch == "" || supportedArch == destArch {
						destPlatforms = append(destPlatforms, supportedPlatformArr)
					}
				}
			}
		}
	}
	if len(destPlatforms) < 1 {
		log.Printf("WARNING: no valid platforms specified")
	}
	return destPlatforms
}

func GetOutDestRoot(appName string, artifactsDestSetting string, workingDirectory string) string {
	var outDestRoot string
	if artifactsDestSetting != "" {
		outDestRoot = artifactsDestSetting
	} else {
		gobin := os.Getenv("GOBIN")
		if gobin == "" {
			gopath := GetGoPathElement(workingDirectory)
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

func ContainsString(h []string, n string) bool {
	for _, e := range h {
		if e == n {
			return true
		}
	}
	return false
}
