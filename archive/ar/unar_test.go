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
	"log"
	"os"
	"runtime"
	"path/filepath"
	"testing"
)
//TODO: this is commented out because it's an unfinished proof-of-concept.
func NoTestUnAr(t *testing.T) {
	nr, err := os.Open(filepath.Join(runtime.GOROOT(), "pkg/linux_386/runtime.a"))
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
	}
}

