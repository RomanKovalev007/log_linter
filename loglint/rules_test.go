package loglint

import (
	"go/ast"
	"go/parser"
	"testing"
)

func mustParseExpr(t *testing.T, src string) ast.Expr {
	t.Helper()
	expr, err := parser.ParseExpr(src)
	if err != nil {
		t.Fatalf("failed to parse expression %q: %v", src, err)
	}
	return expr
}

func TestIsUppercaseStart(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{"uppercase ascii", "Starting server", true},
		{"lowercase ascii", "starting server", false},
		{"digit start", "123 items", false},
		{"empty string", "", false},
		{"uppercase cyrillic", "Запуск", true},
		{"lowercase cyrillic", "запуск", false},
		{"space start", " hello", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUppercaseStart(tt.msg)
			if got != tt.want {
				t.Errorf("isUppercaseStart(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestHasNonEnglish(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{"english only", "starting server on port 8080", false},
		{"cyrillic", "запуск сервера", true},
		{"mixed", "hello" + " мир", true},
		{"digits and spaces", "123 456", false},
		{"empty", "", false},
		{"chinese", "hello world 你好", true},
		{"special chars no letters", "!@#$%", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasNonEnglish(tt.msg)
			if got != tt.want {
				t.Errorf("hasNonEnglish(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestHasSpecialChars(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{"clean message", "server started", false},
		{"with digits", "port 8080", false},
		{"exclamation", "server started!", true},
		{"exclamation", "server" + " started!", true},
		{"colon", "warning: something", true},
		{"ellipsis", "loading...", true},
		{"emoji", "done \U0001F680", true},
		{"hyphen", "re-connect", true},
		{"underscore", "log_message", true},
		{"empty", "", false},
		{"only letters", "helloworld", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasSpecialChars(tt.msg)
			if got != tt.want {
				t.Errorf("hasSpecialChars(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestContainsSensitiveKeyword(t *testing.T) {
	tests := []struct {
		name     string
		literals []string
		want     bool
	}{
		{"password keyword", []string{"user password: "}, true},
		{"token keyword", []string{"token: "}, true},
		{"api_key keyword", []string{"api_key="}, true},
		{"apikey keyword", []string{"apikey "}, true},
		{"secret keyword", []string{"my secret "}, true},
		{"credential keyword", []string{"credential "}, true},
		{"bearer keyword", []string{"bearer "}, true},
		{"session_id keyword", []string{"session_id "}, true},
		{"clean message", []string{"user authenticated"}, false},
		{"empty", []string{}, false},
		{"case insensitive", []string{"PASSWORD: "}, true},
		{"keyword in second literal", []string{"value is ", "password="}, true},
		{"no keywords", []string{"hello ", " world"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsSensitiveKeyword(tt.literals, defaultSensitiveKeywords)
			if got != tt.want {
				t.Errorf("containsSensitiveKeyword(%v) = %v, want %v", tt.literals, got, tt.want)
			}
		})
	}
}

func TestCollectStringLiterals(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"single literal", `"hello"`, 1},
		{"concat two literals", `"hello" + "world"`, 2},
		{"concat with ident", `"hello" + x`, 1},
		{"identifier only", `x`, 0},
		{"triple concat", `"a" + x + "b"`, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := mustParseExpr(t, tt.src)
			got := collectStringLiterals(expr)
			if len(got) != tt.want {
				t.Errorf("collectStringLiterals(%q) returned %d literals, want %d", tt.src, len(got), tt.want)
			}
		})
	}
}

func TestHasNonLiteralParts(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{"single literal", `"hello"`, false},
		{"concat two literals", `"hello" + "world"`, false},
		{"concat with ident", `"hello" + x`, true},
		{"just ident", `x`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := mustParseExpr(t, tt.src)
			got := hasNonLiteralParts(expr)
			if got != tt.want {
				t.Errorf("hasNonLiteralParts(%q) = %v, want %v", tt.src, got, tt.want)
			}
		})
	}
}
