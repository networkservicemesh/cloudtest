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
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Package is an *ast.Package with suites lookup
type Package struct {
	resolvedImports map[string]*Package // import path -> *Package
	suites          map[string]*Suite   // name -> *Suite
	importPath      string
	Files           []*File
	once            *sync.Once
	resolveErr      error
}

// NewPackage creates a new Package from *ast.Package
func NewPackage(pkg *ast.Package, resolvedImports map[string]*Package) *Package {
	p := &Package{
		resolvedImports: resolvedImports,
		suites:          make(map[string]*Suite),
	}

	for _, f := range pkg.Files {
		p.Files = append(p.Files, newFile(f, p.resolvedImports, p.suites, p))
	}

	return p
}

func newImport(importPath string, resolvedImports map[string]*Package) *Package {
	return &Package{
		resolvedImports: resolvedImports,
		suites:          make(map[string]*Suite),
		importPath:      importPath,
		once:            new(sync.Once),
	}
}

func (p *Package) resolve() error {
	pkg, err := build.Default.Import(p.importPath, ".", build.FindOnly)
	if err != nil {
		return err
	}

	pkgNodes, err := parser.ParseDir(token.NewFileSet(), pkg.Dir, func(fileInfo os.FileInfo) bool {
		return !strings.HasSuffix(fileInfo.Name(), "_test.go")
	}, 0)
	if err != nil {
		return err
	}
	if len(pkgNodes) != 1 {
		return errors.Errorf("found more than 1 package in directory: %s", p.importPath)
	}

	for _, pkgNode := range pkgNodes {
		for _, f := range pkgNode.Files {
			p.Files = append(p.Files, newFile(f, p.resolvedImports, p.suites, p))
		}
	}

	return nil
}

// Lookup looks up for a `name` suite
func (p *Package) Lookup(name string) (suite *Suite, err error) {
	if p.once != nil {
		p.once.Do(func() {
			p.resolveErr = p.resolve()
		})
	}
	if p.resolveErr != nil {
		return nil, p.resolveErr
	}

	var ok bool
	suite, ok = p.suites[name]
	if ok {
		return suite, nil
	}

	suite = new(Suite)
	if suite.parent, err = p.findSuiteParent(name); suite.parent == nil || err != nil {
		return nil, err
	}
	suite.tests = p.findSuiteTests(name)

	p.suites[name] = suite

	return suite, nil
}

func (p *Package) findSuiteParent(name string) (parentSuite *Suite, err error) {
	for i := 0; parentSuite == nil && i < len(p.Files); i++ {
		if parentSuite, err = p.Files[i].findSuiteParent(name); err != nil {
			return nil, err
		}
	}
	return parentSuite, nil
}

func (p *Package) findSuiteTests(name string) (tests []string) {
	for _, file := range p.Files {
		tests = append(tests, file.findSuiteTests(name)...)
	}
	return tests
}
