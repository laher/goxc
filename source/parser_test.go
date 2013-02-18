package source

// +build ignore

import (
	"path/filepath"
	"testing"
)

func TestLearning(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "goxc.go"))
	if err != nil {
		t.Logf("Glob error %v", err)
	} else {
		files, err := LoadFiles(matches)
		if err != nil {
			t.Logf("%v", err)
			return
		}
		for _, f := range files {
			version := FindConstantValue(f, "PKG_VERSION")
			t.Logf("Version = %v", version)
			name := FindConstantValue(f, "PKG_NAME")
			t.Logf("Name = %v", name)
		}
	}
}
