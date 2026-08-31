package gitproxy

import (
	"context"
	stderrors "errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	moduerrors "github.com/qimaotech/modu/internal/errors"
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

// TestFetch_CanceledContextPreservesCause 验证 fetch 错误同时保留业务错误码和上下文原因。
func TestFetch_CanceledContextPreservesCause(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	g := &GitProxy{}
	err := g.Fetch(ctx, t.TempDir())
	if err == nil {
		t.Fatal("Fetch() error = nil, want canceled error")
	}
	if !stderrors.Is(err, moduerrors.ErrGitExec) {
		t.Fatalf("Fetch() error = %v, want ErrGitExec", err)
	}
	if !stderrors.Is(err, context.Canceled) {
		t.Fatalf("Fetch() error = %v, want context.Canceled", err)
	}
}

// TestCreateWorktreeFromRemoteBranch_FetchesRestrictedRefspec 验证受限 refspec 下仍能拉取目标分支并创建 tracking worktree。
func TestCreateWorktreeFromRemoteBranch_FetchesRestrictedRefspec(t *testing.T) {
	unsetRepositoryGitEnvironment(t)
	remotePath, repoPath := setupRestrictedRefspecRepo(t)
	worktreePath := filepath.Join(filepath.Dir(repoPath), "feature-worktree")

	verifyCmd := exec.Command("git", "-C", repoPath, "rev-parse", "--verify", "refs/remotes/origin/feature/test")
	if err := verifyCmd.Run(); err == nil {
		t.Fatal("restricted clone unexpectedly contains origin/feature/test")
	}

	g := &GitProxy{}
	if err := g.CreateWorktreeFromRemoteBranch(context.Background(), repoPath, "feature/test", worktreePath); err != nil {
		t.Fatalf("CreateWorktreeFromRemoteBranch() error = %v, remote = %s", err, remotePath)
	}

	branchOut, err := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatalf("read worktree branch: %v, output: %s", err, string(branchOut))
	}
	if got := string(branchOut); got != "feature/test\n" {
		t.Fatalf("worktree branch = %q, want %q", got, "feature/test\n")
	}

	upstreamOut, err := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", "@{upstream}").CombinedOutput()
	if err != nil {
		t.Fatalf("read worktree upstream: %v, output: %s", err, string(upstreamOut))
	}
	if got := string(upstreamOut); got != "origin/feature/test\n" {
		t.Fatalf("worktree upstream = %q, want %q", got, "origin/feature/test\n")
	}
	if err := g.Fetch(context.Background(), repoPath); err != nil {
		t.Fatalf("Fetch() after adding branch refspec error = %v", err)
	}
}

// unsetRepositoryGitEnvironment 隔离 pre-commit 注入的仓库级 Git 环境变量。
func unsetRepositoryGitEnvironment(t *testing.T) {
	t.Helper()

	for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE"} {
		value, exists := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		if exists {
			t.Cleanup(func() {
				if err := os.Setenv(key, value); err != nil {
					t.Errorf("restore %s: %v", key, err)
				}
			})
		}
	}
}

func TestGetBranchPushStatus_Pushed(t *testing.T) {
	repoPath := setupPushStatusRepo(t, "feature/test")

	g := &GitProxy{}
	status, err := g.GetBranchPushStatus(context.Background(), repoPath, "feature/test")
	if err != nil {
		t.Fatalf("GetBranchPushStatus() unexpected error: %v", err)
	}
	if !status.IsPushed {
		t.Fatalf("IsPushed = false, want true: %+v", status)
	}
	if status.AheadCount != 0 {
		t.Fatalf("AheadCount = %d, want 0", status.AheadCount)
	}
}

func TestGetBranchPushStatus_AheadOfRemote(t *testing.T) {
	repoPath := setupPushStatusRepo(t, "feature/test")
	writeCommit(t, repoPath, "local.txt", "local change", "local change")

	g := &GitProxy{}
	status, err := g.GetBranchPushStatus(context.Background(), repoPath, "feature/test")
	if err != nil {
		t.Fatalf("GetBranchPushStatus() unexpected error: %v", err)
	}
	if status.IsPushed {
		t.Fatalf("IsPushed = true, want false: %+v", status)
	}
	if status.AheadCount != 1 {
		t.Fatalf("AheadCount = %d, want 1", status.AheadCount)
	}
}

func TestGetBranchPushStatus_MissingRemoteBranch(t *testing.T) {
	repoPath := setupPushStatusRepo(t, "feature/test")
	runGit(t, repoPath, "checkout", "-b", "local-only")
	writeCommit(t, repoPath, "local-only.txt", "local only", "local only")

	g := &GitProxy{}
	status, err := g.GetBranchPushStatus(context.Background(), repoPath, "local-only")
	if err != nil {
		t.Fatalf("GetBranchPushStatus() unexpected error: %v", err)
	}
	if status.IsPushed {
		t.Fatalf("IsPushed = true, want false: %+v", status)
	}
	if status.Reason == "" {
		t.Fatalf("Reason should explain missing remote branch: %+v", status)
	}
}

func setupPushStatusRepo(t *testing.T, branch string) string {
	t.Helper()

	tmpDir := t.TempDir()
	remotePath := filepath.Join(tmpDir, "remote.git")
	repoPath := filepath.Join(tmpDir, "repo")

	runGit(t, "", "init", "--bare", remotePath)
	runGit(t, "", "clone", remotePath, repoPath)
	runGit(t, repoPath, "config", "user.name", "modu test")
	runGit(t, repoPath, "config", "user.email", "modu@example.com")
	runGit(t, repoPath, "checkout", "-b", branch)
	writeCommit(t, repoPath, "README.md", "initial", "initial")
	runGit(t, repoPath, "push", "-u", "origin", branch)

	return repoPath
}

// setupRestrictedRefspecRepo 创建仅默认拉取 main 分支的本地 clone。
func setupRestrictedRefspecRepo(t *testing.T) (string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	remotePath := filepath.Join(tmpDir, "remote.git")
	seedPath := filepath.Join(tmpDir, "seed")
	repoPath := filepath.Join(tmpDir, "repo")

	runGit(t, "", "init", "--bare", remotePath)
	runGit(t, "", "init", seedPath)
	runGit(t, seedPath, "config", "user.name", "modu test")
	runGit(t, seedPath, "config", "user.email", "modu@example.com")
	runGit(t, seedPath, "checkout", "-b", "main")
	writeCommit(t, seedPath, "README.md", "main", "initial main")
	runGit(t, seedPath, "remote", "add", "origin", remotePath)
	runGit(t, seedPath, "push", "-u", "origin", "main")
	runGit(t, seedPath, "checkout", "-b", "feature/test")
	writeCommit(t, seedPath, "feature.txt", "feature", "add feature")
	runGit(t, seedPath, "push", "-u", "origin", "feature/test")
	runGit(t, "", "clone", "--single-branch", "--branch", "main", remotePath, repoPath)

	return remotePath, repoPath
}

func writeCommit(t *testing.T, repoPath, fileName, content, message string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(repoPath, fileName), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoPath, "add", fileName)
	runGit(t, repoPath, "commit", "-m", message)
}

func runGit(t *testing.T, repoPath string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	if repoPath != "" {
		cmd.Dir = repoPath
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v, output: %s", args, err, string(out))
	}
}
