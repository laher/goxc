package typeutils

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

// deletes a given index
func StringSliceDelIndex(a []string, i int) []string {
	a = append(a[:i], a[i+1:]...)
	// OR  a = a[:i+copy(a[i:], a[i+1:])]
	return a
}

// deletes first occurence of a given string
func StringSliceDel(slice []string, value string) []string {
	p := StringSlicePos(slice, value)
	if p > -1 {
		return StringSliceDelIndex(slice, p)
	}
	return slice
}

// deletes all occurences of a given string
func StringSliceDelAll(slice []string, value string) []string {
	p := StringSlicePos(slice, value)
	for p > -1 {
		slice = StringSliceDelIndex(slice, p)
		p = StringSlicePos(slice, value)
	}
	return slice
}

// returns first index of a given string
func StringSlicePos(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

// returns true if a slice contains given string
func StringSliceContains(slice []string, value string) bool {
	return StringSlicePos(slice, value) > -1
}

func StringSliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, c := range a {
		if c != b[i] {
			return false
		}
	}
	return true
}

//this was taken from golang `bytes.Compare`.
func StringSliceCompare(a, b []string) int {
	m := len(a)
	if m > len(b) {
		m = len(b)
	}
	for i, ac := range a[0:m] {
		bc := b[i]
		switch {
		case ac > bc:
			return 1
		case ac < bc:
			return -1
		}
	}
	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	}
	return 0
}
