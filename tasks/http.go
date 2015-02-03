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

// TASK_PUBLISH_HTTP is the task name to be used in the CLI
const TASK_PUBLISH_HTTP = "publish-http"

type httpTaskConfig struct {
	includePatterns string
	excludePatterns string
	username        string
	password        string
	exists          string
	urlTemplate     *template.Template
	client          *http.Client
}

func init() {
	Register(Task{TASK_PUBLISH_HTTP,
		"Upload artifacts to an HTTP server using a PUT request. Configuration required, see `goxc -h http` output.",
		httpRunTask,
		map[string]interface{}{
			"url-template":  "",
			"username":      "",
			"password":      "",
			"include":       "*.zip,*.tar.gz,*.deb",
			"exclude":       "*.orig.tar.gz,data.tar.gz,control.tar.gz,*.debian.tar.gz,*-dev_*.deb",
			"exists-action": "fail",
		}})
}

func httpRunTask(tp TaskParams) error {
	urlTemplateString := tp.Settings.GetTaskSettingString(TASK_PUBLISH_HTTP, "url-template")
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
		exists:          strings.ToLower(tp.Settings.GetTaskSettingString(TASK_PUBLISH_HTTP, "exists-action")),
		includePatterns: tp.Settings.GetTaskSettingString(TASK_PUBLISH_HTTP, "include"),
		excludePatterns: tp.Settings.GetTaskSettingString(TASK_PUBLISH_HTTP, "exclude"),
		username:        tp.Settings.GetTaskSettingString(TASK_PUBLISH_HTTP, "username"),
		password:        tp.Settings.GetTaskSettingString(TASK_PUBLISH_HTTP, "password"),
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

func httpURLTemplateContext(tp TaskParams, fi os.FileInfo) map[string]interface{} {
	return map[string]interface{}{
		"AppName":        tp.Settings.AppName,
		"Version":        tp.Settings.GetFullVersionName(),
		"Arch":           tp.Settings.Arch,
		"Os":             tp.Settings.Os,
		"PackageVersion": tp.Settings.PackageVersion,
		"BranchName":     tp.Settings.BranchName,
		"PrereleaseInfo": tp.Settings.PrereleaseInfo,
		"BuildName":      tp.Settings.BuildName,
		"FormatVersion":  tp.Settings.FormatVersion,
		"FileName":       fi.Name(),
		"FileSize":       fi.Size(),
		"FileModTime":    fi.ModTime(),
		"FileMode":       fi.Mode(),
	}
}

func httpExistsFile(config *httpTaskConfig, url string) (bool, error) {
	res, err := config.client.Head(url)
	if err != nil {
		return false, err
	}
	switch res.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, fmt.Errorf("Unexpected status code %v when testing for existence of %v", res.StatusCode, url)
	}
}

func httpDeleteFile(config *httpTaskConfig, url string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	res, err := config.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("Unable to delete %v: %v", url, res.Status)
	}
	return nil
}

func httpUploadFile(config *httpTaskConfig, fullPath string, fi os.FileInfo, tp TaskParams) error {
	var urlb bytes.Buffer
	err := config.urlTemplate.Execute(&urlb, httpURLTemplateContext(tp, fi))
	if err != nil {
		return err
	}
	var url = urlb.String()
	exists, err := httpExistsFile(config, url)
	if err != nil {
		return err
	}
	if exists {
		switch config.exists {
		case "replace":
			if !tp.Settings.IsQuiet() {
				log.Printf("Deleting existent file %v at %v", fi.Name(), url)
			}
			if err := httpDeleteFile(config, url); err != nil {
				return err
			}
		case "omit":
			if !tp.Settings.IsQuiet() {
				log.Printf("Omitting existent file %v at %v", fi.Name(), url)
			}
			return nil
		case "fail":
			return fmt.Errorf("Failing http publish of %v because it already exists at %v", fi.Name(), url)
		}
	}
	if !tp.Settings.IsQuiet() {
		log.Printf("Putting %s to %s", fi.Name(), url)
	}
	b, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url, b)
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
