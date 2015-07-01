package config

import (
	"log"
	"path/filepath"
	"testing"
)

func TestStripEmpties(t *testing.T) {
	js := []byte(`{
	"Verbosity" : "",
	"PackageVersion" : "0.1.1",
	"zipArchives" : false,
	"ArtifactsDest" : "../goxc-pages/",
	"platforms": ["windows/386"],
	"blah" : []
	 }`)

	outjs, err := StripEmpties(js, true)
	if err != nil {
		t.Fatalf("Error returned from StripEmpties: %v", err)
	}
	log.Printf("stripped: %v", string(outjs))
}
func TestLoadSettings(t *testing.T) {
	js := []byte(`{
	"Verbosity" : "v",
	"PackageVersion" : "0.1.1",
	"zipArchives" : false,
	"ArtifactsDest" : "../goxc-pages/",
	"platforms": "windows/386"
	}`)
	settings, err := readJson(js)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	if !settings.IsVerbose() {
		t.Fatalf("Verbose flag not set!")
	}
}

func TestLoadJsonConfigsInvalid(t *testing.T) {
	_, err := LoadJsonConfigs("", []string{filepath.Join("testdata", "invalid.goxc.json")}, false)
	if err == nil {
		t.Fatalf("invalid.goxc.json was loaded without an error, despite being invalid")
	}
}

/*
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
*/
