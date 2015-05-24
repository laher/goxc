package httpc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type HttpError struct {
	StatusCode int
	message    string
}

func (e HttpError) Error() string {
	return fmt.Sprintf("Error code: %d, message: %s", e.StatusCode, e.message)
}

func DoHttp(method, url, _deprecated, user, apikey, contentType string, requestReader io.Reader, requestLength int64, isVerbose bool) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, requestReader)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(user, apikey)
	if requestLength > 0 {
		if isVerbose {
			log.Printf("Adding Header - Content-Length: %s", strconv.FormatInt(requestLength, 10))
		}
		req.ContentLength = requestLength
	}
	if contentType != "" {
		if isVerbose {
			log.Printf("Adding Header - Content-Type: %s", contentType)
		}

		req.Header.Add("Content-Type", contentType)
	}
	//log.Printf("req: %v", req)
	if isVerbose {
		log.Printf("%s to %s", method, url)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if isVerbose {
		log.Printf("Http response received")
	}

	return resp, nil
}

func ParseSlice(resp *http.Response, isVerbose bool) ([]map[string]interface{}, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		log.Printf("Error closing response body: %v", err)
	}
	//200 is OK, 201 is Created, etc
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("Error code: %s", resp.Status)
		log.Printf("Error body: %s", body)
		return nil, HttpError{resp.StatusCode, resp.Status}
	}
	if isVerbose {
		log.Printf("Response status: '%s', Body: %s", resp.Status, body)
	}
	var b []map[string]interface{}
	if len(body) > 0 {
		err = json.Unmarshal(body, &b)
		if err != nil {
			return nil, err
		}
	}
	return b, err
}

func ParseMap(resp *http.Response, isVerbose bool) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		log.Printf("Error closing response body: %v", err)
	}
	//200 is OK, 201 is Created, etc
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("Error code: %s", resp.Status)
		log.Printf("Error body: %s", body)
		return nil, HttpError{resp.StatusCode, resp.Status}
	}
	if isVerbose {
		log.Printf("Response status: '%s', Body: %s", resp.Status, body)
	}
	var b map[string]interface{}
	if len(body) > 0 {
		err = json.Unmarshal(body, &b)
		if err != nil {
			return nil, err
		}
	}
	return b, err
}

func GetContentType(text string) string {
	contentType := "text/plain"
	if strings.HasSuffix(text, ".zip") {
		contentType = "application/zip"
	} else if strings.HasSuffix(text, ".deb") {
		contentType = "application/vnd.debian.binary-package"
	} else if strings.HasSuffix(text, ".tar.gz") {
		contentType = "application/x-gzip"
	}
	return contentType
}

func UploadFile(method, url, subject, user, apikey, fullPath, relativePath, contentType string, isVerbose bool) (map[string]interface{}, error) {
	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("Error reading file for upload: %v", err)
		return nil, err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		log.Printf("Error statting file for upload: %v", err)
		return nil, err
	}
	resp, err := DoHttp(method, url, subject, user, apikey, contentType, file, fi.Size(), isVerbose)
	r, err := ParseMap(resp, isVerbose)
	return r, err
}
