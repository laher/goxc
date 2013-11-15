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
	"fmt"
	//Tip for Forkers: please 'clone' from my url and then 'pull' from your url. That way you wont need to change the import path.
	//see https://groups.google.com/forum/?fromgroups=#!starred/golang-nuts/CY7o2aVNGZY
	"github.com/laher/goxc/source"
	"go/parser"
	"go/token"
	"go/printer"
	"log"
	"os"
	"path/filepath"
)

const (
	TASK_INTERPOLATE_SOURCE = "interpolate-source"
)

//runs automatically
func init() {
	Register(Task{
		TASK_INTERPOLATE_SOURCE,
		"Replaces a given constant/var value with the current version.",
		runTaskInterpolateSource,
		map[string]interface{}{"varnameVersion" : "VERSION"}})
}

func runTaskInterpolateSource(tp TaskParams) (error) {
	err := writeSource(tp)
	return err
}

func writeSource(tp TaskParams) (err error) {

	varname := tp.Settings.GetTaskSettingString(TASK_INTERPOLATE_SOURCE, "varnameVersion")
	versionName := fmt.Sprintf("\"%s\"", tp.Settings.GetFullVersionName())
	if varname != "" {
		matches, err := filepath.Glob(filepath.Join(tp.WorkingDirectory,"**.go"))
		if err != nil {
			return err
		}
		log.Printf("Source files: %v", matches)
		fset := token.NewFileSet() // positions are relative to fset
		found := false
		for _, match := range matches {
			f, err := parser.ParseFile(fset, match, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			//find version var
			versionVar := source.FindValue(f, varname, []token.Token{token.CONST, token.VAR})
			if versionVar != nil {
				found = true
				log.Printf("Version '%s' = %v -> %s", varname, versionVar.Value, versionName)
				versionVar.Value = versionName
				fw, err := os.OpenFile(match, os.O_WRONLY | os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				defer fw.Close()
				err = printer.Fprint(fw, fset, f)
				if err != nil {
					return err
				}
				err = fw.Close()
				if err != nil {
					return err
				}
			}
		}
		if !found {
			log.Printf("Version var '%s' not found", varname)
		}

	}
	if err != nil {
		return err
	}
	return err
}

