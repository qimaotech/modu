package gitproxy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"codeup.aliyun.com/qimao/public/devops/modu/internal/errors"
)

// GitProxy Git 操作真实实现
type GitProxy struct{}

// New 创建 Git 代理
func New() GitClient {
	return &GitProxy{}
}

// Clone 克隆仓库
func (g *GitProxy) Clone(ctx context.Context, url, path string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", url, path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git clone] failed to clone %s to %s: %w, output: %s", url, path, errors.ErrGitExec, string(out))
	}
	return nil
}

// CreateWorktree 创建工作树
func (g *GitProxy) CreateWorktree(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error {
	// 先 fetch 获取最新
	if err := g.Fetch(ctx, repoPath); err != nil {
		return err
	}

	// 创建 worktree
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "worktree", "add", "-b", branch, worktreePath, baseBranch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git worktree add] failed to create worktree at %s: %w, output: %s", worktreePath, errors.ErrGitExec, string(out))
	}
	return nil
}

// GetStatus 获取目录状态
func (g *GitProxy) GetStatus(ctx context.Context, path string) (Status, error) {
	// 检查目录是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Status{}, fmt.Errorf("[git status] path does not exist: %s, %w", path, errors.ErrModuleNotFound)
	}

	cmd := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Status{}, fmt.Errorf("[git status] failed to get status for %s: %w, output: %s", path, errors.ErrGitExec, string(out))
	}

	return parseStatus(ctx, string(out), path)
}

// RemoveWorktree 删除工作树
func (g *GitProxy) RemoveWorktree(ctx context.Context, path string) error {
	// 先用 git worktree remove 移除
	cmd := exec.CommandContext(ctx, "git", "worktree", "remove", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// 如果 worktree remove 失败，尝试直接删除目录
		if rmErr := os.RemoveAll(path); rmErr != nil {
			return fmt.Errorf("[git worktree remove] failed to remove worktree at %s: %w, output: %s", path, errors.ErrGitExec, string(out))
		}
		return nil
	}
	return nil
}

// ListWorktrees 列出所有工作树
func (g *GitProxy) ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "worktree", "list", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("[git worktree list] failed: %w, output: %s", errors.ErrGitExec, string(out))
	}

	return parseWorktreeList(string(out))
}

// Fetch 从远程获取最新
func (g *GitProxy) Fetch(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "fetch", "--all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git fetch] failed to fetch in %s: %w, output: %s", repoPath, errors.ErrGitExec, string(out))
	}
	return nil
}

// BranchExists 检查分支是否存在
func (g *GitProxy) BranchExists(ctx context.Context, repoPath, branch string) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "--verify", branch)
	return cmd.Run() == nil
}

// parseStatus 解析 git status --porcelain 输出
func parseStatus(ctx context.Context, output, path string) (Status, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	files := make([]FileStatus, 0, len(lines))
	isDirty := false

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		status := rune(line[0])
		name := strings.TrimSpace(line[3:])
		if name == "" {
			name = line[2:]
		}

		// 非空状态表示有变更
		if status != ' ' && status != '?' {
			isDirty = true
		}
		// ?? 表示未跟踪文件，也是脏
		if line[0] == '?' && line[1] == '?' {
			isDirty = true
		}

		files = append(files, FileStatus{
			Name:   name,
			Status: status,
		})
	}

	// 获取当前分支
	branch := ""
	cmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err == nil {
		branch = strings.TrimSpace(string(out))
	}

	return Status{
		IsDirty: isDirty,
		Branch:  branch,
		Files:   files,
	}, nil
}

// parseWorktreeList 解析 git worktree list 输出
func parseWorktreeList(output string) ([]WorktreeInfo, error) {
	var worktrees []WorktreeInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	var current WorktreeInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			current.Path = path
		} else if strings.HasPrefix(line, "HEAD ") {
			// HEAD 行不需要处理
		} else if branch, ok := strings.CutPrefix(line, "branch refs/heads/"); ok {
			current.Branch = branch
		} else if line == "" {
			// 空行表示一个 worktree 结束
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = WorktreeInfo{}
			}
		}
	}

	// 处理最后一个
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// ExecGit 执行 git 命令并返回输出
func ExecGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if err != nil {
		return output, fmt.Errorf("git %v failed: %w, stderr: %s", args, err, stderr.String())
	}
	return output, nil
}
