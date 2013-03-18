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
	"path/filepath"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
)


func MoveBinaryToZIP(outDir, binPath, appName string, resources []string, isRemoveBinary bool, settings config.Settings) (zipFilename string, err error) {
	if settings.PackageVersion != "" && settings.PackageVersion != config.PACKAGE_VERSION_DEFAULT {
		// v0.1.6 using appname_version_platform. See issue 3
		zipFilename = appName + "_" + settings.PackageVersion + "_" + filepath.Base(filepath.Dir(binPath)) + ".zip"
	} else {
		zipFilename = appName + "_" + filepath.Base(filepath.Dir(binPath)) + ".zip"
	}
	zf, err := os.Create(filepath.Join(outDir, zipFilename))
	if err != nil {
		return
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)

	addFileToZIP(zw, binPath)
	if err != nil {
		zw.Close()
		return
	}
	//resources
	for _, resource := range resources {
		addFileToZIP(zw, resource)
		if err != nil {
			zw.Close()
			return
		}
	}
	err = zw.Close()
	if err != nil {
		return
	}
	if isRemoveBinary {
		// Remove binary and its directory.
		err = os.Remove(binPath)
		if err != nil {
			return
		}
		err = os.Remove(filepath.Dir(binPath))
		if err != nil {
			return
		}
	}
	return
}

func addFileToZIP(zw *zip.Writer, path string) (err error) {
	binfo, err := os.Stat(path)
	if err != nil {
		return
	}
	header, err := zip.FileInfoHeader(binfo)
	if err != nil {
		return
	}
	header.Method = zip.Deflate
	w, err := zw.CreateHeader(header)
	if err != nil {
		zw.Close()
		return
	}
	bf, err := os.Open(path)
	if err != nil {
		return
	}
	defer bf.Close()
	_, err = io.Copy(w, bf)
	return
}

