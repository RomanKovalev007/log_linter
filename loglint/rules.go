package loglint

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

// isUppercaseStart returns true if the message starts with an uppercase letter.
func isUppercaseStart(msg string) bool {
	if len(msg) == 0 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(msg)
	return unicode.IsUpper(r)
}

// hasNonEnglish returns true if the message contains non-ASCII letters.
func hasNonEnglish(msg string) bool {
	for _, r := range msg {
		if unicode.IsLetter(r) && r > unicode.MaxASCII {
			return true
		}
	}
	return false
}

// hasSpecialChars returns true if the message contains special characters or emoji.
// Allowed characters: letters, digits and spaces.
func hasSpecialChars(msg string) bool {
	for _, r := range msg {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != ' ' {
			return true
		}
	}
	return false
}

// containsSensitiveKeyword returns true if any literal contains a sensitive keyword.
func containsSensitiveKeyword(literals []string, keywords []string) bool {
	for _, lit := range literals {
		lower := strings.ToLower(lit)
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return true
			}
		}
	}
	return false
}

func checkLowercaseStart(pass *analysis.Pass, expr ast.Expr, msg string) {
	if isUppercaseStart(msg) {
		pass.Reportf(expr.Pos(), "log message should start with a lowercase letter")
	}
}

func checkEnglishOnly(pass *analysis.Pass, expr ast.Expr, msg string) bool {
	if hasNonEnglish(msg) {
		pass.Reportf(expr.Pos(), "log message should be in English only")
		return true
	}
	return false
}

func checkNoSpecialChars(pass *analysis.Pass, expr ast.Expr, msg string) bool {
	if hasSpecialChars(msg) {
		pass.Reportf(expr.Pos(), "log message should not contain special characters or emoji")
		return true
	}
	return false
}

func checkSensitiveData(pass *analysis.Pass, expr ast.Expr, keywords []string) {
	binExpr, ok := expr.(*ast.BinaryExpr)
	if !ok || binExpr.Op != token.ADD {
		return
	}

	if !hasNonLiteralParts(expr) {
		return
	}

	if containsSensitiveKeyword(collectStringLiterals(expr), keywords) {
		pass.Reportf(expr.Pos(), "log message should not contain sensitive data")
	}
}

// check if expr has non literal parts
func hasNonLiteralParts(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return false
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			return hasNonLiteralParts(e.X) || hasNonLiteralParts(e.Y)
		}
	}
	return true
}

// create slice of all string literals
func collectStringLiterals(expr ast.Expr) []string {
	var result []string
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			val, err := strconv.Unquote(e.Value)
			if err == nil {
				result = append(result, val)
			}
		}
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			result = append(result, collectStringLiterals(e.X)...)
			result = append(result, collectStringLiterals(e.Y)...)
		}
	}
	return result
}
