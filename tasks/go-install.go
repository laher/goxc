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
	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/executils"
)

//runs automatically
func init() {
	Register(Task{
		TASK_GO_INSTALL,
		"runs `go install`. installs a version consistent with goxc-built binaries.",
		runTaskGoInstall,
		nil})
}

func runTaskGoInstall(tp TaskParams) error {
	for _, mainDir := range tp.MainDirs {
		err := executils.InvokeGo(mainDir, "install", []string{}, []string{}, tp.Settings)
		if err != nil {
			return err
		}
	}
	return nil
}
