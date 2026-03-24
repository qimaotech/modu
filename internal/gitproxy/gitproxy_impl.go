package gitproxy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/qimaotech/modu/internal/errors"
	"github.com/qimaotech/modu/internal/logger"
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

// CreateWorktreeFromExistingBranch 从现有分支创建 worktree（不创建新分支）
func (g *GitProxy) CreateWorktreeFromExistingBranch(ctx context.Context, repoPath, branch, worktreePath string) error {
	// 先 fetch 获取最新
	if err := g.Fetch(ctx, repoPath); err != nil {
		return err
	}

	// 从现有分支创建 worktree（不带 -b 参数）
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "worktree", "add", worktreePath, branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git worktree add] failed to create worktree from branch %s at %s: %w, output: %s", branch, worktreePath, errors.ErrGitExec, string(out))
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

// branchToFeatureDirSlug 与 engine.featureToDirName 一致：分支名 -> feature 目录 slug
func branchToFeatureDirSlug(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

// RemoveWorktreeAndBranch 删除 worktree；仅当当前检出分支的 slug 与 featureDirName 一致时才删除该分支
func (g *GitProxy) RemoveWorktreeAndBranch(ctx context.Context, repoPath, worktreePath, featureDirName string) error {
	logger.Debug("RemoveWorktreeAndBranch: repo=%s, featureDirName=%s, path=%s", repoPath, featureDirName, worktreePath)

	status, err := g.GetStatus(ctx, worktreePath)
	branchToDelete := ""
	if err != nil {
		logger.Warn("无法读取 worktree 分支状态，跳过删除分支: path=%s, err=%v", worktreePath, err)
		fmt.Printf("Warning: skip branch delete (cannot read status): %s\n", worktreePath)
	} else {
		b := strings.TrimSpace(status.Branch)
		if b == "" || b == "HEAD" {
			logger.Warn("worktree 无有效分支名（detached HEAD 等），跳过删除分支: path=%s", worktreePath)
			fmt.Printf("Warning: skip branch delete (detached or unknown HEAD): %s\n", worktreePath)
		} else if branchToFeatureDirSlug(b) != featureDirName {
			logger.Warn("当前分支 %s（slug=%s）与 feature 目录名 %s 不一致，跳过删除分支", b, branchToFeatureDirSlug(b), featureDirName)
			fmt.Printf("Warning: skip branch delete: branch %q slug does not match feature dir %q\n", b, featureDirName)
		} else {
			branchToDelete = b
		}
	}

	// 先用 git worktree remove 移除
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "worktree", "remove", "--force", worktreePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Warn("git worktree remove 失败，尝试直接删除目录: path=%s, error=%s", worktreePath, string(out))
		// 如果 worktree remove 失败，尝试直接删除目录
		if rmErr := os.RemoveAll(worktreePath); rmErr != nil {
			logger.Error("删除目录失败: path=%s, error=%v", worktreePath, rmErr)
			return fmt.Errorf("[git worktree remove] failed to remove worktree at %s: %w, output: %s", worktreePath, errors.ErrGitExec, string(out))
		}
		logger.Info("直接删除目录成功: %s", worktreePath)
	}

	// 先 prune 清理过期的 worktree 引用
	cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "worktree", "prune")
	if err := cmd.Run(); err != nil {
		logger.Warn("git worktree prune 失败: %v", err)
	} else {
		logger.Info("git worktree prune 成功")
	}

	// 再删除对应的分支（仅在与目录 slug 一致时）
	if branchToDelete != "" {
		logger.Info("删除分支: repo=%s, branch=%s", repoPath, branchToDelete)
		cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "branch", "-D", branchToDelete)
		out, err = cmd.CombinedOutput()
		if err != nil {
			logger.Warn("删除分支失败（可能不存在）: branch=%s, error=%s", branchToDelete, string(out))
			fmt.Printf("Warning: failed to delete branch %s: %s\n", branchToDelete, string(out))
		} else {
			logger.Info("删除分支成功: %s", branchToDelete)
		}
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

// Rebase 在当前路径下执行 fetch 后 rebase origin/<当前分支>
func (g *GitProxy) Rebase(ctx context.Context, path string) error {
	// fetch 在 path 对应的仓库
	if err := g.Fetch(ctx, path); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("[git rev-parse] failed to get branch in %s: %w", path, err)
	}
	branch := strings.TrimSpace(string(branchOut))
	if branch == "" || branch == "HEAD" {
		return fmt.Errorf("[rebase] detached HEAD in %s", path)
	}
	rebaseCmd := exec.CommandContext(ctx, "git", "-C", path, "rebase", "origin/"+branch)
	out, err := rebaseCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git rebase] failed in %s: %w, output: %s", path, errors.ErrGitExec, string(out))
	}
	return nil
}

// FetchAndSwitchBranch fetch 并切换到指定分支
func (g *GitProxy) FetchAndSwitchBranch(ctx context.Context, repoPath, branch string) error {
	// fetch 所有远程
	if err := g.Fetch(ctx, repoPath); err != nil {
		return err
	}

	// 检查分支是否存在于本地
	exists := g.BranchExists(ctx, repoPath, branch)
	if !exists {
		// 本地不存在，尝试 checkout 到远程分支
		cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "checkout", "-b", branch, "origin/"+branch)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("[git checkout] failed to create branch %s in %s: %w, output: %s", branch, repoPath, errors.ErrGitExec, string(out))
		}
		return nil
	}

	// 本地已存在，直接 checkout
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "checkout", branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git checkout] failed to switch to branch %s in %s: %w, output: %s", branch, repoPath, errors.ErrGitExec, string(out))
	}

	// rebase 到远程分支
	rebaseCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rebase", "origin/"+branch)
	rebaseOut, err := rebaseCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git rebase] failed in %s: %w, output: %s", repoPath, errors.ErrGitExec, string(rebaseOut))
	}

	return nil
}

// BranchExists 检查分支是否存在
func (g *GitProxy) BranchExists(ctx context.Context, repoPath, branch string) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "--verify", branch)
	return cmd.Run() == nil
}

// RemoteBranchExists 检查远端仓库是否存在指定分支
func (g *GitProxy) RemoteBranchExists(ctx context.Context, repoURL, branch string) bool {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--heads", repoURL, "refs/heads/"+branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	// 如果输出不为空，说明分支存在
	return len(strings.TrimSpace(string(out))) > 0
}

// CreateWorktreeFromRemoteBranch 从远程分支创建 worktree（不创建新分支）
func (g *GitProxy) CreateWorktreeFromRemoteBranch(ctx context.Context, repoPath, branch, worktreePath string) error {
	// 先 fetch 确保远程分支信息最新
	if err := g.Fetch(ctx, repoPath); err != nil {
		return err
	}
	// 直接从远程分支创建 worktree，不创建新分支
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "worktree", "add", worktreePath, "origin/"+branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git worktree add] failed to create worktree from remote branch %s at %s: %w, output: %s", branch, worktreePath, errors.ErrGitExec, string(out))
	}
	return nil
}

// CheckBranchWorktreeStatus 检查分支是否已被 worktree 使用
func (g *GitProxy) CheckBranchWorktreeStatus(ctx context.Context, repoPath, branch string) (bool, error) {
	worktrees, err := g.ListWorktrees(ctx, repoPath)
	if err != nil {
		return false, fmt.Errorf("failed to list worktrees: %w", err)
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return true, nil
		}
	}
	return false, nil
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
