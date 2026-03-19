package gitproxy

import (
	"context"
)

// GitClient Git 操作接口
type GitClient interface {
	// Clone 克隆仓库到指定路径
	Clone(ctx context.Context, url, path string) error
	// CreateWorktree 创建工作树
	CreateWorktree(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error
	// GetStatus 获取目录状态
	GetStatus(ctx context.Context, path string) (Status, error)
	// RemoveWorktree 删除工作树
	RemoveWorktree(ctx context.Context, path string) error
	// RemoveWorktreeAndBranch 删除工作树并删除对应分支
	RemoveWorktreeAndBranch(ctx context.Context, repoPath, branch, worktreePath string) error
	// ListWorktrees 列出所有工作树
	ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error)
	// Fetch 从远程获取最新
	Fetch(ctx context.Context, repoPath string) error
	// Rebase 在当前路径下执行 git rebase origin/<当前分支>
	Rebase(ctx context.Context, path string) error
	// FetchAndSwitchBranch fetch 并切换到指定分支
	FetchAndSwitchBranch(ctx context.Context, repoPath, branch string) error
	// BranchExists 检查分支是否存在
	BranchExists(ctx context.Context, repoPath, branch string) bool
	// CheckBranchWorktreeStatus 检查分支是否已被 worktree 使用
	CheckBranchWorktreeStatus(ctx context.Context, repoPath, branch string) (bool, error)
	// CreateWorktreeFromExistingBranch 从现有分支创建 worktree（不创建新分支）
	CreateWorktreeFromExistingBranch(ctx context.Context, repoPath, branch, worktreePath string) error
	// RemoteBranchExists 检查远端仓库是否存在指定分支
	RemoteBranchExists(ctx context.Context, repoURL, branch string) bool
	// CreateWorktreeFromRemoteBranch 从远程分支创建 worktree（不创建新分支）
	CreateWorktreeFromRemoteBranch(ctx context.Context, repoPath, branch, worktreePath string) error
}

// Status Git 状态
type Status struct {
	IsDirty bool         // 是否存在未提交修改
	Branch  string       // 当前分支
	Files   []FileStatus // 变更文件列表
}

// FileStatus 文件变更状态
type FileStatus struct {
	Name   string // 文件名
	Status rune   // 状态字符 (M, A, D, ?? 等)
}

// WorktreeInfo 工作树信息
type WorktreeInfo struct {
	Path   string
	Branch string
}
