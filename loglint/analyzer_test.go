package loglint_test

import (
	loglint "github.com/RomanKovalev007/log_linter/loglint"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, loglint.Analyzer, "testcases")
}
