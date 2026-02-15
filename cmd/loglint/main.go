package main

import (
	loglint "github.com/RomanKovalev007/log_linter/loglint"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(loglint.Analyzer)
}
