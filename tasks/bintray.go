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
	"github.com/laher/goxc/typeutils"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
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
			"apihost":         "https://api.bintray.com/",
			"downloadshost":          "https://dl.bintray.com/",
			"isdownloadspage": true, "downloadspage": "bintray.md",
			"include": "*.tar.gz,*.deb,*.zip", "exclude": "*.exe"}})
}

func runTaskBintray(tp taskParams) error {
	subject := tp.settings.GetTaskSettingString(TASK_BINTRAY, "subject")
	apikey := tp.settings.GetTaskSettingString(TASK_BINTRAY, "apikey")
	repository := tp.settings.GetTaskSettingString(TASK_BINTRAY, "repository")
	pkg := tp.settings.GetTaskSettingString(TASK_BINTRAY, "package")
	apiHost := tp.settings.GetTaskSettingString(TASK_BINTRAY, "apihost")
	downloadsHost := tp.settings.GetTaskSettingString(TASK_BINTRAY, "downloadshost")
	versionDir := filepath.Join(tp.outDestRoot, tp.settings.GetFullVersionName())

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
	filename := tp.settings.GetTaskSettingString(TASK_BINTRAY, "downloadspage")
	reportFilename := filepath.Join(tp.outDestRoot, tp.settings.GetFullVersionName(), filename)
	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	f, err := os.OpenFile(reportFilename, flags, 0600)
	if err != nil {
		return err
	}

	fileInfos, err := ioutil.ReadDir(versionDir)
	if err != nil {
		return err
	}
	defer f.Close()
	fileheader := tp.settings.GetTaskSettingString(TASK_DOWNLOADS_PAGE, "fileheader")
	if fileheader != "" {
		_, err = fmt.Fprintf(f, "%s\n\n", fileheader)
	}
	_, err = fmt.Fprintf(f, "%s downloads (version %s)\n-------------\n", tp.appName, tp.settings.GetFullVersionName())
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
					relativePath := fi.Name() + "/" + fi2.Name()
					fullPath := filepath.Join(versionDir, relativePath)
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
					url := apiHost + "/content/" + subject + "/" + repository + "/" + pkg + "/" + tp.settings.GetFullVersionName() + "/" + relativePath
					// for some reason there's no /pkg/ level in the downloads url.
					downloadsUrl := downloadsHost + "/content/" + subject + "/" + repository + "/" + tp.settings.GetFullVersionName() + "/" + relativePath + "?direct"
					resp, err := doMultipartFile("PUT", url, subject, apikey, fullPath, relativePath)
					if err != nil {
						return err
					}
					log.Printf("File uploaded. %v", resp)
					//TODO: move out of loop. Only once per batch? LOLWUT I don't know what this comment means any more.
					_, err = fmt.Fprintf(f, " [[%s](%s)]", text, downloadUrl)
					if err != nil {
						return err
					}
					err = publish(apiHost, apikey, subject, repository, pkg, tp.settings.GetFullVersionName())
					if err != nil {
						return err
					}

				}
			}

		}
	}
	return err
}

func publish(apihost, apikey, subject, repository, pkg, version string) error {
	resp, err := doHttp("POST", apihost+"/content/"+subject+"/"+repository+"/"+pkg+"/"+version+"/publish", subject, apikey, nil, 0)
	if err == nil {
		log.Printf("File uploaded. %v", resp)
	}
	return err
}

//PUT /content/:subject/:repo/:package/:version/:path
func uploadFile(apihost, apikey, subject, repository, pkg, version, fullPath, relativePath string) error {
	url := apihost + "/content/" + subject + "/" + repository + "/" + pkg + "/" + version + "/" + relativePath
	/*
		file, err := os.Open(fullPath)
		if err != nil {
			log.Printf("Error reading file for upload: %v", err)
			return err
		}
		defer file.Close()
		fi, err := file.Stat()
		if err != nil {
			log.Printf("Error reading file for upload: %v", err)
			return err
		}
		resp, err := doHttp("PUT", url, subject, apikey, file, fi.Size())
	*/
	resp, err := doMultipartFile("PUT", url, subject, apikey, fullPath, relativePath)
	if err == nil {
		log.Printf("File uploaded. %v", resp)
	}
	return err
}

//POST /packages/:subject/:repo/:package/versions
func createVersion(apihost, apikey, subject, repository, pkg, version string) error {
	req := map[string]interface{}{"name": version, "release_notes": "built by goxc", "release_url": "http://x.x.x/x/x"}
	requestData, err := json.Marshal(req)
	if err != nil {
		return err
	}
	requestLength := len(requestData)
	reader := bytes.NewReader(requestData)
	resp, err := doHttp("POST", apihost+"/packages/"+subject+"/"+repository+"/"+pkg+"/versions", subject, apikey, reader, int64(requestLength))
	if err == nil {
		log.Printf("Created new version. %v", resp)
	}
	return err
}

func doMultipartFile(method, url, subject, apikey, fullPath, relativePath string) (*http.Response, error) {
	bodyBuf := bytes.NewBufferString("")
	bodyWriter := multipart.NewWriter(bodyBuf)

	// write the Part headers to the buffer
	_, err := bodyWriter.CreateFormFile("upfile", relativePath)
	if err != nil {
		fmt.Println("error writing to buffer")
		return nil, err
	}

	// write the file data
	fh, err := os.Open(fullPath)
	if err != nil {
		fmt.Println("error opening file")
		return nil, err
	}
	// boundary
	boundary := bodyWriter.Boundary()
	closeBuf := bytes.NewBufferString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	// multi-reader defers the reading of the file data
	requestReader := io.MultiReader(bodyBuf, fh, closeBuf)
	fi, err := fh.Stat()
	if err != nil {
		fmt.Printf("Error Stating file: %s", fullPath)
		return nil, err
	}
	req, err := http.NewRequest(method, url, requestReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.SetBasicAuth(subject, apikey)
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundary)
	req.ContentLength = fi.Size() + int64(bodyBuf.Len()) + int64(closeBuf.Len())

	return http.DefaultClient.Do(req)
}

func doHttp(method, url, subject, apikey string, requestReader io.Reader, requestLength int64) (map[string]interface{}, error) {
	client := &http.Client{}
	log.Printf("reader: %+v", requestReader)
	req, err := http.NewRequest(method, url, requestReader)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(subject, apikey)
	if requestLength > 0 {
		log.Printf("Adding Header - Content-Length: %s", strconv.FormatInt(requestLength, 10))
		req.ContentLength = requestLength
	}
	//log.Printf("req: %v", req)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	log.Printf("Http complete")
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
	log.Printf("Body: %s", body)
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
