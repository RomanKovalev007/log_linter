package loglint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if !cfg.isLowercaseEnabled() {
		t.Error("lowercase should be enabled by default")
	}
	if !cfg.isEnglishOnlyEnabled() {
		t.Error("english_only should be enabled by default")
	}
	if !cfg.isNoSpecialEnabled() {
		t.Error("no_special_chars should be enabled by default")
	}
	if !cfg.isSensitiveDataEnabled() {
		t.Error("sensitive_data should be enabled by default")
	}
	if len(cfg.sensitiveKeywords()) == 0 {
		t.Error("sensitive keywords should not be empty by default")
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	cfg, err := loadConfig("")
	if err != nil {
		t.Fatalf("loadConfig empty path: %v", err)
	}
	if !cfg.isLowercaseEnabled() {
		t.Error("expected default config with all rules enabled")
	}
}

func TestLoadConfigDisableRules(t *testing.T) {
	content := `
rules:
  lowercase: false
  english_only: false
  no_special_chars: true
  sensitive_data: false
`
	path := writeTempFile(t, content)

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.isLowercaseEnabled() {
		t.Error("lowercase should be disabled")
	}
	if cfg.isEnglishOnlyEnabled() {
		t.Error("english_only should be disabled")
	}
	if !cfg.isNoSpecialEnabled() {
		t.Error("no_special_chars should be enabled")
	}
	if cfg.isSensitiveDataEnabled() {
		t.Error("sensitive_data should be disabled")
	}
}

func TestLoadConfigCustomKeywords(t *testing.T) {
	content := `
sensitive_keywords:
  - ssn
  - credit_card
`
	path := writeTempFile(t, content)

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	keywords := cfg.sensitiveKeywords()
	if len(keywords) != 2 {
		t.Fatalf("expected 2 keywords, got %d", len(keywords))
	}
	if keywords[0] != "ssn" || keywords[1] != "credit_card" {
		t.Errorf("unexpected keywords: %v", keywords)
	}
}

func TestLoadConfigPartial(t *testing.T) {
	content := `
rules:
  lowercase: false
`
	path := writeTempFile(t, content)

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.isLowercaseEnabled() {
		t.Error("lowercase should be disabled")
	}
	// unspecified rules should remain enabled by default
	if !cfg.isEnglishOnlyEnabled() {
		t.Error("english_only should be enabled when not specified")
	}
	if !cfg.isNoSpecialEnabled() {
		t.Error("no_special_chars should be enabled when not specified")
	}
	if !cfg.isSensitiveDataEnabled() {
		t.Error("sensitive_data should be enabled when not specified")
	}
}

func TestLoadConfigInvalidPath(t *testing.T) {
	_, err := loadConfig("/nonexistent/.loglint.yml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	path := writeTempFile(t, "{{invalid yaml")

	_, err := loadConfig(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".loglint.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}
