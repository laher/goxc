// GOXC IS NOT READY FOR USE AS AN API - function names and packages will continue to change until version 1.0
package platforms

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
	//"github.com/laher/goxc/archive"
	//"github.com/laher/goxc/config"
	"log"
	"runtime"
	"strings"
)

const (
	AMD64   = "amd64"
	X86     = "386"
	ARM     = "arm"
	DARWIN  = "darwin"
	LINUX   = "linux"
	FREEBSD = "freebsd"
	NETBSD  = "netbsd"
	OPENBSD = "openbsd"
	WINDOWS = "windows"
	PLAN9   = "plan9"
)

var (
	OSES                    = []string{DARWIN, LINUX, FREEBSD, NETBSD, OPENBSD, PLAN9, WINDOWS}
	ARCHS                   = []string{X86, AMD64, ARM}
	SUPPORTED_PLATFORMS_1_0 = [][]string{
		{DARWIN, X86},
		{DARWIN, AMD64},
		{LINUX, X86},
		{LINUX, AMD64},
		{LINUX, ARM},
		{FREEBSD, X86},
		{FREEBSD, AMD64},
		// {FREEBSD, ARM},
		// couldnt build toolchain for netbsd using a linux 386 host: 2013-02-19
		//	{NETBSD, X86},
		//	{NETBSD, AMD64},
		{OPENBSD, X86},
		{OPENBSD, AMD64},
		{WINDOWS, X86},
		{WINDOWS, AMD64}}
	NEW_PLATFORMS_1_1 = [][]string{
		{FREEBSD, ARM},
		{NETBSD, X86},
		{NETBSD, AMD64},
		{NETBSD, ARM},
		{PLAN9, X86}}

	SUPPORTED_PLATFORMS_1_1 = append(append([][]string{}, SUPPORTED_PLATFORMS_1_0...), NEW_PLATFORMS_1_1...)
)

func getSupportedPlatforms() [][]string {
	if strings.HasPrefix(runtime.Version(), "go1.0") {
		return SUPPORTED_PLATFORMS_1_0
	}
	return SUPPORTED_PLATFORMS_1_1
}

//0.5 add support for space delimiters (similar to BuildConstraints)
//0.5 add support for different oses/services
func GetDestPlatforms(specifiedOses string, specifiedArches string) [][]string {
	destOses := strings.FieldsFunc(specifiedOses, func(r rune) bool { return r == ',' || r == ' ' })
	destArchs := strings.FieldsFunc(specifiedArches, func(r rune) bool { return r == ',' || r == ' ' })

	for _, o := range destOses {
		supported := false
		for _, supportedPlatformArr := range getSupportedPlatforms() {
			supportedOs := supportedPlatformArr[0]
			if o == supportedOs {
				supported = true
			}
		}
		if !supported {
			log.Printf("WARNING: Operating System '%s' is unsupported", o)
		}
	}
	for _, o := range destArchs {
		supported := false
		for _, supportedPlatformArr := range getSupportedPlatforms() {
			supportedArch := supportedPlatformArr[1]
			if o == supportedArch {
				supported = true
			}
		}
		if !supported {
			log.Printf("WARNING: Architecture '%s' is unsupported", o)
		}
	}
	if len(destOses) == 0 {
		destOses = []string{""}
	}
	if len(destArchs) == 0 {
		destArchs = []string{""}
	}
	var destPlatforms [][]string
	for _, supportedPlatformArr := range getSupportedPlatforms() {
		supportedOs := supportedPlatformArr[0]
		supportedArch := supportedPlatformArr[1]
		for _, destOs := range destOses {
			if destOs == "" || supportedOs == destOs {
				for _, destArch := range destArchs {
					if destArch == "" || supportedArch == destArch {
						destPlatforms = append(destPlatforms, supportedPlatformArr)
					}
				}
			}
		}
	}
	if len(destPlatforms) < 1 {
		log.Printf("WARNING: no valid platforms specified")
	}
	return destPlatforms
}
