package main

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
	"os/exec"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"

)


// 0.3.1
func InvokeGo(workingDirectory string, args []string) error {
	log.Printf("invoking 'go %v' on '%s'", args, workingDirectory)
	cmd := exec.Command("go")
	cmd.Args = append(cmd.Args, args...)
	if args[0] == config.TASK_INSTALL || args[0] == "build" || args[0] == config.TASK_TEST {
		addLdFlagVersion(settings, cmd)
	}
	cmd.Dir = workingDirectory
	f, err := redirectIO(cmd)
	if err != nil {
		log.Printf("Error redirecting IO: %s", err)
		return err
	}
	if f != nil {
		defer f.Close()
	}
	if settings.IsVerbose() {
		log.Printf("'go' args: %v", cmd.Args)
		log.Printf("'go' working directory: %s", cmd.Dir)
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("Launch error: %s", err)
		return err
	} else {
		err = cmd.Wait()
		if err != nil {
			log.Printf("invocation error: %s", err)
			return err
		} else {
			log.Printf("go succeeded")
		}
	}
	return nil
}


// this function copied from 'https://github.com/laher/mkdo'
func redirectIO(cmd *exec.Cmd) (*os.File, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println(err)
	}
	if settings.IsVerbose() {
		log.Printf("Redirecting output")
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	//direct. Masked passwords work OK!
	cmd.Stdin = os.Stdin
	return nil, err
}
