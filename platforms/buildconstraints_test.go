package platforms

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	testBCs := map[string]string{
		"freebsd linux,!arm": "[{freebsd 386} {freebsd amd64} {linux 386} {linux amd64}]",
		"windows":            "[{windows 386} {windows amd64}]",
		"windows,!386":       "[{windows amd64}]",
		"!windows":           "[{darwin 386} {darwin amd64} {linux 386} {linux amd64} {linux arm} {freebsd 386} {freebsd amd64} {openbsd 386} {openbsd amd64}]",
		"!386":               "[{darwin amd64} {linux amd64} {linux arm} {freebsd amd64} {openbsd amd64} {windows amd64}]",
		"":                   "[{darwin 386} {darwin amd64} {linux 386} {linux amd64} {linux arm} {freebsd 386} {freebsd amd64} {openbsd 386} {openbsd amd64} {windows 386} {windows amd64}]",
	}
	for buildConstraints, expectedPlatforms := range testBCs {
		targets := ApplyBuildConstraints(buildConstraints, SUPPORTED_PLATFORMS_1_0)
		t.Logf("build: %s. targets: %v", buildConstraints, targets)
		targetsAsString := fmt.Sprintf("%v", targets)
		if targetsAsString != expectedPlatforms {
			t.Fatalf("unexpected result %v != %v", expectedPlatforms, targets)
		}
	}
}
