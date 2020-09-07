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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/networkservicemesh/cloudtest/pkg/model"
)

const noneParseFlag = 0

func Find(root string) ([]*model.Suite, error) {
	var testPaths []string
	var suiteTests = make(map[string][]string)
	var suiteEntryPoints = make(map[string][]string)
	var result []*model.Suite

	if err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, "_test.go") {
				testPaths = append(testPaths, path)
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()

	for _, p := range testPaths {
		b, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}
		f, err := parser.ParseFile(fileSet, p, string(b), noneParseFlag)
		if err != nil {
			return nil, err
		}
		forEachTest(f, func(x *ast.FuncDecl) {
			if x.Recv != nil && len(x.Type.Params.List) == 0 {
				if v, ok := x.Recv.List[0].Type.(*ast.StarExpr); ok {
					suiteTests[v.X.(*ast.Ident).Name] = append(suiteTests[v.X.(*ast.Ident).Name], x.Name.Name)
				}
			}
			if len(x.Type.Params.List) == 1 {
				if v, ok := x.Type.Params.List[0].Type.(*ast.StarExpr); ok {
					if v.X.(*ast.SelectorExpr).X.(*ast.Ident).Name == "testing" {
						if suiteName, ok := findSuiteNameInBody(x.Body); ok {
							suiteEntryPoints[suiteName] = append(suiteEntryPoints[suiteName], x.Name.Name)
						}
					}
				}
			}
		})
	}
	for k, v := range suiteTests {
		rootTests := suiteEntryPoints[k]
		if len(rootTests) == 0 {
			continue
		}
		for _, name := range rootTests {
			result = append(result, &model.Suite{
				Name:  name,
				Tests: append([]string{}, v...),
			})
		}
	}
	return result, nil
}

func forEachTest(node ast.Node, applier func(decl *ast.FuncDecl)) {
	ast.Inspect(node, func(n ast.Node) bool {
		if decl, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(decl.Name.Name, "Test") {
				applier(decl)
			}
		}
		return true
	})
}

func findSuiteNameInBody(body *ast.BlockStmt) (string, bool) {
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
			if name := findExpressionName(call.Args[1]); name != "" {
				return name, true
			}
		}
	}
	return "", false
}
func findExpressionName(exp ast.Expr) string {
	if exp == nil {
		return ""
	}
	switch v := exp.(type) {
	case *ast.CallExpr:
		if funName := findExpressionName(v.Fun); funName == "new" {
			return findExpressionName(v.Args[0])
		}
	case *ast.CompositeLit:
		return findExpressionName(v.Type)
	case *ast.UnaryExpr:
		return findExpressionName(v.X)
	case *ast.Ident:
		if v.Obj == nil {
			return v.Name
		}
		if spec, ok := v.Obj.Decl.(*ast.ValueSpec); ok {
			for _, val := range spec.Values {
				name := findExpressionName(val)
				if name != "" {
					return name
				}
			}
		}
		if assign, ok := v.Obj.Decl.(*ast.AssignStmt); ok {
			for _, r := range assign.Rhs {
				name := findExpressionName(r.(ast.Expr))
				if name != "" {
					return name
				}
			}
		}
		return v.Name
	}
	return ""
}
