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
	"archive/zip"
	"io"
	"os"
	"strings"
)

//TODO: folder support
func Zip(zipFilename string, itemsToArchive []ArchiveItem) error {
	zf, err := os.Create(zipFilename)
	if err != nil {
		return err
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	defer zw.Close()

	//resources
	for _, item := range itemsToArchive {
		err = addFileToZIP(zw, item)
		if err != nil {
			return err
		}
	}
	err = zw.Close()
	return err
}

func addFileToZIP(zw *zip.Writer, item ArchiveItem) (err error) {
	binfo, err := os.Stat(item.FileSystemPath)
	if err != nil {
		return
	}
	header, err := zip.FileInfoHeader(binfo)
	if err != nil {
		return
	}
	header.Method = zip.Deflate
	//always use forward slashes even on Windows
	header.Name = strings.Replace(item.ArchivePath, "\\", "/", -1)
	w, err := zw.CreateHeader(header)
	if err != nil {
		zw.Close()
		return
	}
	bf, err := os.Open(item.FileSystemPath)
	if err != nil {
		return
	}
	defer bf.Close()
	_, err = io.Copy(w, bf)
	return
}
