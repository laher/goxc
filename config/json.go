package config

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
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/laher/goxc/core"
	"github.com/laher/goxc/typeutils"
)

const (
	GOXC_CONFIG_VERSION = "0.9"
)

var GOXC_CONFIG_SUPPORTED = []string{"0.5.0", "0.6", "0.8", "0.9"}

//0.6 REMOVED JsonSettings struct.

//Loads a config file and merges results with any 'override' files.
//0.8 using new inheritance rules. More flexibility, particular for people wanting different rules for different platforms
//0.10.x adding parameter isWriteLocal
func LoadJsonConfigOverrideable(dir string, configName string, isRead, isWriteLocal, verbose bool) (Settings, error) {
	var configs []string
	if isRead {
		configs = []string{configName + core.GOXC_LOCAL_FILE_EXT, configName + core.GOXC_FILE_EXT,
			core.GOXC_CONFIGNAME_BASE + core.GOXC_LOCAL_FILE_EXT,
			core.GOXC_CONFIGNAME_BASE + core.GOXC_FILE_EXT}
	} else {
		if isWriteLocal {
			configs = []string{configName + core.GOXC_LOCAL_FILE_EXT}
		} else {
			configs = []string{configName + core.GOXC_FILE_EXT}
		}
	}
	return LoadJsonConfigs(dir, configs, verbose)
}

func LoadJsonConfigs(dir string, configs []string, verbose bool) (Settings, error) {
	//most-important first
	mergedSettingsMap := map[string]interface{}{}
	for _, jsonFile := range configs {
		settingsMap, err := loadJsonFileAsMap(jsonFile, verbose)
		if err != nil {
			if os.IsNotExist(err) {
				//continue onto next file
			} else {
				//parse error. Stop right there.
				return Settings{}, err
			}
		} else {
			mergedSettingsMap = typeutils.MergeMaps(mergedSettingsMap, settingsMap)
		}
	}
	return loadSettingsSection(mergedSettingsMap)
}

//0.5.6 provide more detail about errors (syntax errors for now)
func printErrorDetails(rawJson []byte, err error) {
	switch typedErr := err.(type) {
	case *json.SyntaxError:
		lineNumber := 1
		colNumber := 0
		for i, b := range rawJson {
			if int64(i) == typedErr.Offset {
				log.Printf("JSON syntax error on line %d, column %d", lineNumber, colNumber)
				return
			}
			if b == '\n' {
				lineNumber = lineNumber + 1
				colNumber = 0
			} else {
				colNumber = colNumber + 1
			}
		}
	}
}

func validateRawJson(rawJson []byte, fileName string) []error {
	var f interface{}
	err := json.Unmarshal(rawJson, &f)
	if err != nil {
		log.Printf("ERROR (%s): invalid json!", fileName)
		printErrorDetails(rawJson, err)
		return []error{err}
	}
	var errs []error
	m := f.(map[string]interface{})
	rejectOldTaskDefinitions := false
	//it was known as 'FormatVersion' until v0.9
	if fv, keyExists := m["FormatVersion"]; keyExists {
		m["ConfigVersion"] = fv
	}
	if fv, keyExists := m["ConfigVersion"]; keyExists {
		configVersion := fv.(string)
		if -1 == typeutils.StringSlicePos(GOXC_CONFIG_SUPPORTED, configVersion) {
			log.Printf("WARNING (%s): is an old config file. File version: %s. Current version %v", fileName, configVersion, GOXC_CONFIG_VERSION)
			rejectOldTaskDefinitions = true
		}
	} else {
		log.Printf("WARNING (%s): format version not specified. Please ensure this file format is up to date.", fileName)
		rejectOldTaskDefinitions = true
	}
	if s, keyExists := m["Settings"]; keyExists {
		settings := s.(map[string]interface{})
		errs = validateSettingsSection(settings, fileName, rejectOldTaskDefinitions)
	} else {
		errs = validateSettingsSection(m, fileName, rejectOldTaskDefinitions)
	}
	return errs
}

func validateSettingsSection(settings map[string]interface{}, fileName string, rejectOldTaskDefinitions bool) (errs []error) {
	if _, keyExists := settings["ArtifactTypes"]; keyExists {
		msg := "'ArtifactTypes' setting is deprecated. Please use tasks instead (By default goxc zips the binary ('archive' task) and then deletes the binary ('rmbin' task)."
		log.Printf("ERROR (%s): %s", fileName, msg)
		errs = append(errs, errors.New(msg))
	}
	if _, keyExists := settings["Codesign"]; keyExists {
		msg := "'Codesign' setting is deprecated. Please use setting \"Settings\" : { \"TaskSettings\" : { \"codesign\" : { \"id\" : \"blah\" } } }."
		log.Printf("ERROR (%s): %s", fileName, msg)
		errs = append(errs, errors.New(msg))
	}
	if rejectOldTaskDefinitions {
		if val, keyExists := settings["Tasks"]; keyExists {
			valArr := val.([]interface{})
			if len(valArr) == 1 && valArr[0] == core.TASK_BUILD_TOOLCHAIN {
				//build-toolchain hasn't changed. Continue.
			} else {
				msg := "task definitions have changed in version 0.5.0. Please refer to latest docs and update your config file to version 0.5.0 accordingly."
				log.Printf("ERROR (%s): %s", fileName, msg)
				errs = append(errs, errors.New(msg))
			}
		}
	}
	return errs
}

func loadFile(jsonFile string, verbose bool) ([]byte, error) {
	file, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		if os.IsNotExist(err) {
			if verbose { //not a problem.
				log.Printf("%s not found", jsonFile)
			}
		} else {
			//always log because it's unexpected
			log.Printf("File error: %v", err)
		}
	} else {
		if verbose {
			log.Printf("Found %s", jsonFile)
		}
	}
	return file, err
}

func loadJsonFileAsMap(jsonFile string, verbose bool) (map[string]interface{}, error) {
	var f map[string]interface{}
	rawJson, err := loadFile(jsonFile, verbose)
	if err != nil {
		return f, err
	}
	errs := validateRawJson(rawJson, jsonFile)
	if errs != nil && len(errs) > 0 {
		return f, errs[0]
	}
	//TODO: super-verbose option for logging file content? log.Printf("%s\n", string(file))
	err = json.Unmarshal(rawJson, &f)
	if err != nil {
		log.Printf("ERROR (%s): invalid json!", jsonFile)
		printErrorDetails(rawJson, err)
	} else {
		//it was known as FormatVersion until v0.9
		if fv, keyExists := f["FormatVersion"]; keyExists {
			f["ConfigVersion"] = fv
		}
		//fill all the way down.
		if fv, keyExists := f["ConfigVersion"]; keyExists {
			if s, keyExists := f["Settings"]; keyExists {
				//Support for old versions, up until version 0.5.
				settingsSection, err := typeutils.ToMap(s, "Settings")
				if err != nil {
					return f, err
				} else {
					if _, keyExists = settingsSection["ConfigVersion"]; !keyExists {
						//set from jsonSettings
						//it was previously known as FormatVersion ...
						formatVersion, err := typeutils.ToString(fv, "FormatVersion")
						if err != nil {
							return f, err
						}
						settingsSection["ConfigVersion"] = formatVersion
					}
					return settingsSection, err
				}
				//settings, err := loadSettingsSection(settingsSection)
				//return settings, err
			}
		} else {
			return f, errors.New("File format version not specified!")
		}
	}

	return f, err
}

// load json file. Glob for goxc
//0.5.6 parse from an interface{} instead of JsonSettings.
//More flexible & better error reporting possible.
// 0.6.4 deprecated in favour of loadJsonFileAsMap
func loadJsonFile(jsonFile string, verbose bool) (Settings, error) {
	m, err := loadJsonFileAsMap(jsonFile, verbose)
	if err != nil {
		return Settings{}, err
	}
	return loadSettingsSection(m)
}

func loadSettingsSection(settingsSection map[string]interface{}) (settings Settings, err error) {
	settings = Settings{}
	for k, v := range settingsSection {
		//try to match key
		switch k {
		case "Tasks":
			settings.Tasks, err = typeutils.ToStringSlice(v, k)
		case "TasksExclude":
			settings.TasksExclude, err = typeutils.ToStringSlice(v, k)
		case "TasksAppend":
			settings.TasksAppend, err = typeutils.ToStringSlice(v, k)
		case "TasksPrepend":
			settings.TasksPrepend, err = typeutils.ToStringSlice(v, k)
		case "AppName":
			settings.AppName, err = typeutils.ToString(v, k)
		case "ArtifactsDest":
			settings.ArtifactsDest, err = typeutils.ToString(v, k)
		case "OutPath":
			settings.OutPath, err = typeutils.ToString(v, k)
		case "Arch":
			settings.Arch, err = typeutils.ToString(v, k)
		case "Os":
			settings.Os, err = typeutils.ToString(v, k)
		case "BuildConstraints":
			settings.BuildConstraints, err = typeutils.ToString(v, k)
		case "ResourcesInclude":
			settings.ResourcesInclude, err = typeutils.ToString(v, k)
		case "ResourcesExclude":
			settings.ResourcesExclude, err = typeutils.ToString(v, k)
		case "MainDirsExclude":
			settings.MainDirsExclude, err = typeutils.ToString(v, k)
		//deprecated
		case "Resources":
			for k2, v2 := range v.(map[string]interface{}) {
				switch k2 {
				case "Include":
					settings.ResourcesInclude, err = typeutils.ToString(v2, k+":"+k2)
				case "Exclude":
					settings.ResourcesExclude, err = typeutils.ToString(v2, k+":"+k2)
				}
			}
		case "PackageVersion":
			settings.PackageVersion, err = typeutils.ToString(v, k)
		case "BranchName":
			settings.BranchName, err = typeutils.ToString(v, k)
		case "PrereleaseInfo":
			settings.PrereleaseInfo, err = typeutils.ToString(v, k)
		case "BuildName":
			settings.BuildName, err = typeutils.ToString(v, k)
		case "Verbosity":
			settings.Verbosity, err = typeutils.ToString(v, k)
		case "TaskSettings":
			settings.TaskSettings, err = typeutils.ToMapStringMapStringInterface(v, k)
		case "FormatVersion":
			settings.GoxcConfigVersion, err = typeutils.ToString(v, k)
		case "ConfigVersion":
			settings.GoxcConfigVersion, err = typeutils.ToString(v, k)
		case "BuildSettings":
			var m map[string]interface{}
			m, err = typeutils.ToMap(v, k)
			if err == nil {
				settings.BuildSettings, err = buildSettingsFromMap(m)
				if err == nil {
					//log.Printf("Parsed build settings OK (%+v)", settings.BuildSettings)
				}
			}
		case "Env":
			settings.Env, err = typeutils.ToStringSlice(v, k)
		default:
			log.Printf("Warning!! Unrecognised Setting '%s' (value %v)", k, v)
		}
		if err != nil {
			return settings, err
		}
	}
	return settings, err
}

func WriteJsonConfig(dir string, settings Settings, configName string, isLocal bool) error {
	settings.GoxcConfigVersion = GOXC_CONFIG_VERSION
	bs := BuildSettings{}
	if settings.BuildSettings != nil && bs.Equals(*settings.BuildSettings) {
		settings.BuildSettings = nil
	}
	if isLocal {
		jsonFile := filepath.Join(dir, configName+core.GOXC_LOCAL_FILE_EXT)
		return writeJsonFile(settings, jsonFile)
	}
	jsonFile := filepath.Join(dir, configName+core.GOXC_FILE_EXT)
	return writeJsonFile(settings, jsonFile)
}

func writeJsonFile(settings Settings, jsonFile string) error {
	data, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		log.Printf("Could NOT marshal json")
		return err
	}
	//0.6 StripEmpties no longer required (use omitempty tag instead)

	if settings.IsVerbose() {
		log.Printf("Writing file %s", jsonFile)
	}
	return ioutil.WriteFile(jsonFile, data, 0644)
}

//use json from string
//0.6 DEPRECATED (unused)
func readJson(js []byte) (Settings, error) {
	var settings Settings
	err := json.Unmarshal(js, &settings)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	if settings.IsVerbose() {
		log.Printf("Settings: %+v", settings)
	}
	return settings, err
}

//0.6 DEPRECATED (unused)
func writeJson(m Settings) ([]byte, error) {
	return json.MarshalIndent(m, "", "\t")
}

//0.6 DEPRECATED (in favour of omitempty tag)
func StripEmpties(rawJson []byte, verbose bool) ([]byte, error) {
	var f interface{}
	err := json.Unmarshal(rawJson, &f)
	if err != nil {
		log.Printf("Warning: invalid json. returning")
		return rawJson, err
	}
	m := f.(map[string]interface{})
	ret := make(map[string]interface{})
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			if v != "" {
				ret[k] = vv
			} else {
				if verbose {
					log.Println("Stripping empty string", k)
				}
			}
		case []interface{}:
			if len(vv) > 0 {
				ret[k] = vv
			} else {
				if verbose {
					log.Println("Stripping empty array", k)
				}
			}
		case map[string]interface{}:
			bytes, err := json.Marshal(vv)
			if err != nil {
				log.Printf("Error marshalling inner map: %v", err)
				return rawJson, err
			}
			strippedInner, err := StripEmpties(bytes, verbose)
			if err != nil {
				log.Printf("Error stripping inner map: %v", err)
				return rawJson, err
			}
			var innerf interface{}
			err = json.Unmarshal(strippedInner, &innerf)
			if err != nil {
				log.Printf("Error unmarshalling inner map: %v", err)
				return rawJson, err
			}
			ret[k] = innerf
		case nil:
			if verbose {
				log.Println("Stripping null value", k)
			}
		default:
			ret[k] = vv
		}
	}
	return json.MarshalIndent(ret, "", "\t")
}
