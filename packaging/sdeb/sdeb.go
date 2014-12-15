package sdeb

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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/laher/goxc/archive"
	"github.com/laher/goxc/core"
	//"text/template"
)

//TODO: unfinished: need to discover root dir to determine which dirs to pre-make.
func SdebGetSourcesAsArchiveItems(codeDir, prefix string) ([]archive.ArchiveItem, error) {
	goPathRoot := core.GetGoPathElement(codeDir)
	goPathRootResolved, err := filepath.EvalSymlinks(goPathRoot)
	if err != nil {
		log.Printf("Could not evaluate symlinks for %s", goPathRoot)
		goPathRootResolved = goPathRoot
	}
	log.Printf("Code dir %s (using goPath element %s)", codeDir, goPathRootResolved)
	return sdebGetSourcesAsArchiveItems(goPathRootResolved, codeDir, prefix)
}

//
func sdebGetSourcesAsArchiveItems(goPathRoot, codeDir, prefix string) ([]archive.ArchiveItem, error) {
	sources := []archive.ArchiveItem{}
	//1. Glob for files in this dir
	//log.Printf("Globbing %s", codeDir)
	matches, err := filepath.Glob(filepath.Join(codeDir, "*.go"))
	if err != nil {
		return sources, err
	}
	for _, match := range matches {
		relativeMatch, err := filepath.Rel(goPathRoot, match)
		if err != nil {
			return nil, errors.New("Error finding go sources " + err.Error())
		}
		destName := filepath.Join(prefix, relativeMatch)
		//log.Printf("Putting file %s in %s", match, destName)
		sources = append(sources, archive.ArchiveItemFromFileSystem(match, destName))
	}

	//2. Recurse into subdirs
	fis, err := ioutil.ReadDir(codeDir)
	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != DIRNAME_TEMP {
			additionalItems, err := sdebGetSourcesAsArchiveItems(goPathRoot, filepath.Join(codeDir, fi.Name()), prefix)
			sources = append(sources, additionalItems...)
			if err != nil {
				return sources, err
			}
		}
	}
	return sources, err
}

func SdebCopySourceRecurse(codeDir, destDir string) (err error) {
	log.Printf("Globbing %s", codeDir)
	//get all files and copy into destDir
	matches, err := filepath.Glob(filepath.Join(codeDir, "*.go"))
	if err != nil {
		return err
	}
	if len(matches) > 0 {
		err = os.MkdirAll(destDir, 0777)
		if err != nil {
			return err
		}
	}
	for _, match := range matches {
		//TODO copy files
		log.Printf("copying %s into %s", match, filepath.Join(destDir, filepath.Base(match)))
		r, err := os.Open(match)
		if err != nil {
			return err
		}
		defer func() {
			err := r.Close()
			if err != nil {
				panic(err)
			}
		}()
		w, err := os.Create(filepath.Join(destDir, filepath.Base(match)))
		if err != nil {
			return err
		}
		defer func() {
			err := w.Close()
			if err != nil {
				panic(err)
			}
		}()

		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
	}
	fis, err := ioutil.ReadDir(codeDir)
	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != DIRNAME_TEMP {
			err = SdebCopySourceRecurse(filepath.Join(codeDir, fi.Name()), filepath.Join(destDir, fi.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
// prepare folders and debian/ files.
// (everything except copying source)
func SdebPrepare(workingDirectory, appName, maintainer, version, arches, description, buildDepends string, metadataDeb map[string]interface{}) (err error) {
	//make temp dir & subfolders
	tmpDir := filepath.Join(workingDirectory, DIRNAME_TEMP)
	debianDir := filepath.Join(tmpDir, "debian")
	err = os.MkdirAll(filepath.Join(debianDir, "source"), 0777)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(tmpDir, "src"), 0777)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(tmpDir, "bin"), 0777)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(tmpDir, "pkg"), 0777)
	if err != nil {
		return err
	}
	//write control file and related files
	tpl, err := template.New("rules").Parse(TEMPLATE_DEBIAN_RULES)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(debianDir, "rules"))
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()
	err = tpl.Execute(file, appName)
	if err != nil {
		return err
	}
	sdebControlFile := getSdebControlFileContent(appName, maintainer, version, arches, description, buildDepends, metadataDeb)
	ioutil.WriteFile(filepath.Join(debianDir, "control"), sdebControlFile, 0666)
	//copy source into folders
	//call dpkg-build, if available
	return err
}
*/
func getSdebControlFileContent(appName, maintainer, version, arches, description, buildDepends string, metadataDeb map[string]interface{}) []byte {
	control := fmt.Sprintf("Source: %s\nPriority: extra\n", appName)
	if maintainer != "" {
		control = fmt.Sprintf("%sMaintainer: %s\n", control, maintainer)
	}
	if buildDepends == "" {
		buildDepends = BUILD_DEPENDS_DEFAULT
	}
	control = fmt.Sprintf("%sBuildDepends: %s\n", control, buildDepends)
	control = fmt.Sprintf("%sStandards-Version: %s\n", control, STANDARDS_VERSION_DEFAULT)

	//TODO - homepage?

	control = fmt.Sprintf("%sVersion: %s\n", control, version)

	control = fmt.Sprintf("%sPackage: %s\n", control, appName)

	//mandatory
	control = fmt.Sprintf("%sArchitecture: %s\n", control, arches)
	for k, v := range metadataDeb {
		control = fmt.Sprintf("%s%s: %s\n", control, k, v)
	}
	control = fmt.Sprintf("%sDescription: %s\n", control, description)
	return []byte(control)
}
