// Utilities for invoking exec. Mainly focused on 'go build' and cross-compilation
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
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/platforms"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	BUILD_COMMANDS = []string{"build", "install"}
	)

// get list of args to be used in variable interpolation
// ldflags are used in any to any build-related go task (install,build,test)
/*
func GetLdFlagVersionArgs(fullVersionName string) string {
	input := map[string]interface{}{"main.BUILD_DATE": time.Now().Format(time.RFC3339)}
	if fullVersionName != "" {
		input["main.VERSION"] = fullVersionName
	}
	return GetInterpolationFlags(input, "-X")
}
*/
func buildInterpolationVars(args map[string]interface{}, fullVersionName string) map[string]interface{} {
	ret := map[string]interface{}{}
	for k, v := range args {
		switch typedV := v.(type) {
		case string:
			switch k {
			case "Version":
				ret[typedV] = fullVersionName
			case "TimeNow":
				ret[typedV] = time.Now().Format(time.RFC3339)
			}

		default:
			//error here?
		}
	}
	return ret
}

// get list of args to be used in e.g. ldflags variable interpolation
// v0.9 changed from ldflags-specific to more general flag building
func buildFlags(args map[string]interface{}, flag string) string {
	if len(args) < 1 {
		return ""
	}
	//ret := make([]string, len(args))
	var buf bytes.Buffer
	for k, v := range args {
		switch typedV := v.(type) {
		case string:
			buf.WriteString(flag + " " + k + " '" + typedV + "' ")
		default:
			buf.WriteString(fmt.Sprintf("%s %s '%v' ", flag, k, typedV))
		}
	}
	return buf.String()
}

func isBuildCommand(subCmd string) bool {
	for _, buildCmd := range BUILD_COMMANDS {
		if subCmd == buildCmd {
			return true
		}
	}
	return false
}

// invoke the go command via the os/exec package
// 0.3.1
// v0.9 changed signature
func InvokeGo(workingDirectory string, subCmd string, subCmdArgs []string, env []string, settings config.Settings) error {
	isVerbose := settings.IsVerbose()
	fullVersionName := settings.GetFullVersionName()
	var buildSettings config.BuildSettings
	if settings.BuildSettings != nil {
		buildSettings = *settings.BuildSettings
	} else {
		buildSettings = config.BuildSettingsDefault()
	}
	goRoot := runtime.GOROOT()

	//log.Printf("using build settings (%+v)", buildSettings)
	if buildSettings.GoRoot != "" {
		goRoot = buildSettings.GoRoot
	}
	cmdPath := filepath.Join(goRoot, "bin", "go")
	cmd := exec.Command(cmdPath)
	RedirectIO(cmd)
	args := []string{subCmd}
	//these commands only apply to `go build` & `go install`
	if isBuildCommand(subCmd) {
		if buildSettings.Processors > 0 {
			args = append(args, "-p", string(buildSettings.Processors))
		}
		if buildSettings.Race {
			args = append(args, "-race")
		}
		if buildSettings.Verbose {
			args = append(args, "-v")
		}
		if buildSettings.PrintCommands {
			args = append(args, "-x")
		}
		if buildSettings.CcFlags != "" {
			args = append(args, "-ccflags", buildSettings.CcFlags)
		}
		if buildSettings.Compiler != "" {
			args = append(args, "-compiler", buildSettings.Compiler)
		}
		if buildSettings.GccGoFlags != "" {
			args = append(args, "-gccgoflags", buildSettings.GccGoFlags)
		}
		if buildSettings.GcFlags != "" {
			args = append(args, "-gcflags", buildSettings.GcFlags)
		}
		if buildSettings.InstallSuffix != "" {
			args = append(args, "-installsuffix", buildSettings.InstallSuffix)
		}
		ldflags := ""
		if buildSettings.LdFlags != "" {
			ldflags = buildSettings.LdFlags
		}
		if buildSettings.LdFlagsXVars != nil {
			//TODO!
			ldflags = ldflags + " " + buildFlags(buildInterpolationVars(*buildSettings.LdFlagsXVars, fullVersionName), "-X")
		}
		if ldflags != "" {
			args = append(args, "-ldflags", ldflags)
		}
		if buildSettings.Tags != "" {
			args = append(args, "-tags", buildSettings.Tags)
		}
	}
	args = append(args, subCmdArgs...)
	err := PrepareCmd(cmd, workingDirectory, args, env, isVerbose)
	if err != nil {
		return err
	}
	log.Printf("invoking '%s %v' from '%s'", cmdPath, PrintableArgs(args), workingDirectory)
	err = cmd.Start()
	if err != nil {
		log.Printf("Launch error: %s", err)
		return err
	} else {
		err = cmd.Wait()
		if err != nil {
			log.Printf("'go' returned error: %s", err)
			return err
		} else {
			if isVerbose {
				log.Printf("'go' completed successfully")
			}
		}
	}
	return nil

}

func PrepareCmd(cmd *exec.Cmd, workingDirectory string, args []string, env []string, isVerbose bool) error {
	cmd.Args = append(cmd.Args, args...)
	cmd.Dir = workingDirectory

	//if f != nil {
		//defer f.Close()
	//}
	//0.7.4 env replaces os.Environ
	cmd.Env = append(cmd.Env, env...)
	for _, thisProcessEnvItem := range os.Environ() {
		thisProcessEnvItemSplit := strings.Split(thisProcessEnvItem, "=")
		key := thisProcessEnvItemSplit[0]
		exists := false
		for _, specifiedEnvItem := range env {
			specifiedEnvItemSplit := strings.Split(specifiedEnvItem, "=")
			specifiedEnvKey := specifiedEnvItemSplit[0]
			if specifiedEnvKey == key {
				log.Printf("Overriding ENV variable (%s replaces %s)", specifiedEnvItem, thisProcessEnvItem)
				exists = true
			}
		}
		if !exists {
			cmd.Env = append(cmd.Env, thisProcessEnvItem)
		}
	}
	if isVerbose {
		log.Printf("(verbose!) all env vars for 'go': %s", cmd.Env)
	}
	if env != nil && len(env) > 0 {
		log.Printf("specified env vars for 'go': %s", env)
	}
	return nil
}

// returns a list of printable args
func PrintableArgs(args []string) string {
	ret := ""
	for _, arg := range args {
		if len(ret) > 0 {
			ret = ret + " "
		}
		if strings.Contains(arg, " ") {
			ret = ret + "\"" + arg + "\""
		} else {
			ret = ret + arg
		}
	}
	return ret
}
func RedirectIO(cmd *exec.Cmd) {
	RedirectIOTo(cmd, os.Stdin, os.Stdout, os.Stderr)
}

func RedirectIOTo(cmd *exec.Cmd, myin io.Reader, myout, myerr io.Writer) {
// redirect IO
	cmd.Stdout = myout
	cmd.Stderr = myerr
	cmd.Stdin = myin
	//return nil, err
}

// check if cgoEnabled is required.
//0.2.4 refactored this out
// TODO not needed for go1.1+. Remove this once go1.0 reaches end of life. (when is that?)
func CgoEnabled(goos, arch string) string {
	var cgoEnabled string
	if goos == runtime.GOOS && arch == runtime.GOARCH {
		//note: added conditional in line with Dave Cheney, but this combination is not yet supported.
		if runtime.GOOS == platforms.FREEBSD && runtime.GOARCH == platforms.ARM {
			cgoEnabled = "0"
		} else {
			cgoEnabled = "1"
		}
	} else {
		cgoEnabled = "0"
	}
	return cgoEnabled
}
