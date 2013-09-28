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
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/exefileparse"
	"github.com/laher/goxc/executils"
	"github.com/laher/goxc/platforms"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

//runs automatically
func init() {
	//GOARM=6 (this is the default for go1.1
	Register(Task{
		"xc",
		"Cross compile. Builds executables for other platforms.",
		runTaskXC,
		map[string]interface{}{"GOARM": ""}})
}

func runTaskXC(tp TaskParams) error {
	if len(tp.DestPlatforms) == 0 {
		return errors.New("No valid platforms specified")
	}
	success := 0
	var err error
	appName := core.GetAppName(tp.WorkingDirectory)
	outDestRoot := core.GetOutDestRoot(appName, tp.Settings.ArtifactsDest, tp.WorkingDirectory)
	log.Printf("mainDirs : %v", tp.MainDirs)
	for _, dest := range tp.DestPlatforms {
		for _, mainDir := range tp.MainDirs {
			exeName := filepath.Base(mainDir)
			absoluteBin, err := xcPlat(dest.Os, dest.Arch, mainDir, tp.Settings, outDestRoot, exeName)
			if err != nil {
				log.Printf("Error: %v", err)
				log.Printf("Have you run `goxc -t` for this platform???")
				return err
			} else {
				success = success + 1
				err = exefileparse.Test(absoluteBin, dest.Arch, dest.Os)
				if err != nil {
					log.Printf("Error: %v", err)
					log.Printf("Something fishy is going on: have you run `goxc -t` for this platform???")
					return err
				}
			}
		}
	}
	//0.6 return error if no platforms succeeded.
	if success < 1 {
		log.Printf("No successes!")
		return err
	}
	return nil
}

func validatePlatToolchain(goos, arch string) error {
	err := validatePlatToolchainBinExists(goos, arch)
	if err != nil {
		return err
	}
	return nil
}

func validatePlatToolchainBinExists(goos, arch string) error {
	goroot := runtime.GOROOT()
	platGoBin := filepath.Join(goroot , "bin", goos+"_"+arch, "go")
	if goos == runtime.GOOS && arch == runtime.GOARCH {

		platGoBin = filepath.Join(goroot , "bin", "go")
	}
	_, err := os.Stat(platGoBin)
	return err
}

// xcPlat: Cross compile for a particular platform
// 0.3.0 - breaking change - changed 'call []string' to 'workingDirectory string'.
func xcPlat(goos, arch string, workingDirectory string, settings config.Settings, outDestRoot string, exeName string) (string, error) {
	err := validatePlatToolchain(goos, arch)
	if err != nil {
		log.Printf("Toolchain not ready. Re-building toolchain. (%v)", err)
		err = buildToolchain(goos, arch, settings)
		if err != nil {
			return "", err
		}
	}
	log.Printf("building %s for platform %s_%s.", exeName, goos, arch)
	relativeDir := filepath.Join(settings.GetFullVersionName(), goos+"_"+arch)

	outDir := filepath.Join(outDestRoot, relativeDir)
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return "", err
	}
	args := []string{"build"}
	relativeBin := core.GetRelativeBin(goos, arch, exeName, false, settings.GetFullVersionName())
	absoluteBin := filepath.Join(outDestRoot, relativeBin)
	args = append(args, executils.GetLdFlagVersionArgs(settings.GetFullVersionName())...)
	args = append(args, "-o", absoluteBin, ".")
	//log.Printf("building %s", exeName)
	//v0.8.5 no longer using CGO_ENABLED
	envExtra := []string{"GOOS=" + goos, "GOARCH=" + arch}
	if goos == platforms.LINUX && arch == platforms.ARM {
		// see http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go
		goarm := settings.GetTaskSettingString(TASK_XC, "GOARM")
		if goarm != "" {
			envExtra = append(envExtra, "GOARM="+goarm)
		}
	}
	err = executils.InvokeGo(workingDirectory, args, envExtra, settings.IsVerbose())
	return absoluteBin, err
}
