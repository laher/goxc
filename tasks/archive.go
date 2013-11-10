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
	"github.com/laher/goxc/archive"
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/platforms"
	"log"
	"os"
	"path/filepath"
)

//runs automatically
func init() {
	Register(Task{
		"archive-zip",
		"Create a zip archive. By default, 'zip' format is used for all platforms except Linux",
		runTaskArchiveZip,
		map[string]interface{}{"platforms": "!linux", "include-top-level-dir": "!windows"}})
	Register(Task{
		"archive-tar-gz",
		"Create a compressed archive. Linux-only by default",
		runTaskArchiveTarGz,
		map[string]interface{}{"platforms": "linux", "include-top-level-dir": "!windows"}})

}

func runTaskArchiveZip(tp TaskParams) error {
	return runTaskArchive(tp, "archive-zip")
}

func runTaskArchiveTarGz(tp TaskParams) error {
	return runTaskArchive(tp, "archive-tar-gz")
}

func runTaskArchive(tp TaskParams, taskName string) error {
	//for previous versions (until 2.0.3) ...
	//osOptions := settings.GetTaskSettingMap(TASK_ARCHIVE, "os")
	if _, keyExists := tp.Settings.TaskSettings["os"]; keyExists {
		return errors.New("Option 'os' no longer supported! Please use 'tar-gz' instead, specified as a 'build contraint'. e.g. 'linux,386'")
	}
	//for Windows, include topleveldir
	bc := tp.Settings.GetTaskSettingString(taskName, "platforms")
	destPlatforms := platforms.ApplyBuildConstraints(bc, tp.DestPlatforms)
	bcTopLevelDir := tp.Settings.GetTaskSettingString(taskName, "include-top-level-dir")
	destPlatformsTopLevelDir := platforms.ApplyBuildConstraints(bcTopLevelDir, tp.DestPlatforms)
	var ending string
	var archiver archive.Archiver
	switch taskName {
	case "archive-tar-gz":
		ending = "tar.gz"
		archiver = archive.TarGz
	case "archive-zip":
		ending = "zip"
		archiver = archive.Zip
	default:
		return errors.New("Unrecognised task name!")
	}
	for _, dest := range destPlatforms {
		isIncludeTopLevelDir := platforms.ContainsPlatform(destPlatformsTopLevelDir, dest)
		err := archivePlat(dest.Os, dest.Arch, tp.MainDirs, tp.AppName, tp.WorkingDirectory, tp.OutDestRoot, tp.Settings, ending, archiver, isIncludeTopLevelDir)
		if err != nil {
			//TODO - 'force' option?
			return err
		}
	}
	//TODO return error?
	return nil
}

func archivePlat(goos, arch string, mainDirs []string, appName, workingDirectory, outDestRoot string, settings config.Settings, ending string, archiver archive.Archiver, includeTopLevelDir bool) error {
	resources := core.ParseIncludeResources(workingDirectory, settings.ResourcesInclude, settings.ResourcesExclude, settings.IsVerbose())
	exes := []string{}
	for _, mainDir := range mainDirs {
		exeName := filepath.Base(mainDir)
		relativeBin := core.GetRelativeBin(goos, arch, exeName, false, settings.GetFullVersionName())
		exes = append(exes, filepath.Join(outDestRoot, relativeBin))
	}
	outDir := filepath.Join(outDestRoot, settings.GetFullVersionName())
	err := os.MkdirAll(outDir, 0777)
	if err != nil {
		return err
	}
	archivePath, err := archive.ArchiveBinariesAndResources(outDir, goos+"_"+arch,
		exes, appName, resources, settings, archiver, ending, includeTopLevelDir)
	if err != nil {
		log.Printf("ZIP error: %s", err)
		return err
	} else {
		log.Printf("Artifact(s) archived to %s", archivePath)
	}
	return nil
}
