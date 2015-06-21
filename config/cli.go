package config

import (
	"errors"
	"strings"
)

//tasks (and task settings) are defined as args after the main flags
func ParseCliTasksAndTaskSettings(args []string) ([]string, map[string]map[string]interface{}, error) {
	tasks := []string{}
	taskSettings := map[string]map[string]interface{}{}
	lastTask := ""
	lastTaskSettingKey := ""
	for _, arg := range args {
		if lastTaskSettingKey != "" {
			taskSettings[lastTask][lastTaskSettingKey] = arg
			lastTaskSettingKey = ""
		} else if strings.HasPrefix(arg, "-") {
			_, exists := taskSettings[lastTask]
			if !exists {
				taskSettings[lastTask] = map[string]interface{}{}
			}
			if strings.Contains(arg, "=") {
				splut := strings.Split(arg, "=")
				key := splut[0][1:]
				//strip double-hyphen
				if strings.HasPrefix(key, "-") {
					key = key[1:]
				}
				val := splut[1]
				taskSettings[lastTask][key] = val
			} else {
				key := arg[1:]
				//strip double-hyphen
				if strings.HasPrefix(key, "-") {
					key = key[1:]
				}
				lastTaskSettingKey = key
			}

		} else {
			tasks = append(tasks, arg)
			lastTask = arg
		}
	}
	if lastTaskSettingKey != "" {
		return tasks, taskSettings, errors.New("Received a task setting with no value. Please at least use empty quotes")
	}
	//log.Printf("TaskSettings: %+v", taskSettings)
	return tasks, taskSettings, nil
}
