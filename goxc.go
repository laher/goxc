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
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const VERSION = "0.0.4"

// Platform names.
const (
	AMD64   = "amd64"
	X86     = "386"
	ARM     = "arm"
	DARWIN  = "darwin"
	LINUX   = "linux"
	FREEBSD = "freebsd"
	WINDOWS = "windows"
)

var PLATFORMS = [][]string{
	{DARWIN, X86},
	{DARWIN, AMD64},
	{FREEBSD, X86},
	{FREEBSD, AMD64},
	//tried to add FREEBSD/ARM but not working for me yet. 2013-02-15
	//{ FREEBSD, ARM },
	{LINUX, X86},
	{LINUX, AMD64},
	{LINUX, ARM},
	{WINDOWS, X86},
	{WINDOWS, AMD64},
}

var (
	flagSet          = flag.NewFlagSet("goxc", flag.ExitOnError)
	verbose          bool
	isHelp           bool
	isVersion        bool
	isBuildToolchain bool
	aos              string
	aarch            string
	artifactVersion  string
	artifactsDest    string
)

func redirectIO(cmd *exec.Cmd) (*os.File, error) {
	// this function copied from 'https://github.com/laher/mkdo'
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println(err)
	}
	if verbose {
		log.Printf("Redirecting output")
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	//direct. Masked passwords work OK!
	cmd.Stdin = os.Stdin
	return nil, err
}

func BuildToolchain(goos string, arch string) {
	goroot := os.Getenv("GOROOT")
	gohostos := runtime.GOOS
	gohostarch := runtime.GOARCH
	if verbose {
		log.Printf("Host OS = %s", gohostos)
	}
	var scriptname string
	if gohostos == WINDOWS {
		scriptname = `\src\make.bat`
	} else {
		scriptname = "/src/make.bash"
	}
	cmd := exec.Command(goroot + scriptname)
	cmd.Dir = goroot + string(os.PathSeparator) + "src"
	cmd.Args = append(cmd.Args, "--no-clean")
	var cgoEnabled string
	if goos == gohostos && arch == gohostarch {
		//note: added conditional in line with Dave Cheney, but this combination is not yet supported.
		if gohostos == FREEBSD && gohostarch == ARM {
			cgoEnabled = "0"
		} else {
			cgoEnabled = "1"
		}
	} else {
		cgoEnabled = "0"
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS="+goos)
	cmd.Env = append(cmd.Env, "CGO_ENABLED="+cgoEnabled)
	cmd.Env = append(cmd.Env, "GOARCH="+arch)
	if goos == LINUX && arch == ARM {
		// see http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go
		cmd.Env = append(cmd.Env, "GOARM=5")
	}
	if verbose {
		log.Printf("'make.bash' env: GOOS=%s, CGO_ENABLED=%s, GOARCH=%s, GOROOT=%s", goos, cgoEnabled, arch, goroot)
		log.Printf("'make.bash' args: %s", cmd.Args)
		log.Printf("'make.bash' working directory: %s", cmd.Dir)
	}
	f, err := redirectIO(cmd)
	if err != nil {
		log.Printf("Error redirecting IO: %s", err)
	}
	if f != nil {
		defer f.Close()
	}

	err = cmd.Start()
	if err != nil {
		log.Printf("Launch error: %s", err)
		// return 1, err
	} else {
		err = cmd.Wait()
		if err != nil {
			log.Printf("Wait error: %s", err)
		} else {
			if verbose {
				log.Printf("Complete")
			}
		}
	}
}

func XCPlat(goos string, arch string, call []string, isFirst bool) string {
	log.Printf("building for platform %s_%s.", goos, arch)

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Printf("GOPATH env variable not set! Using '.'")
		gopath = "."
	}
	gohostos := runtime.GOOS
	app_dirname, err := filepath.Abs(call[0])
	if err != nil {
		log.Printf("Error: %v", err)
	}
	appName := filepath.Base(app_dirname)

	relativeDir := artifactVersion + string(os.PathSeparator) + goos + "_" + arch

	var outDestRoot string
	if artifactsDest != "" {
		outDestRoot = artifactsDest
	} else {
		gobin := os.Getenv("GOBIN")
		if gobin == "" {
			gobin = gopath + string(os.PathSeparator) + "bin"
		}
		outDestRoot = gobin + string(os.PathSeparator) + appName + "-xc"
	}

	outDir := outDestRoot + string(os.PathSeparator) + relativeDir
	os.MkdirAll(outDir, 0755)

	cmd := exec.Command("go")
	cmd.Args = append(cmd.Args, "build")
	cmd.Args = append(cmd.Args, "-o")
	cmd.Dir = call[0]
	var ending = ""
	if goos == WINDOWS {
		ending = ".exe"
	}
	relativeBinForMarkdown := goos + "_" + arch + string(os.PathSeparator) + appName + ending
	relativeBin := relativeDir + string(os.PathSeparator) + appName + ending
	cmd.Args = append(cmd.Args, outDestRoot+string(os.PathSeparator)+relativeBin)
	cmd.Args = append(cmd.Args, ".") //relative to pwd (specified in call[0])

	cmd.Env = os.Environ()

	var cgoEnabled string
	if goos == gohostos {
		cgoEnabled = "1"
	} else {
		cgoEnabled = "0"
	}

	cmd.Env = append(cmd.Env, "GOOS="+goos)
	cmd.Env = append(cmd.Env, "CGO_ENABLED="+cgoEnabled)
	cmd.Env = append(cmd.Env, "GOARCH="+arch)
	if verbose {
		log.Printf("'go' env: GOOS=%s, CGO_ENABLED=%s, GOARCH=%s", goos, cgoEnabled, arch)
		log.Printf("'go' args: %s", cmd.Args)
		log.Printf("'go' working directory: %s", cmd.Dir)
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("Launch error: %s", err)
	} else {
		err = cmd.Wait()
		if err != nil {
			log.Printf("Compiler error: %s", err)
		} else {
			log.Printf("Artifact generated OK")
			reportFilename := outDestRoot + string(os.PathSeparator) + artifactVersion + string(os.PathSeparator) + "downloads.md"
			var flags int
			if isFirst {
				log.Printf("Creating %s", reportFilename)
				flags = os.O_WRONLY | os.O_TRUNC | os.O_CREATE
			} else {
				flags = os.O_APPEND | os.O_WRONLY

			}
			f, err := os.OpenFile(reportFilename, flags, 0600)
			if err == nil {
				defer f.Close()
				if isFirst {
					_, err = fmt.Fprintf(f, "%s downloads (%s)\n------------\n\n", appName, artifactVersion)
				}
				_, err = fmt.Fprintf(f, " * [%s %s](%s)\n", goos, arch, relativeBinForMarkdown)
			}
			if err != nil {
				log.Printf("Could not report to '%s': %s", reportFilename, err)
			}
		}
	}
	return relativeBin
}

func printHelp() {
	fmt.Fprint(os.Stderr, "`goxc` [options] <directory_name>\n")
	fmt.Fprintf(os.Stderr, " Version %s. Options:\n", VERSION)
	flagSet.PrintDefaults()
}

func printVersion() {
	fmt.Fprintf(os.Stderr, " goxc version %s\n", VERSION)
}

func GOXC(call []string) (int, error) {
	e := flagSet.Parse(call[1:])
	if e != nil {
		return 1, e
	}
	remainder := flagSet.Args()
	if isHelp {
		printHelp()
		return 0, nil
	} else if isVersion {
		printVersion()
		return 0, nil
	} else if isBuildToolchain {
		//no need for remaining args
	} else if len(remainder) < 1 {
		printHelp()
		return 1, nil
	}

	isFirst := true
	for _, v := range PLATFORMS {
		if aos == "" || v[0] == aos {
			if aarch == "" || v[1] == aarch {
				if isBuildToolchain {
					BuildToolchain(v[0], v[1])
				} else {
					XCPlat(v[0], v[1], remainder, isFirst)
				}
				isFirst = false
			}
		}
	}
	return 0, nil
}

func main() {
	log.SetPrefix("[goxc] ")
	flagSet.StringVar(&aos, "os", "", "Specify OS (linux/darwin/windows). Compiles all by default")
	flagSet.StringVar(&aarch, "arch", "", "Specify Arch (386/x64/arm). Compiles all by default")
	flagSet.StringVar(&artifactVersion, "av", "latest", "Artifact version (default='latest')")
	flagSet.StringVar(&artifactsDest, "d", "", "Destination root directory (default=$GOBIN)")
	flagSet.BoolVar(&isBuildToolchain, "t", false, "Build cross-compiler toolchain(s)")
	flagSet.BoolVar(&isHelp, "h", false, "Show this help")
	flagSet.BoolVar(&isVersion, "version", false, "version info")
	flagSet.BoolVar(&verbose, "v", false, "verbose")
	GOXC(os.Args)
}
