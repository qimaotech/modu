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
	// RemoveWorktreeAndBranch 删除工作树；若 worktree 当前分支经 /→- 规范化后与 featureDirName 一致则再删除该分支，否则通过结果返回警告并跳过删分支（防误删）
	RemoveWorktreeAndBranch(ctx context.Context, repoPath, worktreePath, featureDirName string) (RemoveWorktreeResult, error)
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
	// GetBranchPushStatus 检查本地分支是否已完整推送到远端分支
	GetBranchPushStatus(ctx context.Context, repoPath, branch string) (BranchPushStatus, error)
}

// RemoveWorktreeResult 描述 worktree 删除成功时需要由调用方展示的非致命警告。
type RemoveWorktreeResult struct {
	Warnings []string
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

// BranchPushStatus 记录本地分支相对远端分支的推送状态。
type BranchPushStatus struct {
	Branch     string `json:"branch"`
	RemoteRef  string `json:"remoteRef"`
	IsPushed   bool   `json:"isPushed"`
	AheadCount int    `json:"aheadCount"`
	Reason     string `json:"reason,omitempty"`
}
