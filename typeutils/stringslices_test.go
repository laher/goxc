package typeutils

import (
	"testing"
)

func TestStringSliceEquals(t *testing.T) {
	expected := []string{"-l", "-X"}
	actual := []string{"-l", "-X"}
	if !StringSliceEquals(actual, expected) {
		t.Fatalf("unexpected result %v != %v", actual, expected)
	}
	actual = append(actual, "3")
	if StringSliceEquals(actual, expected) {
		t.Fatalf("unexpected result %v == %v", actual, expected)
	}

	actual2 := []string{"sdfssdfsdfsdf", "-dsdsdsfds werwer"}
	if StringSliceEquals(actual2, expected) {
		t.Fatalf("unexpected result %v == %v", actual2, expected)
	}
}
