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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const FORMAT_VERSION = "0.5.0"

type JsonSettings struct {
	Settings Settings
	//Format for goxc.json files
	FormatVersion string
	//TODO??: InheritFiles []string
}

func WrapJsonSettings(settings Settings) JsonSettings {
	return JsonSettings{Settings: settings, FormatVersion: FORMAT_VERSION}
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

//beginnings of taskSettings merging
func getTaskSettings(rawJson []byte, fileName string) (map[string]interface{}, error) {
	var f interface{}
	err := json.Unmarshal(rawJson, &f)
	if err != nil {
		log.Printf("ERROR (%s): invalid json!", fileName)
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
		return []error{err}
	}
	errs := []error{}
	m := f.(map[string]interface{})
	rejectOldTaskDefinitions := false
	if formatVersion, keyExists := m["FormatVersion"]; keyExists {
		if formatVersion != FORMAT_VERSION {
			log.Printf("WARNING (%s): is an old config file. File version: %s. Expected version %s", fileName, formatVersion, FORMAT_VERSION)
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
			if _, keyExists := settings["Tasks"]; keyExists {
				msg := "task definitions have changed in version 0.5.0. Please refer to latest docs and update your config file to version 0.5.0 accordingly."
				log.Printf("ERROR (%s): %s", fileName, msg)
				errs = append(errs, errors.New(msg))
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
func loadJsonFile(jsonFile string, verbose bool) (JsonSettings, error) {
	var settings JsonSettings
	rawJson, err := loadFile(jsonFile, verbose)
	if err != nil {
		return settings, err
	}
	errs := validateRawJson(rawJson, jsonFile)
	if errs != nil && len(errs) > 0 {
		return settings, errs[0]
	}

	//TODO: super-verbose option for logging file content? log.Printf("%s\n", string(file))
	json.Unmarshal(rawJson, &settings)
	if err != nil {
		log.Printf("Unmarshal error: %s", err)
		return settings, err
	} else {
		//settings.Settings.TaskSettings, err= getTaskSettings(rawJson, jsonFile)
		if verbose {
			log.Printf("unmarshalled settings OK")
		}
	}
	//TODO: verbosity here? log.Printf("Results: %v", settings)
	return settings, nil
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
	stripped, err := StripEmpties(data, settings.Settings.IsVerbose())
	if err == nil {
		data = stripped
	} else {
		log.Printf("Error stripping empty config keys - %s", err)
	}
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
