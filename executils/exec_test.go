package executils

import (
	"github.com/laher/goxc/typeutils"
	"testing"
)

func TestGetLdFlagVersionArgs(t *testing.T) {
	actual := GetLdFlagVersionArgs("1.1")
	expected := []string{"-ldargs", "-X main.Version 1.1"}
	if typeutils.StringSliceEquals(actual, expected) {
		t.Fatalf("unexpected result %v != %v", actual, expected)
	}
}
