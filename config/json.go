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
	"fmt"
)

//TODO!!
/*
type Config struct {
	Name            string
	Version         string
	ManifestVersion string
	Author          string
	Description     string
	Url             string
	License         string
	//Platforms       Platforms
}
*/

//use json
func ReadSettings(js []byte) (Settings, error) {
	var settings Settings
	err := json.Unmarshal(js, &settings)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", settings)
	return settings, err
}

func WriteSettings(m Settings) ([]byte, error) {
	return json.Marshal(m)
}
