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

	"log"
)

//runs automatically
func init() {
	Register(Task{
		TASK_GO_TEST,
		"runs `go test ./...`. (dir is configurable).",
		runTaskGoTest,
		map[string]interface{}{"dir": "./...", "i": false, "short": false}})
}

func runTaskGoTest(tp TaskParams) error {
	dir := tp.Settings.GetTaskSettingString(TASK_GO_TEST, "dir")
	i := tp.Settings.GetTaskSettingBool(TASK_GO_TEST, "i") //this should be false by default! leaving it exposed for invocation as a flag
	short := tp.Settings.GetTaskSettingBool(TASK_GO_TEST, "short")
	args := []string{}
	if i {
		args = append(args, "-i")
	}
	if short {
		args = append(args, "-short")
	}
	args = append(args, dir)
	if tp.Settings.IsVerbose() {
		log.Printf("Running `go test` with args: %v", args)
	}
	err := executils.InvokeGo(tp.WorkingDirectory, "test", args, []string{}, tp.Settings)
	return err
}
