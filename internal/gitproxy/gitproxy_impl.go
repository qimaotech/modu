package gitproxy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
func (g *GitProxy) Clone(ctx context.Context, url, path string, options CloneOptions) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("clone destination already exists: %s: %w", path, errors.ErrInvalidOperation)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	temporaryPath, err := os.MkdirTemp(filepath.Dir(path), ".modu-clone-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporaryPath)
	args := []string{"clone", "--progress"}
	if options.Filter != "" {
		args = append(args, "--filter="+options.Filter)
	}
	args = append(args, "--", url, temporaryPath)
	out, err := runProgress(ctx, "克隆", args...)
	if err != nil {
		return fmt.Errorf("[git clone] %s: %w: %w, output: %s", path, errors.ErrGitExec, commandCause(ctx, err), out)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("clone destination already exists: %s: %w", path, errors.ErrInvalidOperation)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("finish clone %s: %w", path, err)
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
	cmd := gitCommand(ctx, "-C", repoPath, "worktree", "add", "-b", branch, worktreePath, baseBranch)
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
	cmd := gitCommand(ctx, "-C", repoPath, "worktree", "add", worktreePath, branch)
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

	cmd := gitCommand(ctx, "-C", path, "status", "--porcelain=v2", "--branch", "-z")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Status{}, fmt.Errorf("[git status] failed to get status for %s: %w, output: %s: %w", path, errors.ErrGitExec, string(out), commandCause(ctx, err))
	}

	return parseStatus(string(out)), nil
}

// RemoveWorktree 删除工作树
func (g *GitProxy) RemoveWorktree(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// 先用 git worktree remove 移除
	cmd := gitCommand(ctx, "worktree", "remove", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		// 如果 worktree remove 失败，尝试直接删除目录
		if rmErr := os.RemoveAll(path); rmErr != nil {
			return fmt.Errorf("[git worktree remove] failed to remove worktree at %s: %w, output: %s: %w", path, errors.ErrGitExec, string(out), commandCause(ctx, err))
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
	if err := ctx.Err(); err != nil {
		return err
	}
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
	cmd := gitCommand(ctx, "-C", repoPath, "worktree", "remove", "--force", worktreePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Warn("git worktree remove 失败，尝试直接删除目录: path=%s, error=%s", worktreePath, string(out))
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		// 如果 worktree remove 失败，尝试直接删除目录
		if rmErr := os.RemoveAll(worktreePath); rmErr != nil {
			logger.Error("删除目录失败: path=%s, error=%v", worktreePath, rmErr)
			return fmt.Errorf("[git worktree remove] failed to remove worktree at %s: %w, output: %s", worktreePath, errors.ErrGitExec, string(out))
		}
		logger.Info("直接删除目录成功: %s", worktreePath)
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	// 先 prune 清理过期的 worktree 引用
	cmd = gitCommand(ctx, "-C", repoPath, "worktree", "prune")
	if err := cmd.Run(); err != nil {
		logger.Warn("git worktree prune 失败: %v", err)
	} else {
		logger.Info("git worktree prune 成功")
	}

	// 再删除对应的分支（仅在与目录 slug 一致时）
	if err := ctx.Err(); err != nil {
		return err
	}
	if branchToDelete != "" {
		logger.Info("删除分支: repo=%s, branch=%s", repoPath, branchToDelete)
		cmd = gitCommand(ctx, "-C", repoPath, "branch", "-D", branchToDelete)
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
	cmd := gitCommand(ctx, "-C", repoPath, "worktree", "list", "--porcelain", "-z")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("[git worktree list] failed: %w, output: %s", errors.ErrGitExec, string(out))
	}

	return parseWorktreeList(string(out))
}

// Fetch 从远程获取最新
func (g *GitProxy) Fetch(ctx context.Context, repoPath string) error {
	return g.fetch(ctx, repoPath, "--all")
}

func (g *GitProxy) fetch(ctx context.Context, repoPath, remote string) error {
	stage := "拉取 " + remote
	if remote == "--all" {
		stage = "拉取远端"
	}
	out, err := runProgress(ctx, stage, "-C", repoPath, "fetch", "--progress", remote)
	if err != nil {
		return fmt.Errorf("[git fetch] %s: %w: %w, output: %s", repoPath, errors.ErrGitExec, commandCause(ctx, err), out)
	}
	return nil
}

// Rebase 在当前路径下执行 fetch 后 rebase origin/<当前分支>
func (g *GitProxy) Rebase(ctx context.Context, path string) error {
	// fetch 在 path 对应的仓库
	if err := g.fetch(ctx, path, "origin"); err != nil {
		return err
	}
	cmd := gitCommand(ctx, "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("[git rev-parse] failed to get branch in %s: %w", path, commandCause(ctx, err))
	}
	branch := strings.TrimSpace(string(branchOut))
	if branch == "" || branch == "HEAD" {
		return fmt.Errorf("[rebase] detached HEAD in %s", path)
	}
	out, err := runProgress(ctx, "rebase", "-C", path, "rebase", "origin/"+branch)
	if err != nil {
		return fmt.Errorf("[git rebase] failed in %s: %w, output: %s: %w", path, errors.ErrGitExec, string(out), commandCause(ctx, err))
	}
	return nil
}

// FetchAndSwitchBranch fetch 并切换到指定分支
func (g *GitProxy) FetchAndSwitchBranch(ctx context.Context, repoPath, branch string) error {
	// 更新仅依赖 origin，避免访问无关远端。
	if err := g.fetch(ctx, repoPath, "origin"); err != nil {
		return err
	}

	// 检查分支是否存在于本地
	exists := g.BranchExists(ctx, repoPath, branch)
	if !exists {
		// 本地不存在，尝试 checkout 到远程分支
		out, err := runProgress(ctx, "切换分支", "-C", repoPath, "checkout", "-b", branch, "origin/"+branch)
		if err != nil {
			return fmt.Errorf("[git checkout] failed to create branch %s in %s: %w: %w, output: %s", branch, repoPath, errors.ErrGitExec, commandCause(ctx, err), out)
		}
		return nil
	}

	// 本地已存在，直接 checkout
	out, err := runProgress(ctx, "切换分支", "-C", repoPath, "checkout", branch)
	if err != nil {
		return fmt.Errorf("[git checkout] failed to switch to branch %s in %s: %w: %w, output: %s", branch, repoPath, errors.ErrGitExec, commandCause(ctx, err), out)
	}

	// rebase 到远程分支
	rebaseOut, err := runProgress(ctx, "rebase", "-C", repoPath, "rebase", "origin/"+branch)
	if err != nil {
		return fmt.Errorf("[git rebase] failed in %s: %w: %w, output: %s", repoPath, errors.ErrGitExec, commandCause(ctx, err), rebaseOut)
	}

	return nil
}

// BranchExists 检查分支是否存在
func (g *GitProxy) BranchExists(ctx context.Context, repoPath, branch string) bool {
	cmd := gitCommand(ctx, "-C", repoPath, "rev-parse", "--verify", branch)
	return cmd.Run() == nil
}

// RemoteBranchExists 检查远端仓库是否存在指定分支
func (g *GitProxy) RemoteBranchExists(ctx context.Context, repoURL, branch string) bool {
	cmd := gitCommand(ctx, "ls-remote", "--heads", repoURL, "refs/heads/"+branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	// 如果输出不为空，说明分支存在
	return len(strings.TrimSpace(string(out))) > 0
}

// CreateWorktreeFromRemoteBranch 从 origin 远程分支创建带跟踪关系的本地 worktree 分支。
func (g *GitProxy) CreateWorktreeFromRemoteBranch(ctx context.Context, repoPath, branch, worktreePath string) error {
	if err := g.fetchOriginBranch(ctx, repoPath, branch); err != nil {
		return err
	}
	// worktree 需要本地分支，否则会处于 detached HEAD，无法正常执行后续 update。
	cmd := gitCommand(ctx, "-C", repoPath, "worktree", "add", "--track", "-b", branch, worktreePath, "origin/"+branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git worktree add] failed to create worktree from remote branch %s at %s: %w: %w, output: %s", branch, worktreePath, errors.ErrGitExec, commandCause(ctx, err), string(out))
	}
	return nil
}

func (g *GitProxy) fetchOriginBranch(ctx context.Context, repoPath, branch string) error {
	// 显式拉取目标分支，兼容 single-branch clone 或受限的 remote fetch refspec。
	remoteBranch := "origin/" + branch
	remoteRef := "refs/remotes/" + remoteBranch
	refspec := "refs/heads/" + branch + ":" + remoteRef
	if err := g.ensureOriginTracksBranch(ctx, repoPath, branch, refspec); err != nil {
		return err
	}
	fetchOut, err := runProgress(ctx, "拉取需求分支", "-C", repoPath, "fetch", "--progress", "origin", "+"+refspec)
	if err != nil {
		return fmt.Errorf("[git fetch] failed to fetch remote branch %s in %s: %w: %w, output: %s", branch, repoPath, errors.ErrGitExec, commandCause(ctx, err), string(fetchOut))
	}
	return nil
}

// ensureOriginTracksBranch 确保 origin 的 fetch refspec 能将目标分支识别为可跟踪的远程分支。
func (g *GitProxy) ensureOriginTracksBranch(ctx context.Context, repoPath, branch, refspec string) error {
	cmd := gitCommand(ctx, "-C", repoPath, "remote", "set-branches", "--add", "origin", branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git remote set-branches] failed to track remote branch %s in %s: %w: %w, output: %s", branch, repoPath, errors.ErrGitExec, commandCause(ctx, err), string(out))
	}

	// set-branches 由 Git 维护 remote.origin.fetch；refspec 仅用于校验命令结果是目标映射。
	verifyCmd := gitCommand(ctx, "-C", repoPath, "config", "--get-all", "--fixed-value", "remote.origin.fetch", "+"+refspec)
	verifyOut, err := verifyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[git config] failed to verify remote branch %s in %s: %w: %w, output: %s", branch, repoPath, errors.ErrGitExec, commandCause(ctx, err), string(verifyOut))
	}
	if strings.TrimSpace(string(verifyOut)) == "" {
		return fmt.Errorf("[git config] target refspec not found for remote branch %s in %s: %w", branch, repoPath, errors.ErrGitExec)
	}
	return nil
}

// commandCause 在命令被上下文中断时优先返回可判断的 context 错误。
func commandCause(ctx context.Context, commandErr error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return fmt.Errorf("git command context: %w", ctxErr)
	}
	return commandErr
}

// GetBranchPushStatus 检查本地分支是否已完整推送到远端分支。
func (g *GitProxy) GetBranchPushStatus(ctx context.Context, repoPath, branch string) (BranchPushStatus, error) {
	branch = strings.TrimSpace(branch)
	status := BranchPushStatus{Branch: branch}
	if branch == "" || branch == "HEAD" {
		status.Reason = "invalid branch"
		return status, nil
	}

	if err := g.Fetch(ctx, repoPath); err != nil {
		status.Reason = "fetch failed"
		return status, fmt.Errorf("[git fetch] failed to inspect push status for branch %s in %s: %w", branch, repoPath, err)
	}

	remoteRef, found, err := g.resolveBranchRemoteRef(ctx, repoPath, branch)
	if err != nil {
		status.Reason = "remote lookup failed"
		return status, err
	}
	if !found {
		status.Reason = "remote branch not found"
		return status, nil
	}
	status.RemoteRef = remoteRef

	aheadCount, err := g.countBranchAhead(ctx, repoPath, remoteRef, branch)
	if err != nil {
		status.Reason = "ahead count failed"
		return status, err
	}
	status.AheadCount = aheadCount
	status.IsPushed = aheadCount == 0
	if aheadCount > 0 {
		status.Reason = fmt.Sprintf("%d local commits not pushed", aheadCount)
	}
	return status, nil
}

// resolveBranchRemoteRef 优先使用 upstream，否则回退到 origin/<branch>。
func (g *GitProxy) resolveBranchRemoteRef(ctx context.Context, repoPath, branch string) (string, bool, error) {
	upstreamArg := branch + "@{upstream}"
	cmd := gitCommand(ctx, "-C", repoPath, "rev-parse", "--abbrev-ref", upstreamArg)
	out, err := cmd.CombinedOutput()
	if err == nil {
		remoteRef := strings.TrimSpace(string(out))
		if remoteRef != "" {
			return remoteRef, true, nil
		}
	}

	fallbackRemoteRef := "origin/" + branch
	verifyRef := "refs/remotes/" + fallbackRemoteRef
	cmd = gitCommand(ctx, "-C", repoPath, "rev-parse", "--verify", "--quiet", verifyRef)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", false, nil
	}
	if strings.TrimSpace(string(out)) == "" {
		return "", false, nil
	}
	return fallbackRemoteRef, true, nil
}

// countBranchAhead 返回 local branch 相对 remoteRef 的领先提交数。
func (g *GitProxy) countBranchAhead(ctx context.Context, repoPath, remoteRef, branch string) (int, error) {
	revisionRange := remoteRef + ".." + branch
	cmd := gitCommand(ctx, "-C", repoPath, "rev-list", "--count", revisionRange)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("[git rev-list] failed to count ahead commits for %s in %s: %w, output: %s", branch, repoPath, errors.ErrGitExec, string(out))
	}
	aheadCount, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("[git rev-list] invalid ahead count for %s in %s: %w, output: %s", branch, repoPath, err, string(out))
	}
	return aheadCount, nil
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

// parseStatus 读取 porcelain v2 的 NUL 分隔记录，保留空格、换行及重命名路径。
func parseStatus(output string) Status {
	result := Status{}
	records := strings.Split(output, "\x00")
	for i := 0; i < len(records); i++ {
		record := records[i]
		if branch, ok := strings.CutPrefix(record, "# branch.head "); ok {
			result.Branch = branch
			if branch == "(detached)" {
				result.Branch = "HEAD"
			}
			continue
		}
		if len(record) < 2 {
			continue
		}
		var file FileStatus
		switch record[0] {
		case '?':
			file = FileStatus{Name: record[2:], Status: '?'}
		case '1', '2', 'u':
			fieldsCount := 9
			if record[0] == '2' {
				fieldsCount = 10
			} else if record[0] == 'u' {
				fieldsCount = 11
			}
			fields := strings.SplitN(record, " ", fieldsCount)
			if len(fields) != fieldsCount || len(fields[1]) < 2 {
				continue
			}
			file = FileStatus{Name: fields[fieldsCount-1], Status: rune(fields[1][0])}
			if file.Status == '.' {
				file.Status = rune(fields[1][1])
			}
			if record[0] == '2' {
				i++
			}
		default:
			continue
		}
		result.IsDirty = true
		result.Files = append(result.Files, file)
	}
	return result
}

// parseWorktreeList 解析 git worktree list 输出
func parseWorktreeList(output string) ([]WorktreeInfo, error) {
	var worktrees []WorktreeInfo
	separator := "\n"
	if strings.Contains(output, "\x00") {
		separator = "\x00"
	}
	lines := strings.Split(output, separator)

	var current WorktreeInfo
	for _, line := range lines {
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			current.Path = path
		} else if strings.HasPrefix(line, "HEAD ") {
			// HEAD 行不需要处理
		} else if branch, ok := strings.CutPrefix(line, "branch refs/heads/"); ok {
			current.Branch = branch
		} else if line == "detached" {
			current.Branch = "HEAD"
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
	cmd := gitCommand(ctx, args...)
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
