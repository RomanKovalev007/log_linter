package main

import (
	loglint "github.com/RomanKovalev007/log_linter/loglint"

	"golang.org/x/tools/go/analysis"
)

type analyzerPlugin struct{}

func (*analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{loglint.Analyzer}
}

// AnalyzerPlugin is the entry point for golangci-lint.
var AnalyzerPlugin analyzerPlugin
