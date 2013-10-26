package ar

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
	"io"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/executils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

//TODO: this is commented out because it's an unfinished proof-of-concept.
func NoTestUnAr(t *testing.T) {
	goroot := runtime.GOROOT()
	goos := "linux"
	arch := "arm"
	platPkgFileRuntime := filepath.Join(goroot, "pkg", goos+"_"+arch, "runtime.a")
	nr, err := os.Open(platPkgFileRuntime)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	tr, err := NewReader(nr)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	for {
		h, err := tr.Next()
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		log.Printf("Header: %+v", h)
		if h.Name == "__.PKGDEF" {
			firstLine := make([]byte, 50)
			if _, tr.err = io.ReadFull(tr.r, firstLine); tr.err != nil {
				t.Fatalf("failed: %v", err)
			}
			tr.nb -= 50
			log.Printf("pkgdef first part: '%s'", string(firstLine))
			expectedPrefix := "go object " + goos + " " + arch + " "
			if !strings.HasPrefix(string(firstLine), expectedPrefix) {
				t.Fatalf("failed: does not match '%s'", expectedPrefix)
			}
			parts := strings.Split(string(firstLine), " ")
			compiledVersion := parts[4]
			log.Printf("Compiled version: %s", compiledVersion)
			runtimeVersion := runtime.Version()
			log.Printf("Runtime version: %s", runtimeVersion)
			cmd := exec.Command("go")
			args := []string{"version"}
			err = executils.PrepareCmd(cmd, ".", args, []string{}, false)
			out, err := cmd.Output()
			if err != nil {
				log.Printf("`go version` failed: %v", err)
			}
			log.Printf("output: %s", string(out))
			goParts := strings.Split(string(out), " ")
			goVersion := goParts[2]
			log.Printf("Go version: %s", goVersion)
			if compiledVersion != goVersion {
				t.Fatalf("Package version does NOT match!")
			}
		}
	}
}
