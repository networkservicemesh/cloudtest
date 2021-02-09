// Copyright (c) 2020 Doc.ai and/or its affiliates.
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

package suites

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/networkservicemesh/cloudtest/pkg/model"
	"github.com/networkservicemesh/cloudtest/pkg/suites/lookup"
)

const noneParseFlag = 0

// Find finds go test suites recursively in root
func Find(root string) (suites []*model.Suite, err error) {
	var testPaths []string
	if err = filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				testPaths = append(testPaths, path)
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()
	resolvedImports := lookup.ResolvedImports()

	for _, p := range testPaths {
		pkgNodes, err := parser.ParseDir(fileSet, p, func(info os.FileInfo) bool {
			return strings.HasSuffix(info.Name(), "_test.go")
		}, noneParseFlag)
		if err != nil {
			return nil, err
		}

		for _, pkgNode := range pkgNodes {
			pkg := lookup.NewPackage(pkgNode, resolvedImports)
			for _, file := range pkg.Files {
				forEachTest(file, func(funcDecl *ast.FuncDecl) {
					var suite *lookup.Suite
					pkgName, suiteName := findSuiteNameInBody(funcDecl.Body)
					if pkgName != "" {
						suite = file.Lookup(pkgName, suiteName)
					} else if suiteName != "" {
						suite = pkg.Lookup(suiteName)
					}
					if suite != nil {
						suites = append(suites, &model.Suite{
							Name:  funcDecl.Name.Name,
							Tests: suite.GetTests(),
						})
					}
				})
			}
		}
	}

	return suites, nil
}

func forEachTest(file *lookup.File, applier func(funcDecl *ast.FuncDecl)) {
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			// func Test...(... *testing.T) {...}
			if !strings.HasPrefix(funcDecl.Name.Name, "Test") {
				continue
			}
			params := funcDecl.Type.Params.List
			if len(params) != 1 {
				continue
			}
			ptrTestingT, ok := params[0].Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			testingT, ok := ptrTestingT.X.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			testing, ok := testingT.X.(*ast.Ident)
			if !ok {
				continue
			}
			if testing.Name != "testing" || testingT.Sel.Name != "T" {
				continue
			}
			applier(funcDecl)
		}
	}
}

func findSuiteNameInBody(body *ast.BlockStmt) (pkgName, name string) {
	for _, l := range body.List {
		var expr *ast.ExprStmt
		var ok bool
		if expr, ok = l.(*ast.ExprStmt); !ok {
			continue
		}
		var call *ast.CallExpr
		if call, ok = expr.X.(*ast.CallExpr); !ok {
			continue
		}
		if len(call.Args) != 2 {
			continue
		}
		var selector *ast.SelectorExpr
		if selector, ok = call.Fun.(*ast.SelectorExpr); !ok {
			continue
		}
		if selector.X.(*ast.Ident).Name == "suite" && selector.Sel.Name == "Run" {
			return findExpressionName(call.Args[1])
		}
	}
	return "", ""
}

func findExpressionName(exp ast.Expr) (pkgName, name string) {
	if exp == nil {
		return "", ""
	}
	switch v := exp.(type) {
	case *ast.CallExpr:
		if _, funName := findExpressionName(v.Fun); funName == "new" {
			return findExpressionName(v.Args[0])
		}
	case *ast.CompositeLit:
		return findExpressionName(v.Type)
	case *ast.UnaryExpr:
		return findExpressionName(v.X)
	case *ast.SelectorExpr:
		if x, ok := v.X.(*ast.Ident); ok {
			return x.Name, v.Sel.Name
		}
	case *ast.Ident:
		if v.Obj == nil {
			return "", v.Name
		}
		if spec, ok := v.Obj.Decl.(*ast.ValueSpec); ok {
			for _, val := range spec.Values {
				pkgName, name = findExpressionName(val)
				if name != "" {
					return pkgName, name
				}
			}
		}
		if assign, ok := v.Obj.Decl.(*ast.AssignStmt); ok {
			for _, r := range assign.Rhs {
				pkgName, name = findExpressionName(r.(ast.Expr))
				if name != "" {
					return pkgName, name
				}
			}
		}
		return "", v.Name
	}
	return "", ""
}
