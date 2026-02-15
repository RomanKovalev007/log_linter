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


// containsSensitiveKeyword returns true if any of the values contains a sensitive keyword.
func containsSensitiveKeyword(values []string, keywords []string) bool {
	for _, val := range values {
		lower := strings.ToLower(val)
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return true
			}
		}
	}
	return false
}

// litValues extracts string values from BasicLit nodes.
func litValues(lits []*ast.BasicLit) []string {
	result := make([]string, 0, len(lits))
	for _, lit := range lits {
		val, err := strconv.Unquote(lit.Value)
		if err == nil {
			result = append(result, val)
		}
	}
	return result
}


func checkSensitiveData(pass *analysis.Pass, expr ast.Expr, keywords []string) {
	binExpr, ok := expr.(*ast.BinaryExpr)
	if !ok || binExpr.Op != token.ADD {
		return
	}

	if !hasNonLiteralParts(expr) {
		return
	}

	lits := collectLits(expr)
	if containsSensitiveKeyword(litValues(lits), keywords) {
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

// collectLits returns all string BasicLit nodes from the expression tree.
func collectLits(expr ast.Expr) []*ast.BasicLit {
	var lits []*ast.BasicLit
	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch node := e.(type) {
		case *ast.BasicLit:
			if node.Kind == token.STRING {
				lits = append(lits, node)
			}
		case *ast.BinaryExpr:
			if node.Op == token.ADD {
				walk(node.X)
				walk(node.Y)
			}
		}
	}
	walk(expr)
	return lits
}


// toLowercaseStart returns the message with the first letter lowercased.
func toLowercaseStart(msg string) string {
	if len(msg) == 0 {
		return msg
	}
	r, size := utf8.DecodeRuneInString(msg)
	return string(unicode.ToLower(r)) + msg[size:]
}

// stripSpecialChars removes all characters that are not letters, digits or spaces.
func stripSpecialChars(msg string) string {
	var b strings.Builder
	for _, r := range msg {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// suggestedFix creates a SuggestedFix that replaces a BasicLit with newText.
func suggestedFix(message string, lit *ast.BasicLit, newText string) []analysis.SuggestedFix {
	return []analysis.SuggestedFix{{
		Message: message,
		TextEdits: []analysis.TextEdit{{
			Pos:     lit.Pos(),
			End:     lit.End(),
			NewText: []byte(strconv.Quote(newText)),
		}},
	}}
}
