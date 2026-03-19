package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/qimaotech/modu/internal/config"
	"github.com/qimaotech/modu/internal/core"
	"github.com/qimaotech/modu/internal/engine"
)

// TestMainProjectEntry_IsMainProject 主项目列表项返回 true
func TestMainProjectEntry_IsMainProject(t *testing.T) {
	e := &MainProjectEntry{&engine.MainProjectStatus{Name: "main"}}
	if !e.IsMainProject() {
		t.Error("MainProjectEntry.IsMainProject() 应为 true")
	}
}

// TestMainProjectEntry_GetName 主项目返回名称
func TestMainProjectEntry_GetName(t *testing.T) {
	e := &MainProjectEntry{&engine.MainProjectStatus{Name: "my-main"}}
	if got := e.GetName(); got != "my-main" {
		t.Errorf("GetName() = %q, 期望 my-main", got)
	}
}

// TestMainProjectEntry_GetDirtyCount_Clean 主项目干净时返回 0
func TestMainProjectEntry_GetDirtyCount_Clean(t *testing.T) {
	e := &MainProjectEntry{&engine.MainProjectStatus{IsDirty: false}}
	if got := e.GetDirtyCount(); got != 0 {
		t.Errorf("GetDirtyCount() = %d, 期望 0", got)
	}
}

// TestMainProjectEntry_GetDirtyCount_Dirty 主项目脏时返回 1
func TestMainProjectEntry_GetDirtyCount_Dirty(t *testing.T) {
	e := &MainProjectEntry{&engine.MainProjectStatus{IsDirty: true}}
	if got := e.GetDirtyCount(); got != 1 {
		t.Errorf("GetDirtyCount() = %d, 期望 1", got)
	}
}

// TestFeatureEntry_IsMainProject feature 列表项返回 false
func TestFeatureEntry_IsMainProject(t *testing.T) {
	e := &FeatureEntry{&core.WorktreeEnv{Name: "feat-a"}}
	if e.IsMainProject() {
		t.Error("FeatureEntry.IsMainProject() 应为 false")
	}
}

// TestFeatureEntry_GetName feature 返回名称
func TestFeatureEntry_GetName(t *testing.T) {
	e := &FeatureEntry{&core.WorktreeEnv{Name: "feat-x"}}
	if got := e.GetName(); got != "feat-x" {
		t.Errorf("GetName() = %q, 期望 feat-x", got)
	}
}

// TestFeatureEntry_GetDirtyCount_AllClean 无脏模块时返回 0
func TestFeatureEntry_GetDirtyCount_AllClean(t *testing.T) {
	e := &FeatureEntry{&core.WorktreeEnv{
		Modules: []core.ModuleStatus{
			{Name: "m1", IsDirty: false},
			{Name: "m2", IsDirty: false},
		},
	}}
	if got := e.GetDirtyCount(); got != 0 {
		t.Errorf("GetDirtyCount() = %d, 期望 0", got)
	}
}

// TestFeatureEntry_GetDirtyCount_SomeDirty 部分脏模块时返回脏数量
func TestFeatureEntry_GetDirtyCount_SomeDirty(t *testing.T) {
	e := &FeatureEntry{&core.WorktreeEnv{
		Modules: []core.ModuleStatus{
			{Name: "m1", IsDirty: true},
			{Name: "m2", IsDirty: false},
			{Name: "m3", IsDirty: true},
		},
	}}
	if got := e.GetDirtyCount(); got != 2 {
		t.Errorf("GetDirtyCount() = %d, 期望 2", got)
	}
}

// TestFeatureEntry_GetDirtyCount_NoModules 无模块时返回 0
func TestFeatureEntry_GetDirtyCount_NoModules(t *testing.T) {
	e := &FeatureEntry{&core.WorktreeEnv{Modules: nil}}
	if got := e.GetDirtyCount(); got != 0 {
		t.Errorf("GetDirtyCount() = %d, 期望 0", got)
	}
}

// TestApp_listEntryCount_NoMain 无主项目时仅统计 Envs
func TestApp_listEntryCount_NoMain(t *testing.T) {
	app := &App{Envs: []core.WorktreeEnv{{Name: "a"}, {Name: "b"}}}
	if got := app.listEntryCount(); got != 2 {
		t.Errorf("listEntryCount() = %d, 期望 2", got)
	}
}

// TestApp_listEntryCount_WithMain 有主项目时为主项目数 + Envs 数
func TestApp_listEntryCount_WithMain(t *testing.T) {
	app := &App{
		mainProject: &engine.MainProjectStatus{Name: "main"},
		Envs:        []core.WorktreeEnv{{Name: "f1"}},
	}
	if got := app.listEntryCount(); got != 2 {
		t.Errorf("listEntryCount() = %d, 期望 2", got)
	}
}

// TestApp_listEntryCount_Empty 无主项目且无 Envs 时为 0
func TestApp_listEntryCount_Empty(t *testing.T) {
	app := &App{Envs: nil}
	if got := app.listEntryCount(); got != 0 {
		t.Errorf("listEntryCount() = %d, 期望 0", got)
	}
}

// TestApp_selectedListEntry_WithMain_SelectFirst 有主项目且选中第一项时返回主项目
func TestApp_selectedListEntry_WithMain_SelectFirst(t *testing.T) {
	main := &engine.MainProjectStatus{Name: "main"}
	app := &App{mainProject: main, Envs: []core.WorktreeEnv{{Name: "f1"}}, selected: 0}
	entry := app.selectedListEntry()
	if entry == nil || !entry.IsMainProject() || entry.GetName() != "main" {
		t.Errorf("selectedListEntry() 应为主项目, got %v", entry)
	}
}

// TestApp_selectedListEntry_WithMain_SelectFeature 有主项目且选中第二项时返回 feature
func TestApp_selectedListEntry_WithMain_SelectFeature(t *testing.T) {
	app := &App{
		mainProject: &engine.MainProjectStatus{Name: "main"},
		Envs:        []core.WorktreeEnv{{Name: "feat-a"}},
		selected:    1,
	}
	entry := app.selectedListEntry()
	if entry == nil || entry.IsMainProject() || entry.GetName() != "feat-a" {
		t.Errorf("selectedListEntry() 应为 feat-a, got %v", entry)
	}
}

// TestApp_selectedListEntry_NoMain_SelectFeature 无主项目时按索引选 feature
func TestApp_selectedListEntry_NoMain_SelectFeature(t *testing.T) {
	app := &App{
		Envs:     []core.WorktreeEnv{{Name: "f1"}, {Name: "f2"}},
		selected: 1,
	}
	entry := app.selectedListEntry()
	if entry == nil || entry.GetName() != "f2" {
		t.Errorf("selectedListEntry() 应为 f2, got %v", entry)
	}
}

// TestApp_selectedListEntry_OutOfRange 选中超出范围时返回 nil
func TestApp_selectedListEntry_OutOfRange(t *testing.T) {
	app := &App{mainProject: &engine.MainProjectStatus{}, Envs: []core.WorktreeEnv{}, selected: 5}
	entry := app.selectedListEntry()
	if entry != nil {
		t.Errorf("selectedListEntry() 超出范围时应为 nil, got %v", entry)
	}
}

// TestApp_selectedFeatureEnv_WithMain_SelectFirst 选中主项目时返回 nil
func TestApp_selectedFeatureEnv_WithMain_SelectFirst(t *testing.T) {
	app := &App{
		mainProject: &engine.MainProjectStatus{Name: "main"},
		Envs:        []core.WorktreeEnv{{Name: "f1"}},
		selected:    0,
	}
	env := app.selectedFeatureEnv()
	if env != nil {
		t.Errorf("选中主项目时 selectedFeatureEnv() 应为 nil, got %v", env)
	}
}

// TestApp_selectedFeatureEnv_WithMain_SelectFeature 选中 feature 时返回对应环境
func TestApp_selectedFeatureEnv_WithMain_SelectFeature(t *testing.T) {
	app := &App{
		mainProject: &engine.MainProjectStatus{Name: "main"},
		Envs:        []core.WorktreeEnv{{Name: "feat-x"}},
		selected:    1,
	}
	env := app.selectedFeatureEnv()
	if env == nil || env.Name != "feat-x" {
		t.Errorf("selectedFeatureEnv() 应为 feat-x, got %v", env)
	}
}

// TestApp_selectedFeatureEnv_NoMain 无主项目时按 selected 索引返回 Env
func TestApp_selectedFeatureEnv_NoMain(t *testing.T) {
	app := &App{
		Envs:     []core.WorktreeEnv{{Name: "a"}, {Name: "b"}},
		selected: 1,
	}
	env := app.selectedFeatureEnv()
	if env == nil || env.Name != "b" {
		t.Errorf("selectedFeatureEnv() 应为 b, got %v", env)
	}
}

// TestNewModuleSelector_EmptyModules 空模块列表
func TestNewModuleSelector_EmptyModules(t *testing.T) {
	sel := NewModuleSelector(nil, nil, nil)
	if sel == nil || len(sel.modules) != 0 {
		t.Error("NewModuleSelector(nil, nil) 应返回空模块列表")
	}
}

// TestNewModuleSelector_PreSelectExisting 已存在模块应被预先选中
func TestNewModuleSelector_PreSelectExisting(t *testing.T) {
	modules := []config.Module{
		{Name: "m1", URL: "u1"},
		{Name: "m2", URL: "u2"},
		{Name: "m3", URL: "u3"},
	}
	existing := []string{"m2"}
	sel := NewModuleSelector(modules, existing, nil)
	if len(sel.selected) != 3 {
		t.Fatalf("selected 长度应为 3, got %d", len(sel.selected))
	}
	if !sel.selected[1] {
		t.Error("m2 应被预先选中")
	}
	if sel.selected[0] || sel.selected[2] {
		t.Error("m1、m3 不应被选中")
	}
}

// TestModuleSelector_SelectedModules_None 未选中任何模块时返回空切片
func TestModuleSelector_SelectedModules_None(t *testing.T) {
	modules := []config.Module{{Name: "m1"}, {Name: "m2"}}
	sel := NewModuleSelector(modules, nil, nil) // existing 为空，全部未选
	got := sel.SelectedModules()
	if len(got) != 0 {
		t.Errorf("SelectedModules() 应为空, got %v", got)
	}
}

// TestModuleSelector_SelectedModules_All 全部选中时返回全部模块
func TestModuleSelector_SelectedModules_All(t *testing.T) {
	modules := []config.Module{{Name: "a"}, {Name: "b"}}
	sel := NewModuleSelector(modules, []string{"a", "b"}, nil)
	got := sel.SelectedModules()
	if len(got) != 2 {
		t.Fatalf("SelectedModules() 长度应为 2, got %d", len(got))
	}
	if got[0].Name != "a" || got[1].Name != "b" {
		t.Errorf("SelectedModules() = %v", got)
	}
}

// TestModuleSelector_Update_Quit 按 q 退出
func TestModuleSelector_Update_Quit(t *testing.T) {
	modules := []config.Module{{Name: "m1"}}
	sel := NewModuleSelector(modules, nil, nil)
	_, cmd := sel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("按 q 应返回 Quit 命令")
	}
	if !sel.quitting {
		t.Error("quitting 应为 true")
	}
}

// TestModuleSelector_Update_Enter 回车确认退出
func TestModuleSelector_Update_Enter(t *testing.T) {
	sel := NewModuleSelector([]config.Module{{Name: "m1"}}, nil, nil)
	_, cmd := sel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("回车应返回 Quit 命令")
	}
}

// TestModuleSelector_Update_Space 空格切换选中状态
func TestModuleSelector_Update_Space(t *testing.T) {
	modules := []config.Module{{Name: "m1"}, {Name: "m2"}}
	sel := NewModuleSelector(modules, nil, nil)
	if sel.selected[0] {
		t.Fatal("初始 m1 应未选中")
	}
	sel.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !sel.selected[0] {
		t.Error("空格后 m1 应被选中")
	}
	sel.Update(tea.KeyMsg{Type: tea.KeySpace})
	if sel.selected[0] {
		t.Error("再按空格 m1 应取消选中")
	}
}

// TestModuleSelector_View_NonEmpty 有模块时 View 包含模块名
func TestModuleSelector_View_NonEmpty(t *testing.T) {
	sel := NewModuleSelector([]config.Module{{Name: "mod-a"}}, nil, nil)
	view := sel.View()
	if view == "" || len(view) < 4 {
		t.Error("View() 应返回非空字符串")
	}
}

// TestApp_View_Loading 状态为 loading 时显示 Loading
func TestApp_View_Loading(t *testing.T) {
	app := &App{state: "loading"}
	view := app.View()
	if view != "Loading..." {
		t.Errorf("View() = %q, 期望 Loading...", view)
	}
}

// TestApp_View_List_Empty 无条目时列表提示用 CLI 创建
func TestApp_View_List_Empty(t *testing.T) {
	app := &App{state: "list", Envs: nil}
	view := app.renderList()
	if view == "" {
		t.Error("renderList() 不应为空")
	}
}

// TestApp_View_Confirm 确认删除视图包含 feature 名
func TestApp_View_Confirm(t *testing.T) {
	app := &App{state: "confirm", feature: "feat-to-delete"}
	view := app.renderConfirm()
	if view == "" || len(view) < 8 {
		t.Error("renderConfirm() 应包含确认文案")
	}
}

// TestApp_View_Error 错误状态视图包含错误信息
func TestApp_View_Error(t *testing.T) {
	app := &App{state: "error", err: nil}
	view := app.renderError()
	if view == "" {
		t.Error("renderError() 不应为空")
	}
}

// TestNew_App 创建 App 时 state 为 loading
func TestNew_App(t *testing.T) {
	cfg := &config.Config{Workspace: "/ws", WorktreeRoot: "/wt", DefaultBase: "develop"}
	app := New(cfg)
	if app == nil || app.Engine == nil || app.state != "loading" {
		t.Errorf("New(cfg) 应返回 state=loading 的 App, state=%q", app.state)
	}
}

// TestApp_View_List_WithMainAndEnvs 有主项目和 feature 时列表渲染
func TestApp_View_List_WithMainAndEnvs(t *testing.T) {
	app := &App{
		state:       "list",
		mainProject: &engine.MainProjectStatus{Name: "main", IsDirty: false, Branch: "main"},
		Envs:        []core.WorktreeEnv{{Name: "feat-a", Modules: []core.ModuleStatus{{Name: "m1", IsDirty: true}}}},
		selected:    0,
	}
	view := app.renderList()
	if view == "" {
		t.Fatal("renderList() 不应为空")
	}
	if !strings.Contains(view, "main") || !strings.Contains(view, "feat-a") {
		t.Errorf("renderList() 应包含 main 和 feat-a: %s", view)
	}
}

// TestApp_View_Default 未知 state 时 View 返回空字符串
func TestApp_View_Default(t *testing.T) {
	app := &App{state: "unknown"}
	view := app.View()
	if view != "" {
		t.Errorf("未知 state 时 View() 应为空, got %q", view)
	}
}
