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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

const TASK_BINTRAY = "bintray"

//runs automatically
func init() {
	register(Task{
		TASK_BINTRAY,
		"Upload artifacts to bintray",
		runTaskBintray,
		map[string]interface{}{"subject": "", "apikey": "", "repository": "",
			"apihost": "https://api.bintray.com/"}})
}

func runTaskBintray(tp taskParams) error {
	subject := tp.settings.GetTaskSetting(TASK_BINTRAY, "subject").(string)
	apikey := tp.settings.GetTaskSetting(TASK_BINTRAY, "apikey").(string)
	repo := tp.settings.GetTaskSetting(TASK_BINTRAY, "repository").(string)
	pkg := tp.settings.GetTaskSetting(TASK_BINTRAY, "package").(string)
	apihost := tp.settings.GetTaskSetting(TASK_BINTRAY, "apihost").(string)
	versionDir := filepath.Join(tp.outDestRoot, tp.settings.GetFullVersionName())

	missing := []string{}

	if subject == "" {
		missing = append(missing, "subject")
	}
	if apikey == "" {
		missing = append(missing, "apikey")
	}
	if repo == "" {
		missing = append(missing, "repository")
	}
	if pkg == "" {
		missing = append(missing, "package")
	}
	if apihost == "" {
		missing = append(missing, "apihost")
	}
	if len(missing) > 0 {
		return errors.New(fmt.Sprintf("bintray configuration missing (%v)", missing))
	}
	versions, err := getVersions(apihost, apikey, subject, repo, pkg)
	if err != nil {
		return err
	}
	log.Printf("Existing versions: %v", versions)
	if core.StringSlicePos(versions, tp.settings.GetFullVersionName()) == -1 {
		//add version
		log.Printf("Note: version %s doesnt exist yet", tp.settings.GetFullVersionName())
		//	err = createVersion(apihost, apikey, subject, repo, pkg, tp.settings.GetFullVersionName())
		//	if err != nil {
		//		return err
		//	}
	}

	// /packages/:subject/:repo/:package
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
					fmt.Sprintf("Platform *%s", platform)
					for _, fi2 := range fileInfos2 {
						relativeLink := fi.Name() + "/" + fi2.Name()
						text := strings.Replace(fi2.Name(), "_", "\\_", -1)
						if strings.HasSuffix(fi2.Name(), ".zip") {
							text = "zip"
						} else if strings.HasSuffix(fi2.Name(), ".deb") {
							text = "deb"
						} else if strings.HasSuffix(fi2.Name(), ".tar.gz") {
							text = "tar.gz"
						} else if fi2.Name() == tp.appName || fi2.Name() == tp.appName+".exe" {
							text = "executable"
						}
						log.Printf("Text: %s", text)
						err = uploadFile(apihost, apikey, subject, repo, pkg, tp.settings.GetFullVersionName(), relativeLink)
						if err != nil {
							return err
						}
						//TODO: move out of loop. Only once per batch.
						err = publish(apihost, apikey, subject, repo, pkg, tp.settings.GetFullVersionName())
						if err != nil {
							return err
						}

					}
				}

			}
		}
	}
	return err
}

func publish(apihost, apikey, subject, repository, pkg, version string) error {
	resp, err := doHttp("POST", apihost+"/content/"+subject+"/"+repository+"/"+pkg+"/"+version+"/publish", subject, apikey, map[string]interface{}{})
	if err == nil {
		log.Printf("File uploaded. %v", resp)
	}
	return err
}

//PUT /content/:subject/:repo/:package/:version/:path
func uploadFile(apihost, apikey, subject, repository, pkg, version, path string) error {
	resp, err := doHttp("PUT", apihost+"/content/"+subject+"/"+repository+"/"+pkg+"/"+version+"/"+path, subject, apikey, nil)
	if err == nil {
		log.Printf("File uploaded. %v", resp)
	}
	return err
}

//POST /packages/:subject/:repo/:package/versions
func createVersion(apihost, apikey, subject, repository, pkg, version string) error {
	req := map[string]interface{}{"name": version, "release_notes": "built by goxc", "release_url": "http://x.x.x/x/x"}
	resp, err := doHttp("POST", apihost+"/packages/"+subject+"/"+repository+"/"+pkg+"/versions", subject, apikey, req)
	if err == nil {
		log.Printf("Created new version. %v", resp)
	}
	return err
}

func doHttp(method, url, subject, apikey string, requestData map[string]interface{}) (map[string]interface{}, error) {
	client := &http.Client{}
	var reader io.Reader
	requestLength := 0
	if requestData != nil {
		postData, err := json.Marshal(requestData)
		if err != nil {
			return nil, err
		}
		requestLength = len(postData)
		reader = bytes.NewReader(postData)
	}
	log.Printf("reader: %+v", reader)
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	if requestLength > 0 {
		log.Printf("adding Content-Length: %s", strconv.Itoa(requestLength))
		req.Header.Add("Content-Length", strconv.Itoa(requestLength))
	}
	req.SetBasicAuth(subject, apikey)
	//log.Printf("req: %v", req)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Error code: %s", resp.Status)
		log.Printf("Error body: %s", body)
		return nil, errors.New(fmt.Sprintf("Invalid HTTP status: %s", resp.Status))
	}
	log.Printf("Response Body: %s", body)
	var b map[string]interface{}
	err = json.Unmarshal(body, &b)
	if err != nil {
		return nil, err
	}
	return b, err
}

func getVersions(apihost, apikey, subject, repository, pkg string) ([]string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apihost+"/packages/"+subject+"/"+repository+"/"+pkg, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(subject, apikey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	var b map[string]interface{}
	err = json.Unmarshal(body, &b)
	if err != nil {
		return nil, err
	}
	log.Printf("Body: %s", body)
	if versions, keyExists := b["versions"]; keyExists {
		versionsSlice, err := config.FromJsonStringSlice(versions, "versions")

		return versionsSlice, err
	}
	return nil, errors.New("Versions not listed!")
}
