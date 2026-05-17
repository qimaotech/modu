package gitproxy

import (
	"context"
	stderrors "errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	errs "github.com/qimaotech/modu/internal/errors"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name          string
		output        string
		path          string
		wantIsDirty   bool
		wantBranch    string
		wantFileCount int
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
			name: "multiple worktrees",
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

func TestRemoteBranchExists_BranchExists(t *testing.T) {
	// 创建临时目录作为测试仓库
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// 初始化一个 git 仓库
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "init", repoPath)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// 创建一个提交
	if err := os.WriteFile(filepath.Join(repoPath, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "-C", repoPath, "add", ".")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "initial")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// 创建一个分支
	cmd = exec.Command("git", "-C", repoPath, "branch", "test-branch")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	g := &GitProxy{}
	exists := g.RemoteBranchExists(context.Background(), repoPath, "test-branch")
	if !exists {
		t.Error("expected branch to exist")
	}
}

func TestRemoteBranchExists_BranchNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// 初始化一个 git 仓库
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "init", repoPath)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// 创建一个提交
	if err := os.WriteFile(filepath.Join(repoPath, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "-C", repoPath, "add", ".")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "initial")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	g := &GitProxy{}
	exists := g.RemoteBranchExists(context.Background(), repoPath, "nonexistent-branch")
	if exists {
		t.Error("expected branch to not exist")
	}
}

func TestRemoteBranchExists_RepoNotExists(t *testing.T) {
	g := &GitProxy{}
	exists := g.RemoteBranchExists(context.Background(), "/nonexistent/repo", "main")
	if exists {
		t.Error("expected false for nonexistent repo")
	}
}

func TestRemoteBranchExists_NetworkError(t *testing.T) {
	g := &GitProxy{}
	// 使用无效的网络地址
	exists := g.RemoteBranchExists(context.Background(), "git@192.0.2.1:nonexistent/repo.git", "main")
	if exists {
		t.Error("expected false for network error")
	}
}

func TestCreateWorktree_FetchFailureIsDistinguishable(t *testing.T) {
	repoPath := initGitProxyTestRepo(t)
	runGitProxyTestCommand(t, repoPath, "remote", "add", "origin", filepath.Join(t.TempDir(), "missing-remote"))

	g := &GitProxy{}
	err := g.CreateWorktree(context.Background(), repoPath, "feature-fetch-fails", "main", filepath.Join(t.TempDir(), "feature"))

	if err == nil {
		t.Fatal("expected fetch failure")
	}
	if !IsFetchError(err) {
		t.Fatalf("expected IsFetchError to be true, got %v", err)
	}
	if !stderrors.Is(err, errs.ErrGitExec) {
		t.Fatalf("expected ERR_GIT_EXEC wrapping, got %v", err)
	}
}

func TestCreateWorktreeNoFetch_CreatesNewBranchWithBrokenRemote(t *testing.T) {
	repoPath := initGitProxyTestRepo(t)
	worktreePath := filepath.Join(t.TempDir(), "feature")
	runGitProxyTestCommand(t, repoPath, "remote", "add", "origin", filepath.Join(t.TempDir(), "missing-remote"))

	g := &GitProxy{}
	err := g.CreateWorktreeNoFetch(context.Background(), repoPath, "feature/local", "main", worktreePath)

	if err != nil {
		t.Fatalf("CreateWorktreeNoFetch returned error: %v", err)
	}
	status, err := g.GetStatus(context.Background(), worktreePath)
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if status.Branch != "feature/local" {
		t.Fatalf("expected branch feature/local, got %s", status.Branch)
	}
}

func TestCreateWorktreeFromExistingBranchNoFetch_ReusesBranchWithBrokenRemote(t *testing.T) {
	repoPath := initGitProxyTestRepo(t)
	worktreePath := filepath.Join(t.TempDir(), "feature")
	runGitProxyTestCommand(t, repoPath, "branch", "feature/existing")
	runGitProxyTestCommand(t, repoPath, "remote", "add", "origin", filepath.Join(t.TempDir(), "missing-remote"))

	g := &GitProxy{}
	err := g.CreateWorktreeFromExistingBranchNoFetch(context.Background(), repoPath, "feature/existing", worktreePath)

	if err != nil {
		t.Fatalf("CreateWorktreeFromExistingBranchNoFetch returned error: %v", err)
	}
	status, err := g.GetStatus(context.Background(), worktreePath)
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if status.Branch != "feature/existing" {
		t.Fatalf("expected branch feature/existing, got %s", status.Branch)
	}
}

func initGitProxyTestRepo(t *testing.T) string {
	t.Helper()

	repoPath := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatal(err)
	}
	runGitProxyTestCommand(t, repoPath, "init", "-b", "main")
	runGitProxyTestCommand(t, repoPath, "config", "user.email", "test@example.com")
	runGitProxyTestCommand(t, repoPath, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGitProxyTestCommand(t, repoPath, "add", ".")
	runGitProxyTestCommand(t, repoPath, "commit", "-m", "initial")
	return repoPath
}

func runGitProxyTestCommand(t *testing.T, repoPath string, args ...string) {
	t.Helper()

	cmdArgs := args
	if len(args) == 0 || args[0] != "init" {
		cmdArgs = append([]string{"-C", repoPath}, args...)
	} else {
		cmdArgs = append(append([]string{}, args...), repoPath)
	}
	cmd := exec.Command("git", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v, output: %s", cmdArgs, err, string(out))
	}
}
