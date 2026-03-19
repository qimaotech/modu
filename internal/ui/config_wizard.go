package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/qimaotech/modu/internal/config"
)

// 配置初始化向导样式
var (
	wizardHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	wizardItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	wizardInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("208"))

	wizardHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

// ConfigWizard 配置初始化向导状态
type ConfigWizard struct {
	step       int    // 当前步骤: 0=workspace, 1=worktree-root, 2=default-base, 3=confirm
	workspace  string // workspace仓库目录
	worktree   string // worktree 目录
	base       string // 默认基准分支
	modules    []config.Module
	moduleName string
	moduleURL  string
	inputField int // 0=moduleName, 1=moduleURL
	err        error
	quitting   bool
	saved      bool
}

// NewConfigWizard 创建配置向导
func NewConfigWizard() *ConfigWizard {
	return &ConfigWizard{
		step:       0,
		workspace:  "/tmp/workspace",
		worktree:   "/tmp/worktrees",
		base:       "develop",
		modules:    []config.Module{},
		inputField: 0,
	}
}

// Init 实现 tea.Model
func (m *ConfigWizard) Init() tea.Cmd {
	return nil
}

// Update 实现 tea.Model
func (m *ConfigWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case configSavedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.saved = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ConfigWizard) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit
	case "enter":
		return m.handleEnter()
	case "backspace":
		return m.handleBackspace()
	case "tab":
		return m, nil
	}

	// 文字输入
	if len(key) == 1 {
		return m.handleChar(key[0])
	}
	return m, nil
}

func (m *ConfigWizard) handleChar(ch byte) (tea.Model, tea.Cmd) {
	switch m.step {
	case 0:
		m.workspace += string(ch)
	case 1:
		m.worktree += string(ch)
	case 2:
		m.base += string(ch)
	}
	return m, nil
}

func (m *ConfigWizard) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0:
		if m.workspace == "" {
			m.workspace = "."
		}
		m.step = 1
	case 1:
		if m.worktree == "" {
			m.worktree = "../worktrees"
		}
		m.step = 2
	case 2:
		if m.base == "" {
			m.base = "develop"
		}
		m.step = 3
	case 3:
		return m, m.doSaveConfig
	}
	return m, nil
}

func (m *ConfigWizard) handleBackspace() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0:
		if len(m.workspace) > 0 {
			m.workspace = m.workspace[:len(m.workspace)-1]
		}
	case 1:
		if len(m.worktree) > 0 {
			m.worktree = m.worktree[:len(m.worktree)-1]
		}
	case 2:
		if len(m.base) > 0 {
			m.base = m.base[:len(m.base)-1]
		}
	}
	return m, nil
}

func (m *ConfigWizard) doSaveConfig() tea.Msg {
	// 1. 创建 workspace 目录（如果不存在）
	if err := os.MkdirAll(m.workspace, 0755); err != nil {
		return configSavedMsg{err: fmt.Errorf("failed to create workspace directory: %w", err)}
	}

	// 2. 确保 workspace 是 git 仓库（如果不是则初始化）
	if err := m.ensureGitRepo(m.workspace, m.base); err != nil {
		return configSavedMsg{err: err}
	}

	// 3. 创建 worktree-root 目录（如果不存在）
	worktreeRootAbs := m.worktree
	if !filepath.IsAbs(worktreeRootAbs) {
		worktreeRootAbs = filepath.Join(m.workspace, m.worktree)
	}
	if err := os.MkdirAll(worktreeRootAbs, 0755); err != nil {
		return configSavedMsg{err: fmt.Errorf("failed to create worktree directory: %w", err)}
	}

	// 4. 保存配置文件到 workspace 目录
	cfg := &config.Config{
		Workspace:    m.workspace,
		WorktreeRoot: m.worktree,
		DefaultBase:  m.base,
		Concurrency:  5,
		AutoFetch:    true,
		StrictDirty:  true,
		Modules:      m.modules,
	}

	configPath := filepath.Join(m.workspace, ".modu.yaml")

	if err := config.SaveConfig(cfg, configPath); err != nil {
		return configSavedMsg{err: err}
	}

	return configSavedMsg{
		err:        nil,
		configPath: configPath,
		workspace:  m.workspace,
		worktree:   m.worktree,
		base:       m.base,
	}
}

// ensureGitRepo 检查并初始化 git 仓库
// 如果 workspace 不是 git 仓库，执行 git init 并创建 base 分支
func (m *ConfigWizard) ensureGitRepo(workspacePath, baseBranch string) error {
	gitDir := filepath.Join(workspacePath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// .git 目录或文件已存在，已经是 git 仓库
		// 检查是否有至少一个提交
		ctx := context.Background()
		logCmd := exec.CommandContext(ctx, "git", "rev-list", "--count", "HEAD")
		logCmd.Dir = workspacePath
		if err := logCmd.Run(); err != nil {
			// 没有提交，创建 base 分支和初始提交
			checkoutCmd := exec.CommandContext(ctx, "git", "checkout", "-b", baseBranch)
			checkoutCmd.Dir = workspacePath
			if err := checkoutCmd.Run(); err != nil {
				return fmt.Errorf("git checkout -b %s failed: %w", baseBranch, err)
			}
			if err := createInitialCommit(workspacePath); err != nil {
				return err
			}
		}
		return nil
	}

	// 不是 git 仓库，执行 git init
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = workspacePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// 创建并切换到 base 分支
	checkoutCmd := exec.CommandContext(ctx, "git", "checkout", "-b", baseBranch)
	checkoutCmd.Dir = workspacePath
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("git checkout -b %s failed: %w", baseBranch, err)
	}

	// 创建初始提交
	if err := createInitialCommit(workspacePath); err != nil {
		return err
	}

	return nil
}

// createInitialCommit 创建初始提交
func createInitialCommit(workspacePath string) error {
	ctx := context.Background()

	// 创建一个空提交
	cmd := exec.CommandContext(ctx, "git", "commit", "--allow-empty", "-m", "Initial commit")
	cmd.Dir = workspacePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}
	return nil
}

type configSavedMsg struct {
	err        error
	configPath string
	workspace  string
	worktree   string
	base       string
}

// View 实现 tea.Model
func (m *ConfigWizard) View() string {
	if m.quitting {
		return "已退出配置向导\n"
	}

	var s strings.Builder

	switch m.step {
	case 0:
		s.WriteString(wizardHeaderStyle.Render("📁 步骤 1/3: workspace仓库目录"))
		s.WriteString("\n\n")
		s.WriteString(wizardItemStyle.Render("使用modu实现多项目管理时，需要一个文件夹统一存放所有 Git 工程代码；\n" +
			"该文件夹是为 Git 版本控制目录（如果不是，将自动帮您初始化）；\n\n" +
			"请输入："))
		s.WriteString("\n\n")
		s.WriteString(wizardInputStyle.Render("> " + m.workspace + "_"))
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("团队使用推荐使用环境变量，个人使用推荐填写绝对路径"))
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("提示: 按 Enter 继续，Esc 退出"))

	case 1:
		s.WriteString(wizardHeaderStyle.Render("📂 步骤 2/3: Worktree 目录"))
		s.WriteString("\n\n")
		s.WriteString(wizardItemStyle.Render("Worktree 目录，用来存放各个特性分支代码的目录；\n\n" +
			"请输入："))
		s.WriteString("\n\n")
		s.WriteString(wizardInputStyle.Render("> " + m.worktree + "_"))
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("团队使用推荐使用环境变量，个人使用推荐填写绝对路径"))
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("按 Enter 继续，Esc 退出"))

	case 2:
		s.WriteString(wizardHeaderStyle.Render("🌿 步骤 3/3: 默认分支"))
		s.WriteString("\n\n")
		s.WriteString(wizardItemStyle.Render("默认分支表示创建新的特性分支的源分支；\n\n" +
			"注意各子模块也需要有此分支。\n\n" +
			"请输入："))
		s.WriteString("\n\n")
		s.WriteString(wizardInputStyle.Render("> " + m.base + "_"))
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("比如: develop, main, master"))
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("按 Enter 继续，Esc 退出"))

	case 3:
		s.WriteString(wizardHeaderStyle.Render("✅ 确认配置"))
		s.WriteString("\n\n")
		s.WriteString(wizardItemStyle.Render("请确认以下配置："))
		s.WriteString("\n\n")
		s.WriteString(wizardItemStyle.Render(fmt.Sprintf("  Workspace目录: %s\t", m.workspace)))
		s.WriteString("\n")
		s.WriteString(wizardItemStyle.Render(fmt.Sprintf("  Worktree目录:  %s\t", m.worktree)))
		s.WriteString("\n")
		s.WriteString(wizardItemStyle.Render(fmt.Sprintf("  默认分支:      %s\t", m.base)))
		s.WriteString("\n")
		s.WriteString(wizardItemStyle.Render(fmt.Sprintf("  模块数量:      %d\t", len(m.modules))))
		s.WriteString("\n")
		s.WriteString("\n")
		s.WriteString(wizardHelpStyle.Render("配置完成后请手动执行：\n☑️ 1. 将您已有的 Git 项目移动到 Workspace 目录\n☑️ 2. 执行 modu config scan"))
		s.WriteString("\n")
		s.WriteString("\n")
		s.WriteString(wizardHelpStyle.Render("🚀 按 Enter 保存配置"))

	default:
		s.WriteString("未知状态\n")
	}

	return s.String()
}

// RunConfigWizard 运行配置向导
// 返回保存的配置信息（如果成功）以及错误
func RunConfigWizard() (*SavedConfigInfo, error) {
	p := tea.NewProgram(NewConfigWizard(), tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run config wizard: %w", err)
	}

	// 从 model 中获取保存结果
	if wizard, ok := model.(*ConfigWizard); ok {
		if wizard.saved {
			return &SavedConfigInfo{
				ConfigPath: filepath.Join(wizard.workspace, ".modu.yaml"),
				Workspace:  wizard.workspace,
				Worktree:   wizard.worktree,
				Base:       wizard.base,
			}, nil
		}
	}
	return nil, nil
}

// SavedConfigInfo 保存配置后的信息
type SavedConfigInfo struct {
	ConfigPath string
	Workspace  string
	Worktree   string
	Base       string
}
