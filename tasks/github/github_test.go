package github

import (
	"flag"
	"os"
	"testing"

	"github.com/laher/goxc/tasks"
	"github.com/laher/goxc/tasks/httpc"
)

var apikey = flag.String("api-key", "", "API key")

var (
	version   = "v0.16.0-testx2"
	tagName   = "v0.16.0-testx"
	apihost   = "https://api.github.com"
	owner     = "laher"
	repo      = "goxc"
	isVerbose = true
)

func TestCreateRelease(t *testing.T) {
	if *apikey == "" {
		t.Skip("api-key is required to run this integration test")
	}
	t.Logf("create release")
	err := createRelease(apihost, owner, *apikey, repo, tagName, version, "Built by goxc", true, isVerbose)
	if err != nil {
		t.Errorf("Error creating release %v", err)
	}
}

///repos/:owner/:repo/releases/tags/:tag
func TestGetTagRelease(t *testing.T) {
	if *apikey == "" {
		t.Skip("api-key is required to run this integration test")
	}
	id, err := ghGetReleaseForTag(apihost, owner, *apikey, repo, tagName, isVerbose)
	if err != nil {
		t.Errorf("Error getting release %v", err)
	}
	t.Logf("ID: %s", id)
}

func TestGetReleases(t *testing.T) {
	if *apikey == "" {
		t.Skip("api-key is required to run this integration test")
	}
	r, err := httpc.DoHttp("GET", apihost+"/repos/"+owner+"/"+repo+"/releases", "", owner, *apikey, "", nil, 0, isVerbose)
	if err != nil {
		t.Errorf("Error getting release %v", err)
	}
	a, err := httpc.ParseSlice(r, isVerbose)
	if err != nil {
		t.Errorf("Error getting release %v", err)
	}
	for _, i := range a {
		id := i["id"]
		name := i["name"]
		t.Logf("ID: %0.f  Name: %s", id, name)
		//		for k, v := range i {
		//			t.Logf("Entry: %s %+v", k, v)
		//		}
	}
	//t.Logf("Response data: %v", a)
}

func TestGhDoUpload(t *testing.T) {

	if *apikey == "" {
		t.Skip("api-key is required to run this integration test")
	}
	t.Logf("do upload")
	apihost := "https://uploads.github.com"
	version := "v0.16.0-testx"
	relativePath := "tasks.go"
	fullPath := "tasks.go"
	contentType := "text/plain"
	err := ghDoUpload(apihost, *apikey, owner, repo, version, relativePath, fullPath, contentType, true, false)
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
