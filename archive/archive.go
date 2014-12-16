// archive features for goxc. Limited support for zip, tar.gz and ar archiving
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
	"path/filepath"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
)

// type definition representing a file to be archived. Details location on filesystem and destination filename inside archive.
type ArchiveItem struct {
	//if FileSystemPath is empty, use Data instead
	FileSystemPath string
	ArchivePath    string
	Data           []byte
}

func ArchiveItemFromFileSystem(fileSystemPath, archivePath string) ArchiveItem {
	return ArchiveItem{fileSystemPath, archivePath, nil}
}

func ArchiveItemFromBytes(data []byte, archivePath string) ArchiveItem {
	return ArchiveItem{"", archivePath, data}
}

// type definition for different archiving implementations
type Archiver func(archiveFilename string, itemsToArchive []ArchiveItem) error

// goxc function to archive a binary along with supporting files (e.g. README or LICENCE).
func ArchiveBinariesAndResources(outDir, platName string, binPaths []string, appName string, resources []string, settings config.Settings, archiver Archiver, ending string, includeTopLevelDir bool) (zipFilename string, err error) {
	var zipName string
	if settings.PackageVersion != "" && settings.PackageVersion != core.PACKAGE_VERSION_DEFAULT {
		//0.1.6 using appname_version_platform. See issue 3
		zipName = appName + "_" + settings.GetFullVersionName() + "_" + platName
	} else {
		zipName = appName + "_" + platName
	}
	zipFilename = filepath.Join(outDir, zipName+"."+ending)
	var zipDir string
	if includeTopLevelDir {
		zipDir = zipName
	} else {
		zipDir = ""
	}
	toArchive := []ArchiveItem{}
	for _, binPath := range binPaths {
		destFile := filepath.Base(binPath)
		if zipDir != "" {
			destFile = filepath.Join(zipDir, destFile)
		}
		toArchive = append(toArchive, ArchiveItemFromFileSystem(binPath, destFile))
	}
	for _, resource := range resources {
		destFile := resource
		if zipDir != "" {
			destFile = filepath.Join(zipDir, destFile)
		}
		toArchive = append(toArchive, ArchiveItemFromFileSystem(resource, destFile))
	}
	err = archiver(zipFilename, toArchive)
	return
}
