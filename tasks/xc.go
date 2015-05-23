package tasks

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
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/argo/ar"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/executils"
	"github.com/laher/goxc/exefileparse"
	"github.com/laher/goxc/platforms"
)

//runs automatically
func init() {
	//GOARM=6 (this is the default for go1.1
	RegisterParallelizable(ParallelizableTask{
		TASK_XC,
		"Cross compile. Builds executables for other platforms.",
		setupXc,
		runXc,
		nil,
		map[string]interface{}{"GOARM": "",
			//"validation" : "tcBinExists,exeParse",
			"validateToolchain":    false,
			"verifyExe":            false,
			"autoRebuildToolchain": false}})
}

func setupXc(tp TaskParams) ([]platforms.Platform, error) {

	if len(tp.DestPlatforms) == 0 {
		return []platforms.Platform{}, errors.New("No valid platforms specified")
	}

	isValidateToolchain := tp.Settings.GetTaskSettingBool(TASK_XC, "validateToolchain")
	goroot := tp.Settings.GoRoot
	for _, dest := range tp.DestPlatforms {
		if isValidateToolchain {
			err := validateToolchain(dest, goroot, tp.Settings.IsVerbose())
			if err != nil {
				log.Printf("Toolchain not ready for %v. Re-building toolchain. (%v)", dest, err)
				isAutoToolchain := tp.Settings.GetTaskSettingBool(TASK_XC, "autoRebuildToolchain")
				if isAutoToolchain {
					err = buildToolchain(dest.Os, dest.Arch, tp.Settings)
				}
				if err != nil {
					return nil, err
				}
			}
		}
	}
	//check for duplicate exePaths
	exePaths := []string{}
	for _, mainDir := range tp.MainDirs {
		var exeName string
		if len(tp.MainDirs) == 1 {
			exeName = tp.Settings.AppName
		} else {
			exeName = filepath.Base(mainDir)
		}
		for _, dest := range tp.DestPlatforms {
			absoluteBin, err := core.GetAbsoluteBin(dest.Os, dest.Arch, tp.Settings.AppName, exeName, tp.WorkingDirectory, tp.Settings.GetFullVersionName(), tp.Settings.OutPath, tp.Settings.ArtifactsDest)
			if err != nil {
				return nil, err
			}
			for _, existingPath := range exePaths {
				if existingPath == absoluteBin {
					return []platforms.Platform{}, errors.New("The xc task will attempt to compile multiple binaries to the same path (" + absoluteBin + "). Please make sure {{.Os}} and {{.Arch}} variables are used in the OutPath. Currently the template is " + tp.Settings.OutPath)
				}
			}
			exePaths = append(exePaths, absoluteBin)
		}
	}
	return tp.DestPlatforms, nil
}

func runXc(tp TaskParams, dest platforms.Platform, errchan chan error) {
	/*
		//outDestRoot, err := core.GetOutDestRoot(tp.AppName, tp.WorkingDirectory, tp.Settings.ArtifactsDest)
		if err != nil {
			log.Printf("Error: %v", err)
			errchan <- err
			return
		}
	*/
	if tp.Settings.IsVerbose() {
		log.Printf("mainDirs : %v", tp.MainDirs)
	}
	for _, mainDir := range tp.MainDirs {
		var exeName string
		var packagePath string
		if len(tp.MainDirs) == 1 {
			exeName = tp.Settings.AppName
		} else {
			exeName = filepath.Base(mainDir)
		}
		packagePath = mainDir
		absoluteBin, err := xcPlat(dest, tp, exeName, packagePath)
		if err != nil {
			log.Printf("Error: %v", err)
			log.Printf("Have you run `goxc -t` for this platform (%s,%s)???", dest.Arch, dest.Os)
			errchan <- err
			return
		} else {
			isVerifyExe := tp.Settings.GetTaskSettingBool(TASK_XC, "verifyExe")
			if isVerifyExe {
				err = exefileparse.Test(absoluteBin, dest.Arch, dest.Os, tp.Settings.IsVerbose())
				if err != nil {
					log.Printf("Error: %v", err)
					log.Printf("Something fishy is going on: have you run `goxc -t` for this platform (%s,%s)???", dest.Arch, dest.Os)
					errchan <- err
					return
				}
			}
		}
	}
	errchan <- nil
}

func validateToolchain(dest platforms.Platform, goroot string, verbose bool) error {
	err := validatePlatToolchainBinExists(dest, goroot)
	if err != nil {
		return err
	}
	err = validatePlatToolchainPackageVersion(dest, goroot, verbose)
	if err != nil {
		return err
	}

	return nil
}

func validatePlatToolchainPackageVersion(dest platforms.Platform, goroot string, verbose bool) error {
	platPkgFileRuntime := filepath.Join(goroot, "pkg", dest.Os+"_"+dest.Arch, "runtime.a")
	nr, err := os.Open(platPkgFileRuntime)
	if err != nil {
		log.Printf("Could not validate toolchain version: %v", err)
	}
	tr, err := ar.NewReader(nr)
	if err != nil {
		log.Printf("Could not validate toolchain version: %v", err)
	}
	for {
		h, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				log.Printf("Could not validate toolchain version: %v", err)
				return nil
			}
			log.Printf("Could not validate toolchain version: %v", err)
			return err
		}
		//log.Printf("Header: %+v", h)
		if h.Name == "__.PKGDEF" {
			firstLine, err := tr.NextString(50)
			if err != nil {
				log.Printf("failed to read first line of PKGDEF: %v", err)
				return nil
			}
			//log.Printf("pkgdef first part: '%s'", firstLine)
			expectedPrefix := "go object " + dest.Os + " " + dest.Arch + " "
			if !strings.HasPrefix(firstLine, expectedPrefix) {
				log.Printf("first line of __.PKGDEF does not match expected pattern: %v", expectedPrefix)
				return nil
			}
			parts := strings.Split(firstLine, " ")
			compiledVersion := parts[4]
			//runtimeVersion := runtime.Version()
			//log.Printf("Runtime version: %s", runtimeVersion)
			cmdPath := filepath.Join(goroot, "bin", "go")
			cmd := exec.Command(cmdPath)
			args := []string{"version"}
			err = executils.PrepareCmd(cmd, ".", args, []string{}, false)
			if err != nil {
				log.Printf("`go version` failed: %v", err)
				return nil
			}
			goVersionOutput, err := cmd.Output()
			if err != nil {
				log.Printf("`go version` failed: %v", err)
				return nil
			}
			//log.Printf("output: %s", string(out))
			goVersionOutputParts := strings.Split(string(goVersionOutput), " ")
			goVersion := goVersionOutputParts[2]
			if compiledVersion != goVersion {
				return errors.New("static library version '" + compiledVersion + "' does NOT match `go version` '" + goVersion + "'!")
			}
			if verbose {
				log.Printf("Toolchain version '%s' verified against 'go %s' for %v", compiledVersion, goVersion, dest)
			}
			return nil
		}
	}
}

func validatePlatToolchainBinExists(dest platforms.Platform, goroot string) error {
	platGoBin := filepath.Join(goroot, "bin", dest.Os+"_"+dest.Arch, "go")
	if dest.Os == runtime.GOOS && dest.Arch == runtime.GOARCH {

		platGoBin = filepath.Join(goroot, "bin", "go")
	}
	if dest.Os == platforms.WINDOWS {
		platGoBin += ".exe"
	}
	_, err := os.Stat(platGoBin)
	return err
}

// xcPlat: Cross compile for a particular platform
// 0.3.0 - breaking change - changed 'call []string' to 'workingDirectory string'.
func xcPlat(dest platforms.Platform, tp TaskParams, exeName string, packagePath string) (string, error) {
	if tp.Settings.IsVerbose() {
		log.Printf("building %s for platform %v.", exeName, dest)
	}
	args := []string{}
	absoluteBin, err := core.GetAbsoluteBin(dest.Os, dest.Arch, tp.Settings.AppName, exeName, tp.WorkingDirectory, tp.Settings.GetFullVersionName(), tp.Settings.OutPath, tp.Settings.ArtifactsDest)
	if err != nil {
		return "", err
	}
	outDir := filepath.Dir(absoluteBin)
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return "", err
	}
	args = append(args, "-o", absoluteBin, ".")
	//log.Printf("building %s", exeName)
	//v0.8.5 no longer using CGO_ENABLED
	envExtra := []string{"GOOS=" + dest.Os, "GOARCH=" + dest.Arch}
	if dest.Os == platforms.LINUX && dest.Arch == platforms.ARM {
		// see http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go
		goarm := tp.Settings.GetTaskSettingString(TASK_XC, "GOARM")
		if goarm != "" {
			envExtra = append(envExtra, "GOARM="+goarm)
		}
	}
	err = executils.InvokeGo(packagePath, "build", args, envExtra, tp.Settings)
	return absoluteBin, err
}
