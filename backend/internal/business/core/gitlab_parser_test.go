package core

import "testing"

func TestParseTaskID(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
		source string
		want   int64
	}{
		{"branch with key", "F5", "feature/f5-42-add-login", 42},
		{"branch numeric", "F5", "feature/128-refactor", 128},
		{"mr title", "F5", "Fix flaky tests [F5-7]", 7},
		{"commit message underscore", "TASK", "TASK_15 hotfix", 15},
		{"no key", "F5", "chore/update-deps", 0},
		{"other prefix ignored", "F5", "feature/AB-42-login", 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseTaskID(tc.prefix, tc.source); got != tc.want {
				t.Fatalf("ParseTaskID(%q, %q) = %d, want %d", tc.prefix, tc.source, got, tc.want)
			}
		})
	}
}

func TestBuildBranchName(t *testing.T) {
	got := BuildBranchName("F5", 42, "Починить авторизацию")
	want := "feature/f5-42-pochinit-avtorizaciyu"
	if got != want {
		t.Fatalf("BuildBranchName() = %q, want %q", got, want)
	}

	if got := BuildBranchName("", 3, "!!!"); got != "feature/f5-3" {
		t.Fatalf("BuildBranchName() fallback = %q", got)
	}
}

func TestNormalizeEventType(t *testing.T) {
	cases := map[string]string{
		"Push Hook":          eventPush,
		"merge_request":      eventMergeRequest,
		"Pipeline Hook":      eventPipeline,
		"Note Hook":          eventNote,
		"Deployment Hook":    "",
		"Merge Request Hook": eventMergeRequest,
	}

	for header, want := range cases {
		if got := normalizeEventType(header); got != want {
			t.Fatalf("normalizeEventType(%q) = %q, want %q", header, got, want)
		}
	}
}
