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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const FORMAT_VERSION = "0.2.0"

type JsonSettings struct {
	Settings Settings
	//Format for goxc.json files
	FormatVersion string
	//TODO??: InheritFiles []string
}

func WrapJsonSettings(settings Settings) JsonSettings {
	return JsonSettings{Settings: settings, FormatVersion: FORMAT_VERSION}
}

func LoadJsonCascadingConfig(dir string, configName string, useLocal bool, verbose bool) (Settings, error) {
	jsonFile := filepath.Join(dir, configName+".json")
	jsonLocalFile := filepath.Join(dir, configName+".local.json")
	var err error
	if useLocal {
		localSettings, err := LoadJsonFile(jsonLocalFile, verbose)
		if err != nil {
			//load non-local file only.
			settings, err := LoadJsonFile(jsonFile, verbose)
			return settings.Settings, err
		} else {
			settings, err := LoadJsonFile(jsonFile, verbose)
			if err != nil {
				return localSettings.Settings, nil
			} else {
				return Merge(localSettings.Settings, settings.Settings), nil
			}
		}
	} else {
		//load non-local file only.
		settings, err := LoadJsonFile(jsonFile, verbose)
		return settings.Settings, err
	}
	//unreachable but required by go compiler
	//return localSettings.Settings, err
	return Settings{}, err
}

// load json file. Glob for goxc
func LoadJsonFile(jsonFile string, verbose bool) (JsonSettings, error) {
	var settings JsonSettings
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
		return settings, err
	} else {
		if verbose {
			log.Printf("Found %s : %+v", jsonFile, settings.Settings)
		}
	}
	//TODO: super-verbose option for logging file content? log.Printf("%s\n", string(file))
	json.Unmarshal(file, &settings)
	if err != nil {
		log.Printf("Unmarshal error: %s", err)
		return settings, err
	} else {
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
		return WriteJsonFile(settings, jsonFile)
	}
	jsonFile := filepath.Join(dir, configName+".json")
	return WriteJsonFile(settings, jsonFile)
}

func WriteJsonFile(settings JsonSettings, jsonFile string) error {
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
func ReadJson(js []byte) (JsonSettings, error) {
	var settings JsonSettings
	err := json.Unmarshal(js, &settings)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	log.Printf("Settings: %+v", settings)
	return settings, err
}

func WriteJson(m JsonSettings) ([]byte, error) {
	return json.Marshal(m)
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
