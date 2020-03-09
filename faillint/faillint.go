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

const (
	// ignoreKey is used in a faillint directive to ignore a line-based problem.
	ignoreKey     = "ignore"
	// fileIgnoreKey is used in a faillint directive to ignore a whole file.
	fileIgnoreKey = "file-ignore"
)

// pathsRegexp represents a regexp that is used to parse -paths flag.
// It parses flag content in set of 3 subgroups:
//
// * import: Mandatory part. Go import path in URL format to be unwanted or have unwanted declarations.
// * declarations: Optional declarations in `{ }`. If set, using the import is allowed expect give declarations.
// * suggestion: Optional suggestion to print when unwanted import or declaration is found.
//
var pathsRegexp = regexp.MustCompile(`(?P<import>[\w/.-]+[\w])(\.?{(?P<declarations>[\w-,]+)}|)(=(?P<suggestion>[\w/.-]+[\w](\.?{[\w-,]+}|))|)`)

type faillint struct {
	paths       string // -paths flag
	ignoretests bool   // -ignore-tests flag
}

// Analyzer is a global instance of the linter.
// DEPRECATED: Use faillint.New instead.
var Analyzer = NewAnalyzer()

// NewAnalyzer create a faillint analyzer.
func NewAnalyzer() *analysis.Analyzer {
	f := faillint{
		paths:       "",
		ignoretests: false,
	}
	a := &analysis.Analyzer{
		Name:             "faillint",
		Doc:              "Report unwanted import path or exported declaration usages",
		Run:              f.run,
		RunDespiteErrors: true,
	}

	a.Flags.StringVar(&f.paths, "paths", "", `import paths or exported declarations (i.e: functions, constant, types or variables) to fail.
E.g. errors=github.com/pkg/errors,fmt.{Errorf}=github.com/pkg/errors.{Errorf},fmt.{Println,Print,Printf},github.com/prometheus/client_golang/prometheus.{DefaultGatherer,MustRegister}`)
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
	imp   string
	decls []string
	sugg  string
}

func parsePaths(paths string) []path {
	pathGroups := pathsRegexp.FindAllStringSubmatch(trimAllWhitespaces(paths), -1)

	parsed := make([]path, 0, len(pathGroups))
	for _, group := range pathGroups {
		p := path{}
		for i, name := range pathsRegexp.SubexpNames() {
			switch name {
			case "import":
				p.imp = group[i]
			case "suggestion":
				p.sugg = group[i]
			case "declarations":
				if group[i] == "" {
					break
				}
				p.decls = strings.Split(group[i], ",")
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
		if hasDirective(pass, file.Doc, fileIgnoreKey) {
			continue
		}
		commentMap := ast.NewCommentMap(pass.Fset, file, file.Comments)
		for _, path := range parsePaths(f.paths) {
			specs := importSpec(file, path.imp)
			if len(specs) == 0 {
				continue
			}
			for _, spec := range specs {
				if usageHasDirective(pass, commentMap, spec, spec.Pos(), ignoreKey) {
					continue
				}
				usages := importUsages(pass, commentMap, file, spec)
				if len(usages) == 0 {
					continue
				}

				if _, ok := usages[unspecifiedUsage]; ok || len(path.decls) == 0 {
					// File using unwanted import. Report.
					msg := fmt.Sprintf("package %q shouldn't be imported", importPath(spec))
					if path.sugg != "" {
						msg += fmt.Sprintf(", suggested: %q", path.sugg)
					}
					pass.Reportf(spec.Path.Pos(), msg)
					continue
				}

				// Not all usages are forbidden. Report only unwanted declarations.
				for _, declaration := range path.decls {
					positions, ok := usages[declaration]
					if !ok {
						continue
					}
					msg := fmt.Sprintf("declaration %q from package %q shouldn't be used", declaration, importPath(spec))
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

// importUsages reports all exported declaration used for a given import.
func importUsages(pass *analysis.Pass, commentMap ast.CommentMap, f *ast.File, spec *ast.ImportSpec) map[string][]token.Pos {
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
			if usageHasDirective(pass, commentMap, n, sel.Sel.NamePos, ignoreKey) {
				return true
			}
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

func parseDirective(s string) (option string, reason string) {
	if !strings.HasPrefix(s, "//faillint:") {
		return "", ""
	}
	s = strings.TrimPrefix(s, "//faillint:")
	fields := strings.SplitN(s, " ", 2)
	switch len(fields) {
	case 0:
		return "", ""
	case 1:
		return fields[0], ""
	default:
		return fields[0], fields[1]
	}
}

func hasDirective(pass *analysis.Pass, cg *ast.CommentGroup, option string) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		o, reason := parseDirective(c.Text)
		if (o == ignoreKey || o == fileIgnoreKey) && reason == "" {
			pass.Reportf(c.Pos(), "missing reason on faillint directive")
		}
		if o == fileIgnoreKey && option == ignoreKey {
			pass.Reportf(c.Pos(), "%s option on faillint directive must be in package docs", fileIgnoreKey)
		}
		if o == option {
			return true
		}
	}
	return false
}

func usageHasDirective(pass *analysis.Pass, cm ast.CommentMap, n ast.Node, p token.Pos, option string) bool {
	for _, cg := range cm[n] {
		if hasDirective(pass, cg, ignoreKey) {
			return true
		}
	}
	// Try to find an "enclosing" node which the ast.CommentMap will
	// thus have associated comments to this field selector.
	for node := range cm {
		if p >= node.Pos() && p <= node.End() {
			for _, cg := range cm[node] {
				if hasDirective(pass, cg, option) {
					return true
				}
			}
		}
	}
	return false
}
