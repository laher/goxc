package goxc

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
	"fmt"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/archive"
	"github.com/laher/goxc/config"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func runTaskXC(destPlatforms [][]string, workingDirectory string, settings config.Settings) error {
	for _, platformArr := range destPlatforms {
		destOs := platformArr[0]
		destArch := platformArr[1]
		xcPlat(destOs, destArch, workingDirectory, settings)
	}
	return nil
}

// xcPlat: Cross compile for a particular platform
// 'isFirst' is used simply to determine whether to start a new downloads.md page
// 0.3.0 - breaking change - changed 'call []string' to 'workingDirectory string'.
func xcPlat(goos, arch string, workingDirectory string, settings config.Settings) string {
	log.Printf("building for platform %s_%s.", goos, arch)
	relativeDir := filepath.Join(settings.GetFullVersionName(), goos+"_"+arch)

	appName := getAppName(workingDirectory)

	outDestRoot := getOutDestRoot(appName, settings)
	outDir := filepath.Join(outDestRoot, relativeDir)
	os.MkdirAll(outDir, 0755)

	cmd := exec.Command("go")
	cmd.Args = append(cmd.Args, "build")
	if settings.GetFullVersionName() != "" {
		cmd.Args = append(cmd.Args, "-ldflags", "-X main.VERSION "+settings.GetFullVersionName()+"")
	}
	cmd.Dir = workingDirectory
	//relativeBinForMarkdown := getRelativeBin(goos, arch, appName, true)
	relativeBin := getRelativeBin(goos, arch, appName, false, settings)
	cmd.Args = append(cmd.Args, "-o", filepath.Join(outDestRoot, relativeBin), workingDirectory)
	f, err := redirectIO(cmd)
	if err != nil {
		log.Printf("Error redirecting IO: %s", err)
	}
	if f != nil {
		defer f.Close()
	}
	cgoEnabled := cgoEnabled(goos, arch)
	cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED="+cgoEnabled, "GOARCH="+arch)
	if settings.IsVerbose() {
		log.Printf("'go' env: GOOS=%s CGO_ENABLED=%s GOARCH=%s", goos, cgoEnabled, arch)
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
	return relativeBin
}

func runTaskCodesign(destPlatforms [][]string, outDestRoot string, appName string, settings config.Settings) error {
	for _, platformArr := range destPlatforms {
		destOs := platformArr[0]
		destArch := platformArr[1]
		relativeBin := getRelativeBin(destOs, destArch, appName, false, settings)
		codesignPlat(destOs, destArch, outDestRoot, relativeBin, settings)
	}
	//TODO return error
	return nil
}

func codesignPlat(goos, arch string, outDestRoot string, relativeBin string, settings config.Settings) {
	// settings.codesign only works on OS X for binaries generated for OS X.
	id:= settings.GetTaskSetting("codesign", "id", "")
	if id != "" && runtime.GOOS == DARWIN && goos == DARWIN {
		if err := signBinary(filepath.Join(outDestRoot, relativeBin), settings); err != nil {
			log.Printf("codesign failed: %s", err)
		} else {
			log.Printf("Signed with ID: %q", id)
		}
	}
}

func signBinary(binPath string, settings config.Settings) error {
	cmd := exec.Command("codesign")
	id:= settings.GetTaskSetting("codesign", "id", "")
	cmd.Args = append(cmd.Args, "-s", id.(string), binPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

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
	resources := parseIncludeResources(workingDirectory, settings.Resources.Include, settings)
	//0.4.0 use a new task type instead of artifact type
	if settings.IsTask(config.TASK_ARCHIVE) {
		// Create ZIP archive.
		relativeBin := getRelativeBin(goos, arch, appName, false, settings)
		zipPath, err := archive.ZipBinaryAndResources(
			filepath.Join(outDestRoot, settings.GetFullVersionName(), goos + "_" + arch),
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

func runTaskRmBin(destPlatforms [][]string, appName, outDestRoot string, settings config.Settings) error {
	for _, platformArr := range destPlatforms {
		destOs := platformArr[0]
		destArch := platformArr[1]
		rmBinPlat(destOs, destArch, appName, outDestRoot, settings)
	}
	//TODO return error
	return nil
}

func rmBinPlat(goos, arch, appName, outDestRoot string, settings config.Settings) {
	relativeBin := getRelativeBin(goos, arch, appName, false, settings)
	archive.RemoveArchivedBinary(filepath.Join(outDestRoot, relativeBin))
}

func runTaskDownloadsPage(destPlatforms [][]string, appName, workingDirectory string, outDestRoot string, settings config.Settings) error {
	filename := settings.GetTaskSetting(config.TASK_DOWNLOADS_PAGE, "filename", "downloads.md")
	reportFilename := filepath.Join(outDestRoot, settings.GetFullVersionName(), filename.(string))
	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	f, err := os.OpenFile(reportFilename, flags, 0600)
	if err == nil {
		defer f.Close()
		header := settings.GetTaskSetting(config.TASK_DOWNLOADS_PAGE, "header", "")
		if header != "" {
			_, err = fmt.Fprintf(f, "%s\n\n", header)
		}
		_, err = fmt.Fprintf(f, "%s downloads (%s)\n-------------\n", appName, settings.GetFullVersionName())
		versionDir := filepath.Join(outDestRoot, settings.GetFullVersionName())
		fileInfos, err := ioutil.ReadDir(versionDir)
		if err == nil {
			if settings.IsVerbose() {
				log.Printf("Read directory %s", versionDir)
			}
			for _, fi := range fileInfos {
				if fi.IsDir() {
					folderName := filepath.Join(versionDir, fi.Name())
					if settings.IsVerbose() {
						log.Printf("Read directory %s", folderName)
					}
					fileInfos2, err := ioutil.ReadDir(folderName)
					if err == nil {
						platform := strings.Replace(fi.Name(), "_", "/", -1)
						fmt.Fprintf(f, "\n * %s:", platform)
						for _, fi2 := range fileInfos2 {
							relativeLink := fi.Name() + "/" + fi2.Name()
							text := strings.Replace(fi2.Name(), "_", "\\_", -1)
							_, err = fmt.Fprintf(f, " [[%s](%s)],", text, relativeLink)
						}
					}
				} else {
					if fi.Name() != filename {
						relativeLink := fi.Name()
						_, err = fmt.Fprintf(f, " * [%s](%s)\n", relativeLink, relativeLink)
					}
					//log.Printf("ignore file %s", fi.Name())
				}
			}
		}
	}
	if err != nil {
		log.Printf("Could not report to '%s': %s", reportFilename, err)
	}
	return err
}

func runTaskToolchain(destPlatforms [][]string, settings config.Settings) error {
	for _, platformArr := range destPlatforms {
		destOs := platformArr[0]
		destArch := platformArr[1]
		buildToolchain(destOs, destArch, settings)
	}
	return nil
}

func RunTasks(workingDirectory string, settings config.Settings) {
	if settings.IsVerbose() {
		log.Printf("looping through each platform")
	}
	destOses := strings.Split(settings.Os, ",")
	destArchs := strings.Split(settings.Arch, ",")
	var destPlatforms [][]string
	for _, supportedPlatformArr := range PLATFORMS {
		supportedOs := supportedPlatformArr[0]
		supportedArch := supportedPlatformArr[1]
		for _, destOs := range destOses {
			if destOs == "" || supportedOs == destOs {
				for _, destArch := range destArchs {
					if destArch == "" || supportedArch == destArch {
						destPlatforms = append(destPlatforms, supportedPlatformArr)
					}
				}
			}
		}
	}
	appName := getAppName(workingDirectory)
	outDestRoot := getOutDestRoot(appName, settings)
	for _, task := range settings.Tasks {
		err := runTask(task, destPlatforms, appName, workingDirectory, outDestRoot, settings)
		if err != nil {
			// TODO: implement 'force' option.
			return
		}
	}
}

func runTask(task string, destPlatforms [][]string, appName, workingDirectory, outDestRoot string, settings config.Settings) error {
	// 0.3.1 added clean, vet, test, install etc
	switch task {
	case config.TASK_CLEAN:
		err := invokeGo(workingDirectory, []string{config.TASK_CLEAN}, settings)
		if err != nil {
			log.Printf("Clean failed! %s", err)
		}
		return err
	case config.TASK_VET:
		err := invokeGo(workingDirectory, []string{config.TASK_VET}, settings)
		if err != nil {
			log.Printf("Vet failed! %s", err)
		}
		return err
	case config.TASK_TEST:
		err := invokeGo(workingDirectory, []string{config.TASK_TEST, "./..."}, settings)
		if err != nil {
			log.Printf("Test failed! %s", err)
		}
		return err
	case config.TASK_FMT:
		err := invokeGo(workingDirectory, []string{config.TASK_FMT}, settings)
		if err != nil {
			log.Printf("Fmt failed! %s", err)
		}
		return err
	case config.TASK_INSTALL:
		err := invokeGo(workingDirectory, []string{config.TASK_INSTALL}, settings)
		if err != nil {
			log.Printf("Install failed! %s", err)
		}
		return err
	case config.TASK_CODESIGN:
		return runTaskCodesign(destPlatforms, appName, outDestRoot, settings)
	case config.TASK_BUILD_TOOLCHAIN:
		return runTaskToolchain(destPlatforms, settings)
	case config.TASK_XC:
		return runTaskXC(destPlatforms, workingDirectory, settings)
	case config.TASK_ARCHIVE:
		return runTaskZip(destPlatforms, appName, workingDirectory, outDestRoot, settings)
	case config.TASK_REMOVE_BIN:
		return runTaskRmBin(destPlatforms, appName, outDestRoot, settings)
	case config.TASK_DOWNLOADS_PAGE:
		return runTaskDownloadsPage(destPlatforms, appName, workingDirectory, outDestRoot, settings)
	}
	// TODO: custom tasks
	log.Printf("Unrecognised task '%s'", task)
	return fmt.Errorf("Unrecognised task '%s'", task)
}
