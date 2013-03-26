package executils

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
	"github.com/laher/goxc/core"
	"runtime"
	"strings"
)

// ldflags relate to any build task ...
func GetLdFlagVersionArgs(fullVersionName string) []string {
	if fullVersionName != "" {
		return []string{"-ldflags", "-X main.VERSION "+fullVersionName+""}
	}
	return []string{}
}

// 0.3.1
func InvokeGo(workingDirectory string, args []string, envExtra []string, isVerbose bool) error {
	cmd := exec.Command("go")
	cmd.Args = append(cmd.Args, args...)
	cmd.Dir = workingDirectory
	f, err := RedirectIO(cmd)
	if err != nil {
		log.Printf("Error redirecting IO: %s", err)
	}
	if f != nil {
		defer f.Close()
	}

	cmd.Env = append([]string{}, os.Environ()...)
	cmd.Env = append(cmd.Env, envExtra...)

	if envExtra != nil && len(envExtra) > 0 {
		log.Printf("'go' extra env vers: %s", envExtra)
	}
	log.Printf("invoking 'go %v' from '%s'", printableArgs(args), workingDirectory)
	err = cmd.Start()
	if err != nil {
		log.Printf("Launch error: %s", err)
		return err
	} else {
		err = cmd.Wait()
		if err != nil {
			log.Printf("Go returned error: %s", err)
			return err
		} else {
			if isVerbose {
				log.Printf("go succeeded")
			}
		}
	}
	return nil
}

func printableArgs(args []string) string {
	ret:= ""
	for _, arg := range args {
		if strings.Contains(arg, " ") {
			ret = ret + " \"" + arg + "\""
		} else {
			ret = ret + " " + arg
		}
	}
	return ret
}

// this function copied from 'https://github.com/laher/mkdo'
func RedirectIO(cmd *exec.Cmd) (*os.File, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println(err)
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	//direct. Masked passwords work OK!
	cmd.Stdin = os.Stdin
	return nil, err
}

//0.2.4 refactored this out
func CgoEnabled(goos, arch string) string {
	var cgoEnabled string
	if goos == runtime.GOOS && arch == runtime.GOARCH {
		//note: added conditional in line with Dave Cheney, but this combination is not yet supported.
		if runtime.GOOS == core.FREEBSD && runtime.GOARCH == core.ARM {
			cgoEnabled = "0"
		} else {
			cgoEnabled = "1"
		}
	} else {
		cgoEnabled = "0"
	}
	return cgoEnabled
}
