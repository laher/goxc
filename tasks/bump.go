package tasks

/*
   Copyright 2013 Am Laher

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"errors"
	"log"
	"strconv"
	"strings"

	// Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	// see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
)

const TASK_BUMP = "bump"

//runs automatically
func init() {
	Register(Task{
		TASK_BUMP,
		"bump package version in .goxc.json. By default, the patch number (after the second dot) is increased by one. You can specify major or minor instead with -dot=0 or -dot=1",
		bump,
		map[string]interface{}{
			"dot": "2"}})
}

func bump(tp TaskParams) error {
	c, err := config.LoadJsonConfigs(tp.WorkingDirectory, []string{core.GOXC_CONFIGNAME_BASE + core.GOXC_FILE_EXT}, !tp.Settings.IsQuiet())
	if err != nil {
		return nil
	}
	pv := c.PackageVersion
	if pv == core.PACKAGE_VERSION_DEFAULT || pv == "" {
		//go from 'default' version to 0.0.1 (or 0.1.0 or 1.0.0)
		pv = "0.0.0"
	}
	pvparts := strings.Split(pv, ".")
	partToBumpStr := tp.Settings.GetTaskSettingString(TASK_BUMP, "dot")
	partToBump, err := strconv.Atoi(partToBumpStr)
	if err != nil {
		return err
	}
	if partToBump < 0 {
		return errors.New("Could not determine which part of the version number to bump")
	}
	if len(pvparts) > partToBump {
		thisPart := pvparts[partToBump]
		thisPartNum, err := strconv.Atoi(thisPart)
		if err != nil {
			return err
		}
		thisPartNum += 1
		pvparts[partToBump] = strconv.Itoa(thisPartNum)
		for i, p := range pvparts[partToBump+1:] {
			_, err := strconv.Atoi(p)
			if err != nil {
				break
			} else {
				//reset smaller parts to 0
				pvparts[i+partToBump+1] = "0"
			}

		}
		pvNew := strings.Join(pvparts, ".")
		c.PackageVersion = pvNew
		if !tp.Settings.IsQuiet() {
			log.Printf("Bumping from %s to %s", pv, c.PackageVersion)
		}
		tp.Settings.PackageVersion = pvNew
		return config.WriteJsonConfig(tp.WorkingDirectory, c, "", false)
	} else {
		return errors.New("PackageVersion does not contain enough dots to bump this part of the version number")
	}
}
