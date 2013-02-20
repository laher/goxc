package config

import (
	"testing"
	"log"
)

func TestLoadSettings(t *testing.T) {
	js := []byte(`{ "Settings" : {
	"Verbosity" : "v",
	"artifactVersion" : "0.1.1",
	"zipArchives" : false,
	"ArtifactsDest" : "../goxc-pages/",
	"platforms": "windows/386"
	} }`)
	settings, err := ReadJson(js)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	if !settings.Settings.IsVerbose() {
		t.Fatalf("Verbose flag not set!")
	}
	if settings.Settings.IsZip() {
		t.Fatalf("Zip Archives flag not set!")
	}
}
func TestLoadFile(t *testing.T) {
	file := "goxc.json"
	settings, err := LoadJsonFile(file, true)
	if err != nil {
		t.Fatalf("Err: %v", err)
	} else {
		log.Printf("settings: %v", settings)
	}
}

func TestLoadParentFile(t *testing.T) {
	file := "../goxc.json"
	settings, err := LoadJsonFile(file, true)
	if err != nil {
		t.Fatalf("Err: %v", err)
	} else {
		log.Printf("settings: %v", settings)
	}
}
