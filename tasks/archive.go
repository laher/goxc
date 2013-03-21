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
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/archive"
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"log"
	"path/filepath"
)

func runTaskZip(destPlatforms [][]string, appName, workingDirectory, outDestRoot string, settings config.Settings) error {
	for _, platformArr := range destPlatforms {
		destOs := platformArr[0]
		destArch := platformArr[1]
		err := zipPlat(destOs, destArch, appName, workingDirectory, outDestRoot, settings)
		if err != nil {
			//TODO - 'force' option
			//return err
		}
	}
	//TODO return error?
	return nil
}

func zipPlat(goos, arch, appName, workingDirectory, outDestRoot string, settings config.Settings) error {
	resources := core.ParseIncludeResources(workingDirectory, settings.Resources.Include, settings.IsVerbose())
	//0.4.0 use a new task type instead of artifact type
	if settings.IsTask(config.TASK_ARCHIVE) {
		// Create ZIP archive.
		relativeBin := core.GetRelativeBin(goos, arch, appName, false, settings.GetFullVersionName())
		zipPath, err := archive.ZipBinaryAndResources(
			filepath.Join(outDestRoot, settings.GetFullVersionName(), goos+"_"+arch),
			filepath.Join(outDestRoot, relativeBin), appName, resources, settings)
		if err != nil {
			log.Printf("ZIP error: %s", err)
			return err
		} else {
			log.Printf("Artifact %s zipped to %s", relativeBin, zipPath)
		}
	}
	return nil
}


