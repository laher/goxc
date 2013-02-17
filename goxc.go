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
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const VERSION = "0.1.4"

const (
	AMD64   = "amd64"
	X86     = "386"
	ARM     = "arm"
	DARWIN  = "darwin"
	LINUX   = "linux"
	FREEBSD = "freebsd"
	WINDOWS = "windows"
)

const MSG_INSTALL_GO_FROM_SOURCE = "goxc requires Go to be installed from Source. Please follow instructions at http://golang.org/doc/install/source"

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
	zipArchives      bool
	aos              string
	aarch            string
	artifactVersion  string
	artifactsDest    string
	codesign         string
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

func sanityCheck() error {
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		return errors.New("GOROOT environment variable is NOT set.")
	} else {
		log.Printf("Found GOROOT=%s", goroot)
	}
	scriptpath := getMakeScriptPath()
	_, err := os.Stat(scriptpath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New(fmt.Sprintf("Make script ('%s') does not exist!", scriptpath))
		} else {
			return errors.New(fmt.Sprintf("Error reading make script ('%s'): %v", scriptpath, err))
		}
	}
	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func getMakeScriptPath() string {
	goroot := os.Getenv("GOROOT")
	gohostos := runtime.GOOS
	var scriptname string
	if gohostos == WINDOWS {
		scriptname = "make.bat"
	} else {
		scriptname = "make.bash"
	}
	return filepath.Join(goroot, "src", scriptname)
}

func BuildToolchain(goos string, arch string) {
	goroot := os.Getenv("GOROOT")
	gohostos := runtime.GOOS
	gohostarch := runtime.GOARCH
	if verbose {
		log.Printf("Host OS = %s", gohostos)
	}
	scriptpath := getMakeScriptPath()
	cmd := exec.Command(scriptpath)
	cmd.Dir = filepath.Join(goroot, "src")
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

	cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED="+cgoEnabled, "GOARCH="+arch)
	if goos == LINUX && arch == ARM {
		// see http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go
		cmd.Env = append(cmd.Env, "GOARM=5")
	}
	if verbose {
		log.Printf("'make' env: GOOS=%s, CGO_ENABLED=%s, GOARCH=%s, GOROOT=%s", goos, cgoEnabled, arch, goroot)
		log.Printf("'make' args: %s", cmd.Args)
		log.Printf("'make' working directory: %s", cmd.Dir)
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
		return
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Wait error: %s", err)
		return
	}
	if verbose {
		log.Printf("Complete")
	}
}

func moveBinaryToZIP(outDir, binPath, appName string) (zipFilename string, err error) {
	zipFileBaseName := appName + "_" + filepath.Base(filepath.Dir(binPath))
	if artifactVersion != "" && artifactVersion != "latest" {
		zipFilename = zipFileBaseName + "_" + artifactVersion + ".zip"
	} else {
		zipFilename = zipFileBaseName + ".zip"
	}
	zf, err := os.Create(filepath.Join(outDir, zipFilename))
	if err != nil {
		return
	}
	defer zf.Close()
	binfo, err := os.Stat(binPath)
	if err != nil {
		return
	}
	bf, err := os.Open(binPath)
	if err != nil {
		return
	}
	defer bf.Close()
	zw := zip.NewWriter(zf)
	header, err := zip.FileInfoHeader(binfo)
	if err != nil {
		return
	}
	header.Method = zip.Deflate
	w, err := zw.CreateHeader(header)
	if err != nil {
		zw.Close()
		return
	}
	_, err = io.Copy(w, bf)
	if err != nil {
		zw.Close()
		return
	}
	err = zw.Close()
	if err != nil {
		return
	}
	// Remove binary and its directory.
	err = os.Remove(binPath)
	if err != nil {
		return
	}
	err = os.Remove(filepath.Dir(binPath))
	if err != nil {
		return
	}
	return
}

func signBinary(binPath string) error {
	cmd := exec.Command("codesign")
	cmd.Args = append(cmd.Args, "-s", codesign, binPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func XCPlat(goos, arch string, call []string, isFirst bool) string {
	log.Printf("building for platform %s_%s.", goos, arch)

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Printf("GOPATH env variable not set! Using '.'")
		gopath = "."
	}
	gohostos := runtime.GOOS
	appDirname, err := filepath.Abs(call[0])
	if err != nil {
		log.Printf("Error: %v", err)
	}
	appName := filepath.Base(appDirname)

	relativeDir := filepath.Join(artifactVersion, goos+"_"+arch)

	var outDestRoot string
	if artifactsDest != "" {
		outDestRoot = artifactsDest
	} else {
		gobin := os.Getenv("GOBIN")
		if gobin == "" {
			gobin = filepath.Join(gopath, "bin")
		}
		outDestRoot = filepath.Join(gobin, appName+"-xc")
	}

	outDir := filepath.Join(outDestRoot, relativeDir)
	os.MkdirAll(outDir, 0755)

	cmd := exec.Command("go")
	cmd.Args = append(cmd.Args, "build", "-o")
	cmd.Dir = call[0]
	var ending = ""
	if goos == WINDOWS {
		ending = ".exe"
	}
	relativeBinForMarkdown := filepath.Join(goos+"_"+arch, appName+ending)
	relativeBin := filepath.Join(relativeDir, appName+ending)
	cmd.Args = append(cmd.Args, filepath.Join(outDestRoot, relativeBin), ".")

	var cgoEnabled string
	if goos == gohostos {
		cgoEnabled = "1"
	} else {
		cgoEnabled = "0"
	}

	cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED="+cgoEnabled, "GOARCH="+arch)
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
			// Codesign only works on OS X for binaries generated for OS X.
			if codesign != "" && gohostos == DARWIN && goos == DARWIN {
				if err := signBinary(filepath.Join(outDestRoot, relativeBin)); err != nil {
					log.Printf("codesign failed: %s", err)
				} else {
					log.Printf("Signed with ID: %q", codesign)
				}
			}

			if zipArchives {
				// Create ZIP archive.
				zipPath, err := moveBinaryToZIP(
					filepath.Join(outDestRoot, artifactVersion),
					filepath.Join(outDestRoot, relativeBin), appName)
				if err != nil {
					log.Printf("ZIP error: %s", err)
				} else {
					relativeBinForMarkdown = zipPath
					log.Printf("Artifact zipped OK")
				}
			}

			// Output report
			reportFilename := filepath.Join(outDestRoot, artifactVersion, "downloads.md")
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

func GOXC(call []string) {
	if err := flagSet.Parse(call[1:]); err != nil {
		log.Printf("Error parsing arguments: %s", err)
		os.Exit(1)
	}
	if isHelp {
		printHelp()
		os.Exit(0)
	}
	if isVersion {
		printVersion()
		os.Exit(0)
	}

	if err := sanityCheck(); err != nil {
		log.Printf("Error: %s", err)
		log.Printf(MSG_INSTALL_GO_FROM_SOURCE)
		os.Exit(1)
	}

	remainder := flagSet.Args()
	if !isBuildToolchain && len(remainder) < 1 {
		printHelp()
		os.Exit(1)
	}

	destOses := strings.Split(aos, ",")
	destArchs := strings.Split(aarch, ",")

	isFirst := true
	for _, v := range PLATFORMS {
		for _, dOs := range destOses {
			if dOs == "" || v[0] == dOs {
				for _, dArch := range destArchs {
					if dArch == "" || v[1] == dArch {
						if isBuildToolchain {
							BuildToolchain(v[0], v[1])
						} else {
							XCPlat(v[0], v[1], remainder, isFirst)
						}
						isFirst = false
					}
				}
			}
		}
	}
}

func main() {
	log.SetPrefix("[goxc] ")
	flagSet.StringVar(&aos, "os", "", "Specify OS (linux,darwin,windows,freebsd). Compiles all by default")
	flagSet.StringVar(&aarch, "arch", "", "Specify Arch (386,amd64,arm). Compiles all by default")
	flagSet.StringVar(&artifactVersion, "av", "latest", "Artifact version (default='latest')")
	flagSet.StringVar(&artifactsDest, "d", "", "Destination root directory (default=$GOBIN)")
	flagSet.StringVar(&codesign, "codesign", "", "identity to sign darwin binaries with (only when host OS is OS X)")
	flagSet.BoolVar(&isBuildToolchain, "t", false, "Build cross-compiler toolchain(s)")
	flagSet.BoolVar(&isHelp, "h", false, "Show this help")
	flagSet.BoolVar(&isVersion, "version", false, "version info")
	flagSet.BoolVar(&verbose, "v", false, "verbose")
	flagSet.BoolVar(&zipArchives, "z", true, "create ZIP archives instead of folders")

	GOXC(os.Args)
}
