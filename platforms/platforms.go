// Support for different target platforms (Operating Systems and Architectures) supported by the Go compiler
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
	"log"
	"runtime"
	"strings"
)

const (
	AMD64     = "amd64"
	AMD64P32  = "amd64p32"
	X86       = "386"
	ARM       = "arm"
	DARWIN    = "darwin"
	DRAGONFLY = "dragonfly"
	FREEBSD   = "freebsd"
	LINUX     = "linux"
	NACL      = "nacl"
	NETBSD    = "netbsd"
	OPENBSD   = "openbsd"
	PLAN9     = "plan9"
	SOLARIS   = "solaris"
	WINDOWS   = "windows"
)

// represents a target compilation platform
type Platform struct {
	Os   string
	Arch string
}

var (
	OSES                    = []string{DARWIN, LINUX, FREEBSD, NETBSD, OPENBSD, PLAN9, WINDOWS, SOLARIS, DRAGONFLY, NACL}
	ARCHS                   = []string{X86, AMD64, ARM}
	SUPPORTED_PLATFORMS_1_0 = []Platform{
		Platform{DARWIN, X86},
		Platform{DARWIN, AMD64},
		Platform{LINUX, X86},
		Platform{LINUX, AMD64},
		Platform{LINUX, ARM},
		Platform{FREEBSD, X86},
		Platform{FREEBSD, AMD64},
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
	NEW_PLATFORMS_1_3 = []Platform{
		Platform{DRAGONFLY, X86},
		Platform{DRAGONFLY, AMD64},
		Platform{NACL, X86},
		Platform{NACL, AMD64P32},
		Platform{SOLARIS, AMD64}}
	NEW_PLATFORMS_1_4 = []Platform{
		Platform{NACL, ARM},
	}

	SUPPORTED_PLATFORMS_1_1 = append(append([]Platform{}, SUPPORTED_PLATFORMS_1_0...), NEW_PLATFORMS_1_1...)
	SUPPORTED_PLATFORMS_1_3 = append(SUPPORTED_PLATFORMS_1_1, NEW_PLATFORMS_1_3...)
	SUPPORTED_PLATFORMS_1_4 = append(SUPPORTED_PLATFORMS_1_3, NEW_PLATFORMS_1_4...)
)

func getSupportedPlatforms() []Platform {
	if strings.HasPrefix(runtime.Version(), "go1.4") {
		return SUPPORTED_PLATFORMS_1_4
	}
	if strings.HasPrefix(runtime.Version(), "go1.3") {
		return SUPPORTED_PLATFORMS_1_3
	}
	// otherwise default to <= go1.2
	return SUPPORTED_PLATFORMS_1_1
}

func ContainsPlatform(haystack []Platform, needle Platform) bool {
	for _, p := range haystack {
		if p.Os == needle.Os && p.Arch == needle.Arch {
			return true
		}
	}
	return false
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
