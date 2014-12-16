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
	"path/filepath"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/archive"
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/platforms"
)

//runs automatically
func init() {
	RegisterParallelizable(ParallelizableTask{
		TASK_ZIP,
		"Create a zip archive. By default, 'zip' format is used for all platforms except Linux",
		setupZip,
		runTaskZip,
		nil,
		map[string]interface{}{"platforms": "!linux", "include-top-level-dir": "!windows"}})
	RegisterParallelizable(ParallelizableTask{
		TASK_TARGZ,
		"Create a compressed archive. Linux-only by default",
		setupTarGz,
		runTaskTarGz,
		nil,
		map[string]interface{}{"platforms": "linux", "include-top-level-dir": "!windows"}})

}

const (
	TASK_ZIP   = "archive-zip"
	TASK_TARGZ = "archive-tar-gz"
)

func setupTarGz(tp TaskParams) ([]platforms.Platform, error) {
	//for previous versions ...
	//osOptions := settings.GetTaskSettingMap(TASK_ARCHIVE, "os")
	if _, keyExists := tp.Settings.TaskSettings[TASK_TARGZ]["os"]; keyExists {
		return []platforms.Platform{}, errors.New("Option 'os' is no longer supported! Please use 'platforms' instead, specified as a 'build contraint'. e.g. 'linux,386'")
	}
	bc := tp.Settings.GetTaskSettingString(TASK_TARGZ, "platforms")
	destPlatforms := platforms.ApplyBuildConstraints(bc, tp.DestPlatforms)
	return destPlatforms, nil
}
func setupZip(tp TaskParams) ([]platforms.Platform, error) {
	//for previous versions ...
	//osOptions := settings.GetTaskSettingMap(TASK_ARCHIVE, "os")
	if _, keyExists := tp.Settings.TaskSettings[TASK_ZIP]["os"]; keyExists {
		return []platforms.Platform{}, errors.New("Option 'os' is no longer supported! Please use 'platforms' instead, specified as a 'build contraint'. e.g. 'linux,386'")
	}
	bc := tp.Settings.GetTaskSettingString(TASK_ZIP, "platforms")
	destPlatforms := platforms.ApplyBuildConstraints(bc, tp.DestPlatforms)
	return destPlatforms, nil
}

func runTaskTarGz(tp TaskParams, dest platforms.Platform, errchan chan error) {
	bcTopLevelDir := tp.Settings.GetTaskSettingString(TASK_TARGZ, "include-top-level-dir")
	destPlatforms := platforms.ApplyBuildConstraints(bcTopLevelDir, []platforms.Platform{dest})
	isIncludeTopLevelDir := platforms.ContainsPlatform(destPlatforms, dest)
	runArchiveTask(tp, dest, errchan, "tar.gz", archive.TarGz, isIncludeTopLevelDir)
}

func runTaskZip(tp TaskParams, dest platforms.Platform, errchan chan error) {
	bcTopLevelDir := tp.Settings.GetTaskSettingString(TASK_ZIP, "include-top-level-dir")
	destPlatforms := platforms.ApplyBuildConstraints(bcTopLevelDir, []platforms.Platform{dest})
	isIncludeTopLevelDir := platforms.ContainsPlatform(destPlatforms, dest)
	runArchiveTask(tp, dest, errchan, "zip", archive.Zip, isIncludeTopLevelDir)
}

func runArchiveTask(tp TaskParams, dest platforms.Platform, errchan chan error, ending string, archiver archive.Archiver, isIncludeTopLevelDir bool) {
	err := archivePlat(dest.Os, dest.Arch, tp.MainDirs, tp.WorkingDirectory, tp.OutDestRoot, tp.Settings, ending, archiver, isIncludeTopLevelDir)
	if err != nil {
		//TODO - 'force' option?
		errchan <- err
		return
	}
	//always notify completion
	errchan <- nil
}

func archivePlat(goos, arch string, mainDirs []string, workingDirectory, outDestRoot string, settings *config.Settings, ending string, archiver archive.Archiver, includeTopLevelDir bool) error {
	resources := core.ParseIncludeResources(workingDirectory, settings.ResourcesInclude, settings.ResourcesExclude, !settings.IsQuiet())
	//log.Printf("Resources: %v", resources)
	exes := []string{}
	for _, mainDir := range mainDirs {
		var exeName string
		if len(mainDirs) == 1 {
			exeName = settings.AppName
		} else {
			exeName = filepath.Base(mainDir)
		}
		binPath, err := core.GetAbsoluteBin(goos, arch, settings.AppName, exeName, workingDirectory, settings.GetFullVersionName(), settings.OutPath, settings.ArtifactsDest)

		if err != nil {
			return err
		}
		exes = append(exes, binPath)
	}
	outDir := filepath.Join(outDestRoot, settings.GetFullVersionName())
	err := os.MkdirAll(outDir, 0777)
	if err != nil {
		return err
	}
	archivePath, err := archive.ArchiveBinariesAndResources(outDir, goos+"_"+arch,
		exes, settings.AppName, resources, *settings, archiver, ending, includeTopLevelDir)
	if err != nil {
		log.Printf("ZIP error: %s", err)
		return err
	} else {
		if !settings.IsQuiet() {
			log.Printf("Artifact(s) archived to %s", archivePath)
		}
	}
	return nil
}
