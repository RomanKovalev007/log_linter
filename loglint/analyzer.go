package loglint

import (
	"flag"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var slogMethods = map[string]int{
	"Debug": 0, "Info": 0, "Warn": 0, "Error": 0,
	"DebugContext": 1, "InfoContext": 1, "WarnContext": 1, "ErrorContext": 1,
	"Log": 2, "LogAttrs": 2,
}

var zapMethods = map[string]bool{
	"Debug": true, "Info": true, "Warn": true, "Error": true,
	"DPanic": true, "Panic": true, "Fatal": true,
	"Debugf": true, "Infof": true, "Warnf": true, "Errorf": true,
	"DPanicf": true, "Panicf": true, "Fatalf": true,
	"Debugw": true, "Infow": true, "Warnw": true, "Errorw": true,
	"DPanicw": true, "Panicw": true, "Fatalw": true,
	"Debugln": true, "Infoln": true, "Warnln": true, "Errorln": true,
	"DPanicln": true, "Panicln": true, "Fatalln": true,
}

var Analyzer = &analysis.Analyzer{
	Name:     "loglint",
	Doc:      "checks log messages for style and security issues",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

var configPath string

func init() {
	Analyzer.Flags = flag.FlagSet{}
	Analyzer.Flags.StringVar(&configPath, "config", "", "path to .loglint.yml config file")
}

func run(pass *analysis.Pass) (interface{}, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		obj := pass.TypesInfo.Uses[selector.Sel]
		if obj == nil {
			return
		}

		fn, ok := obj.(*types.Func)
		if !ok {
			return
		}

		pkg := fn.Pkg()
		if pkg == nil {
			return
		}

		pkgPath := pkg.Path()
		methodName := fn.Name()
		var msgIndex int

		switch pkgPath {
		case "log/slog":
			idx, found := slogMethods[methodName]
			if !found {
				return
			}
			msgIndex = idx
		case "go.uber.org/zap":
			if !zapMethods[methodName] {
				return
			}
			msgIndex = 0
		default:
			return
		}

		if msgIndex >= len(call.Args) {
			return
		}

		msgArg := call.Args[msgIndex]
		lits := collectLits(msgArg)
		values := litValues(lits)

		if cfg.isLowercaseEnabled() && len(values) > 0 && values[0] != "" {
			if isUppercaseStart(values[0]) {
				d := analysis.Diagnostic{
					Pos:     msgArg.Pos(),
					Message: "log message should start with a lowercase letter",
				}
				if lit, ok := msgArg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					fixed := toLowercaseStart(values[0])
					if cfg.isNoSpecialEnabled() && hasSpecialChars(fixed) {
						fixed = stripSpecialChars(fixed)
					}
					d.SuggestedFixes = suggestedFix("fix log message", lit, fixed)
				}
				pass.Report(d)
			}
		}

		if cfg.isEnglishOnlyEnabled() {
			for _, val := range values {
				if hasNonEnglish(val) {
					pass.Reportf(msgArg.Pos(), "log message should be in English only")
					break
				}
			}
		}

		if cfg.isNoSpecialEnabled() {
			for i, val := range values {
				if !hasSpecialChars(val) {
					continue
				}

				d := analysis.Diagnostic{
					Pos:     msgArg.Pos(),
					Message: "log message should not contain special characters or emoji",
				}

				needsFix := true
				if i == 0 && cfg.isLowercaseEnabled() && len(values) > 0 && isUppercaseStart(values[0]) {
					needsFix = false
				}

				if needsFix && i < len(lits) {
					d.SuggestedFixes = suggestedFix("remove special characters", lits[i], stripSpecialChars(val))
				}
				pass.Report(d)
				break
			}
		}

		if cfg.isSensitiveDataEnabled() {
			checkSensitiveData(pass, msgArg, cfg.sensitiveKeywords())
		}
	})

	return nil, nil
}

