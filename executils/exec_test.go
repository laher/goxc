package executils

/*
func TestGetLdFlagVersionArgs(t *testing.T) {
	actual := GetLdFlagVersionArgs("1.1")
	expected0 := "-ldflags"
	if len(actual) != 2 {
		t.Fatalf("unexpected result length != 2 (%v)", actual)
	}
	if actual[0] != expected0 {
		t.Fatalf("unexpected result length != 2 (%v)", actual[0])
	}
}

func TestGetInterpolationLdFlags(t *testing.T) {
	v := map[string]string{"main.VERSION": "1.0", "main.BUILD_DATE": "1-1-1970"}
	actual := GetInterpolationLdFlags(v)
	expected := []string{"-ldflags", "-X main.VERSION '1.0' -X main.BUILD_DATE '1-1-1970' "}

	if !typeutils.StringSliceEquals(actual, expected) {
		t.Fatalf("unexpected result %v != %v", actual, expected)
	}
}
*/
