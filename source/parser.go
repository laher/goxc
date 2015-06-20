// Limited support for parsing Go source.
package source

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
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func FindMainDirs(root string, excludingGlobs []string, isVerbose bool) ([]string, error) {
	return FindSourceDirs(root, "main", excludingGlobs, isVerbose)
}

func FindSourceDirs(root string, packageNameFilter string, excludingGlobs []string, isVerbose bool) ([]string, error) {
	mainDirs := []string{}
	sourceFiles := []string{}
	root, err := filepath.Abs(root)
	if err != nil {
		log.Printf("Error resolving root dir: %v", err)
		return []string{}, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		log.Printf("Error resolving root dir (EvalSymlinks): %v", err)
		return []string{}, err
	}
	err = filepath.Walk(root, func(path string, info os.FileInfo, inerr error) error {

		var err error
		//find files ending with .go
		//check for 'package main'
		//if strings.Contains(path, string(filepath.Separator) + ".") {
		//skip '.hidden' dirs
		if strings.HasPrefix(filepath.Base(path), ".") {
			finfo, err := os.Stat(path)
			if err != nil {
				log.Printf("Error stat'ing %s", path)
				return err
			}
			if finfo.IsDir() {
				return filepath.SkipDir
			} else {
				//only log if it's a go file
				if strings.HasSuffix(path, ".go") && isVerbose {
					log.Printf("Ignoring '.hidden' file %s", path)
				}
			}
		} else if strings.HasSuffix(path, ".go") {
			//read file and check package
			sourceFiles = append(sourceFiles, path)
		}
		return err
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return mainDirs, err
	}
	parsedMap, err := LoadFilesMap(sourceFiles)
	if err != nil {
		log.Printf("Error: %v", err)
		return mainDirs, err
	}
	for name, file := range parsedMap {
		if packageNameFilter == "" || file.Name.Name == packageNameFilter {
			mainDir, err := filepath.Abs(filepath.Dir(name))
			if err != nil {
				log.Printf("Abs error: %s: %v", filepath.Dir(name), err)
			} else {
				excluded := false
				for _, exclGlob := range excludingGlobs {
					//log.Printf("Glob testing: %s matches %s", filepath.Join(root, exclGlob), mainDir)
					matches, err := filepath.Match(filepath.Join(root, exclGlob), mainDir)
					if err != nil {
						//ignore this exclusion glob
						log.Printf("Glob error: %s: %s", exclGlob, err)
					} else if matches {
						if isVerbose {
							log.Printf("Main dir '%s' excluded by glob '%s'", mainDir, exclGlob)
						}
						excluded = true
					} else {
						absExcl, err := filepath.Abs(filepath.Join(root, exclGlob))
						if err != nil {
							//ignore
							log.Printf("Abs error: %s: %v", filepath.Join(root, exclGlob), err)
						} else if strings.HasPrefix(mainDir, absExcl) {
							if isVerbose {
								log.Printf("Main dir '%s' excluded because it is in '%s'", mainDir, absExcl)
							}
							excluded = true
						} else {
							//log.Printf("Main dir '%s' is NOT in '%s'", mainDir, absExcl)
						}
					}

				}
				if !excluded {
					alreadyThere := false
					for _, v := range mainDirs {
						if v == mainDir {
							alreadyThere = true
						}
					}
					if !alreadyThere {
						mainDirs = append(mainDirs, mainDir)
					}
				}
			}
		}
	}
	return mainDirs, err
}

func LoadFilesMap(filenames []string) (map[string]*ast.File, error) {
	filesMap := map[string]*ast.File{}
	fset := token.NewFileSet() // positions are relative to fset
	for _, match := range filenames {
		f, err := parser.ParseFile(fset, match, nil, 0)
		if err != nil {
			log.Printf("Source parser error %v", err)
		} else {
			filesMap[match] = f
		}
	}
	return filesMap, nil
}
func LoadFiles(filenames []string) ([]*ast.File, error) {
	files := []*ast.File{}
	fset := token.NewFileSet() // positions are relative to fset
	for _, match := range filenames {
		f, err := parser.ParseFile(fset, match, nil, 0)
		if err != nil {
			log.Printf("Source parser error %v", err)
		} else {
			files = append(files, f)
		}
	}
	return files, nil
}

func FindConstantValue(f *ast.File, name string, isVerbose bool) *ast.BasicLit {
	return FindValue(f, name, []token.Token{token.CONST}, isVerbose)
}

//TODO: refactor to more idiomatic version of this (e.g. use Visit?)
func FindValue(f *ast.File, name string, toks []token.Token, isVerbose bool) *ast.BasicLit {
	isName := false
	isTok := false
	value := ""
	var ret *ast.BasicLit
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			isTok = false
			for _, tok := range toks {
				if x.Tok == tok {
					isTok = true
				}
			}
		case *ast.BasicLit:
			if isName && isTok {
				//strip quotes
				value = strings.Replace(x.Value, "\"", "", -1)
				if isVerbose {
					log.Printf("Found value (%s)", value)
				}
				ret = x
				isName = false
				return false //break out
			}
		case *ast.Ident:
			isName = false
			if isTok {
				//log.Printf("Matching token type, named %s, %+v", x.Name, x)
				if x.Name == name {
					isName = true
				}
			}
		}
		return true
	})
	//return value, found
	return ret
}
