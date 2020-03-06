// Package faillint defines an Analyzer that fails when a package is imported
// that matches against a set of defined packages.
package faillint

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var (
	pathRegexp = regexp.MustCompile(`(?P<import>[\w/.-]+[\w])(\.?{(?P<functions>[\w-,]+)}|)(=(?P<suggestion>[\w/.-]+[\w](\.?{[\w-,]+}|))|)`)
)

type faillint struct {
	paths       string // -paths flag
	ignoretests bool   // -ignore-tests flag
}

// New create a faillint analyzer.
func New() *analysis.Analyzer {
	f := faillint{
		paths:       "",
		ignoretests: false,
	}
	a := &analysis.Analyzer{
		Name:             "faillint",
		Doc:              "Report unwanted import path, and function usages",
		Run:              f.run,
		RunDespiteErrors: true,
	}

	a.Flags.StringVar(&f.paths, "paths", "", `import paths, functions or methods to fail.
E.g: foo,github.com/foo/bar,github.com/foo/bar/foo.{A}=github.com/foo/bar/bar.{C},github.com/foo/bar/foo.{D,C}`)
	a.Flags.BoolVar(&f.ignoretests, "ignore-tests", false, "ignore all _test.go files")
	return a
}

func trimAllWhitespaces(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

type path struct {
	imp  string
	fn   []string
	sugg string
}

func parsePaths(paths string) []path {
	pathGroups := pathRegexp.FindAllStringSubmatch(trimAllWhitespaces(paths), -1)

	parsed := make([]path, 0, len(pathGroups))
	for _, group := range pathGroups {
		p := path{}
		for i, name := range pathRegexp.SubexpNames() {
			switch name {
			case "import":
				p.imp = group[i]
			case "suggestion":
				p.sugg = group[i]
			case "functions":
				if group[i] == "" {
					break
				}
				p.fn = strings.Split(group[i], ",")
			}
		}
		parsed = append(parsed, p)
	}
	return parsed
}

// run is the runner for an analysis pass.
func (f *faillint) run(pass *analysis.Pass) (interface{}, error) {
	if f.paths == "" {
		return nil, nil
	}
	for _, file := range pass.Files {
		if f.ignoretests && strings.Contains(pass.Fset.File(file.Package).Name(), "_test.go") {
			continue
		}
		for _, path := range parsePaths(f.paths) {
			specs := importSpec(file, path.imp)
			if len(specs) == 0 {
				continue
			}
			for _, spec := range specs {
				usages := importUsages(file, spec)
				if len(usages) == 0 {
					continue
				}

				if _, ok := usages[unspecifiedUsage]; ok || len(path.fn) == 0 {
					// File using unwanted import. Report.
					msg := fmt.Sprintf("package %q shouldn't be imported", importPath(spec))
					if path.sugg != "" {
						msg += fmt.Sprintf(", suggested: %q", path.sugg)
					}
					pass.Reportf(spec.Path.Pos(), msg)
					continue
				}

				// Not all usages are forbidden. Report only unwanted functions.
				for _, fn := range path.fn {
					positions, ok := usages[fn]
					if !ok {
						continue
					}
					msg := fmt.Sprintf("function %q from package %q shouldn't be used", fn, importPath(spec))
					if path.sugg != "" {
						msg += fmt.Sprintf(", suggested: %q", path.sugg)
					}
					for _, pos := range positions {
						pass.Reportf(pos, msg)
					}
				}
			}
		}
	}

	return nil, nil
}

const unspecifiedUsage = "unspecified"

// importUsages reports all usages of a given import.
func importUsages(f *ast.File, spec *ast.ImportSpec) map[string][]token.Pos {
	importRef := spec.Name.String()
	switch importRef {
	case "<nil>":
		importRef, _ = strconv.Unquote(spec.Path.Value)
		// If the package importRef is not explicitly specified,
		// make an educated guess. This is not guaranteed to be correct.
		lastSlash := strings.LastIndex(importRef, "/")
		if lastSlash != -1 {
			importRef = importRef[lastSlash+1:]
		}
	case "_", ".":
		// Not sure if this import is used - on the side of caution, report special "unspecified" usage.
		return map[string][]token.Pos{unspecifiedUsage: nil}
	}
	usages := map[string][]token.Pos{}

	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if isTopName(sel.X, importRef) {
			usages[sel.Sel.Name] = append(usages[sel.Sel.Name], sel.Sel.NamePos)
		}
		return true
	})
	return usages
}

// importSpecs returns all import specs for f import statements importing path.
func importSpec(f *ast.File, path string) (imports []*ast.ImportSpec) {
	for _, s := range f.Imports {
		if importPath(s) == path {
			imports = append(imports, s)
		}
	}
	return imports
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
