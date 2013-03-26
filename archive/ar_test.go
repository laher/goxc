package archive

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


/*
import (
	"io/ioutil"
	"path/filepath"
	"testing"
)
//TODO: this is commented out because it's an unfinished proof-of-concept.
func TestArForDeb(t *testing.T) {
	ioutil.WriteFile(filepath.Join("test_data","debian-binary"), []byte("2.0\n"), 0644)
	ioutil.WriteFile(filepath.Join("test_data","control"), []byte(`Package: goxc
Priority: extra
Maintainer: Am Laher
Architecture: i386
Version: 0.5.2
Depends: golang
Provides: goxc
Description: Cross-compiler utility for Go
`), 0644)
	TarGz(filepath.Join("test_data","control.tar.gz"), [][]string{[]string{ filepath.Join("test_data","control"), "control"} })
	TarGz(filepath.Join("test_data","data.tar.gz"), [][]string{[]string{ filepath.Join("test_data","goxc"), "/usr/bin/goxc"} })
	targetFile := filepath.Join("test_data","goxc_0.5.2_i386.deb")
	inputs := [][]string{
	 []string{filepath.Join("test_data","debian-binary"),"debian-binary"},
	 []string{filepath.Join("test_data","control.tar.gz"),"control.tar.gz"},
	 []string{filepath.Join("test_data","data.tar.gz"),"data.tar.gz"}}
	err := ArForDeb(targetFile, inputs)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
*/
