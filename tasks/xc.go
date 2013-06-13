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
	"github.com/laher/goxc/executils"
	"log"
	"os"
	"path/filepath"
)

//runs automatically
func init() {
	Register(Task{
		"xc",
		"Cross compile. Builds executables for other platforms.",
		runTaskXC,
		nil})
}

func runTaskXC(tp TaskParams) error {
	if len(tp.DestPlatforms) == 0 {
		return errors.New("No valid platforms specified")
	}
	success := 0
	var err error
	for _, dest := range tp.DestPlatforms {
		err = xcPlat(dest.Os, dest.Arch, tp.WorkingDirectory, tp.Settings)
		if err != nil {
			log.Printf("Error: %v", err)
		} else {
			success = success + 1
		}
	}
	//0.6 return error if no platforms succeeded.
	if success < 1 {
		log.Printf("No successes!")
		return err
	}
	return nil
}

// xcPlat: Cross compile for a particular platform
// 0.3.0 - breaking change - changed 'call []string' to 'workingDirectory string'.
func xcPlat(goos, arch string, workingDirectory string, settings config.Settings) error {
	log.Printf("building for platform %s_%s.", goos, arch)
	relativeDir := filepath.Join(settings.GetFullVersionName(), goos+"_"+arch)

	appName := core.GetAppName(workingDirectory)

	outDestRoot := core.GetOutDestRoot(appName, settings.ArtifactsDest, workingDirectory)
	outDir := filepath.Join(outDestRoot, relativeDir)
	os.MkdirAll(outDir, 0755)

	args := []string{"build"}
	relativeBin := core.GetRelativeBin(goos, arch, appName, false, settings.GetFullVersionName())
	args = append(args, executils.GetLdFlagVersionArgs(settings.GetFullVersionName())...)
	args = append(args, "-o", filepath.Join(outDestRoot, relativeBin), ".")
	//TODO: use runtime.Version() to detect whether this is needed (unnecessary from 1.1+)
	cgoEnabled := executils.CgoEnabled(goos, arch)
	envExtra := []string{"GOOS=" + goos, "CGO_ENABLED=" + cgoEnabled, "GOARCH=" + arch}
	err := executils.InvokeGo(workingDirectory, args, envExtra, settings.IsVerbose(), settings.PrependCurrentEnv)
	return err
}
