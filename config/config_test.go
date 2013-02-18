package main

import (
	"testing"
)

func TestLoadManifest(t *testing.T) {
	js := []byte(`{
	"platforms": {
		"windows": [ "386", "amd64" ],
		"linux": ["amd64"]
	}}`)
	ReadManifest(js)
}
