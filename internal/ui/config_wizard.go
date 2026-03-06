package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"codeup.aliyun.com/qimao/public/devops/modu/internal/config"
)

// 配置初始化向导样式
var (
	wizardHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	wizardItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	wizardInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("208"))

	wizardHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

// ConfigWizard 配置初始化向导状态
type ConfigWizard struct {
	step       int    // 当前步骤: 0=workspace, 1=worktree-root, 2=default-base, 3=modules, 4=confirm
	workspace  string // 裸仓库目录
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
		workspace:  "./workspace",
		worktree:   "./worktrees",
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
		// 在模块名称和URL之间切换
		if m.step == 3 && m.inputField == 0 && m.moduleName != "" {
			m.inputField = 1
		}
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
	case 3:
		if m.inputField == 0 {
			m.moduleName += string(ch)
		} else {
			m.moduleURL += string(ch)
		}
	}
	return m, nil
}

func (m *ConfigWizard) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0:
		if m.workspace == "" {
			m.workspace = "./workspace"
		}
		m.step = 1
	case 1:
		if m.worktree == "" {
			m.worktree = "./worktrees"
		}
		m.step = 2
	case 2:
		if m.base == "" {
			m.base = "develop"
		}
		m.step = 3
	case 3:
		// 添加模块
		if m.moduleName != "" && m.moduleURL != "" {
			m.modules = append(m.modules, config.Module{
				Name: m.moduleName,
				URL:  m.moduleURL,
			})
			m.moduleName = ""
			m.moduleURL = ""
			m.inputField = 0
		} else if m.moduleName != "" {
			// 只有名称，跳到URL输入
			m.inputField = 1
		} else {
			// 没有输入，进入确认
			m.step = 4
		}
	case 4:
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
	case 3:
		if m.inputField == 1 && len(m.moduleURL) > 0 {
			m.moduleURL = m.moduleURL[:len(m.moduleURL)-1]
		} else if m.inputField == 0 && len(m.moduleName) > 0 {
			m.moduleName = m.moduleName[:len(m.moduleName)-1]
		}
	}
	return m, nil
}

func (m *ConfigWizard) doSaveConfig() tea.Msg {
	cfg := &config.Config{
		Workspace:    m.workspace,
		WorktreeRoot: m.worktree,
		DefaultBase:  m.base,
		Concurrency:  5,
		AutoFetch:    true,
		StrictDirty:  true,
		Modules:      m.modules,
	}

	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, ".modu.yaml")

	if err := config.SaveConfig(cfg, configPath); err != nil {
		return configSavedMsg{err: err}
	}

	return configSavedMsg{err: nil}
}

type configSavedMsg struct {
	err error
}

// View 实现 tea.Model
func (m *ConfigWizard) View() string {
	if m.quitting {
		return "已退出配置向导\n"
	}

	var s strings.Builder

	switch m.step {
	case 0:
		s.WriteString(wizardHeaderStyle.Render("📁 步骤 1/4: 仓库目录"))
		s.WriteString("\n\n")
		s.WriteString(wizardItemStyle.Render("请输入主仓库所在目录（裸仓库）：\n"))
		s.WriteString(wizardInputStyle.Render("> " + m.workspace + "_"))
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("提示: 包含 .git 的仓库目录，按 Enter 继续"))

	case 1:
		s.WriteString(wizardHeaderStyle.Render("📂 步骤 2/4: Worktree 目录"))
		s.WriteString("\n\n")
		s.WriteString(wizardItemStyle.Render("请输入特性分支代码存放目录：\n"))
		s.WriteString(wizardInputStyle.Render("> " + m.worktree + "_"))
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("按 Enter 继续，Backspace 返回"))

	case 2:
		s.WriteString(wizardHeaderStyle.Render("🌿 步骤 3/4: 默认分支"))
		s.WriteString("\n\n")
		s.WriteString(wizardItemStyle.Render("请输入默认基准分支名称：\n"))
		s.WriteString(wizardInputStyle.Render("> " + m.base + "_"))
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("如: develop, main, master\n按 Enter 继续，Backspace 返回"))

	case 3:
		s.WriteString(wizardHeaderStyle.Render("📦 步骤 4/4: 模块配置"))
		s.WriteString("\n\n")

		if len(m.modules) > 0 {
			s.WriteString(wizardItemStyle.Render("已添加模块:\n"))
			for _, mod := range m.modules {
				s.WriteString(wizardItemStyle.Render(fmt.Sprintf("  • %s -> %s\n", mod.Name, mod.URL)))
			}
			s.WriteString("\n")
		}

		s.WriteString(wizardItemStyle.Render("添加模块（名称 + Enter + URL + Enter）：\n"))
		cursor := "_"
		if m.inputField == 0 {
			s.WriteString(wizardInputStyle.Render("名称: " + m.moduleName + cursor))
		} else {
			s.WriteString(wizardInputStyle.Render("名称: " + m.moduleName + "\n"))
			s.WriteString(wizardInputStyle.Render("URL:  " + m.moduleURL + cursor))
		}
		s.WriteString("\n\n")
		s.WriteString(wizardHelpStyle.Render("无更多模块直接 Enter 确认，Tab 切换名称/URL"))

	case 4:
		s.WriteString(wizardHeaderStyle.Render("✅ 确认配置"))
		s.WriteString("\n\n")
		s.WriteString(wizardItemStyle.Render("请确认以下配置：\n\n"))
		s.WriteString(wizardItemStyle.Render(fmt.Sprintf("  仓库目录:     %s\n", m.workspace)))
		s.WriteString(wizardItemStyle.Render(fmt.Sprintf("  Worktree目录: %s\n", m.worktree)))
		s.WriteString(wizardItemStyle.Render(fmt.Sprintf("  默认分支:     %s\n", m.base)))
		s.WriteString(wizardItemStyle.Render(fmt.Sprintf("  模块数量:     %d\n\n", len(m.modules))))
		s.WriteString(wizardHelpStyle.Render("按 Enter 保存配置"))

	default:
		s.WriteString("未知状态\n")
	}

	return s.String()
}

// RunConfigWizard 运行配置向导
func RunConfigWizard() error {
	p := tea.NewProgram(NewConfigWizard(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run config wizard: %w", err)
	}
	return nil
}
