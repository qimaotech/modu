package ui

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/qimaotech/modu/internal/config"
	"github.com/qimaotech/modu/internal/core"
	"github.com/qimaotech/modu/internal/engine"
)

// 全局样式
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{
			Dark:  "86", // 浅青色 - 深色背景
			Light: "25", // 深蓝色 - 浅色背景
		}).
		Background(lipgloss.AdaptiveColor{
			Dark:  "236", // 深灰 - 深色背景
			Light: "254", // 浅灰 - 浅色背景
		}).
		Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
			Dark:  "252", // 浅灰 - 深色背景
			Light: "238", // 深灰 - 浅色背景
		})

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{
			Dark:  "86", // 浅青色 - 深色背景
			Light: "21", // 蓝色 - 浅色背景
		}).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
			Dark:  "196", // 红色 - 深色背景
			Light: "196", // 红色 - 浅色背景
		})

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
			Dark:  "82", // 绿色 - 深色背景
			Light: "28", // 深绿 - 浅色背景
		})
)

// 错误定义
var (
	ErrNoSelection               = errors.New("未选中任何项目")
	ErrMainProjectNotFound       = errors.New("主项目信息不存在")
	ErrFeatureWithoutMainProject = errors.New("该 feature 无主项目，无法复制路径")
	ErrFeatureCannotOpen         = errors.New("该 feature 无主项目，无法打开")
	ErrFeatureNotFound           = errors.New("未找到 feature")
)

var (
	isAppInstalled = defaultIsAppInstalled
	startCommand   = defaultStartCommand
)

func defaultIsAppInstalled(app string) bool {
	if strings.TrimSpace(app) == "" || runtime.GOOS != "darwin" {
		return false
	}
	return exec.Command("open", "-Ra", app).Run() == nil
}

func defaultStartCommand(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

// ListEntry 列表项统一接口（主项目或 feature）
type ListEntry interface {
	IsMainProject() bool
	GetName() string
	GetDirtyCount() int
}

type operationMenuItem struct {
	label    string
	shortcut string
	run      func() tea.Cmd
}

func (item operationMenuItem) display() string {
	if item.shortcut == "" {
		return item.label
	}
	return fmt.Sprintf("%s (%s)", item.label, item.shortcut)
}

// MainProjectEntry 主项目列表项
type MainProjectEntry struct {
	*engine.MainProjectStatus
}

func (e *MainProjectEntry) IsMainProject() bool { return true }
func (e *MainProjectEntry) GetName() string     { return e.MainProjectStatus.Name }
func (e *MainProjectEntry) GetDirtyCount() int {
	if e.MainProjectStatus.IsDirty {
		return 1
	}
	return 0
}

// FeatureEntry 包装 WorktreeEnv 实现 ListEntry
type FeatureEntry struct {
	*core.WorktreeEnv
}

func (e *FeatureEntry) IsMainProject() bool { return false }
func (e *FeatureEntry) GetName() string     { return e.WorktreeEnv.Name }
func (e *FeatureEntry) GetDirtyCount() int {
	n := 0
	for _, m := range e.WorktreeEnv.Modules {
		if m.IsDirty {
			n++
		}
	}
	return n
}

// TUI App 状态
type App struct {
	Engine           *engine.Engine
	Envs             []core.WorktreeEnv
	mainProject      *engine.MainProjectStatus
	selected         int
	menuSelected     int    // 操作菜单选中项
	state            string // "loading", "list", "menu", "modules", "confirm", "confirm_unpushed", "error"
	feature          string
	err              error
	message          string
	unpushedBranches []engine.UnpushedBranch
	// 配置化 App 打开工具
	availableAppOpeners []config.AppOpener
	appOpenersResolved  bool
	// 模块管理相关字段
	moduleSelector    *ModuleSelector // 模块选择器
	moduleCursor      int             // 模块列表光标位置
	modulesFeature    string          // 当前操作的 feature 名称
	isMainProjectMenu bool            // 当前菜单是否为主项目菜单
	// 创建 feature 相关字段
	createFeatureInput  []rune // feature 名称输入内容
	createFeatureCursor int    // feature 名称输入光标
	createFeatureError  string // feature 名称输入校验提示
	createFeature       string // 当前待创建的 feature 名称
}

// New 创建 TUI App
func New(cfg *config.Config) *App {
	app := &App{
		Engine: engine.New(cfg),
		state:  "loading",
	}
	app.resolveAvailableAppOpeners()
	return app
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
	main, _ := m.Engine.GetMainProject(context.Background())
	return loadedMsg{envs: envs, mainProject: main}
}

type loadedMsg struct {
	envs        []core.WorktreeEnv
	mainProject *engine.MainProjectStatus
}

// updateDoneMsg 更新代码完成；feature 非空表示本次为 feature worktree 更新
type updateDoneMsg struct {
	success int
	failed  map[string]error
	feature string
}

// refreshListMsg 请求重新加载列表（如模块变更后）
type refreshListMsg struct{}

type createFeatureDoneMsg struct {
	feature string
	err     error
}

type errorMsg struct {
	err error
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		m.Envs = msg.envs
		m.mainProject = msg.mainProject
		m.state = "list"
	case updateDoneMsg:
		if len(msg.failed) == 0 {
			if msg.feature != "" {
				m.message = fmt.Sprintf("更新成功: feature %s（主项目 + %d 个模块）", msg.feature, msg.success-1)
			} else if msg.success == 1 {
				m.message = "更新成功: 主项目"
			} else {
				m.message = fmt.Sprintf("更新成功: 主项目 + %d 个模块", msg.success-1)
			}
		} else {
			names := make([]string, 0, len(msg.failed))
			for name := range msg.failed {
				names = append(names, name)
			}
			m.message = fmt.Sprintf("更新成功: %d 个，失败: %d 个 (%s)", msg.success, len(msg.failed), strings.Join(names, ", "))
		}
		m.state = "loading"
		return m, m.loadEnvs
	case refreshListMsg:
		m.state = "loading"
		return m, m.loadEnvs
	case createFeatureDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = "error"
			return m, nil
		}
		m.message = "已创建 feature: " + msg.feature
		m.resetCreateFeature()
		m.state = "loading"
		return m, m.loadEnvs
	case errorMsg:
		m.err = msg.err
		m.state = "error"
	case tea.KeyMsg:
		switch m.state {
		case "list":
			return m.handleListKey(msg)
		case "menu":
			return m.handleMenuKey(msg)
		case "modules":
			return m.handleModulesKey(msg)
		case "create_input":
			return m.handleCreateInputKey(msg)
		case "create_modules":
			return m.handleCreateModulesKey(msg)
		case "confirm":
			return m.handleConfirmKey(msg)
		case "confirm_unpushed":
			return m.handleUnpushedConfirmKey(msg)
		case "error":
			m.state = "list"
			m.err = nil
		}
	}
	return m, nil
}

func (m *App) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	total := m.listEntryCount()
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < total-1 {
			m.selected++
		}
	case "enter":
		if total > 0 {
			m.state = "menu"
			m.menuSelected = 0
			m.isMainProjectMenu = m.selectedListEntry() != nil && m.selectedListEntry().IsMainProject()
		}
	case "d":
		entry := m.selectedListEntry()
		if entry != nil && !entry.IsMainProject() {
			m.state = "confirm"
			m.feature = entry.GetName()
		}
	case "o":
		if entry := m.selectedListEntry(); entry != nil {
			if entry.IsMainProject() {
				path := m.mainProject.Path
				cmd := exec.Command("code", path)
				_ = cmd.Start()
			} else {
				env := m.selectedFeatureEnv()
				if env != nil && env.MainProject != nil {
					cmd := exec.Command("code", env.MainProject.Path)
					_ = cmd.Start()
				} else {
					m.err = fmt.Errorf("该 feature 无主项目，无法打开: %w", ErrFeatureCannotOpen)
					m.state = "error"
				}
			}
		}
	case "x":
		if entry := m.selectedListEntry(); entry != nil {
			path, err := m.getSelectedPath()
			if err != nil {
				m.err = err
				m.state = "error"
			} else {
				cmd := exec.Command("open", "-a", "Codex", path)
				_ = cmd.Start()
			}
		}
	case "m":
		if env := m.selectedFeatureEnv(); env != nil {
			m.initModuleSelector()
			m.state = "modules"
		}
	case "u":
		if entry := m.selectedListEntry(); entry != nil {
			if entry.IsMainProject() {
				m.state = "loading"
				return m, m.executeUpdateCode()
			}
			m.state = "loading"
			return m, m.executeUpdateWorktree(entry.GetName())
		}
	case "c":
		if entry := m.selectedListEntry(); entry != nil {
			path, err := m.getSelectedPath()
			if err != nil {
				m.err = err
				m.state = "error"
			} else {
				if err := clipboard.WriteAll(path); err == nil {
					m.message = "路径已复制: " + path
				} else {
					m.err = fmt.Errorf("复制失败: %w", err)
					m.state = "error"
				}
			}
		}
	case "n":
		m.startCreateFeature()
	case "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

// selectedFeatureEnv 当前选中的 feature 环境（仅当选中 feature 时有效）
func (m *App) selectedFeatureEnv() *core.WorktreeEnv {
	if m.mainProject != nil {
		if m.selected <= 0 {
			return nil
		}
		idx := m.selected - 1
		if idx < len(m.Envs) {
			return &m.Envs[idx]
		}
		return nil
	}
	if m.selected < len(m.Envs) {
		return &m.Envs[m.selected]
	}
	return nil
}

// getSelectedPath 获取选中项的主项目路径
func (m *App) getSelectedPath() (string, error) {
	entry := m.selectedListEntry()
	if entry == nil {
		return "", fmt.Errorf("未选中任何项目: %w", ErrNoSelection)
	}
	if entry.IsMainProject() {
		if m.mainProject == nil {
			return "", fmt.Errorf("主项目信息不存在: %w", ErrMainProjectNotFound)
		}
		return m.mainProject.Path, nil
	}
	env := m.selectedFeatureEnv()
	if env == nil || env.MainProject == nil {
		return "", fmt.Errorf("该 feature 无主项目，无法复制路径: %w", ErrFeatureWithoutMainProject)
	}
	return env.MainProject.Path, nil
}

// getSelectedOpenPath 获取选中项的主项目路径，用于打开工具。
func (m *App) getSelectedOpenPath() (string, error) {
	entry := m.selectedListEntry()
	if entry == nil {
		return "", fmt.Errorf("未选中任何项目: %w", ErrNoSelection)
	}
	if entry.IsMainProject() {
		if m.mainProject == nil {
			return "", fmt.Errorf("主项目信息不存在: %w", ErrMainProjectNotFound)
		}
		return m.mainProject.Path, nil
	}
	env := m.selectedFeatureEnv()
	if env == nil || env.MainProject == nil {
		return "", fmt.Errorf("该 feature 无主项目，无法打开: %w", ErrFeatureCannotOpen)
	}
	return env.MainProject.Path, nil
}

// copyPathAndBack 复制路径并返回列表视图
func (m *App) copyPathAndBack() {
	path, err := m.getSelectedPath()
	if err != nil {
		m.err = err
		m.state = "error"
		return
	}
	if err := clipboard.WriteAll(path); err != nil {
		m.err = fmt.Errorf("复制失败: %w", err)
		m.state = "error"
		return
	}
	m.message = "路径已复制: " + path
	m.state = "list"
}

func (m *App) ensureAvailableAppOpeners() {
	if !m.appOpenersResolved {
		m.resolveAvailableAppOpeners()
	}
}

func (m *App) resolveAvailableAppOpeners() {
	m.availableAppOpeners = nil
	if m.Engine == nil || m.Engine.Config == nil {
		m.appOpenersResolved = true
		return
	}
	for _, opener := range m.Engine.Config.AppOpeners {
		if isAppInstalled(opener.App) {
			m.availableAppOpeners = append(m.availableAppOpeners, opener)
		}
	}
	m.appOpenersResolved = true
}

func (m *App) operationMenuItems() []operationMenuItem {
	m.ensureAvailableAppOpeners()

	items := []operationMenuItem{
		{label: "打开 VS Code", shortcut: "o", run: m.openVSCodeSelected},
		{label: "打开 Codex", shortcut: "x", run: func() tea.Cmd {
			m.openAppByName("Codex")
			return nil
		}},
	}

	items = append(items, m.configuredAppOpenerMenuItems()...)
	items = append(items, operationMenuItem{label: "复制路径", shortcut: "c", run: func() tea.Cmd {
		m.copyPathAndBack()
		return nil
	}})

	if m.isMainProjectMenu {
		items = append(items, operationMenuItem{label: "更新代码", shortcut: "u", run: func() tea.Cmd {
			m.state = "loading"
			return m.executeUpdateCode()
		}})
		return items
	}

	items = append(items,
		operationMenuItem{label: "更新代码", shortcut: "u", run: func() tea.Cmd {
			if env := m.selectedFeatureEnv(); env != nil {
				m.state = "loading"
				return m.executeUpdateWorktree(env.Name)
			}
			return nil
		}},
		operationMenuItem{label: "Modules 管理", shortcut: "m", run: func() tea.Cmd {
			m.initModuleSelector()
			m.state = "modules"
			return nil
		}},
		operationMenuItem{label: "删除", shortcut: "d", run: func() tea.Cmd {
			if env := m.selectedFeatureEnv(); env != nil {
				m.state = "confirm"
				m.feature = env.Name
			}
			return nil
		}},
	)
	return items
}

func (m *App) configuredAppOpenerMenuItems() []operationMenuItem {
	reserved := builtinMenuShortcuts()
	shortcutCounts := make(map[string]int)
	for _, opener := range m.availableAppOpeners {
		key := normalizeMenuShortcut(opener.Shortcut)
		if key != "" {
			shortcutCounts[key]++
		}
	}

	items := make([]operationMenuItem, 0, len(m.availableAppOpeners))
	for _, opener := range m.availableAppOpeners {
		opener := opener
		label := "打开 " + appOpenerLabel(opener)
		shortcut := normalizeMenuShortcut(opener.Shortcut)
		if shortcut != "" && (reserved[shortcut] || shortcutCounts[shortcut] > 1) {
			shortcut = ""
		}
		items = append(items, operationMenuItem{label: label, shortcut: shortcut, run: func() tea.Cmd {
			m.openAppByName(opener.App)
			return nil
		}})
	}
	return items
}

func builtinMenuShortcuts() map[string]bool {
	return map[string]bool{
		"o": true,
		"x": true,
		"c": true,
		"u": true,
		"m": true,
		"d": true,
		"q": true,
	}
}

func normalizeMenuShortcut(shortcut string) string {
	if shortcut == "" {
		return ""
	}
	return strings.ToLower(shortcut)
}

func appOpenerLabel(opener config.AppOpener) string {
	if strings.TrimSpace(opener.Label) != "" {
		return strings.TrimSpace(opener.Label)
	}
	return strings.TrimSpace(opener.App)
}

func (m *App) openVSCodeSelected() tea.Cmd {
	path, err := m.getSelectedOpenPath()
	if err != nil {
		m.err = err
		m.state = "error"
		return nil
	}
	_ = startCommand("code", path)
	m.state = "list"
	return nil
}

func (m *App) openAppByName(appName string) {
	path, err := m.getSelectedOpenPath()
	if err != nil {
		m.err = err
		m.state = "error"
		return
	}
	_ = startCommand("open", "-a", appName, path)
	m.state = "list"
}

func (m *App) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		branches, err := m.Engine.CheckDeleteUnpushedBranches(context.Background(), m.feature)
		if err != nil {
			m.err = err
			m.state = "error"
			return m, nil
		}
		if len(branches) > 0 {
			m.unpushedBranches = branches
			m.state = "confirm_unpushed"
			return m, nil
		}
		return m.deleteFeature(false)
	case "n", "esc":
		m.state = "list"
	}
	return m, nil
}

// handleUnpushedConfirmKey 处理未推送分支风险确认。
func (m *App) handleUnpushedConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		return m.deleteFeature(true)
	case "n", "esc":
		m.unpushedBranches = nil
		m.state = "list"
	}
	return m, nil
}

// deleteFeature 执行 feature 删除并刷新列表。
func (m *App) deleteFeature(allowUnpushedBranches bool) (tea.Model, tea.Cmd) {
	err := m.Engine.DeleteWorktreeWithOptions(context.Background(), m.feature, false, engine.DeleteOptions{
		AllowUnpushedBranches: allowUnpushedBranches,
	})
	if err != nil {
		m.err = err
		m.state = "error"
		return m, nil
	}
	m.message = "已删除 feature: " + m.feature
	m.unpushedBranches = nil
	m.state = "loading"
	m.selected = 0
	return m, m.loadEnvs
}

// initModuleSelector 初始化模块选择器
func (m *App) initModuleSelector() {
	env := m.selectedFeatureEnv()
	if env == nil {
		return
	}
	m.modulesFeature = env.Name
	if env.Branch != "" {
		m.modulesFeature = env.Branch
	}

	existingModules := make([]string, len(env.Modules))
	for i, mod := range env.Modules {
		existingModules[i] = mod.Name
	}

	// 创建模块选择器，预先选中已存在的模块
	m.moduleSelector = NewModuleSelector(m.Engine.Config.Modules, existingModules, nil, nil, "选择模块")
	m.moduleCursor = 0
}

func (m *App) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuItems := m.operationMenuItems()
	if len(menuItems) == 0 {
		m.state = "list"
		return m, nil
	}
	if m.menuSelected >= len(menuItems) {
		m.menuSelected = len(menuItems) - 1
	}

	switch msg.String() {
	case "up", "k":
		if m.menuSelected > 0 {
			m.menuSelected--
		}
	case "down", "j":
		if m.menuSelected < len(menuItems)-1 {
			m.menuSelected++
		}
	case "enter":
		return m, menuItems[m.menuSelected].run()
	case "q", "esc":
		m.state = "list"
	default:
		key := normalizeMenuShortcut(msg.String())
		for _, item := range menuItems {
			if item.shortcut != "" && item.shortcut == key {
				return m, item.run()
			}
		}
	}
	return m, nil
}

func (m *App) handleModulesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.moduleCursor > 0 {
			m.moduleCursor--
		}
	case "down", "j":
		if m.moduleCursor < len(m.moduleSelector.modules)-1 {
			m.moduleCursor++
		}
	case " ":
		// 空格切换选中状态
		if m.moduleCursor < len(m.moduleSelector.selected) {
			m.moduleSelector.selected[m.moduleCursor] = !m.moduleSelector.selected[m.moduleCursor]
		}
	case "enter":
		// 回车确认，执行模块增删
		return m, m.executeModulesChange
	case "q", "esc":
		// 返回操作菜单
		m.state = "menu"
	}
	return m, nil
}

func (m *App) startCreateFeature() {
	m.state = "create_input"
	m.createFeatureInput = nil
	m.createFeatureCursor = 0
	m.createFeatureError = ""
	m.createFeature = ""
	m.moduleSelector = nil
	m.moduleCursor = 0
}

func (m *App) resetCreateFeature() {
	m.createFeatureInput = nil
	m.createFeatureCursor = 0
	m.createFeatureError = ""
	m.createFeature = ""
	m.moduleSelector = nil
	m.moduleCursor = 0
}

func (m *App) cancelCreateFeature() {
	m.resetCreateFeature()
	m.state = "list"
}

func (m *App) handleCreateInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.clampCreateFeatureCursor()

	switch msg.String() {
	case "esc", "ctrl+c":
		m.cancelCreateFeature()
	case "enter":
		feature := strings.TrimSpace(string(m.createFeatureInput))
		if feature == "" {
			m.createFeatureError = "feature 名称不能为空"
			return m, nil
		}
		m.createFeature = feature
		m.createFeatureError = ""
		m.initCreateModuleSelector()
		m.state = "create_modules"
	case "left":
		if m.createFeatureCursor > 0 {
			m.createFeatureCursor--
		}
	case "right":
		if m.createFeatureCursor < len(m.createFeatureInput) {
			m.createFeatureCursor++
		}
	case "home":
		m.createFeatureCursor = 0
	case "end":
		m.createFeatureCursor = len(m.createFeatureInput)
	case "backspace", "ctrl+h":
		if m.createFeatureCursor > 0 {
			m.createFeatureInput = append(m.createFeatureInput[:m.createFeatureCursor-1], m.createFeatureInput[m.createFeatureCursor:]...)
			m.createFeatureCursor--
			m.createFeatureError = ""
		}
	case "delete":
		if m.createFeatureCursor < len(m.createFeatureInput) {
			m.createFeatureInput = append(m.createFeatureInput[:m.createFeatureCursor], m.createFeatureInput[m.createFeatureCursor+1:]...)
			m.createFeatureError = ""
		}
	default:
		if len(msg.Runes) > 0 {
			input := make([]rune, 0, len(m.createFeatureInput)+len(msg.Runes))
			input = append(input, m.createFeatureInput[:m.createFeatureCursor]...)
			input = append(input, msg.Runes...)
			input = append(input, m.createFeatureInput[m.createFeatureCursor:]...)
			m.createFeatureInput = input
			m.createFeatureCursor += len(msg.Runes)
			m.createFeatureError = ""
		}
	}
	return m, nil
}

func (m *App) clampCreateFeatureCursor() {
	if m.createFeatureCursor < 0 {
		m.createFeatureCursor = 0
	}
	if m.createFeatureCursor > len(m.createFeatureInput) {
		m.createFeatureCursor = len(m.createFeatureInput)
	}
}

func (m *App) initCreateModuleSelector() {
	var modules []config.Module
	var defaultSelected []string
	remoteHasBranch := make(map[string]bool)

	if m.Engine != nil && m.Engine.Config != nil {
		modules = m.Engine.Config.Modules
		defaultSelected = m.Engine.Config.DefaultSelectedModules
		var err error
		remoteHasBranch, err = m.Engine.GetModulesWithRemoteBranch(context.Background(), m.createFeature)
		if err != nil {
			remoteHasBranch = make(map[string]bool)
		}
	}

	m.moduleSelector = NewModuleSelector(modules, nil, remoteHasBranch, defaultSelected, "选择要创建的项目/模块")
	m.moduleCursor = 0
}

func (m *App) handleCreateModulesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.moduleCursor > 0 {
			m.moduleCursor--
		}
	case "down", "j":
		if m.moduleSelector != nil && m.moduleCursor < len(m.moduleSelector.modules)-1 {
			m.moduleCursor++
		}
	case " ":
		if m.moduleSelector != nil && m.moduleCursor >= 0 && m.moduleCursor < len(m.moduleSelector.selected) {
			m.moduleSelector.selected[m.moduleCursor] = !m.moduleSelector.selected[m.moduleCursor]
		}
	case "enter":
		m.state = "loading"
		return m, m.executeCreateFeature()
	case "q", "esc", "ctrl+c":
		m.cancelCreateFeature()
	}
	return m, nil
}

// executeUpdateCode 执行主项目+模块更新，返回在后台运行并发送 updateDoneMsg 的 Cmd
func (m *App) executeUpdateCode() tea.Cmd {
	return func() tea.Msg {
		success, failed := m.Engine.UpdateMainProject(context.Background())
		return updateDoneMsg{success: success, failed: failed, feature: ""}
	}
}

// executeUpdateWorktree 执行指定 feature 的 worktree 更新
func (m *App) executeUpdateWorktree(feature string) tea.Cmd {
	return func() tea.Msg {
		success, failed := m.Engine.UpdateWorktree(context.Background(), feature)
		return updateDoneMsg{success: success, failed: failed, feature: feature}
	}
}

func (m *App) executeCreateFeature() tea.Cmd {
	feature := m.createFeature
	var selectedModules []config.Module
	if m.moduleSelector != nil {
		selectedModules = m.moduleSelector.SelectedModules()
	}

	return func() tea.Msg {
		if m.Engine == nil || m.Engine.Config == nil || m.Engine.GitProxy == nil {
			return createFeatureDoneMsg{feature: feature, err: errors.New("TUI 引擎未初始化")}
		}

		createCfg := *m.Engine.Config
		createCfg.Modules = append([]config.Module(nil), selectedModules...)
		createEngine := engine.NewWithClient(&createCfg, m.Engine.GitProxy)
		if err := createEngine.CreateWorktree(context.Background(), feature, createCfg.DefaultBase); err != nil {
			return createFeatureDoneMsg{feature: feature, err: err}
		}
		return createFeatureDoneMsg{feature: feature}
	}
}

func (m *App) executeModulesChange() tea.Msg {
	selectedModules := m.moduleSelector.SelectedModules()

	var env *core.WorktreeEnv
	for i := range m.Envs {
		if m.Envs[i].Name == m.modulesFeature || m.Envs[i].Branch == m.modulesFeature {
			env = &m.Envs[i]
			break
		}
	}
	if env == nil {
		return errorMsg{fmt.Errorf("未找到 feature %s: %w", m.modulesFeature, ErrFeatureNotFound)}
	}
	existingModules := make(map[string]bool)
	for _, mod := range env.Modules {
		existingModules[mod.Name] = true
	}

	// 分类：需要添加的和需要删除的
	var toAdd []string
	var toRemove []string

	for _, mod := range selectedModules {
		if !existingModules[mod.Name] {
			toAdd = append(toAdd, mod.Name)
		}
	}

	for _, mod := range env.Modules {
		// 检查是否在选中列表中
		found := false
		for _, sel := range selectedModules {
			if sel.Name == mod.Name {
				found = true
				break
			}
		}
		if !found {
			toRemove = append(toRemove, mod.Name)
		}
	}

	// 执行添加
	for _, modName := range toAdd {
		err := m.Engine.AddModule(context.Background(), m.modulesFeature, modName)
		if err != nil {
			return errorMsg{fmt.Errorf("添加模块 %s 失败: %w", modName, err)}
		}
	}

	// 执行删除
	for _, modName := range toRemove {
		err := m.Engine.RemoveModule(context.Background(), m.modulesFeature, modName)
		if err != nil {
			return errorMsg{fmt.Errorf("删除模块 %s 失败: %w", modName, err)}
		}
	}

	m.message = fmt.Sprintf("模块已更新: 添加 %d, 删除 %d", len(toAdd), len(toRemove))
	return refreshListMsg{}
}

func (m *App) View() string {
	switch m.state {
	case "loading":
		return "Loading..."
	case "list":
		return m.renderList()
	case "menu":
		return m.renderMenu()
	case "modules":
		return m.renderModules()
	case "create_input":
		return m.renderCreateInput()
	case "create_modules":
		return m.renderCreateModules()
	case "confirm":
		return m.renderConfirm()
	case "confirm_unpushed":
		return m.renderUnpushedConfirm()
	case "error":
		return m.renderError()
	default:
		return ""
	}
}

// listEntryCount 列表总条数（主项目 + features）
func (m *App) listEntryCount() int {
	n := len(m.Envs)
	if m.mainProject != nil {
		n++
	}
	return n
}

// selectedListEntry 当前选中的列表项，可能为主项目或 feature
func (m *App) selectedListEntry() ListEntry {
	if m.mainProject != nil {
		if m.selected == 0 {
			return &MainProjectEntry{m.mainProject}
		}
		idx := m.selected - 1
		if idx < len(m.Envs) {
			return &FeatureEntry{&m.Envs[idx]}
		}
	} else if m.selected < len(m.Envs) {
		return &FeatureEntry{&m.Envs[m.selected]}
	}
	return nil
}

func (m *App) renderList() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("modu - Worktree Manager"))
	s.WriteString("\n\n")
	s.WriteString(itemStyle.Render("↑/↓ 选择  Enter 回车  n 新建 feature  m 管理模块  u 更新代码  c 复制路径\nd 删除  o 打开 VS Code  x 打开 Codex  q/esc 退出"))
	s.WriteString("\n\n")

	total := m.listEntryCount()
	if total == 0 {
		s.WriteString(itemStyle.Render("No features found. Use CLI to create one: modu create <feature>"))
		return s.String()
	}

	row := 0
	if m.mainProject != nil {
		status := successStyle.Render("clean")
		if m.mainProject.IsDirty {
			status = errorStyle.Render("dirty")
		}
		line := fmt.Sprintf("→ %s [主项目] (%s) [%s]", m.mainProject.Name, status, m.mainProject.Branch)
		if m.selected == 0 {
			s.WriteString(selectedItemStyle.Render(line))
		} else {
			s.WriteString(itemStyle.Render(line))
		}
		s.WriteString("\n")
		row++
	}
	for i := range m.Envs {
		env := &m.Envs[i]
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
		prefix := "  "
		if m.selected == row {
			prefix = "→ "
		}
		line := fmt.Sprintf("%s%s (%s)", prefix, env.Name, status)
		if m.selected == row {
			s.WriteString(selectedItemStyle.Render(line))
		} else {
			s.WriteString(itemStyle.Render(line))
		}
		s.WriteString("\n")
		row++
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
	s.WriteString(headerStyle.Render("确认删除"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("确定要删除 feature「%s」吗？\n", m.feature))
	s.WriteString(itemStyle.Render("按 y 确认，n 取消"))
	s.WriteString("\n\n")
	return s.String()
}

func (m *App) renderUnpushedConfirm() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("确认删除未推送分支"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("feature「%s」包含尚未确认推送到远程的本地分支：\n\n", m.feature))
	for _, branch := range m.unpushedBranches {
		reason := branch.Reason
		if reason == "" {
			reason = "not pushed"
		}
		s.WriteString(fmt.Sprintf("- %s: %s (%s)\n", branch.Name, branch.Branch, reason))
	}
	s.WriteString("\n")
	s.WriteString(itemStyle.Render("按 y 删除本地分支，n 取消"))
	s.WriteString("\n\n")
	return s.String()
}

func (m *App) renderMenu() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("操作菜单"))
	s.WriteString("\n\n")

	if m.isMainProjectMenu && m.mainProject != nil {
		s.WriteString(fmt.Sprintf("当前选中: %s [主项目] (<dirty状态>)\n\n", m.mainProject.Name))
	} else {
		if env := m.selectedFeatureEnv(); env != nil {
			s.WriteString(fmt.Sprintf("当前选中: %s\n\n", env.Name))
		}
	}
	menuItems := m.operationMenuItems()
	if m.menuSelected >= len(menuItems) && len(menuItems) > 0 {
		m.menuSelected = len(menuItems) - 1
	}
	for i, item := range menuItems {
		if i == m.menuSelected {
			s.WriteString(selectedItemStyle.Render(fmt.Sprintf("→ %s", item.display())))
		} else {
			s.WriteString(itemStyle.Render(fmt.Sprintf("  %s", item.display())))
		}
		s.WriteString("\n")
	}
	s.WriteString("\n")
	s.WriteString(itemStyle.Render("按 ↑/↓ 选择，Enter 执行，esc 返回"))
	return s.String()
}

func (m *App) renderModules() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("模块管理"))
	s.WriteString("\n\n")

	// 显示当前操作的 feature
	s.WriteString(fmt.Sprintf("当前 feature: %s\n\n", m.modulesFeature))

	if m.moduleSelector == nil {
		s.WriteString(itemStyle.Render("加载中..."))
		return s.String()
	}

	// 显示所有模块
	for i, module := range m.moduleSelector.modules {
		cursor := " "
		if m.moduleCursor == i {
			cursor = ">"
		}

		checkbox := "[ ]"
		if m.moduleSelector.selected[i] {
			checkbox = "[x]"
		}

		if m.moduleCursor == i {
			s.WriteString(selectedItemStyle.Render(fmt.Sprintf("%s %s %s", cursor, checkbox, module.Name)))
		} else {
			s.WriteString(itemStyle.Render(fmt.Sprintf("%s %s %s", cursor, checkbox, module.Name)))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(itemStyle.Render("空格切换选择，Enter 确认，q/esc 返回"))
	return s.String()
}

func (m *App) renderCreateInput() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("新建 feature"))
	s.WriteString("\n\n")
	s.WriteString("请输入 feature 名称:\n\n")
	s.WriteString(selectedItemStyle.Render(m.renderCreateFeatureInputValue()))
	s.WriteString("\n\n")
	if m.createFeatureError != "" {
		s.WriteString(errorStyle.Render(m.createFeatureError))
		s.WriteString("\n\n")
	}
	s.WriteString(itemStyle.Render("Enter 继续，Esc/Ctrl+C 取消"))
	return s.String()
}

func (m *App) renderCreateFeatureInputValue() string {
	m.clampCreateFeatureCursor()
	before := string(m.createFeatureInput[:m.createFeatureCursor])
	after := string(m.createFeatureInput[m.createFeatureCursor:])
	return before + "▌" + after
}

func (m *App) renderCreateModules() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("新建 feature"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("当前 feature: %s\n\n", m.createFeature))
	s.WriteString(successStyle.Render(fmt.Sprintf("[x] 主项目: %s（始终创建）", m.createMainProjectName())))
	s.WriteString("\n")

	if m.moduleSelector == nil || len(m.moduleSelector.modules) == 0 {
		s.WriteString(itemStyle.Render("未配置可选模块，将只创建主项目。"))
		s.WriteString("\n\n")
		s.WriteString(itemStyle.Render("Enter 创建，q/esc 取消"))
		return s.String()
	}

	for i, module := range m.moduleSelector.modules {
		cursor := " "
		if m.moduleCursor == i {
			cursor = ">"
		}

		checkbox := "[ ]"
		if m.moduleSelector.selected[i] {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("%s %s %s", cursor, checkbox, module.Name)
		if m.moduleCursor == i {
			s.WriteString(selectedItemStyle.Render(line))
		} else {
			s.WriteString(itemStyle.Render(line))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(itemStyle.Render("空格切换模块，Enter 创建，q/esc 取消"))
	return s.String()
}

func (m *App) createMainProjectName() string {
	if m.mainProject != nil && m.mainProject.Name != "" {
		return m.mainProject.Name
	}
	if m.Engine != nil && m.Engine.Config != nil && m.Engine.Config.Workspace != "" {
		return filepath.Base(m.Engine.Config.Workspace)
	}
	return "主项目"
}

func (m *App) renderError() string {
	var s strings.Builder
	s.WriteString(headerStyle.Render("错误"))
	s.WriteString("\n\n")
	s.WriteString(errorStyle.Render(fmt.Sprintf("%v", m.err)))
	s.WriteString("\n\n")
	s.WriteString(itemStyle.Render("按任意键继续..."))
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
// remoteHasBranch: 远端是否有该分支的模块，预先选中这些模块
// defaultSelectedModules: 默认选中的模块列表，这些模块会被预先选中
// title: 自定义标题，空字符串使用默认标题
// 返回: 选中的模块列表, 用户是否按 q/ctrl+c 退出
func SelectModules(modules []config.Module, existingModules []string, remoteHasBranch map[string]bool, defaultSelectedModules []string, title string) ([]config.Module, bool, error) {
	if len(modules) == 0 {
		return modules, false, nil
	}

	p := tea.NewProgram(NewModuleSelector(modules, existingModules, remoteHasBranch, defaultSelectedModules, title))
	result, runErr := p.Run()
	if runErr != nil {
		return nil, false, fmt.Errorf("failed to run module selector: %w", runErr)
	}

	// 类型断言一定会成功，因为 Program 已经成功返回
	selector, _ := result.(*ModuleSelector)
	return selector.SelectedModules(), selector.quitting, nil
}

// ModuleSelector 模块选择器
type ModuleSelector struct {
	modules  []config.Module
	selected []bool
	cursor   int
	quitting bool
	title    string // 自定义标题
}

func NewModuleSelector(modules []config.Module, existingModules []string, remoteHasBranch map[string]bool, defaultSelectedModules []string, title string) *ModuleSelector {
	selected := make([]bool, len(modules))

	// 创建已存在模块的 map
	existingMap := make(map[string]bool)
	for _, name := range existingModules {
		existingMap[name] = true
	}

	// 确保 remoteHasBranch 不为 nil，避免空指针
	if remoteHasBranch == nil {
		remoteHasBranch = make(map[string]bool)
	}

	// 创建默认选中模块的 map
	defaultSelectedMap := make(map[string]bool)
	for _, name := range defaultSelectedModules {
		defaultSelectedMap[name] = true
	}

	// 预先选中：已存在的模块 OR 远端有该分支的模块 OR 配置中默认选中的模块
	for i, m := range modules {
		if existingMap[m.Name] || remoteHasBranch[m.Name] || defaultSelectedMap[m.Name] {
			selected[i] = true
		}
	}

	// 默认标题
	if title == "" {
		title = "选择要创建的模块"
	}

	return &ModuleSelector{
		modules:  modules,
		selected: selected,
		title:    title,
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
	s.WriteString(m.title + "（空格选中，回车确认）:\n\n")

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
