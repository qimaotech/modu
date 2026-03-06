package gitproxy

import (
	"context"
	"testing"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		path           string
		wantIsDirty    bool
		wantBranch     string
		wantFileCount  int
	}{
		{
			name:          "clean working tree",
			output:        "",
			path:          "/test/repo",
			wantIsDirty:   false,
			wantBranch:    "main",
			wantFileCount: 0,
		},
		{
			name:          "modified files",
			output:        "M  README.md\n M go.mod",
			path:          "/test/repo",
			wantIsDirty:   true,
			wantBranch:    "main",
			wantFileCount: 2,
		},
		{
			name:          "untracked files",
			output:        "?? newfile.txt",
			path:          "/test/repo",
			wantIsDirty:   true,
			wantBranch:    "main",
			wantFileCount: 1,
		},
		{
			name:          "added files",
			output:        "A  new.go",
			path:          "/test/repo",
			wantIsDirty:   true,
			wantBranch:    "main",
			wantFileCount: 1,
		},
		{
			name:          "deleted files",
			output:        "D  old.go",
			path:          "/test/repo",
			wantIsDirty:   true,
			wantBranch:    "main",
			wantFileCount: 1,
		},
		{
			name:          "mixed status",
			output:        "M  modified.txt\nA  added.go\nD  deleted.go\n?? untracked.txt",
			path:          "/test/repo",
			wantIsDirty:   true,
			wantBranch:    "main",
			wantFileCount: 4,
		},
		{
			name:          "empty output",
			output:        "   ",
			path:          "/test/repo",
			wantIsDirty:   false,
			wantBranch:    "main",
			wantFileCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := parseStatus(context.Background(), tt.output, tt.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if status.IsDirty != tt.wantIsDirty {
				t.Errorf("IsDirty = %v, want %v", status.IsDirty, tt.wantIsDirty)
			}

			if len(status.Files) != tt.wantFileCount {
				t.Errorf("file count = %d, want %d", len(status.Files), tt.wantFileCount)
			}
		})
	}
}

func TestParseWorktreeList(t *testing.T) {
	tests := []struct {
		name            string
		output          string
		wantCount       int
		wantFirstPath   string
		wantFirstBranch string
	}{
		{
			name:            "empty",
			output:          "",
			wantCount:       0,
			wantFirstPath:   "",
			wantFirstBranch: "",
		},
		{
			name:            "single worktree",
			output:          "worktree /home/user/repos/main\nHEAD abcd1234\nbranch refs/heads/main\n\n",
			wantCount:       1,
			wantFirstPath:   "/home/user/repos/main",
			wantFirstBranch: "main",
		},
		{
			name:            "multiple worktrees",
			output: `worktree /home/user/repos/main
HEAD abcd1234
branch refs/heads/main

worktree /home/user/repos/feature-add-auth
HEAD defg5678
branch refs/heads/feature/add-auth

`,
			wantCount:       2,
			wantFirstPath:   "/home/user/repos/main",
			wantFirstBranch: "main",
		},
		{
			name:            "worktree without branch",
			output:          "worktree /home/user/repos/feature-test\nHEAD abcd1234\n\n",
			wantCount:       1,
			wantFirstPath:   "/home/user/repos/feature-test",
			wantFirstBranch: "",
		},
		{
			name:            "single line format",
			output:          "worktree /path/to/worktree\n",
			wantCount:       1,
			wantFirstPath:   "/path/to/worktree",
			wantFirstBranch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worktrees, err := parseWorktreeList(tt.output)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(worktrees) != tt.wantCount {
				t.Errorf("worktree count = %d, want %d", len(worktrees), tt.wantCount)
			}

			if tt.wantCount > 0 {
				if worktrees[0].Path != tt.wantFirstPath {
					t.Errorf("first path = %s, want %s", worktrees[0].Path, tt.wantFirstPath)
				}
				if worktrees[0].Branch != tt.wantFirstBranch {
					t.Errorf("first branch = %s, want %s", worktrees[0].Branch, tt.wantFirstBranch)
				}
			}
		})
	}
}
