package config

import (
	"encoding/json"
	"fmt"
)

//TODO!!
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

//use json
func ReadConfig(js []byte) {
	var config Config
	err := json.Unmarshal(js, &config)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", archs)
}

func WriteConfig(m Manifest) ([]byte, error) {
	return json.Marshal(m)
}
