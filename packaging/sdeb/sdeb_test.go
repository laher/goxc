package sdeb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSdebBuild(t *testing.T) {
	workingDirectory := "."
	err := os.MkdirAll(workingDirectory, 0777)
	if err != nil {
		t.Fatalf("%v", err)
	}
	/* SdebPrepare is old now
	err = SdebPrepare(workingDirectory, "my-app-x", "A L", "1.2.3-alpha", "platform-x", "This package does x", "", *new(map[string]interface{}))
	if err != nil {
		t.Fatalf("%v", err)
	}
	*/
	tmpDir := filepath.Join(workingDirectory, DIRNAME_TEMP)
	destDir := filepath.Join(tmpDir, "src")
	workingDirectory = "../.."
	err = SdebCopySourceRecurse(workingDirectory, destDir)
	if err != nil {
		t.Fatalf("%v", err)
	}
	//TODO: find code & copy
	//ioutil.WriteFile(filepath.Join(debianDir, "control"), sdebControlFile, 0666)
	//TODO: targz
}
