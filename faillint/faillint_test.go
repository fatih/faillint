package faillint_test

import (
	"testing"

	"github.com/fatih/faillint/faillint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()

	var tests = []struct {
		name  string
		paths string
	}{
		{
			name:  "a",
			paths: "errors",
		},
		{
			name:  "b",
			paths: "",
		},
		{
			name:  "c",
			paths: "errors=", // malformed suggestion
		},
		{
			name:  "d",
			paths: "errors=github.com/pkg/errors",
		},
		{
			name:  "e",
			paths: "errors=github.com/pkg/errors,golang.org/x/net/context=context",
		},
	}
	for _, ts := range tests {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			faillint.Analyzer.Flags.Set("paths", ts.paths)
			analysistest.Run(t, testdata, faillint.Analyzer, ts.name)
		})
	}
}
