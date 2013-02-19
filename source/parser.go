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
	"strings"
)

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

func FindConstantValue(f *ast.File, name string) string {
	return FindValue(f, name, token.CONST)
}

//TODO: refactor to more idiomatic version of this (e.g. use Visit?)
func FindValue(f *ast.File, name string, tok token.Token) string {
	isName := false
	isTok := false
	value := ""
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			switch x.Tok {
			case tok:
				isTok = true
			default:
				isTok = false
			}

		case *ast.BasicLit:
			if isName && isTok {
				//strip quotes
				value = strings.Replace(x.Value, "\"", "", -1)
				isName = false
				return false //break out
			}
		case *ast.Ident:
			if isTok && x.Name == name {
				isName = true
			}
		}
		return true
	})
	return value
}
