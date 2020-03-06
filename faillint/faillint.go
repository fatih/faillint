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

var paths string // -paths flag

func init() {
	// seems like using init() is the only way to add our own flags
	Analyzer.Flags.StringVar(&paths, "paths", paths, "import paths to fail")
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
		for _, path := range imports {
			imp, submatch := usesImport(file, path)
			if imp == nil {
				continue
			}

			if submatch {
				importPath, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					continue
				}
				// If user has explicitly specified this path, skip this submatch,
				// and leave to be handled in the full path.
				if _, ok := suggestions[importPath]; ok {
					continue
				}
			}

			var msg string
			if submatch {
				msg = fmt.Sprintf("sub-package of %q shouldn't be imported", path)
			} else {
				msg = fmt.Sprintf("package %q shouldn't be imported", path)
			}
			if s := suggestions[path]; s != "" {
				msg += fmt.Sprintf(", suggested: %q", s)
			}

			pass.Reportf(imp.Path.Pos(), msg)
		}
	}

	return nil, nil
}

// usesImport reports whether a given import package or its subpackage is used.
func usesImport(f *ast.File, path string) (spec *ast.ImportSpec, submatch bool) {
	spec, submatch = importSpec(f, path)
	if spec == nil {
		return nil, false
	}

	name := spec.Name.String()
	switch name {
	case "<nil>":
		// If the package name is not explicitly specified,
		// get the last component from import path.
		impPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, false
		}
		lastSlash := strings.LastIndex(impPath, "/")
		if lastSlash == -1 {
			name = impPath
		} else {
			name = impPath[lastSlash+1:]
		}
	case "_", ".":
		// Not sure if this import is used - err on the side of caution.
		return spec, submatch
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
		return spec, submatch
	}

	return nil, false
}

// importSpec returns the import spec if f imports path or its subpackage,
// or nil otherwise.
func importSpec(f *ast.File, path string) (node *ast.ImportSpec, submatch bool) {
	for _, s := range f.Imports {
		impPath := importPath(s)
		if impPath != "" {
			if strings.HasPrefix(impPath, path) {
				return s, impPath != path
			}
		}
	}
	return nil, false
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
