package config

import (
	"testing"
)

func TestLoadSettings(t *testing.T) {
	js := []byte(`{
	"verbose" : true,
	"artifactVersion" : "0.1.1",
	"zipArchives" : false,
	"ArtifactsDest" : "../goxc-pages/",
	"platforms": "windows/386"
	}`)
	settings, err := ReadSettings(js)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	if !settings.Verbose {
		t.Fatalf("Verbose flag not set!")
	}
	if settings.ZipArchives {
		t.Fatalf("Zip Archives flag not set!")
	}
}
