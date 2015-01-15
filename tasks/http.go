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
	"fmt"
	"net/http"
	"text/template"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/core"
	//"github.com/laher/goxc/typeutils"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// HTTPTask is the task name to be used in the CLI
const HTTPTask = "http"

type httpTaskConfig struct {
	includePatterns string
	excludePatterns string
	username        string
	password        string
	urlTemplate     *template.Template
	client          *http.Client
}

func init() {
	Register(Task{
		HTTPTask,
		"Upload artifacts to an HTTP server using a PUT request. Configuration required, see `goxc -h http` output.",
		httpRunTask,
		map[string]interface{}{
			"url-template": "",
			"username":     "",
			"password":     "",
			"include":      "*.zip,*.tar.gz,*.deb",
			"exclude":      "*.orig.tar.gz,data.tar.gz,control.tar.gz,*.debian.tar.gz,*-dev_*.deb",
		}})
}

func httpRunTask(tp TaskParams) error {
	urlTemplateString := tp.Settings.GetTaskSettingString(HTTPTask, "url-template")
	missing := []string{}
	if urlTemplateString == "" {
		missing = append(missing, "url-template")
	}
	if len(missing) > 0 {
		return fmt.Errorf("HTTP task configuration missing %v", missing)
	}
	template := template.New("url-template")
	if _, err := template.Parse(urlTemplateString); err != nil {
		return err
	}
	config := &httpTaskConfig{
		includePatterns: tp.Settings.GetTaskSettingString(HTTPTask, "include"),
		excludePatterns: tp.Settings.GetTaskSettingString(HTTPTask, "exclude"),
		username:        tp.Settings.GetTaskSettingString(HTTPTask, "username"),
		password:        tp.Settings.GetTaskSettingString(HTTPTask, "password"),
		urlTemplate:     template,
		client:          &http.Client{},
	}
	versionDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName())
	err := filepath.Walk(versionDir, func(path string, info os.FileInfo, e error) error {
		return httpWalk(config, path, info, e, tp)
	})
	if err != nil {
		return err
	}
	return nil
}

func httpWalk(config *httpTaskConfig, fullPath string, fi os.FileInfo, err error, tp TaskParams) error {
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return nil
	}
	versionDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName())
	relativePath := strings.TrimPrefix(strings.Replace(fullPath, versionDir, "", -1), "/")
	if !tp.Settings.IsQuiet() {
		log.Printf("Considering %s", relativePath)
	}
	resourceGlobs := core.ParseCommaGlobs(config.includePatterns)
	excludeGlobs := core.ParseCommaGlobs(config.excludePatterns)
	matches := false
	for _, resourceGlob := range resourceGlobs {
		ok, err := filepath.Match(resourceGlob, fi.Name())
		if err != nil {
			return err
		}
		if ok {
			matches = true
		}
	}
	if matches == false {
		if !tp.Settings.IsQuiet() {
			log.Printf("Not including %s (for include patterns %s)", relativePath, strings.Join(resourceGlobs, ", "))
		}
		return nil
	}
	for _, excludeGlob := range excludeGlobs {
		ok, err := filepath.Match(excludeGlob, fi.Name())
		if err != nil {
			return err
		}
		if ok {
			if !tp.Settings.IsQuiet() {
				log.Printf("Excluding %s (for exclude pattern %s)", relativePath, excludeGlob)
			}
			return nil
		}
	}
	return httpUploadFile(config, fullPath, fi, tp)
}

func httpUploadFile(config *httpTaskConfig, fullPath string, fi os.FileInfo, tp TaskParams) error {
	var url bytes.Buffer
	err := config.urlTemplate.Execute(&url, map[string]interface{}{
		"TaskParams": tp,
		"FileInfo":   fi,
	})
	if err != nil {
		return err
	}
	if !tp.Settings.IsQuiet() {
		log.Printf("Putting %s to %s", fi.Name(), url.String())
	}
	b, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url.String(), b)
	if err != nil {
		return err
	}
	if config.username != "" || config.password != "" {
		req.SetBasicAuth(config.username, config.password)
	}
	res, err := config.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("Unexpected response: %v", res.Status)
	}
	return nil
}
