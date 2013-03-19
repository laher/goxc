package main

import (
	"testing"
)

func TestRemove(t *testing.T) {
	//goroot := runtime.GOROOT()
	arr := []string{"1","2"}
	removed := remove(arr, "1")
	if len(removed) != 1 {
		t.Fatalf("Remove failed!")
	}
	removed = remove(arr, "3")
	if len(removed) != 2 {
		t.Fatalf("Remove failed!")
	}
}
