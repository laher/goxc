package tasks

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
	"log"
	"os"
	"path/filepath"
	"strings"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/core"
)

//runs automatically
func init() {
	Register(Task{
		TASK_COPY_RESOURCES,
		"Copy resources",
		runTaskCopyResources,
		nil})
}

func runTaskCopyResources(tp TaskParams) error {
	resources := core.ParseIncludeResources(tp.WorkingDirectory, tp.Settings.ResourcesInclude, tp.Settings.ResourcesExclude, !tp.Settings.IsQuiet())
	destFolder := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName())
	if !tp.Settings.IsQuiet() {
		log.Printf("resources: %v", resources)
	}
	for _, resource := range resources {
		if strings.HasPrefix(resource, tp.WorkingDirectory) {
			resource = resource[len(tp.WorkingDirectory)+1:]
		}
		//_, resourcebase := filepath.Split(resource)
		sourcePath := filepath.Join(tp.WorkingDirectory, resource)
		destPath := filepath.Join(destFolder, resource)
		finfo, err := os.Lstat(sourcePath)
		if err != nil {
			return err
		}
		if finfo.IsDir() {
			err = os.MkdirAll(destPath, 0777)
			if err != nil && !os.IsExist(err) {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(destPath), 0777)
			if err != nil && !os.IsExist(err) {
				return err
			}
			_, err = copyFile(sourcePath, destPath, !tp.Settings.IsQuiet())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func copyDir(srcDir, destDir string, isVerbose bool) (fileCount int, err error) {
	fileCount = 0
	err = os.MkdirAll(destDir, 0777)
	if err != nil && !os.IsExist(err) {
		return fileCount, err
	}
	err = filepath.Walk(srcDir, func(path string, fi os.FileInfo, err error) error {
		fileCount++
		base := strings.Replace(path, srcDir, "", 1)
		dest := filepath.Join(destDir, base)
		if fi.IsDir() {
			err := os.Mkdir(dest, 0777)
			if os.IsExist(err) {
				return nil
			} else {
				return err
			}
		} else {
			if isVerbose {
				log.Printf("path: %s, base: %s", path, base)
			}
			_, err := copyFile(path, dest, isVerbose)
			return err
		}
	})
	return
}

func copyFile(srcName, dstName string, isVerbose bool) (written int64, err error) {
	if isVerbose {
		log.Printf("Copying file %s to %s", srcName, dstName)
	}
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return
	}
	defer dst.Close()

	return io.Copy(dst, src)
}
