package main

import (
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema/internal/testutil"
)

func TestContains(t *testing.T) {
	t.Parallel()

	if !contains("hello world", "world") {
		t.Fatal("expected substring match")
	}
	if contains("hello", "world") {
		t.Fatal("unexpected substring match")
	}
}

func TestCategorizeErrorsInChinese_PrintsCategories(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, func() {
		categorizeErrorsInChinese(map[string]string{
			"name":  "长度至少为 3",
			"email": "格式无效",
			"age":   "最小值为 0",
		})
	})

	for _, want := range []string{"格式错误:", "范围错误:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}

func TestMain_PrintsMultilingualErrors(t *testing.T) {
	// No t.Parallel(): captures process-wide stdout.
	out := testutil.CaptureStdout(t, main)

	for _, want := range []string{
		"DetailedErrors Multilingual Support Demo",
		"English (Default):",
		"简体中文:",
		"日本語:",
		"Français:",
		"Deutsch:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q in %q", want, out)
		}
	}
}
