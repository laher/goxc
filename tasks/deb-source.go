package tasks

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

/* Nearing completion */
//TODO: various refinements ...
import (
	/*
		"bytes"
		"crypto/md5"
		"crypto/sha1"
		"crypto/sha256"
		"encoding/hex"
		//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
		//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
		"github.com/laher/goxc/archive"
		//"github.com/laher/goxc/packaging/sdeb"
	*/
	"github.com/debber/debber-v0.3/deb"
	"github.com/debber/debber-v0.3/debgen"
	"github.com/laher/goxc/platforms"
	"github.com/laher/goxc/typeutils"
	//	"io"
	//	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	//"strings"
	//	"text/template"
	//	"time"
	"fmt"
)

/*
type Checksum struct {
	Checksum string
	Size     int64
	File     string
}
type TemplateVars struct {
	PackageName      string
	PackageVersion   string
	BuildDepends     string
	Priority         string
	Maintainer       string
	MaintainerEmail  string
	StandardsVersion string
	Architecture     string
	Section          string
	Depends          string
	Description      string
	Other            string
	Status           string
	EntryDate        string
	Format           string
	Files            []Checksum
	ChecksumsSha1    []Checksum
	ChecksumsSha256  []Checksum
}
*/

//runs automatically
func init() {
	Register(Task{
		TASK_DEB_SOURCE,
		"Build a source package. Currently only supports 'source deb' format for Debian/Ubuntu Linux.",
		runTaskPkgSource,
		map[string]interface{}{
			"metadata": map[string]interface{}{
				"maintainer":      "unknown",
				"maintainerEmail": "unknown@example.com",
			},
			"metadata-deb": map[string]interface{}{
				"Depends":       "",
				"Build-Depends": "debhelper (>=4.0.0), golang-go, gcc",
			},
			"rmtemp":              true,
			"go-sources-dir":      ".",
			"other-mappped-files": map[string]interface{}{},
		}})
}

func runTaskPkgSource(tp TaskParams) (err error) {
	var makeSourceDeb bool
	for _, dest := range tp.DestPlatforms {
		if dest.Os == platforms.LINUX {
			makeSourceDeb = true
		}
	}
	//TODO rpm
	if makeSourceDeb {
		if tp.Settings.IsVerbose() {
			log.Printf("Building 'source deb' for Ubuntu/Debian Linux.")
		}
		//	log.Printf("WARNING: 'source deb' functionality requires more documentation and config options to make it properly useful. More coming soon.")
		err = debSourceBuild(tp)
		if err != nil {
			return
		}
	} else {
		if !tp.Settings.IsQuiet() {
			log.Printf("Not building source debs because Linux has not been selected as a target OS")
		}
	}
	//OK
	return
}

/*
func checksums(path, name string) (*Checksum, *Checksum, *Checksum, error) {
	//checksums
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, err
	}

	hashMd5 := md5.New()
	size, err := io.Copy(hashMd5, f)
	if err != nil {
		return nil, nil, nil, err
	}
	checksumMd5 := Checksum{hex.EncodeToString(hashMd5.Sum(nil)), size, name}

	f.Seek(int64(0), 0)
	hash256 := sha256.New()
	size, err = io.Copy(hash256, f)
	if err != nil {
		return nil, nil, nil, err
	}
	checksumSha256 := Checksum{hex.EncodeToString(hash256.Sum(nil)), size, name}

	f.Seek(int64(0), 0)
	hash1 := sha1.New()
	size, err = io.Copy(hash1, f)
	if err != nil {
		return nil, nil, nil, err
	}
	checksumSha1 := Checksum{hex.EncodeToString(hash1.Sum(nil)), size, name}

	err = f.Close()
	if err != nil {
		return nil, nil, nil, err
	}

	return &checksumMd5, &checksumSha1, &checksumSha256, nil

}
*/
func debSourceBuild(tp TaskParams) (err error) {
	metadata := tp.Settings.GetTaskSettingMap(TASK_DEB_SOURCE, "metadata")
	//armArchName := getArmArchName(tp.Settings)
	metadataDebX := tp.Settings.GetTaskSettingMap(TASK_DEB_SOURCE, "metadata-deb")
	otherMappedFilesFromSetting := tp.Settings.GetTaskSettingMap(TASK_DEB_GEN, "other-mapped-files")
	otherMappedFiles, err := calcOtherMappedFiles(otherMappedFilesFromSetting)
	if err != nil {
		return err
	}
	metadataDeb := map[string]string{}
	for k, v := range metadataDebX {
		val, ok := v.(string)
		if ok {
			metadataDeb[k] = val
		}
	}
	rmtemp := tp.Settings.GetTaskSettingBool(TASK_DEB_SOURCE, "rmtemp")
	debDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName()) //v0.8.1 dont use platform dir
	tmpDir := filepath.Join(debDir, ".goxc-temp")

	shortDescription := "?"
	if desc, keyExists := metadata["description"]; keyExists {
		var err error
		shortDescription, err = typeutils.ToString(desc, "description")
		if err != nil {
			return err
		}
	}
	longDescription := " "
	if ldesc, keyExists := metadata["long-description"]; keyExists {
		var err error
		longDescription, err = typeutils.ToString(ldesc, "long-description")
		if err != nil {
			return err
		}
	}
	maintainerName := "?"
	if maint, keyExists := metadata["maintainer"]; keyExists {
		var err error
		maintainerName, err = typeutils.ToString(maint, "maintainer")
		if err != nil {
			return err
		}
	}
	maintainerEmail := "example@example.org"
	if maintEmail, keyExists := metadata["maintainer-email"]; keyExists {
		var err error
		maintainerEmail, err = typeutils.ToString(maintEmail, "maintainer-email")
		if err != nil {
			return err
		}
	}

	build := debgen.NewBuildParams()
	build.DestDir = debDir
	build.TmpDir = tmpDir
	build.Init()
	build.IsRmtemp = rmtemp
	var ctrl *deb.Control
	//Read control data. If control file doesnt exist, use parameters ...
	fi, err := os.Open(filepath.Join(build.DebianDir, "control"))
	if os.IsNotExist(err) {
		log.Printf("WARNING - no debian 'control' file found. Use `debber` to generate proper debian metadata")
		ctrl = deb.NewControlDefault(tp.AppName, maintainerName, maintainerEmail, shortDescription, longDescription, false)
	} else if err != nil {
		return fmt.Errorf("%v", err)
	} else {
		cfr := deb.NewControlFileReader(fi)
		ctrl, err = cfr.Parse()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	goSourcesDir := tp.Settings.GetTaskSettingString(TASK_DEB_SOURCE, "go-sources-dir")
	mappedSources, err := debgen.GlobForGoSources(goSourcesDir, []string{build.DestDir, build.TmpDir})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	for k, v := range mappedSources {
		otherMappedFiles[k] = v
	}
	build.Version = tp.Settings.GetFullVersionName()
	spgen, err := debgen.PrepareSourceDebGenerator(ctrl, build)
	if spgen.OrigFiles == nil {
		spgen.OrigFiles = map[string]string{}
	}
	for k, v := range otherMappedFiles {
		spgen.OrigFiles[k] = v
	}
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	err = spgen.GenerateAllDefault()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if tp.Settings.IsVerbose() {
		log.Printf("Wrote dsc file to %s", filepath.Join(build.DestDir, spgen.SourcePackage.DscFileName))
		log.Printf("Wrote orig file to %s", filepath.Join(build.DestDir, spgen.SourcePackage.OrigFileName))
		log.Printf("Wrote debian file to %s", filepath.Join(build.DestDir, spgen.SourcePackage.DebianFileName))
	}
	return nil
}
