// GOXC IS NOT READY FOR USE AS AN API - function names and packages will continue to change until version 1.0
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
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
)

// function definition
type Archiver func(archiveFilename string, items [][]string) error

func ArchiveBinaryAndResources(outDir, binPath, appName string, resources []string, settings config.Settings, archiver Archiver, ending string) (zipFilename string, err error) {
	if settings.PackageVersion != "" && settings.PackageVersion != core.PACKAGE_VERSION_DEFAULT {
		// v0.1.6 using appname_version_platform. See issue 3
		zipFilename = appName + "_" + settings.GetFullVersionName() + "_" + filepath.Base(filepath.Dir(binPath)) + "." + ending
	} else {
		zipFilename = appName + "_" + filepath.Base(filepath.Dir(binPath)) + "." + ending
	}
	toArchive := [][]string{[]string{binPath, filepath.Base(binPath)}}
	for _, resource := range resources {
		toArchive = append(toArchive, []string{resource, resource})
	}
	err = archiver(filepath.Join(outDir, zipFilename), toArchive)
	return
}
