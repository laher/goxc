package goxc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

//just for test-driving the 'test' task
/*
func TestFail(t *testing.T) {
	t.Fatalf("FAIL")
}
*/

func TestSanityCheck(t *testing.T) {
	//goroot := runtime.GOROOT()
	err := SanityCheck("")
	if err == nil {
		t.Fatalf("sanity check failed! Expected to flag missing GOROOT variable")
	}
	tmpDir, err := ioutil.TempDir("", "goxc_test_sanityCheck")
	defer os.RemoveAll(tmpDir)
	err = goxc.SanityCheck(tmpDir)
	if err == nil {
		t.Fatalf("sanity check failed! Expected to notice missing src folder")
	}

	srcDir := filepath.Join(tmpDir, "src")
	os.Mkdir(srcDir, 0700)
	scriptname := toolchain.GetMakeScriptPath(tmpDir)
	ioutil.WriteFile(scriptname, []byte("1"), 0111)
	err = SanityCheck(tmpDir)
	if err != nil {
		t.Fatalf("sanity check failed! Did not find src folder: %v", err)
	}
	os.Chmod(srcDir, 0600)
	defer os.Chmod(srcDir, 0700)
	err = SanityCheck(tmpDir)
	if err == nil {
		t.Fatalf("sanity check failed! Expected NOT to be able to open src folder")
	}
}
