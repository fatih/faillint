package faillint_test

import (
	"testing"

	"github.com/fatih/faillint/faillint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()

	var tests = []struct {
		name        string
		paths       string
		ignoretests bool
	}{
		{
			name:        "a",
			paths:       "errors",
			ignoretests: false,
		},
		{
			name:        "b",
			paths:       "",
			ignoretests: false,
		},
		{
			name:        "c",
			paths:       "errors=", // malformed suggestion
			ignoretests: false,
		},
		{
			name:        "d",
			paths:       "errors=github.com/pkg/errors",
			ignoretests: false,
		},
		{
			name:        "e",
			paths:       "errors=github.com/pkg/errors,golang.org/x/net/context=context",
			ignoretests: false,
		},
		{
			name:        "f",
			paths:       "errors",
			ignoretests: true,
		},
		{
			name:        "g",
			paths:       "errors",
			ignoretests: true,
		},
		{
			name:        "h",
			paths:       "errors",
			ignoretests: false,
		},
	}
	for _, ts := range tests {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			a := faillint.NewAnalyzer()
			a.Flags.Set("paths", ts.paths)
			if ts.ignoretests {
				a.Flags.Set("ignore-tests", "true")
			}
			analysistest.Run(t, testdata, a, ts.name)
		})
	}
}
