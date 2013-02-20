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
	"path/filepath"
)

type JsonSettings struct {
	Settings Settings
	//Format for goxc.json files
	FormatVersion string
	//TODO??: InheritFiles []string
}

func LoadJsonCascadingConfig(dir string, configName string, verbose bool) (Settings, error) {
	jsonFile := filepath.Join(dir, configName + ".json")
	jsonLocalFile := filepath.Join(dir, configName + ".local.json")
	localSettings, err := LoadJsonFile(jsonLocalFile, verbose)
	if err != nil {
		settings, err := LoadJsonFile(jsonFile, verbose)
		if err != nil {
			if verbose {
				log.Printf("Could NOT load %s: %s", jsonFile, err)
			}
		}
		return settings.Settings, err
	} else {
		settings, err := LoadJsonFile(jsonFile, verbose)
		if err != nil {
			if verbose {
				log.Printf("Could NOT load %s: %s", jsonLocalFile, err)
			}
			return localSettings.Settings, nil
		} else {
			return Merge(localSettings.Settings, settings.Settings), nil
		}
	}
	//unreachable but required by go compiler
	return localSettings.Settings, err
}

// load json file. Glob for goxc
func LoadJsonFile(jsonFile string, verbose bool) (JsonSettings, error) {
	var settings JsonSettings
	file, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		if verbose {
			log.Printf("File error: %v", err)
		}
		return settings, err
	}
	//TODO: super-verbose option for logging file content? log.Printf("%s\n", string(file))
	json.Unmarshal(file, &settings)
	if err != nil {
		if verbose {
			log.Printf("Unmarshal error: %s", err)
		}
		return settings, err
	}
	//TODO: verbosity here? log.Printf("Results: %v", settings)
	return settings, nil
}

//use json from string
func ReadJson(js []byte) (JsonSettings, error) {
	var settings JsonSettings
	err := json.Unmarshal(js, &settings)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	log.Printf("Settings: %v", settings)
	return settings, err
}

func WriteSettings(m JsonSettings) ([]byte, error) {
	return json.Marshal(m)
}
