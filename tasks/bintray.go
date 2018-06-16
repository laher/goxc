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
//TODO: handle conflicts (delete or skip?)
//TODO: own options for downloadspage
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/tasks/httpc"
	"github.com/laher/goxc/typeutils"
)

const TASK_BINTRAY = "bintray"

//runs automatically
func init() {
	Register(Task{
		TASK_BINTRAY,
		"Upload artifacts to bintray.com, and generate a local markdown page of links (bintray registration details required in goxc config. See `goxc -h bintray`)",
		runTaskBintray,
		map[string]interface{}{"subject": "", "apikey": "", "repository": "",
			"apihost":       "https://api.bintray.com/",
			"downloadshost": "https://dl.bintray.com/",
			"downloadspage": "bintray.md",
			"fileheader":    "---\nlayout: default\ntitle: Downloads\n---\nFiles hosted at [bintray.com](https://bintray.com)\n\n",
			"include":       "*.zip,*.tar.gz,*.deb",
			"exclude":       "bintray.md",
			"outputFormat":  "by-file-extension", // use by-file-extension, markdown or html
			"templateText": `---
layout: default
title: Downloads
---
Files hosted at [bintray.com](https://bintray.com)

{{.AppName}} downloads (version {{.Version}})

{{range $k, $v := .Categories}}### {{$k}}

{{range $v}} * [{{.Text}}]({{.RelativeLink}})
{{end}}
{{end}}

{{.ExtraVars.footer}}`,
			"templateFile":      "", //use if populated
			"templateExtraVars": map[string]interface{}{"footer": "Generated by goxc"}}})
}

type BtDownload struct {
	Text         string
	RelativeLink string
}
type BtReport struct {
	AppName    string
	Version    string
	Categories map[string]*[]BtDownload
	ExtraVars  map[string]interface{}
}

func runTaskBintray(tp TaskParams) error {
	subject := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "subject")
	apikey := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "apikey")
	repository := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "repository")
	pkg := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "package")
	apiHost := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "apihost")
	//downloadsHost := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "downloadshost")
	versionDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName())

	missing := []string{}

	if subject == "" {
		missing = append(missing, "subject")
	}
	if apikey == "" {
		missing = append(missing, "apikey")
	}
	if repository == "" {
		missing = append(missing, "repository")
	}
	if pkg == "" {
		missing = append(missing, "package")
	}
	if apiHost == "" {
		missing = append(missing, "apihost")
	}
	if len(missing) > 0 {
		return errors.New(fmt.Sprintf("bintray configuration missing (%v)", missing))
	}
	outFilename := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "downloadspage")
	templateText := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "templateText")
	templateFile := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "templateFile")
	format := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "outputFormat")
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
	templateVars := tp.Settings.GetTaskSettingMap(TASK_DOWNLOADS_PAGE, "templateExtraVars")
	reportFilename := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName(), outFilename)
	_, err := os.Stat(filepath.Dir(reportFilename))
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("No artifacts built for this version yet. Please build some artifacts before running the 'bintray' task")
		} else {
			return err
		}
	}
	report := BtReport{tp.AppName, tp.Settings.GetFullVersionName(), map[string]*[]BtDownload{}, templateVars}
	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	out, err := os.OpenFile(reportFilename, flags, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	//	fileheader := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "fileheader")
	//if fileheader != "" {
	//_, err = fmt.Fprintf(f, "%s\n\n", fileheader)
	//}
	//	_, err = fmt.Fprintf(f, "%s downloads (version %s)\n-------------\n", tp.AppName, tp.Settings.GetFullVersionName())
	//	if !tp.Settings.IsQuiet() {
	//		log.Printf("Read directory %s", versionDir)
	//	}
	//for 'first entry in dir' detection.
	dirs := []string{}
	err = filepath.Walk(versionDir, func(path string, info os.FileInfo, e error) error {
		return walkFunc(path, info, e, outFilename, dirs, tp, format, report)
	})
	if err != nil {
		return err
	}
	err = RunTemplate(reportFilename, templateFile, templateText, out, report, format)
	if err != nil {
		return err
	}
	//close explicitly for return value
	return out.Close()
}

func walkFunc(fullPath string, fi2 os.FileInfo, err error, reportFilename string, dirs []string, tp TaskParams, format string, report BtReport) error {
	if fi2.IsDir() || fi2.Name() == reportFilename {

		return nil
	}
	subject := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "subject")
	user := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "user")
	if user == "" {
		user = subject
	}
	apikey := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "apikey")
	repository := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "repository")
	pkg := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "package")
	apiHost := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "apihost")
	downloadsHost := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "downloadshost")
	includeResources := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "include")
	excludeResources := tp.Settings.GetTaskSettingString(TASK_BINTRAY, "exclude")
	versionDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName())

	relativePath := strings.Replace(fullPath, versionDir, "", -1)
	relativePath = strings.Replace(relativePath, "\\", "/", -1)
	relativePath = strings.TrimPrefix(relativePath, "/")
	fmt.Printf("relative path %s, full path %s\n", relativePath, fullPath)

	resourceGlobs := core.ParseCommaGlobs(includeResources)
	//log.Printf("IncludeGlobs: %v", resourceGlobs)
	excludeGlobs := core.ParseCommaGlobs(excludeResources)
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
		if !tp.Settings.IsQuiet() {
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
			if !tp.Settings.IsQuiet() {
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
	/*
		text := strings.Replace(fi2.Name(), "_", "\\_", -1)
		if strings.HasSuffix(fi2.Name(), ".zip") {
			text = "zip"
		} else if strings.HasSuffix(fi2.Name(), ".deb") {
			text = "deb"
		} else if strings.HasSuffix(fi2.Name(), ".tar.gz") {
			text = "tar.gz"
		} else if fi2.Name() == tp.AppName || fi2.Name() == tp.AppName+".exe" {
			text = "executable"
		}
	*/
	//PUT /content/:subject/:repo/:package/:version/:path
	url := apiHost + "/content/" + subject + "/" + repository + "/" + pkg + "/" + tp.Settings.GetFullVersionName() + "/" + relativePath
	// for some reason there's no /pkg/ level in the downloads url.
	downloadsUrl := downloadsHost + "/content/" + subject + "/" + repository + "/" + relativePath + "?direct"
	contentType := httpc.GetContentType(text)
	resp, err := httpc.UploadFile("PUT", url, subject, user, apikey, fullPath, relativePath, contentType, !tp.Settings.IsQuiet())
	if err != nil {
		if serr, ok := err.(httpc.HttpError); ok {
			if serr.StatusCode == 409 {
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
	if !tp.Settings.IsQuiet() {
		log.Printf("File uploaded. (expected empty map[]): %v", resp)
	}
	//commaIfRequired := ""
	if first {
		first = false
	} else {
		//commaIfRequired = ","
	}
	if format == "markdown" {
		text = strings.Replace(text, "_", "\\_", -1)
	}
	category := GetCategory(relativePath)
	download := BtDownload{text, downloadsUrl}
	v, ok := report.Categories[category]
	var existing []BtDownload
	if !ok {
		existing = []BtDownload{}
	} else {
		existing = *v
	}
	existing = append(existing, download)
	report.Categories[category] = &existing

	//_, err = fmt.Fprintf(f, "%s [[%s](%s)]", commaIfRequired, text, downloadsUrl)
	if err != nil {
		return err
	}
	err = publish(apiHost, user, apikey, subject, repository, pkg, tp.Settings.GetFullVersionName(), !tp.Settings.IsQuiet())
	return err
}

func publish(apihost, user, apikey, subject, repository, pkg, version string, isVerbose bool) error {
	resp, err := httpc.DoHttp("POST", apihost+"/content/"+subject+"/"+repository+"/"+pkg+"/"+version+"/publish", subject, user, apikey, "", nil, 0, isVerbose)
	if err == nil {
		log.Printf("Version published. %v", resp)
	}
	return err
}

//NOTE: not necessary.
//POST /packages/:subject/:repo/:package/versions
func createVersion(apihost, user, apikey, subject, repository, pkg, version string, isVerbose bool) error {
	req := map[string]interface{}{"name": version, "release_notes": "built by goxc", "release_url": "http://x.x.x/x/x"}
	requestData, err := json.Marshal(req)
	if err != nil {
		return err
	}
	requestLength := len(requestData)
	reader := bytes.NewReader(requestData)
	resp, err := httpc.DoHttp("POST", apihost+"/packages/"+subject+"/"+repository+"/"+pkg+"/versions", subject, user, apikey, "", reader, int64(requestLength), isVerbose)
	if err == nil {
		if isVerbose {
			log.Printf("Created new version. %v", resp)
		}
	}
	return err
}

func getVersions(apihost, apikey, subject, repository, pkg string, isVerbose bool) ([]string, error) {
	client := &http.Client{}
	url := apihost + "/packages/" + subject + "/" + repository + "/" + pkg
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(subject, apikey)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error calling %s - %v", url, err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body - %v", err)
		return nil, err
	}
	resp.Body.Close()
	var b map[string]interface{}
	err = json.Unmarshal(body, &b)
	if err != nil {
		log.Printf("Error parsing json body - %v", err)
		log.Printf("Body: %s", body)
		return nil, err
	}
	if isVerbose {
		log.Printf("Body: %s", body)
	}
	if versions, keyExists := b["versions"]; keyExists {
		versionsSlice, err := typeutils.ToStringSlice(versions, "versions")

		return versionsSlice, err
	}
	return nil, errors.New("Versions not listed!")
}

/*
// sample usage
func main() {
  target_url := "http://localhost:8888/"
  filename := "/path/to/file.rtf"
  postFile(filename, target_url)
}
*/
