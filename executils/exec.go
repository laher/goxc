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
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/platforms"
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
			_, err := buf.WriteString(flag + " " + k + " '" + typedV + "' ")
			if err != nil {
				log.Printf("Error writing flags")
			}
		default:
			_, err := buf.WriteString(fmt.Sprintf("%s %s '%v' ", flag, k, typedV))
			if err != nil {
				log.Printf("Error writing flags")
			}
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

func splitEnvVar(asString string) (string, string, error) {
	parts := strings.SplitN(asString, "=", 2)
	if len(parts) > 1 {
		return parts[0], parts[1], nil
	} else {
		return "", "", errors.New("Invalid env variable definition")
	}
}

// invoke the go command via the os/exec package
// 0.3.1
// v0.9 changed signature
func InvokeGo(workingDirectory string, subCmd string, subCmdArgs []string, env []string, settings *config.Settings) error {
	fullVersionName := settings.GetFullVersionName()
	//var buildSettings config.BuildSettings
	buildSettings := settings.BuildSettings
	goRoot := settings.GoRoot
	if settings.IsVerbose() {
		log.Printf("build settings: %s", goRoot)
	}
	cmdPath := filepath.Join(goRoot, "bin", "go")
	args := []string{subCmd}

	//these features only apply to `go build` & `go install`
	if isBuildCommand(subCmd) {
		if buildSettings.Processors != nil {
			args = append(args, "-p", strconv.Itoa(*buildSettings.Processors))
		}
		if buildSettings.Race != nil && *buildSettings.Race {
			args = append(args, "-race")
		}
		if buildSettings.Verbose != nil && *buildSettings.Verbose {
			args = append(args, "-v")
		}
		if buildSettings.PrintCommands != nil && *buildSettings.PrintCommands {
			args = append(args, "-x")
		}
		if buildSettings.CcFlags != nil && *buildSettings.CcFlags != "" {
			args = append(args, "-ccflags", *buildSettings.CcFlags)
		}
		if buildSettings.Compiler != nil && *buildSettings.Compiler != "" {
			args = append(args, "-compiler", *buildSettings.Compiler)
		}
		if buildSettings.GccGoFlags != nil && *buildSettings.GccGoFlags != "" {
			args = append(args, "-gccgoflags", *buildSettings.GccGoFlags)
		}
		if buildSettings.GcFlags != nil && *buildSettings.GcFlags != "" {
			args = append(args, "-gcflags", *buildSettings.GcFlags)
		}
		if buildSettings.InstallSuffix != nil && *buildSettings.InstallSuffix != "" {
			args = append(args, "-installsuffix", *buildSettings.InstallSuffix)
		}
		ldflags := ""
		if buildSettings.LdFlags != nil {
			ldflags = *buildSettings.LdFlags
		}
		if buildSettings.LdFlagsXVars != nil {
			//TODO!
			ldflags = ldflags + " " + buildFlags(buildInterpolationVars(*buildSettings.LdFlagsXVars, fullVersionName), "-X")
		} else {
			log.Printf("WARNING: LdFlagsXVars is nil. Not passing package version into compiler")
		}
		if ldflags != "" {
			args = append(args, "-ldflags", ldflags)
		}
		if buildSettings.Tags != nil && *buildSettings.Tags != "" {
			args = append(args, "-tags", *buildSettings.Tags)
		}
		if len(buildSettings.ExtraArgs) > 0 {
			args = append(args, buildSettings.ExtraArgs...)
		}
	}
	if settings.IsVerbose() {
		log.Printf("Env: %v", settings.Env)
	}
	if len(settings.Env) > 0 {
		vars := struct {
			PS  string
			PLS string
			Env map[string]string
		}{
			string(os.PathSeparator),
			string(os.PathListSeparator),
			map[string]string{},
		}
		for _, val := range os.Environ() {
			k, v, err := splitEnvVar(val)
			if err != nil {
				//ignore invalid env vars from environment
			} else {
				vars.Env[k] = v
			}
		}
		for _, envTpl := range settings.Env {
			if settings.IsVerbose() {
				log.Printf("Processing env var %s", envTpl)
			}
			tpl, err := template.New("envItem").Parse(envTpl)
			if err != nil {
				return err
			}
			var dest bytes.Buffer
			err = tpl.Execute(&dest, vars)
			if err != nil {
				return err
			}
			executed := dest.String()
			if settings.IsVerbose() {
				if envTpl != executed {
					log.Printf("Setting env var (converted from %s to %s)", envTpl, executed)
				} else {
					log.Printf("Setting env var from config: %s", executed)
				}
			}
			env = append(env, dest.String())
			//new address if necessary
			k, v, err := splitEnvVar(dest.String())
			if err != nil {
				//fail on badly specified ENV vars
				return errors.New("Invalid env var defined by settings")
			} else {
				vars.Env[k] = v
			}
		}
	}
	args = append(args, subCmdArgs...)
	cmd, err := NewCmd(cmdPath, workingDirectory, args, env, settings.IsVerbose(), !settings.IsQuiet())
	if err != nil {
		return err
	}
	if settings.IsVerbose() {
		log.Printf("invoking '%s %v' from '%s'", cmdPath, PrintableArgs(args), workingDirectory)
	}

	err = StartAndWait(cmd)
	if err != nil {
		log.Printf("'go' returned error: %s", err)
		return err
	}
	if settings.IsVerbose() {
		log.Printf("'go' completed successfully")
	}
	return nil

}

func NewCmd(cmdPath string, workingDirectory string, args []string, env []string, isVerbose bool, isRedirectToStdout bool) (*exec.Cmd, error) {
	cmd := exec.Command(cmdPath)
	if isRedirectToStdout {
		RedirectIO(cmd)
	}
	return cmd, PrepareCmd(cmd, workingDirectory, args, env, isVerbose)
}

func PrepareCmd(cmd *exec.Cmd, workingDirectory string, args []string, env []string, isVeryVerbose bool) error {
	cmd.Args = append(cmd.Args, args...)
	cmd.Dir = workingDirectory
	cmd.Env = CombineActualEnv(append(cmd.Env, env...), isVeryVerbose)
	return nil
}

// StartAndWait starts the given command and waits for it to exit.  If the
// command started successfully but exited with an error, any output to stderr
// is included in the error message.
func StartAndWait(cmd *exec.Cmd) error {
	stderr := &bytes.Buffer{}
	if cmd.Stderr == nil {
		cmd.Stderr = stderr
	} else {
		cmd.Stderr = io.MultiWriter(cmd.Stderr, stderr)
	}
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("Launch error: %s", err)
	} else {
		err = cmd.Wait()
		if err != nil {
			if stderr.Len() > 0 {
				return fmt.Errorf("Wait error: %s: %s", err, strings.TrimSpace(stderr.String()))
			}
			return fmt.Errorf("Wait error: %s", err)
		}
		return err
	}
}

func CombineActualEnv(env []string, isVeryVerbose bool) []string {
	//0.7.4 env replaces os.Environ
	cmdEnv := []string{}
	cmdEnv = append(cmdEnv, env...)
	for _, thisProcessEnvItem := range os.Environ() {
		thisProcessEnvItemSplit := strings.Split(thisProcessEnvItem, "=")
		key := thisProcessEnvItemSplit[0]
		exists := false
		for _, specifiedEnvItem := range env {
			specifiedEnvItemSplit := strings.Split(specifiedEnvItem, "=")
			specifiedEnvKey := specifiedEnvItemSplit[0]
			if specifiedEnvKey == key {
				if isVeryVerbose {
					log.Printf("Overriding ENV variable (%s replaces %s)", specifiedEnvItem, thisProcessEnvItem)
				}
				exists = true
			}
		}
		if !exists {
			cmdEnv = append(cmdEnv, thisProcessEnvItem)
		}
	}
	if isVeryVerbose {
		log.Printf("(verbose!) all env vars for 'go': %s", cmdEnv)
		if env != nil && len(env) > 0 {
			log.Printf("specified env vars for 'go': %s", env)
		}
	}
	return cmdEnv
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
