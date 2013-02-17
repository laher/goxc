package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestSanityCheck(t *testing.T) {
	goroot := os.Getenv("GOROOT")
	defer os.Setenv("GOROOT", goroot) //just incase it doesnt sort itself out in time for subsequent tests
	os.Setenv("GOROOT", "")
	err := sanityCheck()
	if err == nil {
		t.Fatalf("sanity check failed! Expected to flag missing GOROOT variable")
	}
	tmpDir, err := ioutil.TempDir("", "goxc_test_sanityCheck")
	defer os.RemoveAll(tmpDir)
	os.Setenv("GOROOT", tmpDir)
	err = sanityCheck()
	if err == nil {
		t.Fatalf("sanity check failed! Expected to notice missing src folder")
	}

	srcDir := filepath.Join(tmpDir, "src")
	os.Mkdir(srcDir, 0700)
	scriptname := getMakeScriptPath()
	ioutil.WriteFile(scriptname, []byte("1"), 0111)
	err = sanityCheck()
	if err != nil {
		//os.Remove(srcDir)
		t.Fatalf("sanity check failed! Did not find src folder", err)
	}
	os.Chmod(srcDir, 0600)
	defer os.Chmod(srcDir, 0700)
	err = sanityCheck()
	if err == nil {
		t.Fatalf("sanity check failed! Expected NOT to be able to open src folder")
	}
}
