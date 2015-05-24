package github

import (
	"os"
	"testing"
	"github.com/laher/goxc/tasks"
	"github.com/laher/goxc/tasks/httpc"
)

var (
	 apikey = ""
	version = "v0.16.0-testx2"
	tagName = "v0.16.0-testx"
	apihost = "https://api.github.com"
	owner = "laher"
	repo = "goxc"
	isVerbose = true
)

func TestCreateRelease(t *testing.T) {
	t.Logf("create release")
	err := createRelease(apihost, owner, apikey, repo, tagName, version, isVerbose)
	if err != nil {
		t.Errorf("Error creating release %v", err)
	}
}
func TestGetRelease(t *testing.T) {
	_, err := httpc.DoHttp("GET", apihost+"/repos/"+owner+"/"+repo+"/releases", "", owner, apikey, "", nil, 0, isVerbose)
	if err != nil {
		t.Errorf("Error getting release %v", err)
	}
}

func TestGhDoUpload(t *testing.T) {
	t.Logf("do upload")
	apihost := "https://uploads.github.com"
	version := "v0.16.0-testx"
	relativePath := "tasks.go"
	fullPath := "tasks.go"
	contentType := "text/plain"
	err := ghDoUpload(apihost, apikey, owner, repo, version, relativePath, fullPath, contentType, false)
	if err != nil {
		t.Errorf("Error creating release %v", err)
	}
}

func TestGhWalkFunc(t *testing.T) {
	t.Skip()
	fullPath := ""
	var fi2 os.FileInfo
	var errIn error
	reportFilename := ""
	dirs := []string{}
	var tp tasks.TaskParams
	var report tasks.BtReport
	format := ""
	err := ghWalkFunc(fullPath, fi2, errIn, reportFilename, dirs, tp, format, report)
	if err != nil {

		t.Errorf("Error doing walkFunc %v", err)
	}
}
