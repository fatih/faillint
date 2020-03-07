package faillint

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}, v ...interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)

		var msg string
		if len(v) > 0 {
			msg = fmt.Sprintf(v[0].(string), v[1:]...)
		}

		fmt.Printf("\033[31m%s:%d:"+msg+"\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func TestParsePaths(t *testing.T) {
	for _, tcase := range []struct {
		paths    string
		expected []path
	}{
		{
			paths:    "",
			expected: []path{},
		},
		{
			paths: "errors",
			expected: []path{
				{imp: "errors"},
			},
		},
		{
			// No dedup ):
			paths: "errors,errors",
			expected: []path{
				{imp: "errors"},
				{imp: "errors"},
			},
		},
		{
			paths: "errors=github.com/pkg/errors",
			expected: []path{
				{imp: "errors", sugg: "github.com/pkg/errors"},
			},
		},
		{
			paths: "fmt.{Errorf}=github.com/pkg/errors",
			expected: []path{
				{imp: "fmt", decls: []string{"Errorf"}, sugg: "github.com/pkg/errors"},
			},
		},
		{
			paths: "fmt.{Errorf}=github.com/pkg/errors.{Errorf}",
			expected: []path{
				{imp: "fmt", decls: []string{"Errorf"}, sugg: "github.com/pkg/errors.{Errorf}"},
			},
		},
		{
			paths: "fmt.{Errorf,AnotherFunction}=github.com/pkg/errors.{Errorf,AnotherFunction}",
			expected: []path{
				{imp: "fmt", decls: []string{"Errorf", "AnotherFunction"}, sugg: "github.com/pkg/errors.{Errorf,AnotherFunction}"},
			},
		},
		{
			// Whitespace madness.
			paths: "fmt.{Errorf, AnotherFunction}    =   github.com/pkg/errors.{ Errorf,  AnotherFunction}",
			expected: []path{
				{imp: "fmt", decls: []string{"Errorf", "AnotherFunction"}, sugg: "github.com/pkg/errors.{Errorf,AnotherFunction}"},
			},
		},
		{
			// Without dot it works too.
			paths: "fmt{Errorf,AnotherFunction}=github.com/pkg/errors.{Errorf,AnotherFunction}",
			expected: []path{
				{imp: "fmt", decls: []string{"Errorf", "AnotherFunction"}, sugg: "github.com/pkg/errors.{Errorf,AnotherFunction}"},
			},
		},
		{
			// Suggestion without { }.
			paths: "fmt.{Errorf}=github.com/pkg/errors.Errorf",
			expected: []path{
				{imp: "fmt", decls: []string{"Errorf"}, sugg: "github.com/pkg/errors.Errorf"},
			},
		},
		{
			// TODO(bwplotka): This might be unexpected for users. Probably detect & error (fmt.Errorf) can be valid repository name though.
			paths: "fmt.Errorf=github.com/pkg/errors.Errorf",
			expected: []path{
				{imp: "fmt.Errorf", sugg: "github.com/pkg/errors.Errorf"},
			},
		},
		{
			paths: "foo,github.com/foo/bar,github.com/foo/bar/foo.{A}=github.com/foo/bar/bar.{C},github.com/foo/bar/foo.{D,C}",
			expected: []path{
				{imp: "foo"},
				{imp: "github.com/foo/bar"},
				{imp: "github.com/foo/bar/foo", decls: []string{"A"}, sugg: "github.com/foo/bar/bar.{C}"},
				{imp: "github.com/foo/bar/foo", decls: []string{"D", "C"}},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			equals(t, tcase.expected, parsePaths(tcase.paths))
		})
	}
}

func TestRun(t *testing.T) {
	testdata := analysistest.TestData()

	for _, tcase := range []struct {
		name  string
		dir   string
		paths string

		ignoreTestFiles bool
	}{
		{
			name:  "unwanted errors package present",
			dir:   "a",
			paths: "errors",
		},
		{
			name:  "malformed suggestion, should still work",
			dir:   "a",
			paths: "errors=", // malformed suggestion
		},
		{
			name:  "unwanted errors package present with package rename",
			dir:   "a_with_name",
			paths: "errors",
		},
		{
			name:  "unwanted errors package present even if not used",
			dir:   "a_all",
			paths: "errors",
		},
		{
			name:  "unwanted errors package under different names",
			dir:   "a_many_names",
			paths: "errors",
		},
		{
			name: "no unwanted path specified",
			// Nothing unwanted, so expect no diagnosis.
			dir:   "b",
			paths: "",
		},
		{
			name:  "empty file, no import",
			dir:   "c",
			paths: "errors",
		},
		{
			name:  "unwanted package with suggestion",
			dir:   "d",
			paths: "errors=github.com/pkg/errors",
		},
		{
			name:  "multiple unwanted packages with suggestions",
			dir:   "e",
			paths: "errors=github.com/pkg/errors,golang.org/x/net/context=context",
		},
		{
			name:  "unwanted package with dot in name with suggestion",
			dir:   "f",
			paths: "github.com/fatih/faillint/faillint/testdata/src/fake.NotFunction=not_fake",
		},
		{
			name:  "unwanted functions",
			dir:   "g",
			paths: "fmt.{Errorf}=github.com/pkg/errors.{Errorf},fmt.{Println,Print,Printf}",
		},
		{
			name:  "unwanted functions with multiple on the same line",
			dir:   "g_complex",
			paths: "fmt.{Errorf}=github.com/pkg/errors.{Errorf},fmt.{Println,Print,Printf}",
		},
		{
			name:  "unwanted functions with package rename",
			dir:   "g_with_name",
			paths: "fmt.{Errorf}=github.com/pkg/errors.{Errorf},fmt.{Println,Print,Printf}",
		},
		{
			name:            "unwanted package with suggestion not ignored even if they are tests",
			dir:             "h",
			paths:           "errors=github.com/pkg/errors",
			ignoreTestFiles: false,
		},
		{
			name:            "unwanted package with suggestion ignored as they are tests and flag ignore specified",
			dir:             "h_ignore",
			paths:           "errors=github.com/pkg/errors",
			ignoreTestFiles: true,
		},
		{
			name:  "unwanted constant, variables and types",
			dir:   "i",
			paths: "os.{O_RDONLY,ErrNotExist,File}",
		},
	} {
		t.Run(tcase.name, func(t *testing.T) {
			f := New()
			f.Flags.Set("paths", tcase.paths)
			if tcase.ignoreTestFiles {
				f.Flags.Set("ignore-tests", "true")
			}

			// No assertion on result is required as 'analysistest' is for that.
			// All expected diagnosis should be specified by comment in affected file starting with `// want`.
			_ = analysistest.Run(t, testdata, f, tcase.dir)
		})
	}
}
