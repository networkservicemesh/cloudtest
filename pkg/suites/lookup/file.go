// Copyright (c) 2021 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lookup

import (
	"go/ast"
	"go/token"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// File is an *ast.File with suites lookup
type File struct {
	resolvedImports map[string]*Package // import path -> *Package
	imports         map[string]*Package // name -> *Package
	suites          map[string]*Suite   // name -> *Suite
	pkg             *Package

	*ast.File
}

func newFile(file *ast.File, resolvedImports map[string]*Package, suites map[string]*Suite, pkg *Package) *File {
	f := &File{
		resolvedImports: resolvedImports,
		imports:         make(map[string]*Package),
		suites:          suites,
		pkg:             pkg,
		File:            file,
	}

	f.findImports()

	return f
}

func (f *File) findImports() {
	for _, importSpec := range f.Imports {
		importPath := strings.ReplaceAll(importSpec.Path.Value, "\"", "")

		var name string
		if importSpec.Name != nil {
			name = importSpec.Name.Name
		} else {
			name = path.Base(importPath)
		}

		var pkg *Package
		if resolved, ok := f.resolvedImports[importPath]; ok {
			pkg = resolved
		} else {
			pkg = newImport(importPath, f.resolvedImports)
		}

		f.imports[name] = pkg
	}
}

// Lookup looks up for a `pkgName.name` suite
func (f *File) Lookup(pkgName, name string) (*Suite, error) {
	pkg, ok := f.imports[pkgName]
	if !ok {
		return nil, errors.Errorf("invalid import: %s", pkgName)
	}
	return pkg.Lookup(name)
}

func (f *File) findSuiteParent(name string) (parentSuite *Suite, err error) {
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				typeSpec := spec.(*ast.TypeSpec) // it should be `*ast.TypeSpec` because of `token.TYPE`
				if typeSpec.Name.Name != name {
					continue
				}

				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					for i := 0; parentSuite == nil && i < len(structType.Fields.List); i++ {
						if field := structType.Fields.List[i]; field.Names == nil {
							if parentSuite, err = f.getSuite(field.Type); err != nil {
								return nil, err
							}
						}
					}
				}

				return parentSuite, nil
			}
		}
	}
	return nil, nil
}

func (f *File) getSuite(expr ast.Expr) (*Suite, error) {
	switch v := expr.(type) {
	case *ast.Ident:
		return f.pkg.Lookup(v.Name)
	case *ast.StarExpr:
		return f.getSuite(v.X)
	case *ast.SelectorExpr:
		if x, ok := v.X.(*ast.Ident); ok {
			return f.imports[x.Name].Lookup(v.Sel.Name)
		}
	}
	return nil, nil
}

func (f *File) findSuiteTests(name string) (tests []string) {
	for _, decl := range f.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && isSuiteTest(funcDecl) {
			var recvName string
			switch v := funcDecl.Recv.List[0].Type.(type) {
			case *ast.Ident:
				recvName = v.Name
			case *ast.StarExpr:
				recvName = v.X.(*ast.Ident).Name
			}

			if recvName == name {
				tests = append(tests, funcDecl.Name.Name)
			}
		}
	}
	return tests
}

func isSuiteTest(funcDecl *ast.FuncDecl) bool {
	// func (s Suite) Test...() {...}
	if funcDecl.Recv == nil {
		return false
	}

	if !strings.HasPrefix(funcDecl.Name.Name, "Test") {
		return false
	}

	return len(funcDecl.Type.Params.List) == 0
}
