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
	"fmt"
	"github.com/laher/goxc/core"
	"github.com/laher/goxc/typeutils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	GOXC_CONFIG_VERSION = "0.6"
	GOXC_FILE_EXT       = ".goxc.json"
	GOXC_LOCAL_FILE_EXT = ".goxc.local.json"
)

var GOXC_CONFIG_SUPPORTED = []string{"0.5.0", "0.6"}

//0.6 REMOVED JsonSettings struct.

//Loads a config file and merges results with any 'override' files.
func LoadJsonConfigOverrideable(dir string, configName string, useLocal bool, verbose bool) (Settings, error) {
	jsonFile := filepath.Join(dir, configName+GOXC_FILE_EXT)
	jsonLocalFile := filepath.Join(dir, configName+GOXC_LOCAL_FILE_EXT)
	var err error
	if useLocal {
		localSettings, err := loadJsonFile(jsonLocalFile, verbose)
		if err != nil {
			//load non-local file only.
			settings, err := loadJsonFile(jsonFile, verbose)
			return settings, err
		} else {
			settings, err := loadJsonFile(jsonFile, verbose)
			if err != nil {
				if os.IsNotExist(err) {
					return localSettings, nil
				} else {
					//parse error. Stop right there.
					return settings, err
				}
			} else {
				return Merge(localSettings, settings), nil
			}
		}
	} else {
		//load non-local file only.
		settings, err := loadJsonFile(jsonFile, verbose)
		return settings, err
	}
	//unreachable but required by go compiler
	//return localSettings.Settings, err
	return Settings{}, err
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

//beginnings of taskSettings merging
func getTaskSettings(rawJson []byte, fileName string) (map[string]interface{}, error) {
	var f interface{}
	err := json.Unmarshal(rawJson, &f)
	if err != nil {
		log.Printf("ERROR (%s): invalid json!", fileName)
		printErrorDetails(rawJson, err)
		return nil, err
	}
	m := f.(map[string]interface{})
	if s, keyExists := m["Settings"]; keyExists {
		settings := s.(map[string]interface{})
		if taskSettings, keyExists := settings["TaskSettings"]; keyExists {
			log.Printf("Found TaskSettings field %+v", taskSettings)
			return taskSettings.(map[string]interface{}), err
		} else {
			log.Printf("No TaskSettings field")
		}
	}
	return nil, fmt.Errorf("No TaskSettings defined")
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
	if fv, keyExists := m["FormatVersion"]; keyExists {
		formatVersion := fv.(string)
		if -1 == typeutils.StringSlicePos(GOXC_CONFIG_SUPPORTED, formatVersion) {
			log.Printf("WARNING (%s): is an old config file. File version: %s. Current version %v", fileName, formatVersion, GOXC_CONFIG_VERSION)
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

// load json file. Glob for goxc
//0.5.6 parse from an interface{} instead of JsonSettings.
//More flexible & better error reporting possible.
func loadJsonFile(jsonFile string, verbose bool) (Settings, error) {
	var settings Settings
	rawJson, err := loadFile(jsonFile, verbose)
	if err != nil {
		return settings, err
	}
	errs := validateRawJson(rawJson, jsonFile)
	if errs != nil && len(errs) > 0 {
		return settings, errs[0]
	}
	//TODO: super-verbose option for logging file content? log.Printf("%s\n", string(file))
	var f interface{}
	err = json.Unmarshal(rawJson, &f)
	if err != nil {
		log.Printf("ERROR (%s): invalid json!", jsonFile)
		printErrorDetails(rawJson, err)
		return settings, err
	} else {
		//fill all the way down.
		m := f.(map[string]interface{})
		if fv, keyExists := m["FormatVersion"]; keyExists {
			if s, keyExists := m["Settings"]; keyExists {
				//Support for old versions, up until version 0.5.
				settingsSection, err := typeutils.ToMap(s, "Settings")
				if err != nil {
					return settings, err
				}
				if _, keyExists = settingsSection["FormatVersion"]; !keyExists {
					//set from jsonSettings
					formatVersion, err := typeutils.ToString(fv, "FormatVersion")
					if err != nil {
						return settings, err
					}
					settingsSection["FormatVersion"] = formatVersion
				}
				settings, err := loadSettingsSection(settingsSection)
				return settings, err
			} else {
				//0.6 has no '{ Settings {} }' level. Just use the top level.
				return loadSettingsSection(m)
			}

		} else {
			return settings, errors.New("File format version not specified!")
		}
		//settings.Settings.TaskSettings, err= getTaskSettings(rawJson, jsonFile)
		if verbose {
			log.Printf("unmarshalled settings OK")
		}
	}
	//TODO: verbosity here? log.Printf("Results: %v", settings)
	return settings, nil
}

func loadSettingsSection(settingsSection map[string]interface{}) (settings Settings, err error) {
	settings = Settings{Resources: Resources{}}
	for k, v := range settingsSection {
		//try to match key
		switch k {
		case "Tasks":
			settings.Tasks, err = typeutils.ToStringSlice(v, k)
		case "TasksAppend":
			settings.TasksAppend, err = typeutils.ToStringSlice(v, k)
		case "ArtifactsDest":
			settings.ArtifactsDest, err = typeutils.ToString(v, k)
		case "Arch":
			settings.Arch, err = typeutils.ToString(v, k)
		case "Os":
			settings.Os, err = typeutils.ToString(v, k)
		case "BuildConstraints":
			settings.BuildConstraints, err = typeutils.ToString(v, k)
		case "Resources":
			for k2, v2 := range v.(map[string]interface{}) {
				switch k2 {
				case "Include":
					settings.Resources.Include, err = typeutils.ToString(v2, k+":"+k2)
				case "Exclude":
					settings.Resources.Exclude, err = typeutils.ToString(v2, k+":"+k2)
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
	if isLocal {
		jsonFile := filepath.Join(dir, configName+GOXC_LOCAL_FILE_EXT)
		return writeJsonFile(settings, jsonFile)
	}
	jsonFile := filepath.Join(dir, configName+GOXC_FILE_EXT)
	return writeJsonFile(settings, jsonFile)
}

func writeJsonFile(settings Settings, jsonFile string) error {
	data, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		log.Printf("Could NOT marshal json")
		return err
	}
	//0.6 StripEmpties no longer required (use omitempty tag instead)

	log.Printf("Writing file %s", jsonFile)
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
	log.Printf("Settings: %+v", settings)
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
