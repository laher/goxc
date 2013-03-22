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
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var toolchainTask = Task{
	"toolchain",
	"Build toolchain. Make sure to run this each time you update go source.",
	runTaskToolchain}

//runs automatically
func init() {
	register(toolchainTask)
}

func runTaskToolchain(tp taskParams) error {
	for _, platformArr := range tp.destPlatforms {
		destOs := platformArr[0]
		destArch := platformArr[1]
		buildToolchain(destOs, destArch, tp.settings)
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
	cgoEnabled := core.CgoEnabled(goos, arch)

	cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED="+cgoEnabled, "GOARCH="+arch)
	if goos == core.LINUX && arch == core.ARM {
		// see http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go
		cmd.Env = append(cmd.Env, "GOARM=5")
	}
	if settings.IsVerbose() {
		log.Printf("'make' env: GOOS=%s CGO_ENABLED=%s GOARCH=%s GOROOT=%s", goos, cgoEnabled, arch, goroot)
		log.Printf("'make' args: %s", cmd.Args)
		log.Printf("'make' working directory: %s", cmd.Dir)
	}
	f, err := core.RedirectIO(cmd)
	if err != nil {
		log.Printf("Error redirecting IO: %s", err)
	}
	if f != nil {
		defer f.Close()
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("Launch error: %s", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Wait error: %s", err)
		return err
	}
	if settings.IsVerbose() {
		log.Printf("Complete")
	}
	return err
}
