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

type Platform struct {
	Os   string
	Arch string
}

var (
	OSES                    = []string{DARWIN, LINUX, FREEBSD, NETBSD, OPENBSD, PLAN9, WINDOWS}
	ARCHS                   = []string{X86, AMD64, ARM}
	SUPPORTED_PLATFORMS_1_0 = []Platform{
		Platform{DARWIN, X86},
		Platform{DARWIN, AMD64},
		Platform{LINUX, X86},
		Platform{LINUX, AMD64},
		Platform{LINUX, ARM},
		Platform{FREEBSD, X86},
		Platform{FREEBSD, AMD64},
		// Platform{FREEBSD, ARM},
		// couldnt build toolchain for netbsd using a linux 386 host: 2013-02-19
		//	Platform{NETBSD, X86},
		//	Platform{NETBSD, AMD64},
		Platform{OPENBSD, X86},
		Platform{OPENBSD, AMD64},
		Platform{WINDOWS, X86},
		Platform{WINDOWS, AMD64}}
	NEW_PLATFORMS_1_1 = []Platform{
		Platform{FREEBSD, ARM},
		Platform{NETBSD, X86},
		Platform{NETBSD, AMD64},
		Platform{NETBSD, ARM},
		Platform{PLAN9, X86}}

	SUPPORTED_PLATFORMS_1_1 = append(append([]Platform{}, SUPPORTED_PLATFORMS_1_0...), NEW_PLATFORMS_1_1...)
)

func getSupportedPlatforms() []Platform {
	if strings.HasPrefix(runtime.Version(), "go1.0") {
		return SUPPORTED_PLATFORMS_1_0
	}
	return SUPPORTED_PLATFORMS_1_1
}



// interpret list of destination platforms (based on os & arch settings)
//0.5 add support for space delimiters (similar to BuildConstraints)
//0.5 add support for different oses/services
func GetDestPlatforms(specifiedOses string, specifiedArches string) []Platform {
	destOses := strings.FieldsFunc(specifiedOses, func(r rune) bool { return r == ',' || r == ' ' })
	destArchs := strings.FieldsFunc(specifiedArches, func(r rune) bool { return r == ',' || r == ' ' })

	for _, o := range destOses {
		supported := false
		for _, supportedPlatformArr := range getSupportedPlatforms() {
			supportedOs := supportedPlatformArr.Os
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
			supportedArch := supportedPlatformArr.Arch
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
	var destPlatforms []Platform
	for _, supportedPlatformArr := range getSupportedPlatforms() {
		supportedOs := supportedPlatformArr.Os
		supportedArch := supportedPlatformArr.Arch
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
