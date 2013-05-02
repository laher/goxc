package tasks

import (
	"testing"
)

func TestRegister(t *testing.T) {
	l := len(allTasks)
	Register(Task{"blah", "blah", nil, nil})
	if len(allTasks)-l != 1 {
		t.Fatalf("unexpected result %v should be one more than %v", len(allTasks), l)
	}
}
