package archive

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
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)
/* example ...
!<arch>
debian-binary   1282478016  0     0     100644  4         `
2.0
control.tar.gz  1282478016  0     0     100644  444       `
.....binary-data.....
*/

func ArForDeb(archiveFilename string, items [][]string) error {
	// open output file
	fo, err := os.Create(archiveFilename)
	if err != nil { panic(err) }
	// close fo on exit and check for its returned error
	defer func() {
		err := fo.Close()
		if err != nil {
	            log.Printf("Error closing output file", err)
		}
	}()
	header := "!<arch>\n"
	if _, err := fo.Write([]byte(header)); err != nil {
		log.Printf("Write error: %v", err)
		return err
        } else {
		for _, item := range items {
			fi, err := os.Open(item[0])
			if err != nil {
				return err
			} else {
				defer fi.Close()
				finf, err := fi.Stat()
				if err != nil {
					return err
				} else {
					fmodTimestamp := fmt.Sprintf("%d", finf.ModTime().Unix())
					//use root (for deb). These files are only for dpkg to extract as root anyway
					//this behaviour could be made configurable if 'Ar' gets used for anything else beyond .deb creation
					uid := "0"
					gid := "0"
					//Files only atm (not dirs)
					mode := "100644"
					size := fmt.Sprintf("%d", finf.Size())
					line := fmt.Sprintf("%s%s%s%s%s%s`\n", pad(item[1],16), pad(fmodTimestamp, 12), pad(gid, 6), pad(uid, 6), pad(mode, 8), pad(size, 10))
					if _, err := fo.Write([]byte(line)); err != nil {
						return err
					} else {
						copyFile(fi, fo)
						//0.5.4 bugfix: data section is 2-byte aligned.
						if finf.Size() % 2 == 1 {
							if _, err = fo.Write([]byte("\n")); err != nil {
								return err
							}
						}

					}
				}
			}
		}
	}
	return err
}

func copyFile(fi *os.File, fo *os.File) {
    // make a read buffer
    r := bufio.NewReader(fi)
    // make a write buffer
    w := bufio.NewWriter(fo)

    // make a buffer to keep chunks that are read
    buf := make([]byte, 1024)
    for {
        // read a chunk
        n, err := r.Read(buf)
        if err != nil && err != io.EOF { panic(err) }
        if n == 0 { break }

        // write a chunk
        if _, err := w.Write(buf[:n]); err != nil {
            panic(err)
        }
    }

    if err := w.Flush(); err != nil { panic(err) }
}

func pad(value string, length int) string {
	plen := length - len(value)
	if plen > 0 {
		return value + strings.Repeat(" ", plen)
	}
	return value
}
