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

/* INCOMPLETE!! */
//TODO: Complete sometime during 0.7.x
import (
	"bytes"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/archive"
	"github.com/laher/goxc/packaging/sdeb"
	"github.com/laher/goxc/platforms"
	"github.com/laher/goxc/typeutils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	//"strings"
	"text/template"
	"time"
)

const (
	TASK_PKG_SOURCE = "pkg-source"
)

//runs automatically
func init() {
	Register(Task{
		TASK_PKG_SOURCE,
		"Build a source package. Currently only supports 'source deb' format for Debian/Ubuntu Linux.",
		runTaskPkgSource,
		map[string]interface{}{"metadata": map[string]interface{}{"maintainer": "unknown"}, "metadata-deb": map[string]interface{}{"Depends": "", "Build-Depends": "debhelper (>=4.0.0), golang-go, gcc"}, "rmtemp": true}})
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
		err = debSourceBuild(tp)
		if err != nil {
			return
		}
	}
	//OK
	return
}

func debSourceBuild(tp TaskParams) (err error) {
	metadata := tp.Settings.GetTaskSettingMap(TASK_PKG_SOURCE, "metadata")
	metadataDeb := tp.Settings.GetTaskSettingMap(TASK_PKG_SOURCE, "metadata-deb")
	
	description := ""
	if desc, keyExists := metadata["description"]; keyExists {
		description, err = typeutils.ToString(desc, "description")
		if err != nil {
			return err
		}
	}
	maintainer := "?"
	if maint, keyExists := metadata["maintainer"]; keyExists {
		maintainer, err = typeutils.ToString(maint, "maintainer")
		if err != nil {
			return err
		}
	}
	version := tp.Settings.GetFullVersionName()
	arches := "any"
	items, err := sdeb.SdebGetSourcesAsArchiveItems(tp.WorkingDirectory, tp.AppName+"-"+tp.Settings.GetFullVersionName())
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	
	//build
	//1. generate orig.tar.gz
	//memcached_1.2.5.orig.tar.gz
	destDir := filepath.Join(tp.OutDestRoot, tp.Settings.GetFullVersionName(), "sdeb")
	err = os.MkdirAll(destDir, 0777)
	if err != nil {
		return err
	}
	//TODO add/exclude resources to /usr/share
	origTgzPath := filepath.Join(destDir, tp.AppName+"_"+version+".orig.tar.gz")
	err = archive.TarGz(origTgzPath, items)
	if err != nil {
		return err
	}
	log.Printf("Created %s", origTgzPath)

	//2. generate .debian.tar.gz (just containing debian/ directory)
	//generate debian/control
	templateVars := getTemplateVars(tp.AppName, maintainer, tp.Settings.GetFullVersionName(), arches, description, metadataDeb)
	controlData, err := getDebMetadataFileContent(sdeb.TEMPLATE_SOURCEDEB_CONTROL, templateVars)
	if err != nil {
		return err
	}
	//generate debian/rules
	rulesData := []byte(sdeb.TEMPLATE_DEBIAN_RULES)
	sourceFormatData := []byte(sdeb.FILECONTENT_DEBIAN_SOURCE_FORMAT)
	//generate debian/changelog
	initialChangelogTemplate := sdeb.TEMPLATE_CHANGELOG_HEADER + "\n\n * Initial entry \n" + sdeb.TEMPLATE_CHANGELOG_FOOTER
	changelogData, err := getDebMetadataFileContent(initialChangelogTemplate, templateVars)
	if err != nil {
		return err
	}
	//generate debian/copyright
	copyrightData := []byte{}
	//generate debian/README.Debian
	//TODO: try pulling in README.md etc
	readmeData := []byte{}
	debTgzPath := filepath.Join(destDir, tp.AppName+"_"+version+".debian.tar.gz")
	err = archive.TarGz(debTgzPath,
		[]archive.ArchiveItem{
			archive.ArchiveItemFromBytes(changelogData, "debian/changelog"),
			archive.ArchiveItemFromBytes(copyrightData, "debian/copyright"),
			archive.ArchiveItemFromBytes(controlData, "debian/control"),
			archive.ArchiveItemFromBytes(readmeData, "debian/README.Debian"),
			archive.ArchiveItemFromBytes(rulesData, "debian/rules"),
			archive.ArchiveItemFromBytes(sourceFormatData, "debian/source/format")})
	if err != nil {
		return err
	}
	log.Printf("Created %s", debTgzPath)

	//3. generate .dsc file
	dscPath := filepath.Join(destDir, tp.AppName+"_"+version+".dsc")
	err = ioutil.WriteFile(dscPath, controlData, 0644)
	if err == nil {
		log.Printf("Wrote %s", dscPath)
	}
	return err
}

func getTemplateVars(appName, maintainer, version, arch, description string, metadataDeb map[string]interface{}) (interface{}) {
	vars := struct {
		PackageName string
		BuildDepends  string
		Priority string
		Maintainer string
		MaintainerEmail string
		StandardsVersion string
		Architecture string
		Section string
		Depends string
		Description string
		Other string
		Status string
		EntryDate string
	}{
		appName,
		sdeb.BUILD_DEPENDS_DEFAULT,
		sdeb.PRIORITY_DEFAULT,
		maintainer,
		maintainer,
		sdeb.STANDARDS_VERSION_DEFAULT,
		sdeb.ARCHITECTURE_DEFAULT,
		sdeb.SECTION_DEFAULT,
		"",
		description,
		"",
		"unreleased",
		time.Now().Format("Mon, 2 Jan 2006 15:04:05 -0700"),
	}
	return vars
}

func getDebMetadataFileContent(templateStr string, vars interface{}) ([]byte, error) {

	tpl, err := template.New("control").Parse(templateStr)
	if err != nil {
		return nil, err
	}
	var dest bytes.Buffer
	
	err = tpl.Execute(&dest, vars)
	if err != nil {
		return nil, err
	}
	return dest.Bytes(), nil
}

/*
func getSourceDebControlFileContent(appName, maintainer, version, arch, description string, metadataDeb map[string]interface{}) []byte {
	control := fmt.Sprintf("Source: %s\nPriority: optional\n", appName)
	if maintainer != "" {
		control = fmt.Sprintf("%sMaintainer: %s\n", control, maintainer)
	}
	//mandatory
	control = fmt.Sprintf("%sStandards-Version: %s\n", control, version)

	control = fmt.Sprintf("%s\nPackage: %s\nArchitecture: any\n", control, appName)
	control = fmt.Sprintf("%sArchitecture: %s\n", control, arch)
	//must include Depends and Build-Depends
	for k, v := range metadataDeb {
		control = fmt.Sprintf("%s%s: %s\n", control, k, v)
	}
	control = fmt.Sprintf("%sDescription: %s\n", control, description)
	return []byte(control)
}
*/
