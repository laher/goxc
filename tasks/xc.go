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
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var xcTask = Task{
	"xc",
	"Cross compile. Builds executables for other platforms.",
	runTaskXC,
	nil}

//runs automatically
func init() {
	register(xcTask)
}

func runTaskXC(tp taskParams) error {
	//func runTaskXC(destPlatforms [][]string, workingDirectory string, settings config.Settings) error {
	if len(tp.destPlatforms) == 0 {
		return errors.New("No valid platforms specified")
	}
	for _, platformArr := range tp.destPlatforms {
		destOs := platformArr[0]
		destArch := platformArr[1]
		err := xcPlat(destOs, destArch, tp.workingDirectory, tp.settings)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	return nil
}

// xcPlat: Cross compile for a particular platform
// 'isFirst' is used simply to determine whether to start a new downloads.md page
// 0.3.0 - breaking change - changed 'call []string' to 'workingDirectory string'.
func xcPlat(goos, arch string, workingDirectory string, settings config.Settings) error {
	log.Printf("building for platform %s_%s.", goos, arch)
	relativeDir := filepath.Join(settings.GetFullVersionName(), goos+"_"+arch)

	appName := core.GetAppName(workingDirectory)

	outDestRoot := core.GetOutDestRoot(appName, settings.ArtifactsDest, workingDirectory)
	outDir := filepath.Join(outDestRoot, relativeDir)
	os.MkdirAll(outDir, 0755)

	cmd := exec.Command("go")
	cmd.Args = append(cmd.Args, "build")
	if settings.GetFullVersionName() != "" {
		cmd.Args = append(cmd.Args, "-ldflags", "-X main.VERSION "+settings.GetFullVersionName()+"")
	}
	cmd.Dir = workingDirectory
	relativeBin := core.GetRelativeBin(goos, arch, appName, false, settings.GetFullVersionName())
	cmd.Args = append(cmd.Args, "-o", filepath.Join(outDestRoot, relativeBin), workingDirectory)
	f, err := core.RedirectIO(cmd)
	if err != nil {
		log.Printf("Error redirecting IO: %s", err)
	}
	if f != nil {
		defer f.Close()
	}
	cgoEnabled := core.CgoEnabled(goos, arch)
	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = append(cmd.Env, "GOOS="+goos, "CGO_ENABLED="+cgoEnabled, "GOARCH="+arch)
	if settings.IsVerbose() {
		log.Printf("'go' env: GOOS=%s CGO_ENABLED=%s GOARCH=%s", goos, cgoEnabled, arch)
		log.Printf("'go' env: %s", cmd.Env)
		log.Printf("'go' args: %v", cmd.Args)
		log.Printf("'go' working directory: %s", cmd.Dir)
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("Launch error: %s", err)
	} else {
		err = cmd.Wait()
		if err != nil {
			log.Printf("Compiler error: %s", err)
		} else {
			log.Printf("Artifact generated OK")
		}
	}
	return err
}
