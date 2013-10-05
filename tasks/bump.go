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
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/config"
	"github.com/laher/goxc/core"
	"log"
	"strconv"
	"strings"
)

//runs automatically
func init() {
	Register(Task{
		"bump",
		"bump package version in .goxc.json",
		bump,
		nil})
}

func bump(tp TaskParams) error {
	c, err := config.LoadJsonConfigs(tp.WorkingDirectory, []string{core.GOXC_CONFIGNAME_BASE + core.GOXC_FILE_EXT}, tp.Settings.IsVerbose())
	if err != nil {
		return nil
	}
	pv := c.PackageVersion
	if pv == core.PACKAGE_VERSION_DEFAULT {
		//go from 'default' version to 0.0.1
		c.PackageVersion = "0.0.1"
	} else {
		pvparts := strings.Split(pv, ".")
		lastpart := pvparts[len(pvparts)-1]
		lastPartNum, err := strconv.Atoi(lastpart)
		if err != nil {
			return err
		}
		lastPartNum+=1
		pvparts[len(pvparts)-1] = strconv.Itoa(lastPartNum)
		pvNew := strings.Join(pvparts, ".")
		c.PackageVersion = pvNew
	}
	log.Printf("Bumping from %s to %s", pv, c.PackageVersion)
	return config.WriteJsonConfig(tp.WorkingDirectory, c, "", false)
}

