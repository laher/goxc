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
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/source"
)

const PKG_NAME = "goxc"
const PKG_VERSION = "0.1.5"

const (
	AMD64   = "amd64"
	X86     = "386"
	ARM     = "arm"
	DARWIN  = "darwin"
	LINUX   = "linux"
	FREEBSD = "freebsd"
	WINDOWS = "windows"

	DEFAULT_VERSION   = "latest"
	DEFAULT_RESOURCES = "INSTALL*,README*,LICENSE*"

	MSG_INSTALL_GO_FROM_SOURCE = "goxc requires Go to be installed from Source. Please follow instructions at http://golang.org/doc/install/source"
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

type Settings struct {
	verbose          bool
	isHelp           bool
	isVersion        bool
	isBuildToolchain bool
	zipArchives      bool
	aos              string
	arch             string
	packageVersion   string
	artifactsDest    string
	codesign         string
	includeResources string
}

var (
	flagSet  = flag.NewFlagSet("goxc", flag.ExitOnError)
	settings Settings
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
	if settings.verbose {
		log.Printf("Redirecting output")
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	//direct. Masked passwords work OK!
	cmd.Stdin = os.Stdin
	return nil, err
}

func sanityCheck(goroot string) error {
	if goroot == "" {
		return errors.New("GOROOT environment variable is NOT set.")
	} else {
		if settings.verbose {
			log.Printf("Found GOROOT=%s", goroot)
		}
	}
	scriptpath := getMakeScriptPath(goroot)
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

func getMakeScriptPath(goroot string) string {
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
	goroot := runtime.GOROOT()
	gohostos := runtime.GOOS
	gohostarch := runtime.GOARCH
	if settings.verbose {
		log.Printf("Host OS = %s", gohostos)
	}
	scriptpath := getMakeScriptPath(goroot)
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
	if settings.verbose {
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
	if settings.verbose {
		log.Printf("Complete")
	}
}

func moveBinaryToZIP(outDir, binPath, appName string, resources []string) (zipFilename string, err error) {
	zipFileBaseName := appName + "_" + filepath.Base(filepath.Dir(binPath))
	if settings.packageVersion != "" && settings.packageVersion != DEFAULT_VERSION {
		zipFilename = zipFileBaseName + "_" + settings.packageVersion + ".zip"
	} else {
		zipFilename = zipFileBaseName + ".zip"
	}
	zf, err := os.Create(filepath.Join(outDir, zipFilename))
	if err != nil {
		return
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)

	addFileToZIP(zw, binPath)
	if err != nil {
		zw.Close()
		return
	}
	//resources
	for _, resource := range resources {
		addFileToZIP(zw, resource)
		if err != nil {
			zw.Close()
			return
		}
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

func addFileToZIP(zw *zip.Writer, path string) (err error) {
	binfo, err := os.Stat(path)
	if err != nil {
		return
	}
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
	bf, err := os.Open(path)
	if err != nil {
		return
	}
	defer bf.Close()
	_, err = io.Copy(w, bf)
	return
}

func parseIncludeResources(basedir string, includeResources string) []string {
	allMatches := []string{}
	if includeResources != "" {
		resourceGlobs := strings.Split(includeResources, ",")
		for _, resourceGlob := range resourceGlobs {
			matches, err := filepath.Glob(filepath.Join(basedir, resourceGlob))
			if err == nil {
				allMatches = append(allMatches, matches...)
			} else {
				log.Printf("GLOB error: %s: %s", resourceGlob, err)
			}
		}
	}
	if settings.verbose {
		log.Printf("Resources to include: %v", allMatches)
	}
	return allMatches

}

func signBinary(binPath string) error {
	cmd := exec.Command("settings.codesign")
	cmd.Args = append(cmd.Args, "-s", settings.codesign, binPath)
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

	relativeDir := filepath.Join(settings.packageVersion, goos+"_"+arch)

	var outDestRoot string
	if settings.artifactsDest != "" {
		outDestRoot = settings.artifactsDest
	} else {
		gobin := os.Getenv("GOBIN")
		if gobin == "" {
			// follow usual GO rules for making GOBIN
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
	if settings.verbose {
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

			resources := parseIncludeResources(call[0], settings.includeResources)

			// settings.codesign only works on OS X for binaries generated for OS X.
			if settings.codesign != "" && gohostos == DARWIN && goos == DARWIN {
				if err := signBinary(filepath.Join(outDestRoot, relativeBin)); err != nil {
					log.Printf("settings.codesign failed: %s", err)
				} else {
					log.Printf("Signed with ID: %q", settings.codesign)
				}
			}

			if settings.zipArchives {
				// Create ZIP archive.
				zipPath, err := moveBinaryToZIP(
					filepath.Join(outDestRoot, settings.packageVersion),
					filepath.Join(outDestRoot, relativeBin), appName, resources)
				if err != nil {
					log.Printf("ZIP error: %s", err)
				} else {
					relativeBinForMarkdown = zipPath
					log.Printf("Artifact zipped OK")
				}
			} else {
				//TODO: move resources to folder & add links to downloads.md
			}

			// Output report
			reportFilename := filepath.Join(outDestRoot, settings.packageVersion, "downloads.md")
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
					_, err = fmt.Fprintf(f, "%s downloads (%s)\n------------\n\n", appName, settings.packageVersion)
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
	fmt.Fprintf(os.Stderr, " Version %s. Options:\n", PKG_VERSION)
	flagSet.PrintDefaults()
}

func printVersion() {
	fmt.Fprintf(os.Stderr, " goxc version %s\n", PKG_VERSION)
}

//merge configuration file and parse source
//TODO honour build flags
//TODO merge in configuration file
func mergeConfiguredSettings(dir string) {
	if settings.verbose {
		log.Printf("loading configured settings")
	}
	// flags take precedence
	if settings.packageVersion == DEFAULT_VERSION || settings.packageVersion == "" {
		glob := filepath.Join(dir, "*.go")
		if settings.verbose {
			log.Printf("looking up PKG_VERSION in %s", glob)
		}
		matches, err := filepath.Glob(glob)
		if err != nil {
			log.Printf("Glob error %v", err)
		} else {
			if settings.verbose {
				log.Printf("Glob found %v", matches)
			}
			files, err := source.LoadFiles(matches)
			if err != nil {
				log.Printf("%v", err)
				return
			}
			for _, f := range files {
				pkgVersion := source.FindConstantValue(f, "xPKG_VERSION")
				if pkgVersion != "" {
					settings.packageVersion = pkgVersion
					log.Printf("Package version = %v", pkgVersion)
				}
			}
		}

	}
}

func GOXC(call []string) {
	if err := flagSet.Parse(call[1:]); err != nil {
		log.Printf("Error parsing arguments: %s", err)
		os.Exit(1)
	}
	if settings.isHelp {
		printHelp()
		os.Exit(0)
	}
	if settings.isVersion {
		printVersion()
		os.Exit(0)
	}

	goroot := runtime.GOROOT()
	if err := sanityCheck(goroot); err != nil {
		log.Printf("Error: %s", err)
		log.Printf(MSG_INSTALL_GO_FROM_SOURCE)
		os.Exit(1)
	}

	remainder := flagSet.Args()
	if !settings.isBuildToolchain && len(remainder) < 1 {
		printHelp()
		os.Exit(1)
	}

	//taken from config plus parsed sources
	mergeConfiguredSettings(remainder[0])

	if settings.verbose {
		log.Printf("looping through each platform")
	}
	destOses := strings.Split(settings.aos, ",")
	destArchs := strings.Split(settings.arch, ",")

	isFirst := true
	for _, v := range PLATFORMS {
		for _, dOs := range destOses {
			if dOs == "" || v[0] == dOs {
				for _, dArch := range destArchs {
					if dArch == "" || v[1] == dArch {
						if settings.isBuildToolchain {
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
	flagSet.StringVar(&settings.aos, "os", "", "Specify OS (linux,darwin,windows,freebsd). Compiles all by default")
	flagSet.StringVar(&settings.arch, "arch", "", "Specify Arch (386,amd64,arm). Compiles all by default")
	flagSet.StringVar(&settings.packageVersion, "pv", DEFAULT_VERSION, "Package version (default='latest')")
	flagSet.StringVar(&settings.packageVersion, "av", DEFAULT_VERSION, "Package version (deprecated option name)")
	flagSet.StringVar(&settings.artifactsDest, "d", "", "Destination root directory (default=$GOBIN/(appname)-xc)")
	flagSet.StringVar(&settings.codesign, "codesign", "", "identity to sign darwin binaries with (only when host OS is OS X)")
	flagSet.StringVar(&settings.includeResources, "include", DEFAULT_RESOURCES, "Include resources in zips") //TODO: Add resources to non-zips & downloads.md
	flagSet.BoolVar(&settings.isBuildToolchain, "t", false, "Build cross-compiler toolchain(s)")
	flagSet.BoolVar(&settings.isHelp, "h", false, "Show this help")
	flagSet.BoolVar(&settings.isVersion, "version", false, "version info")
	flagSet.BoolVar(&settings.verbose, "v", false, "verbose")
	flagSet.BoolVar(&settings.zipArchives, "z", true, "create ZIP archives instead of folders")

	GOXC(os.Args)
}
