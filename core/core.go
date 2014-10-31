//Some general utility functions for goxc
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
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

const ()

// get the path to the 'make' batch script within the GO source tree.
// i.e. runtime.GOOS / src / make.bat|bash
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

// Basic system sanity check. Checks GOROOT is set and 'make' batch script exists.
// TODO: in future this could check for existence of gcc/mingw/alternative
func SanityCheck(goroot string) error {
	if goroot == "" {
		goroot = runtime.GOROOT()
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

// simple fileExists method which inspects the error from os.Stat
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ParseCommaGlobs(commaGlob string) []string {
	if commaGlob == "" {
		return []string{}
	}
	globs := strings.Split(commaGlob, ",")
	//normalize
	//treat slashes/backslashes as the same thing ...
	for i, glob := range globs {
		glob := strings.Replace(glob, "/", string(os.PathSeparator), -1)
		glob = strings.Replace(glob, "\\", string(os.PathSeparator), -1)
		globs[i] = glob
	}
	return globs
}
func resolveToFiles(item string) ([]string, error) {
	fi, err := os.Lstat(item)
	if err != nil {
		return []string{}, err
	}
	if fi.IsDir() {
		files, err := dirToFiles(item)
		return files, err
	} else {
		return []string{item}, nil
	}
}
func dirToFiles(dir string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// Glob parser for 'Include resources'
// TODO generalise for exclude resources and any other globs.
func ParseIncludeResources(basedir, includeResources, excludeResources string, isVerbose bool) []string {
	allMatches := []string{}
	if includeResources != "" {
		resourceGlobs := ParseCommaGlobs(includeResources)
		if isVerbose {
			log.Printf("IncludeGlobs: %v", resourceGlobs)
		}
		excludeGlobs := ParseCommaGlobs(excludeResources)
		if isVerbose {
			log.Printf("ExcludeGlobs: %v", excludeGlobs)
		}
		for _, resourceGlob := range resourceGlobs {
			matches, err := filepath.Glob(filepath.Join(basedir, resourceGlob))
			if err != nil {
				//ignore this inclusion glob
				log.Printf("GLOB error: %s: %s", resourceGlob, err)
			} else {
				for _, match := range matches {
					files, err := resolveToFiles(match)

					if err != nil {
						//ignore this match
						log.Printf("dir lookup error: %s: %s", match, err)
					} else {
						for _, file := range files {
							exclude := false
							for _, excludeGlob := range excludeGlobs {
								//incomplete!!
								if !strings.Contains(excludeGlob, string(os.PathSeparator)) {
									excludeGlob = filepath.Join(filepath.Dir(file), excludeGlob)
								}
								excludedThisTime, err := filepath.Match(excludeGlob, file)
								if isVerbose {
									log.Printf("Globbing %s for exclusion %s", file, excludeGlob)
								}
								if err != nil {
									//ignore this exclusion glob
									log.Printf("Exclude-GLOB error: %s: %s", excludeGlob, err)
								}
								if excludedThisTime {
									if isVerbose {
										log.Printf("Excluded: %s with %s", file, excludeGlob)
									}
									exclude = true
								}
							}
							if !exclude {
								//return relative filename
								relativeFilename, err := filepath.Rel(basedir, file)
								if err != nil {
									log.Printf("Warning: file %s is not inside %s", file, basedir)
									allMatches = append(allMatches, file)
								} else {
									allMatches = append(allMatches, relativeFilename)
								}
							}
						}
					}
				}
			}
		}
	}
	if isVerbose {
		log.Printf("Resources to include: %v", allMatches)
	}
	return allMatches

}

// Get application name (defaults to dirname)
func GetAppName(specifiedAppName, workingDirectory string) string {
	if specifiedAppName != "" {
		return specifiedAppName
	}
	appDirname, err := filepath.Abs(workingDirectory)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	appName := filepath.Base(appDirname)
	return appName
}

// Tries to find the most relevant GOPATH element.
// First, tries to find an element which is a parent of the current directory.
// If not, it uses the first one.
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

func GetOutDestRoot(appName string, workingDirectory string, templateText string) (string, error) {
	if templateText == "" {
		templateText = ARTIFACTS_DEST_TEMPLATE_DEFAULT
	}
	var outDestRoot string
	tmpl, err := template.New("rootTemplate").Parse(templateText)
	goBin := GoBin(workingDirectory)
	homeDir := UserHomeDir()
	myGoPath := GetGoPathElement(workingDirectory)
	data := RootDirVars{goBin, myGoPath, homeDir, appName, string(os.PathSeparator)}
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		return "", err
	}
	outDestRoot = out.String()
	//normaliseAbs
	if strings.HasPrefix(outDestRoot, "~/") {
		outDestRoot = strings.Replace(outDestRoot, "~", UserHomeDir(), 1)
	}
	outDestRootAbs, err := filepath.Abs(outDestRoot)
	if err != nil {
		log.Printf("Error resolving absolute filename")
		return outDestRoot, nil
	} else {
		return outDestRootAbs, nil
	}
}

func GoBin(workingDirectory string) string {
	gobin := os.Getenv("GOBIN")
	if gobin == "" {
		gopath := GetGoPathElement(workingDirectory)
		// follow usual GO rules for making GOBIN
		gobin = filepath.Join(gopath, "bin")
	}
	return gobin
}

func UserHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Printf("Could not get home directory: %s", err)
		return os.Getenv("HOME")
	}
	//log.Printf("user dir: %s", usr.HomeDir)
	return usr.HomeDir
}

type RootDirVars struct {
	GoBin    string
	MyGoPath string
	Home     string
	AppName  string
	PS       string
}
type BinNameVars struct {
	RootDirVars
	Dest    string
	ExeName string
	Version string
	Os      string
	Arch    string
	Ext     string
}

func GetAbsoluteBin(goos, arch string, appName, exeName, workingDirectory, fullVersionName, templateText string, artifactsDestSetting string) (string, error) {
	tmpl, err := template.New("binTemplate").Parse(templateText)
	if err != nil {
		return "", err
	}
	var ending = ""
	if goos == WINDOWS {
		ending = ".exe"
	}
	rootTemplate := artifactsDestSetting

	root, err := GetOutDestRoot(appName, workingDirectory, rootTemplate)
	goBin := GoBin(workingDirectory)
	homeDir := UserHomeDir()
	myGoPath := GetGoPathElement(workingDirectory)
	rdv := RootDirVars{goBin, myGoPath, homeDir, appName, string(os.PathSeparator)}
	data := BinNameVars{rdv, root, exeName, fullVersionName, goos, arch, ending}
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

// Check if slice contains a string.
// DEPRECATED: use equivalent func inside typeutils.
func ContainsString(h []string, n string) bool {
	for _, e := range h {
		if e == n {
			return true
		}
	}
	return false
}
