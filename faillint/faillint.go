// Package faillint defines an Analyzer that fails when a package is imported
// that matches against a set of defined packages.
package faillint

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer of the linter
var Analyzer = &analysis.Analyzer{
	Name:             "faillint",
	Doc:              "report unwanted import path usages",
	Run:              run,
	RunDespiteErrors: true,
}

var paths string     // -paths flag
var ignoretests bool // -ignore-tests flag

func init() {
	// seems like using init() is the only way to add our own flags
	Analyzer.Flags.StringVar(&paths, "paths", paths, "import paths to fail")
	Analyzer.Flags.BoolVar(&ignoretests, "ignore-tests", ignoretests, "ignore all _test.go files and packages.")
}

// Run is the runner for an analysis pass
func run(pass *analysis.Pass) (interface{}, error) {
	if paths == "" {
		return nil, nil
	}

	p := strings.Split(paths, ",")

	suggestions := make(map[string]string, len(p))
	imports := make([]string, 0, len(p))

	for _, s := range p {
		imps := strings.Split(s, "=")

		imp := imps[0]
		suggest := ""
		if len(imps) == 2 {
			suggest = imps[1]
		}

		imports = append(imports, imp)
		suggestions[imp] = suggest
	}

	for _, file := range pass.Files {
		if ignoretests && strings.Contains(pass.Fset.File(file.Package).Name(), "_test.go") {
			continue
		}
		for _, path := range imports {
			imp := usesImport(file, path)
			if imp == nil {
				continue
			}

			impPath := importPath(imp)

			msg := fmt.Sprintf("package %q shouldn't be imported", impPath)
			if s := suggestions[impPath]; s != "" {
				msg += fmt.Sprintf(", suggested: %q", s)
			}

			pass.Reportf(imp.Path.Pos(), msg)
		}
	}

	return nil, nil
}

// usesImport reports whether a given import is used.
func usesImport(f *ast.File, path string) *ast.ImportSpec {
	spec := importSpec(f, path)
	if spec == nil {
		return nil
	}

	name := spec.Name.String()
	switch name {
	case "<nil>":
		// If the package name is not explicitly specified,
		// make an educated guess. This is not guaranteed to be correct.
		lastSlash := strings.LastIndex(path, "/")
		if lastSlash == -1 {
			name = path
		} else {
			name = path[lastSlash+1:]
		}
	case "_", ".":
		// Not sure if this import is used - err on the side of caution.
		return spec
	}

	var used bool
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if ok && isTopName(sel.X, name) {
			used = true
		}
		return true
	})

	if used {
		return spec
	}

	return nil
}

// importSpec returns the import spec if f imports path,
// or nil otherwise.
func importSpec(f *ast.File, path string) *ast.ImportSpec {
	for _, s := range f.Imports {
		if importPath(s) == path {
			return s
		}
	}
	return nil
}

// importPath returns the unquoted import path of s,
// or "" if the path is not properly quoted.
func importPath(s *ast.ImportSpec) string {
	t, err := strconv.Unquote(s.Path.Value)
	if err == nil {
		return t
	}
	return ""
}

// isTopName returns true if n is a top-level unresolved identifier with the given name.
func isTopName(n ast.Expr, name string) bool {
	id, ok := n.(*ast.Ident)
	return ok && id.Name == name && id.Obj == nil
}
