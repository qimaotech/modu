package gitproxy

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRemoveWorktreeAndBranch_ReturnsWarningWithoutWritingStdout 验证警告由结果返回而不直接破坏 TUI 输出。
func TestRemoveWorktreeAndBranch_ReturnsWarningWithoutWritingStdout(t *testing.T) {
	indexFile, hadIndexFile := os.LookupEnv("GIT_INDEX_FILE")
	if err := os.Unsetenv("GIT_INDEX_FILE"); err != nil {
		t.Fatalf("清理外层 GIT_INDEX_FILE 失败: %v", err)
	}
	t.Cleanup(func() {
		var err error
		if hadIndexFile {
			err = os.Setenv("GIT_INDEX_FILE", indexFile)
		} else {
			err = os.Unsetenv("GIT_INDEX_FILE")
		}
		if err != nil {
			t.Errorf("恢复 GIT_INDEX_FILE 失败: %v", err)
		}
	})

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")
	worktreePath := filepath.Join(tmpDir, "feature-welfare-feeds-cpr")
	runGit(t, "", "init", repoPath)
	runGit(t, repoPath, "config", "user.name", "modu test")
	runGit(t, repoPath, "config", "user.email", "modu@example.com")
	writeCommit(t, repoPath, "README.md", "initial", "initial")
	runGit(t, repoPath, "worktree", "add", "-b", "feature/dynamic-coin", worktreePath)

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	originalStdout := os.Stdout
	os.Stdout = writer
	removeResult, removeErr := (&GitProxy{}).RemoveWorktreeAndBranch(
		context.Background(),
		repoPath,
		worktreePath,
		"feature-welfare-feeds-cpr",
	)
	closeErr := writer.Close()
	os.Stdout = originalStdout
	stdout, readErr := io.ReadAll(reader)
	_ = reader.Close()

	if removeErr != nil {
		t.Fatalf("RemoveWorktreeAndBranch() unexpected error: %v", removeErr)
	}
	if closeErr != nil || readErr != nil {
		t.Fatalf("读取 stdout 失败: close=%v read=%v", closeErr, readErr)
	}
	if len(stdout) != 0 {
		t.Fatalf("RemoveWorktreeAndBranch() 不应直接写 stdout，got %q", stdout)
	}
	if len(removeResult.Warnings) != 1 || !strings.Contains(removeResult.Warnings[0], "slug does not match") {
		t.Fatalf("应返回分支不匹配警告，got %+v", removeResult.Warnings)
	}
	if !(&GitProxy{}).BranchExists(context.Background(), repoPath, "feature/dynamic-coin") {
		t.Fatal("分支不匹配时应保留原分支")
	}
}

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
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// 使用无效的网络地址
	exists := g.RemoteBranchExists(ctx, "git@192.0.2.1:nonexistent/repo.git", "main")
	if exists {
		t.Error("expected false for network error")
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
	cmd.Env = gitTestEnvironment()
	if repoPath != "" {
		cmd.Dir = repoPath
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v, output: %s", args, err, string(out))
	}
}

// gitTestEnvironment 移除提交钩子注入的仓库环境，避免临时仓库和 linked worktree 误用外层索引。
func gitTestEnvironment() []string {
	repositoryVariables := map[string]struct{}{
		"GIT_COMMON_DIR":       {},
		"GIT_DIR":              {},
		"GIT_INDEX_FILE":       {},
		"GIT_OBJECT_DIRECTORY": {},
		"GIT_PREFIX":           {},
		"GIT_WORK_TREE":        {},
	}
	environment := make([]string, 0, len(os.Environ()))
	for _, item := range os.Environ() {
		name, _, _ := strings.Cut(item, "=")
		if _, blocked := repositoryVariables[name]; blocked {
			continue
		}
		environment = append(environment, item)
	}
	return environment
}
