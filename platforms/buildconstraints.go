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
	"strings"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/typeutils"
)

// parse and filter list of platforms
func ApplyBuildConstraints(buildConstraints string, unfilteredPlatforms []Platform) []Platform {
	ret := []Platform{}
	items := strings.FieldsFunc(buildConstraints, func(r rune) bool { return r == ' ' })
	if len(items) == 0 {
		return unfilteredPlatforms
	}
	for _, item := range items {
		parts := strings.FieldsFunc(item, func(r rune) bool { return r == ',' })
		itemOs := []string{}
		itemNegOs := []string{}
		itemArch := []string{}
		itemNegArch := []string{}
		for _, part := range parts {
			isNeg, modulus := isNegative(part)
			if IsOs(modulus) {
				if isNeg {
					itemNegOs = append(itemNegOs, modulus)
				} else {
					itemOs = append(itemOs, modulus)
				}
			} else if IsArch(modulus) {
				if isNeg {
					itemNegArch = append(itemNegArch, modulus)
				} else {
					itemArch = append(itemArch, modulus)
				}

			} else {
				log.Printf("Unrecognised build constraint! Ignoring '%s'", part)
			}
		}
		ret = append(ret, resolveItem(itemOs, itemNegOs, itemArch, itemNegArch, unfilteredPlatforms)...)
	}
	return ret
}

// check if a string is a valid architecture name
func IsArch(part string) bool {
	return typeutils.StringSlicePos(ARCHS, part) > -1
}

// check if a string is a valid OS name
func IsOs(part string) bool {
	return typeutils.StringSlicePos(OSES, part) > -1
}

func isNegative(part string) (bool, string) {
	isNeg := strings.HasPrefix(part, "!")
	if isNeg {
		return true, part[1:]
	}
	return false, part
}

func resolveItem(itemOses, itemNegOses, itemArchs, itemNegArchs []string, unfilteredPlatforms []Platform) []Platform {
	ret := []Platform{}
	if len(itemOses) == 0 {
		//none specified: add all
		itemOses = getOses(unfilteredPlatforms)
	}
	for _, itemNegOs := range itemNegOses {
		//log.Printf("negos " + itemNegOs)
		itemOses = typeutils.StringSliceDelAll(itemOses, itemNegOs)
	}

	//log.Printf("oses %v", itemOses)
	for _, itemOs := range itemOses {
		itemArchsThisOs := make([]string, len(itemArchs))
		copy(itemArchsThisOs, itemArchs)
		if len(itemArchs) == 0 {
			//none specified: add all
			itemArchsThisOs = getArchsForOs(unfilteredPlatforms, itemOs)
		}
		for _, itemNegArch := range itemNegArchs {
			itemArchsThisOs = typeutils.StringSliceDelAll(itemArchsThisOs, itemNegArch)
		}
		for _, itemArch := range itemArchsThisOs {
			ret = append(ret, Platform{itemOs, itemArch})
		}
	}
	return ret
}

func getArchsForOs(sp []Platform, os string) []string {
	archs := []string{}
	for _, p := range sp {
		if p.Os == os {
			archs = append(archs, p.Arch)
		}
	}
	return archs
}
func getOses(sp []Platform) []string {
	oses := []string{}
	for _, p := range sp {
		if !typeutils.StringSliceContains(oses, p.Os) {
			oses = append(oses, p.Os)
		}
	}
	return oses
}
