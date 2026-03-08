package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"codeup.aliyun.com/qimao/public/devops/modu/internal/config"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/core"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/engine"
)

// 全局样式
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))
)

// TUI App 状态
type App struct {
	Engine   *engine.Engine
	Envs     []core.WorktreeEnv
	selected int
	state    string // "loading", "list", "confirm", "error"
	feature  string
	err      error
	message  string
}

// New 创建 TUI App
func New(cfg *config.Config) *App {
	return &App{
		Engine: engine.New(cfg),
		state:  "loading",
	}
}

// Model 实现 tea.Model 接口
func (m *App) Init() tea.Cmd {
	return m.loadEnvs
}

func (m *App) loadEnvs() tea.Msg {
	envs, err := m.Engine.ListWorktrees(context.Background())
	if err != nil {
		return errorMsg{err}
	}
	return loadedMsg{envs}
}

type loadedMsg struct {
	envs []core.WorktreeEnv
}

type errorMsg struct {
	err error
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		m.Envs = msg.envs
		m.state = "list"
	case errorMsg:
		m.err = msg.err
		m.state = "error"
	case tea.KeyMsg:
		switch m.state {
		case "list":
			return m.handleListKey(msg)
		case "confirm":
			return m.handleConfirmKey(msg)
		}
	}
	return m, nil
}

func (m *App) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.Envs)-1 {
			m.selected++
		}
	case "enter":
		if len(m.Envs) > 0 {
			m.state = "confirm"
			m.feature = m.Envs[m.selected].Name
		}
	case "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m *App) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		err := m.Engine.DeleteWorktree(context.Background(), m.feature, false)
		if err != nil {
			m.err = err
			m.state = "error"
		} else {
			m.message = "Deleted feature: " + m.feature
			m.state = "list"
			m.selected = 0
		}
	case "n", "esc":
		m.state = "list"
	}
	return m, nil
}

func (m *App) View() string {
	switch m.state {
	case "loading":
		return "Loading..."
	case "list":
		return m.renderList()
	case "confirm":
		return m.renderConfirm()
	case "error":
		return m.renderError()
	default:
		return ""
	}
}

func (m *App) renderList() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("modu - Worktree Manager"))
	s.WriteString("\n\n")
	s.WriteString(itemStyle.Render("Use arrow keys to navigate, Enter to select, q to quit"))
	s.WriteString("\n\n")

	if len(m.Envs) == 0 {
		s.WriteString(itemStyle.Render("No features found. Use CLI to create one: modu create <feature>"))
		return s.String()
	}

	for i, env := range m.Envs {
		dirtyCount := 0
		for _, mod := range env.Modules {
			if mod.IsDirty {
				dirtyCount++
			}
		}
		status := successStyle.Render("clean")
		if dirtyCount > 0 {
			status = errorStyle.Render(fmt.Sprintf("%d dirty", dirtyCount))
		}

		if i == m.selected {
			s.WriteString(selectedItemStyle.Render(fmt.Sprintf("→ %s (%s)", env.Name, status)))
		} else {
			s.WriteString(itemStyle.Render(fmt.Sprintf("  %s (%s)", env.Name, status)))
		}
		s.WriteString("\n")
	}

	if m.message != "" {
		s.WriteString("\n")
		s.WriteString(successStyle.Render(m.message))
		m.message = ""
	}

	return s.String()
}

func (m *App) renderConfirm() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("Confirm Delete"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("Are you sure you want to delete feature: %s?\n", m.feature))
	s.WriteString(itemStyle.Render("Press 'y' to confirm, 'n' to cancel"))
	s.WriteString("\n\n")
	return s.String()
}

func (m *App) renderError() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("Error"))
	s.WriteString("\n\n")
	s.WriteString(errorStyle.Render(fmt.Sprintf("%v", m.err)))
	s.WriteString("\n\n")
	s.WriteString(itemStyle.Render("Press any key to continue..."))
	return s.String()
}

// Run 启动 TUI
func Run(cfg *config.Config) error {
	p := tea.NewProgram(New(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	return nil
}

// StartTUI 启动 TUI（从 CLI 调用）
func StartTUI(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return Run(cfg)
}

// SelectModules 让用户选择模块（空格选中，上下键切换，回车确认）
// existingModules: 已存在的模块列表，这些模块会被预先选中
func SelectModules(modules []config.Module, existingModules []string) ([]config.Module, error) {
	if len(modules) == 0 {
		return modules, nil
	}

	p := tea.NewProgram(NewModuleSelector(modules, existingModules))
	result, runErr := p.Run()
	if runErr != nil {
		return nil, fmt.Errorf("failed to run module selector: %w", runErr)
	}

	// 类型断言一定会成功，因为 Program 已经成功返回
	selector, _ := result.(*ModuleSelector)
	return selector.SelectedModules(), nil
}

// ModuleSelector 模块选择器
type ModuleSelector struct {
	modules  []config.Module
	selected []bool
	cursor   int
	quitting bool
}

func NewModuleSelector(modules []config.Module, existingModules []string) *ModuleSelector {
	selected := make([]bool, len(modules))

	// 创建已存在模块的 map
	existingMap := make(map[string]bool)
	for _, name := range existingModules {
		existingMap[name] = true
	}

	// 预先选中已存在的模块
	for i, m := range modules {
		if existingMap[m.Name] {
			selected[i] = true
		}
	}

	return &ModuleSelector{
		modules:  modules,
		selected: selected,
	}
}

func (m *ModuleSelector) Init() tea.Cmd {
	return nil
}

func (m *ModuleSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.modules)-1 {
				m.cursor++
			}
		case " ":
			// 空格切换选中状态
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			// 回车确认，返回选中的模块
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ModuleSelector) View() string {
	var s strings.Builder
	s.WriteString("选择要创建的模块（空格选中，回车确认）:\n\n")

	for i, module := range m.modules {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[x]"
		}

		s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, checkbox, module.Name))
	}

	s.WriteString("\n按 q 退出，空格切换选择，回车确认\n")
	return s.String()
}

func (m *ModuleSelector) SelectedModules() []config.Module {
	var result []config.Module
	for i, module := range m.modules {
		if m.selected[i] {
			result = append(result, module)
		}
	}
	// 如果没有选中任何模块，返回空切片
	return result
}
