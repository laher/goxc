package source

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
