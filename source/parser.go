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

func FindMainDirs(root string) ([]string, error) {
	mainDirs := []string{}
	sourceFiles := []string{}
	root, err := filepath.Abs(root)
	if err != nil {
		log.Printf("Error resolving root dir: %v", err)
		return []string{}, err
	}
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
				if strings.HasSuffix(path, ".go") {
					log.Printf("Ignoring '.hidden' file %s", path)
				}
			}
		} else {
			if strings.HasSuffix(path, ".go") {
				//read file and check package
				sourceFiles = append(sourceFiles, path)
			}
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
		if file.Name.Name == "main" {
			mainDir, err := filepath.Abs(filepath.Dir(name))
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
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

func FindConstantValue(f *ast.File, name string) *ast.BasicLit {
	return FindValue(f, name, []token.Token{token.CONST})
}

//TODO: refactor to more idiomatic version of this (e.g. use Visit?)
func FindValue(f *ast.File, name string, toks []token.Token) *ast.BasicLit {
	isName := false
	isTok := false
	value := ""
	found := false
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
				log.Printf("Found value (%s)", value)
				ret = x
				found = true
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
