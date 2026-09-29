package gitproxy

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/qimaotech/modu/internal/errors"
)

// ValidateBranchName 只接受实际分支名，防止修订表达式或路径进入 worktree 操作。
func ValidateBranchName(ctx context.Context, branch string) error {
	out, err := gitCommand(ctx, "check-ref-format", "--branch", branch).CombinedOutput()
	if err != nil {
		return fmt.Errorf("无效的 feature 分支名 %q: %w: %w", branch, errors.ErrInvalidOperation, commandCause(ctx, err))
	}
	if strings.TrimSpace(string(out)) != branch {
		return fmt.Errorf("请提供完整分支名 %q: %w", branch, errors.ErrInvalidOperation)
	}
	return nil
}

func (g *GitProxy) QueryRemoteBranch(ctx context.Context, repoPath, repoURL, branch string) (bool, error) {
	args := make([]string, 0, 8)
	if repoPath != "" {
		// 避免 Git 向上查找，把普通模块目录误认成主项目仓库。
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
			return false, fmt.Errorf("目录不是已初始化的仓库 %s: %w", repoPath, err)
		}
		args = append(args, "-C", repoPath)
		repoURL = "origin"
	} else if strings.TrimSpace(repoURL) == "" {
		return false, fmt.Errorf("仓库地址为空: %w", errors.ErrConfigInvalid)
	}
	args = append(args, "ls-remote", "--exit-code", "--heads", "--", repoURL, "refs/heads/"+branch)
	out, err := gitCommand(ctx, args...).CombinedOutput()
	if err == nil {
		return true, nil
	}
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	var exitErr *exec.ExitError
	if stderrors.As(err, &exitErr) && exitErr.ExitCode() == 2 {
		return false, nil
	}
	return false, fmt.Errorf("查询远程分支 %s 失败: %w: %w, %s", branch, errors.ErrGitExec, err, out)
}

func (g *GitProxy) CheckoutWorktree(ctx context.Context, repoPath, branch, worktreePath string) (WorktreeCheckout, error) {
	worktrees, err := g.ListWorktrees(ctx, repoPath)
	if err != nil {
		return WorktreeCheckout{}, err
	}
	targetInfo, statErr := os.Stat(worktreePath)
	if statErr != nil && !os.IsNotExist(statErr) {
		return WorktreeCheckout{}, statErr
	}
	for _, worktree := range worktrees {
		if targetInfo != nil {
			info, err := os.Stat(worktree.Path)
			if err == nil && os.SameFile(targetInfo, info) {
				if worktree.Branch != branch {
					return WorktreeCheckout{}, fmt.Errorf("目录 %s 已检出分支 %s，预期 %s: %w", worktreePath, worktree.Branch, branch, errors.ErrFeatureExists)
				}
				commit, err := worktreeCommit(ctx, worktreePath)
				return WorktreeCheckout{Commit: commit, Existing: true}, err
			}
		}
	}
	if _, err := os.Lstat(worktreePath); !os.IsNotExist(err) {
		return WorktreeCheckout{}, fmt.Errorf("目标目录已存在且不是该仓库的目标 worktree: %s: %w", worktreePath, errors.ErrFeatureExists)
	}
	for _, worktree := range worktrees {
		if worktree.Branch == branch {
			return WorktreeCheckout{}, fmt.Errorf("分支 %s 已被 %s 使用: %w", branch, worktree.Path, errors.ErrFeatureExists)
		}
	}
	localRef := "refs/heads/" + branch
	err = gitCommand(ctx, "-C", repoPath, "show-ref", "--verify", "--quiet", localRef).Run()
	if err == nil {
		if err := g.fetchOriginBranch(ctx, repoPath, branch); err != nil {
			return WorktreeCheckout{}, err
		}
		localCommit, err := gitCommand(ctx, "-C", repoPath, "rev-parse", localRef).Output()
		if err != nil {
			return WorktreeCheckout{}, commandCause(ctx, err)
		}
		remoteCommit, err := gitCommand(ctx, "-C", repoPath, "rev-parse", "refs/remotes/origin/"+branch).Output()
		if err != nil {
			return WorktreeCheckout{}, commandCause(ctx, err)
		}
		if string(localCommit) != string(remoteCommit) {
			return WorktreeCheckout{}, fmt.Errorf("本地分支 %s 与 origin/%s 提交不同，请先处理本地分支；已有环境可用 modu update 同步: %w", branch, branch, errors.ErrFeatureExists)
		}
		if out, err := gitCommand(ctx, "-C", repoPath, "branch", "--set-upstream-to=origin/"+branch, branch).CombinedOutput(); err != nil {
			return WorktreeCheckout{}, fmt.Errorf("设置远程跟踪分支失败: %w, %s", commandCause(ctx, err), out)
		}
		if out, err := gitCommand(ctx, "-C", repoPath, "worktree", "add", worktreePath, branch).CombinedOutput(); err != nil {
			return WorktreeCheckout{}, fmt.Errorf("创建 worktree 失败: %w, %s", commandCause(ctx, err), out)
		}
	} else {
		var exitErr *exec.ExitError
		if !stderrors.As(err, &exitErr) || exitErr.ExitCode() != 1 || ctx.Err() != nil {
			return WorktreeCheckout{}, fmt.Errorf("查询本地分支失败: %w", commandCause(ctx, err))
		}
		if err := g.CreateWorktreeFromRemoteBranch(ctx, repoPath, branch, worktreePath); err != nil {
			return WorktreeCheckout{}, err
		}
	}
	commit, err := worktreeCommit(ctx, worktreePath)
	return WorktreeCheckout{Commit: commit}, err
}

func worktreeCommit(ctx context.Context, worktreePath string) (string, error) {
	out, err := gitCommand(ctx, "-C", worktreePath, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("读取 worktree 提交失败: %w", commandCause(ctx, err))
	}
	return strings.TrimSpace(string(out)), nil
}
