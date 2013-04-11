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
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/executils"
	"github.com/laher/goxc/platforms"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
)

var toolchainTask = Task{
	"toolchain",
	"Build toolchain. Make sure to run this each time you update go source.",
	runTaskToolchain,
	nil}

//runs automatically
func init() {
	register(toolchainTask)
}

func runTaskToolchain(tp taskParams) error {
	if len(tp.destPlatforms) < 1 {
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
		for _, platformArr := range tp.destPlatforms {
			busy = true
			destOs := platformArr[0]
			destArch := platformArr[1]
			err = buildToolchain(destOs, destArch, tp.settings)
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				busy = false
				success = success + 1
			}
		}
		if success < 1 {
			log.Printf("No successes!")
			return err
		}
	}
	return nil
}

// Build toolchain for a given target platform
func buildToolchain(goos string, arch string, settings config.Settings) error {
	goroot := runtime.GOROOT()
	scriptpath := core.GetMakeScriptPath(goroot)
	cmd := exec.Command(scriptpath)
	cmd.Dir = filepath.Join(goroot, "src")
	cmd.Args = append(cmd.Args, "--no-clean")
	cgoEnabled := executils.CgoEnabled(goos, arch)
	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = append(cmd.Env, "GOOS="+goos, "CGO_ENABLED="+cgoEnabled, "GOARCH="+arch)
	if goos == platforms.LINUX && arch == platforms.ARM {
		// see http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go
		cmd.Env = append(cmd.Env, "GOARM=5")
	}
	if settings.IsVerbose() {
		log.Printf("'make' env: GOOS=%s CGO_ENABLED=%s GOARCH=%s GOROOT=%s", goos, cgoEnabled, arch, goroot)
	}
	log.Printf("Invoking '%v' from %s", executils.PrintableArgs(cmd.Args), cmd.Dir)
	f, err := executils.RedirectIO(cmd)
	if err != nil {
		log.Printf("Error redirecting IO: %s", err)
	}
	if f != nil {
		defer f.Close()
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("Build toolchain: Launch error: %s", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Build Toolchain: wait error: %s", err)
		return err
	}
	if settings.IsVerbose() {
		log.Printf("Complete")
	}
	return err
}
