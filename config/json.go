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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const GOXC_CONFIG_VERSION = "0.5.0"

type JsonSettings struct {
	Settings Settings
	//Format for goxc.json files
	FormatVersion string
	//TODO??: InheritFiles []string
}

func WrapJsonSettings(settings Settings) JsonSettings {
	settings.GoxcConfigVersion = GOXC_CONFIG_VERSION
	return JsonSettings{Settings: settings, FormatVersion: GOXC_CONFIG_VERSION}
}

func LoadJsonConfigOverrideable(dir string, configName string, useLocal bool, verbose bool) (Settings, error) {
	jsonFile := filepath.Join(dir, configName+".json")
	jsonLocalFile := filepath.Join(dir, configName+".local.json")
	var err error
	if useLocal {
		localSettings, err := loadJsonFile(jsonLocalFile, verbose)
		if err != nil {
			//load non-local file only.
			settings, err := loadJsonFile(jsonFile, verbose)
			return settings.Settings, err
		} else {
			settings, err := loadJsonFile(jsonFile, verbose)
			if err != nil {
				if os.IsNotExist(err) {
					return localSettings.Settings, nil
				} else {
					//parse error. Stop right there.
					return settings.Settings, err
				}
			} else {
				return Merge(localSettings.Settings, settings.Settings), nil
			}
		}
	} else {
		//load non-local file only.
		settings, err := loadJsonFile(jsonFile, verbose)
		return settings.Settings, err
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
	errs := []error{}
	m := f.(map[string]interface{})
	rejectOldTaskDefinitions := false
	if formatVersion, keyExists := m["FormatVersion"]; keyExists {
		if formatVersion != GOXC_CONFIG_VERSION {
			log.Printf("WARNING (%s): is an old config file. File version: %s. Expected version %s", fileName, formatVersion, GOXC_CONFIG_VERSION)
			rejectOldTaskDefinitions = true
		}
	} else {
		log.Printf("WARNING (%s): format version not specified. Please ensure this file format is up to date.", fileName)
		rejectOldTaskDefinitions = true
	}
	if s, keyExists := m["Settings"]; keyExists {
		settings := s.(map[string]interface{})
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
	} else {
		log.Printf("No settings found. Ignoring file.")
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
//v0.5.6 parse from an interface{} instead of JsonSettings.
//More flexible & better error reporting possible.
func loadJsonFile(jsonFile string, verbose bool) (JsonSettings, error) {
	var jsonSettings JsonSettings
	rawJson, err := loadFile(jsonFile, verbose)
	if err != nil {
		return jsonSettings, err
	}
	errs := validateRawJson(rawJson, jsonFile)
	if errs != nil && len(errs) > 0 {
		return jsonSettings, errs[0]
	}
	jsonSettings = JsonSettings{}
	jsonSettings.Settings = Settings{}
	jsonSettings.Settings.Resources = Resources{}
	//TODO: super-verbose option for logging file content? log.Printf("%s\n", string(file))
	var f interface{}
	err = json.Unmarshal(rawJson, &f)
	if err != nil {
		log.Printf("ERROR (%s): invalid json!", jsonFile)
		printErrorDetails(rawJson, err)
		return jsonSettings, err
	} else {
		//fill all the way down.
		m := f.(map[string]interface{})
		if fv, keyExists := m["FormatVersion"]; keyExists {
			//OK
			jsonSettings.FormatVersion = fv.(string)
			if s, keyExists := m["Settings"]; keyExists {
				settingsSection := s.(map[string]interface{})
				for k, v := range settingsSection {
					//try to match key
					switch k {
					case "Tasks":
						jsonSettings.Settings.Tasks, err = fromJsonStringArray(v, k)
					case "TasksAppend":
						jsonSettings.Settings.TasksAppend, err = fromJsonStringArray(v, k)
						if err != nil {
							return jsonSettings, err
						}
					case "ArtifactsDest":
						jsonSettings.Settings.ArtifactsDest, err = fromJsonString(v, k)
					case "Arch":
						jsonSettings.Settings.Arch, err = fromJsonString(v, k)
					case "Os":
						jsonSettings.Settings.Os, err = fromJsonString(v, k)
					case "BuildConstraints":
						jsonSettings.Settings.BuildConstraints, err = fromJsonString(v, k)
					case "Resources":
						for k2, v2 := range v.(map[string]interface{}) {
							switch k2 {
							case "Include":
								jsonSettings.Settings.Resources.Include, err = fromJsonString(v2, k+":"+k2)
							case "Exclude":
								jsonSettings.Settings.Resources.Exclude, err = fromJsonString(v2, k+":"+k2)
							}
						}
					case "PackageVersion":
						log.Printf("Package version %s", v)
						jsonSettings.Settings.PackageVersion, err = fromJsonString(v, k)
					case "BranchName":
						jsonSettings.Settings.BranchName, err = fromJsonString(v, k)
					case "PrereleaseInfo":
						jsonSettings.Settings.PrereleaseInfo, err = fromJsonString(v, k)
					case "BuildName":
						jsonSettings.Settings.BuildName, err = fromJsonString(v, k)
					case "Verbosity":
						jsonSettings.Settings.Verbosity, err = fromJsonString(v, k)
					case "TaskSettings":
						jsonSettings.Settings.TaskSettings, err = fromJsonStringMap(v, k)
					default:
						log.Printf("Warning!! Unrecognised Setting '%s' (value %v)", k, v)
					}
					if err != nil {
						return jsonSettings, err
					}
				}
			}
		} else {
			return jsonSettings, errors.New("File format version not specified!")
		}
		//settings.Settings.TaskSettings, err= getTaskSettings(rawJson, jsonFile)
		if verbose {
			log.Printf("unmarshalled settings OK")
		}
	}
	//TODO: verbosity here? log.Printf("Results: %v", settings)
	return jsonSettings, nil
}
func fromJsonStringArray(v interface{}, k string) ([]string, error) {
	ret := []string{}
	switch typedV := v.(type) {
	case []interface{}:
		for _, i := range typedV {
			ret = append(ret, i.(string))
		}
		return ret, nil
	}
	return ret, fmt.Errorf("%s should be a json array, not a %T", k, v)
}
func fromJsonString(v interface{}, k string) (string, error) {
	switch typedV := v.(type) {
	case string:
		return typedV, nil
	}
	return "", fmt.Errorf("%s should be a json string, not a %T", k, v)
}
func fromJsonStringMap(v interface{}, k string) (map[string]interface{}, error) {
	switch typedV := v.(type) {
	case map[string]interface{}:
		return typedV, nil
	}
	return nil, fmt.Errorf("%s should be a json map, not a %T", k, v)
}
func WriteJsonConfig(dir string, settings JsonSettings, configName string, isLocal bool) error {
	if isLocal {
		jsonFile := filepath.Join(dir, configName+".local.json")
		return writeJsonFile(settings, jsonFile)
	}
	jsonFile := filepath.Join(dir, configName+".json")
	return writeJsonFile(settings, jsonFile)
}

func writeJsonFile(settings JsonSettings, jsonFile string) error {
	data, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		log.Printf("Could NOT marshal json")
		return err
	}
/* StripEmpties no longer required (use omitempty tag)
	stripped, err := StripEmpties(data, settings.Settings.IsVerbose())
	if err == nil {
		data = stripped
	} else {
		log.Printf("Error stripping empty config keys - %s", err)
	}
*/
	log.Printf("Writing file %s", jsonFile)
	return ioutil.WriteFile(jsonFile, data, 0644)
}

//use json from string
func readJson(js []byte) (JsonSettings, error) {
	var settings JsonSettings
	err := json.Unmarshal(js, &settings)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	log.Printf("Settings: %+v", settings)
	return settings, err
}

func writeJson(m JsonSettings) ([]byte, error) {
	return json.MarshalIndent(m, "", "\t")
}

//v0.5.9: DEPRECATED
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
