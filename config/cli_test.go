package config

import (
	"testing"

	"github.com/laher/goxc/typeutils"
)

type TasksAndTaskSettings struct {
	Args         []string
	Tasks        []string
	TaskSettings map[string]map[string]interface{}
	ParseError   error
}

var fixtures = []TasksAndTaskSettings{
	{Args: []string{"task1", "-blah=1", "-wah=2"},
		Tasks: []string{"task1"},
		TaskSettings: map[string]map[string]interface{}{
			"task1": map[string]interface{}{
				"blah": "1",
				"wah":  "2",
			},
		},
	},
	{Args: []string{"task1", "-blah=1", "t2", "-wah=2"},
		Tasks: []string{"task1", "t2"},
		TaskSettings: map[string]map[string]interface{}{
			"task1": map[string]interface{}{
				"blah": "1",
			},
			"t2": map[string]interface{}{
				"wah": "2",
			},
		},
	},
}

func TestParseCliTasksAndSettings(t *testing.T) {
	for _, fixture := range fixtures {
		t.Logf("Fixture: %+v", fixture.Args)
		tasks, taskSettings, err := ParseCliTasksAndTaskSettings(fixture.Args)
		if err != fixture.ParseError {
			t.Errorf("Error: '%v' differs from expected: %v", err, fixture.ParseError)
		}
		t.Logf("Tasks: %v", tasks)
		if !typeutils.StringSliceEquals(tasks, fixture.Tasks) {
			t.Errorf("Tasks %v not equal to expected %v", tasks, fixture.Tasks)
		}
		t.Logf("TaskSettings: %v", taskSettings)

		if !typeutils.AreMapStringMapStringInterfacesEqual(taskSettings, fixture.TaskSettings) {
			t.Errorf("TaskSettings %v not equal to expected %v", taskSettings, fixture.TaskSettings)
		}
	}
}
