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
		imps := usesImports(file, imports)
		if len(imps) == 0 {
			continue
		}

		for _, imp := range imps {
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

// usesImports reports whether a given import is used.
func usesImports(f *ast.File, paths []string) []*ast.ImportSpec {
	specs := importSpecs(f, paths)
	if len(specs) == 0 {
		return nil
	}

	var result []*ast.ImportSpec
	names := make(map[string]*ast.ImportSpec, 0)

	for path, spec := range specs {
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
			result = append(result, spec)
			continue
		}

		// we're going to check them individually
		names[name] = spec
	}

	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		for name, spec := range names {
			if isTopName(sel.X, name) {
				result = append(result, spec)
			}
		}

		return true
	})

	return result
}

// importSpecs returns the import specs if f imports the given paths,
// or nil otherwise.
func importSpecs(f *ast.File, paths []string) map[string]*ast.ImportSpec {
	hasImport := make(map[string]*ast.ImportSpec, 0)
	for _, s := range f.Imports {
		hasImport[importPath(s)] = s
	}

	specs := make(map[string]*ast.ImportSpec, 0)
	for _, path := range paths {
		spec, ok := hasImport[path]
		if !ok {
			continue
		}

		specs[path] = spec
	}

	return specs
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
