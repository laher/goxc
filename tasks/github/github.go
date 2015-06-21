package github

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
//TODO: handle conflicts (delete or skip?)
//TODO: own options for downloadspage
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/tasks"
	"github.com/laher/goxc/tasks/httpc"
)

func RunTaskPubGH(tp tasks.TaskParams) error {
	owner := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "owner")
	apikey := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "apikey")
	repository := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "repository")
	apiHost := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "apihost")
	//downloadsHost := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "downloadshost")
	versionDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName())

	missing := []string{}

	if owner == "" {
		missing = append(missing, "owner")
	}
	if apikey == "" {
		missing = append(missing, "apikey")
	}
	if repository == "" {
		missing = append(missing, "repository")
	}
	if apiHost == "" {
		missing = append(missing, "apihost")
	}
	if len(missing) > 0 {
		return errors.New(fmt.Sprintf("github configuration missing (%v)", missing))
	}
	outFilename := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "downloadspage")
	templateText := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "templateText")
	templateFile := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "templateFile")
	format := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "outputFormat")
	body := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "body")
	preRelease := tp.Settings.GetTaskSettingBool(tasks.TASK_PUBLISH_GITHUB, "prerelease")
	if format == "by-file-extension" {
		if strings.HasSuffix(outFilename, ".md") || strings.HasSuffix(outFilename, ".markdown") {
			format = "markdown"
		} else if strings.HasSuffix(outFilename, ".html") || strings.HasSuffix(outFilename, ".htm") {
			format = "html"
		} else {
			//unknown ...
			format = ""
		}
	}
	templateVars := tp.Settings.GetTaskSettingMap(tasks.TASK_DOWNLOADS_PAGE, "templateExtraVars")
	reportFilename := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName(), outFilename)
	_, err := os.Stat(filepath.Dir(reportFilename))
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("No artifacts built for this version yet. Please build some artifacts before running the 'publish-github' task")
		} else {
			return err
		}
	}
	prefix := tp.Settings.GetTaskSettingString(tasks.TASK_TAG, "prefix")
	tagName := prefix + tp.Settings.GetFullVersionName()
	err = createRelease(apiHost, owner, apikey, repository, tagName, tp.Settings.GetFullVersionName(), body, preRelease, tp.Settings.IsVerbose())
	if err != nil {
		if serr, ok := err.(httpc.HttpError); ok {
			if serr.StatusCode == 422 {
				//existing release. ignore.
				if !tp.Settings.IsQuiet() {
					log.Printf("Note: release already exists. %v", serr)
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}
	report := tasks.BtReport{
		AppName:    tp.AppName,
		Version:    tp.Settings.GetFullVersionName(),
		Categories: map[string]*[]tasks.BtDownload{},
		ExtraVars:  templateVars}
	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	out, err := os.OpenFile(reportFilename, flags, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	//for 'first entry in dir' detection.
	dirs := []string{}
	err = filepath.Walk(versionDir, func(path string, info os.FileInfo, e error) error {
		return ghWalkFunc(path, info, e, outFilename, dirs, tp, format, report)
	})
	if err != nil {
		return err
	}
	err = tasks.RunTemplate(reportFilename, templateFile, templateText, out, report, format)
	if err != nil {
		return err
	}
	//close explicitly for return value
	return out.Close()
}

func ghWalkFunc(fullPath string, fi2 os.FileInfo, err error, reportFilename string, dirs []string, tp tasks.TaskParams, format string, report tasks.BtReport) error {
	excludeResources := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "exclude")
	excludeGlobs := core.ParseCommaGlobs(excludeResources)
	versionDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName())
	relativePath := strings.Replace(fullPath, versionDir, "", -1)
	relativePath = strings.TrimPrefix(relativePath, "/")
	//fmt.Printf("relative path %s, full path %s\n", relativePath, fullPath)
	if fi2.IsDir() {
		//check globs ...
		for _, excludeGlob := range excludeGlobs {
			ok, err := filepath.Match(excludeGlob, fi2.Name())
			if err != nil {
				return err
			}
			if ok {
				if tp.Settings.IsVerbose() {
					log.Printf("Excluded: %s (pattern %v)", relativePath, excludeGlob)
				}
				return filepath.SkipDir
			}
		}
		return nil
	}
	if fi2.Name() == reportFilename {
		return nil
	}
	owner := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "owner")
	user := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "user")
	if user == "" {
		user = owner
	}
	apikey := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "apikey")
	repository := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "repository")
	apiHost := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "apihost")
	uploadApiHost := "https://uploads.github.com"
	downloadsHost := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "downloadshost")
	includeResources := tp.Settings.GetTaskSettingString(tasks.TASK_PUBLISH_GITHUB, "include")

	resourceGlobs := core.ParseCommaGlobs(includeResources)
	//log.Printf("IncludeGlobs: %v", resourceGlobs)
	//log.Printf("ExcludeGlobs: %v", excludeGlobs)
	matches := false
	for _, resourceGlob := range resourceGlobs {
		ok, err := filepath.Match(resourceGlob, fi2.Name())
		if err != nil {
			return err
		}
		if ok {
			matches = true
		}
	}
	if matches == false {
		if tp.Settings.IsVerbose() {
			log.Printf("Not included: %s (pattern %v)", relativePath, includeResources)
		}
		return nil
	}
	for _, excludeGlob := range excludeGlobs {
		ok, err := filepath.Match(excludeGlob, fi2.Name())
		if err != nil {
			return err
		}
		if ok {
			if tp.Settings.IsVerbose() {
				log.Printf("Excluded: %s (pattern %v)", relativePath, excludeGlob)
			}
			return nil
		}
	}
	first := true

	parent := filepath.Dir(relativePath)
	//platform := strings.Replace(parent, "_", "/", -1)
	//fmt.Fprintf(f, "\n * **%s**:", platform)
	for _, d := range dirs {
		if d == parent {
			first = false
		}
	}
	if first {
		dirs = append(dirs, parent)
	}
	//fmt.Printf("relative path %s, platform %s\n", relativePath, parent)
	text := fi2.Name()
	version := tp.Settings.GetFullVersionName()
	isVerbose := tp.Settings.IsVerbose()
	isQuiet := tp.Settings.IsQuiet()
	contentType := httpc.GetContentType(text)

	prefix := tp.Settings.GetTaskSettingString(tasks.TASK_TAG, "prefix")
	tagName := prefix + version
	release, err := ghGetReleaseForTag(apiHost, owner, apikey, repository, tagName, isVerbose)
	if err != nil {
		return err
	}
	err = ghDoUpload(uploadApiHost, apikey, owner, repository, release, relativePath, fullPath, contentType, isVerbose, isQuiet)
	if err != nil {
		return err
	}
	if first {
		first = false
	} else {
		//commaIfRequired = ","
	}
	if format == "markdown" {
		text = strings.Replace(text, "_", "\\_", -1)
	}
	category := tasks.GetCategory(relativePath)
	downloadsUrl := downloadsHost + "/" + owner + "/" + repository + "/releases/download/" + version + "/" + relativePath + ""
	download := tasks.BtDownload{Text: text, RelativeLink: downloadsUrl}
	v, ok := report.Categories[category]
	var existing []tasks.BtDownload
	if !ok {
		existing = []tasks.BtDownload{}
	} else {
		existing = *v
	}
	existing = append(existing, download)
	report.Categories[category] = &existing

	return err
}

func ghGetReleaseForTag(apihost, owner, apikey, repo, tagName string, isVerbose bool) (string, error) {
	r, err := httpc.DoHttp("GET", apihost+"/repos/"+owner+"/"+repo+"/releases/tags/"+tagName, "", owner, apikey, "", nil, 0, isVerbose)
	if err != nil {
		return "", err
	}
	i, err := httpc.ParseMap(r, isVerbose)
	if err != nil {
		return "", err
	}
	var id string
	idI, ok := i["id"]
	if !ok {
		return "", fmt.Errorf("Id not provided")
	}
	switch i := idI.(type) {
	case float64:
		id = fmt.Sprintf("%0.f", i)
	default:
		return "", fmt.Errorf("ID not a float")
	}
	return id, err
}

//POST https://<upload_url>/repos/:owner/:repo/releases/:id/assets?name=foo.zip
func ghDoUpload(apiHost, apikey, owner, repository, release, relativePath, fullPath, contentType string, isVerbose, isQuiet bool) error {
	//POST /repos/:owner/:repo/releases/:id/assets?name=foo.zip
	url := apiHost + "/repos/" + owner + "/" + repository + "/releases/" + release + "/assets?name=" + relativePath
	if !isQuiet {
		log.Printf("Uploading to %v", url)
	}
	resp, err := httpc.UploadFile("POST", url, repository, owner, apikey, fullPath, relativePath, contentType, isVerbose)
	if err != nil {
		if serr, ok := err.(httpc.HttpError); ok {
			if serr.StatusCode == 409 || serr.StatusCode == 422 {
				//conflict. skip
				//continue but dont publish.
				//TODO - provide an option to replace existing artifact
				//TODO - ?check exists before attempting upload?
				log.Printf("WARNING - file already exists. Skipping. %v", resp)
				return nil
			} else {
				return err
			}
		} else {
			return err
		}
	}
	if isVerbose {
		log.Printf("File uploaded. Response: %v", resp)
	} else if !isQuiet {
		log.Printf("File uploaded.")
	}
	return err
}

//POST /repos/:owner/:repo/releases
func createRelease(apihost, owner, apikey, repo, tagName, version, body string, preRelease, isVerbose bool) error {
	req := map[string]interface{}{"tag_name": tagName, "name": version, "body": body, "prerelease": preRelease}
	requestData, err := json.Marshal(req)
	if err != nil {
		return err
	}
	requestLength := len(requestData)
	reader := bytes.NewReader(requestData)
	resp, err := httpc.DoHttp("POST", apihost+"/repos/"+owner+"/"+repo+"/releases", owner, owner, apikey, "", reader, int64(requestLength), isVerbose)
	if err == nil {
		if isVerbose {
			log.Printf("Created new version. %v", resp)
			i, err := httpc.ParseMap(resp, isVerbose)
			if err != nil {
				log.Printf("Error parsing response as map: %v", err)
				return err
			}
			log.Printf("Created new version: %+v", i)

		}
	}
	return err
}
