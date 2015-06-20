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
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TarGz implementation of Archiver.
func TarGz(archiveFilename string, itemsToArchive []ArchiveItem) error {
	// file write
	fw, err := os.Create(archiveFilename)
	if err != nil {
		return err
	}
	defer fw.Close()

	// gzip write
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, item := range itemsToArchive {
		err = addItemToTarGz(item, tw)
		if err != nil {
			return err
		}
	}
	err = tw.Close()
	return err
}

// Write a single file to TarGz
func TarGzWrite(item ArchiveItem, tw *tar.Writer, fi os.FileInfo) (err error) {
	if item.FileSystemPath != "" {
		fr, err := os.Open(item.FileSystemPath)
		if err == nil {
			defer fr.Close()

			h := new(tar.Header)
			h.Name = item.ArchivePath
			h.Size = fi.Size()
			h.Mode = int64(fi.Mode())
			h.ModTime = fi.ModTime()

			err = tw.WriteHeader(h)

			if err == nil {
				_, err = io.Copy(tw, fr)
			}
		}
	} else {
		h := new(tar.Header)
		//backslash-only paths
		h.Name = strings.Replace(item.ArchivePath, "\\", "/", -1)
		h.Size = int64(len(item.Data))
		h.Mode = int64(0644) //? is this ok?
		h.ModTime = time.Now()
		err = tw.WriteHeader(h)
		if err == nil {
			_, err = tw.Write(item.Data)
		}
	}
	return err
}

func addItemToTarGz(item ArchiveItem, tw *tar.Writer) (err error) {
	if item.FileSystemPath != "" {
		fi, err := os.Stat(item.FileSystemPath)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			err = addDirectoryToTarGz(item, tw)
			return err
		}
		err = TarGzWrite(item, tw, fi)
	} else {
		err = TarGzWrite(item, tw, nil)
	}
	return err
}

func addDirectoryToTarGz(dirPath ArchiveItem, tw *tar.Writer) error {
	dir, err := os.Open(dirPath.FileSystemPath)
	if err == nil {
		defer dir.Close()
		fis, err := dir.Readdir(0)
		if err == nil {
			for _, fi := range fis {
				curPath := ArchiveItemFromFileSystem(filepath.Join(dirPath.FileSystemPath, fi.Name()), filepath.Join(dirPath.ArchivePath, fi.Name()))
				err = addItemToTarGz(curPath, tw)
				if err != nil {
					return err
				}
			}
		}
	}
	return err
}
