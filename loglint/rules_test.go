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

func TestToLowercaseStart(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"ascii", "Starting server", "starting server"},
		{"cyrillic", "Запуск", "запуск"},
		{"already lowercase", "starting", "starting"},
		{"empty", "", ""},
		{"single char", "A", "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toLowercaseStart(tt.msg)
			if got != tt.want {
				t.Errorf("toLowercaseStart(%q) = %q, want %q", tt.msg, got, tt.want)
			}
		})
	}
}

func TestStripSpecialChars(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"exclamation", "server started!", "server started"},
		{"multiple", "connection failed!!!", "connection failed"},
		{"colon", "warning: something went wrong", "warning something went wrong"},
		{"ellipsis", "loading...", "loading"},
		{"emoji", "done \U0001F680", "done "},
		{"clean", "server started", "server started"},
		{"empty", "", ""},
		{"only special", "!!!", ""},
		{"mixed", "hello-world_test.go", "helloworldtestgo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripSpecialChars(tt.msg)
			if got != tt.want {
				t.Errorf("stripSpecialChars(%q) = %q, want %q", tt.msg, got, tt.want)
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

func TestCollectLits(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		wantCount  int
		wantValues []string
	}{
		{"single literal", `"hello"`, 1, []string{"hello"}},
		{"concat two literals", `"hello" + "world"`, 2, []string{"hello", "world"}},
		{"concat with ident", `"hello" + x`, 1, []string{"hello"}},
		{"identifier only", `x`, 0, nil},
		{"triple concat", `"a" + x + "b"`, 2, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := mustParseExpr(t, tt.src)
			lits := collectLits(expr)
			if len(lits) != tt.wantCount {
				t.Errorf("collectLits(%q) returned %d literals, want %d", tt.src, len(lits), tt.wantCount)
			}
			if tt.wantValues != nil {
				values := litValues(lits)
				if len(values) != len(tt.wantValues) {
					t.Errorf("litValues returned %d values, want %d", len(values), len(tt.wantValues))
				}
				for i, v := range values {
					if i < len(tt.wantValues) && v != tt.wantValues[i] {
						t.Errorf("litValues[%d] = %q, want %q", i, v, tt.wantValues[i])
					}
				}
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
