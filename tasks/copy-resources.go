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
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/core"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//runs automatically
func init() {
	register(Task{
		TASK_COPY_RESOURCES,
		"Copy resources",
		runTaskCopyResources,
		nil})
}

func runTaskCopyResources(tp taskParams) error {
	resources := core.ParseIncludeResources(tp.workingDirectory, tp.settings.Resources.Include, tp.settings.IsVerbose())
	destFolder := filepath.Join(tp.outDestRoot, tp.settings.GetFullVersionName())
	for _, resource := range resources {
		if strings.HasPrefix(resource, tp.workingDirectory) {
			resource = resource[len(tp.workingDirectory)+1:]
		}
		_, err := copyFile(filepath.Join(destFolder, resource), filepath.Join(tp.workingDirectory, resource))
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(dstName, srcName string) (written int64, err error) {
	log.Printf("Copying file %s to %s", srcName, dstName)
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
