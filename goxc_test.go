package main

import (
	"encoding/json"
	"github.com/laher/goxc/config"
	"testing"
)

func TestRemove(t *testing.T) {
	//goroot := runtime.GOROOT()
	arr := []string{"1", "2"}
	removed := remove(arr, "1")
	if len(removed) != 1 {
		t.Fatalf("Remove failed!")
	}
	removed = remove(arr, "3")
	if len(removed) != 2 {
		t.Fatalf("Remove failed!")
	}
}

func TestPrintJsonDefaults(t *testing.T) {
	settings := config.Settings{}
	settings = fillDefaults(settings)
	jsonSettings := config.WrapJsonSettings(settings)
	jsonD, _ := json.MarshalIndent(jsonSettings, "", "\t")
	t.Logf("JSON defaults: \n%+v", string(jsonD))
}
