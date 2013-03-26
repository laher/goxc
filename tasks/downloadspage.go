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
	"fmt"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/core"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var dpTask = Task{
	core.TASK_DOWNLOADS_PAGE,
	"Generate a downloads page. Currently only supports Markdown format.",
	runTaskDownloadsPage,
	map[string]interface{}{"filename": "downloads.md",
		"fileheader": "---\nlayout: default\ntitle: Downloads\n---\n\n"}}

//runs automatically
func init() {
	register(dpTask)
}

func runTaskDownloadsPage(tp taskParams) error {
	filename := tp.settings.GetTaskSetting(core.TASK_DOWNLOADS_PAGE, "filename").(string)
	reportFilename := filepath.Join(tp.outDestRoot, tp.settings.GetFullVersionName(), filename)
	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	f, err := os.OpenFile(reportFilename, flags, 0600)
	if err == nil {
		defer f.Close()
		fileheader := tp.settings.GetTaskSetting(core.TASK_DOWNLOADS_PAGE, "fileheader").(string)
		if fileheader != "" {
			_, err = fmt.Fprintf(f, "%s\n\n", fileheader)
		}
		_, err = fmt.Fprintf(f, "%s downloads (%s)\n-------------\n", tp.appName, tp.settings.GetFullVersionName())
		versionDir := filepath.Join(tp.outDestRoot, tp.settings.GetFullVersionName())
		fileInfos, err := ioutil.ReadDir(versionDir)
		if err == nil {
			if tp.settings.IsVerbose() {
				log.Printf("Read directory %s", versionDir)
			}
			for _, fi := range fileInfos {
				if fi.IsDir() {
					folderName := filepath.Join(versionDir, fi.Name())
					if tp.settings.IsVerbose() {
						log.Printf("Read directory %s", folderName)
					}
					fileInfos2, err := ioutil.ReadDir(folderName)
					if err == nil {
						platform := strings.Replace(fi.Name(), "_", "/", -1)
						fmt.Fprintf(f, "\n * **%s**:", platform)
						for _, fi2 := range fileInfos2 {
							relativeLink := fi.Name() + "/" + fi2.Name()
							text := strings.Replace(fi2.Name(), "_", "\\_", -1)
							if strings.HasSuffix(fi2.Name(), ".zip") {
								text = "zip"
							}
							if fi2.Name() == tp.appName || fi2.Name() == tp.appName+".exe" {
								text = "executable"
							}
							_, err = fmt.Fprintf(f, " [[%s](%s)]", text, relativeLink)
						}
					}
				}
			}
			_, err = fmt.Fprint(f, "\n")
			//readmes etc will come below main artifacts
			for _, fi := range fileInfos {
				if !fi.IsDir() {
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
		log.Printf("Could not generate downloads page '%s': %s", reportFilename, err)
	}
	return err
}
