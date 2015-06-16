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
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/executils"
	"github.com/laher/goxc/platforms"
)

//runs automatically
func init() {
	Register(Task{
		"toolchain",
		"Build toolchain. Make sure to run this each time you update go source.",
		runTaskToolchain,
		map[string]interface{}{"GOARM": "", "extra-env": []string{}, "no-clean": true}})
}

func runTaskToolchain(tp TaskParams) error {
	if len(tp.DestPlatforms) < 1 {
		return errors.New("No valid platforms specified")
	} else {
		log.Printf("Please do NOT try to quit during a build-toolchain. This can leave your Go toolchain in a non-working state.")
		busy := false
		schan := make(chan os.Signal, 1)
		signal.Notify(schan, os.Interrupt)
		g := func() {
			for sig := range schan {
				// sig is a ^C, handle it
				if busy == true {
					log.Printf("WARNING!!! Received SIGINT (%v) during buildToolchain! DO NOT QUIT DURING BUILD TOOLCHAIN! You may need to run $GOROOT/src/make.bash (or .bat)", sig)
				}
			}
		}
		go g()
		success := 0
		var err error
		for _, dest := range tp.DestPlatforms {
			busy = true
			err = buildToolchain(dest.Os, dest.Arch, tp.Settings)
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				busy = false
				success = success + 1
			}
		}
		if success < 1 {
			log.Printf("No successes!")
			log.Printf("Have you installed Go from source?? If not, please see http://golang.org/doc/install/source")
			return err
		}
	}
	return nil
}

// Build toolchain for a given target platform
func buildToolchain(goos string, arch string, settings *config.Settings) error {
	goroot := settings.GoRoot
	scriptpath := core.GetMakeScriptPath(goroot)
	cmd := exec.Command(scriptpath)

	cmd.Dir = filepath.Join(goroot, "src")

	noClean := settings.GetTaskSettingBool(TASK_BUILD_TOOLCHAIN, "no-clean")
	if noClean {
		cmd.Args = append(cmd.Args, "--no-clean")
	}
	//0.8.5: no longer using cgoEnabled
	env := []string{"GOOS=" + goos, "GOARCH=" + arch}
	extraEnv := settings.GetTaskSettingStringSlice(TASK_BUILD_TOOLCHAIN, "extra-env")
	if settings.IsVerbose() {
		log.Printf("extra-env: %v", extraEnv)
	}
	env = append(env, extraEnv...)
	if goos == platforms.LINUX && arch == platforms.ARM {
		// see http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go
		//NOTE: I don't think it has any effect on fp
		goarm := settings.GetTaskSettingString(TASK_BUILD_TOOLCHAIN, "GOARM")
		if goarm != "" {
			env = append(env, "GOARM="+goarm)
		}
	}

	if settings.IsVerbose() {
		log.Printf("Setting env: %v", env)
	}
	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = append(cmd.Env, env...)
	if settings.IsVerbose() {
		log.Printf("'make' env: GOOS=%s GOARCH=%s GOROOT=%s", goos, arch, goroot)
		log.Printf("Invoking '%v' from %s", executils.PrintableArgs(cmd.Args), cmd.Dir)
	}
	executils.RedirectIO(cmd)
	err := executils.StartAndWait(cmd)
	if err != nil {
		log.Printf("Build toolchain: %s", err)
	}
	if settings.IsVerbose() {
		log.Printf("Complete")
	}
	return err
}
