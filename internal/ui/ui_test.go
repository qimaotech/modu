package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/qimaotech/modu/internal/config"
	"github.com/qimaotech/modu/internal/core"
	"github.com/qimaotech/modu/internal/engine"
	errs "github.com/qimaotech/modu/internal/errors"
	"github.com/qimaotech/modu/internal/gitproxy"
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
	sel := NewModuleSelector(nil, nil, nil, nil, "")
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
	sel := NewModuleSelector(modules, existing, nil, nil, "")
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

// TestNewModuleSelector_PreSelectDefaultSelected 配置中默认选中的模块应被预先选中
func TestNewModuleSelector_PreSelectDefaultSelected(t *testing.T) {
	modules := []config.Module{
		{Name: "m1", URL: "u1"},
		{Name: "m2", URL: "u2"},
		{Name: "m3", URL: "u3"},
	}
	defaultSelected := []string{"m1", "m3"}
	sel := NewModuleSelector(modules, nil, nil, defaultSelected, "")
	if len(sel.selected) != 3 {
		t.Fatalf("selected 长度应为 3, got %d", len(sel.selected))
	}
	if !sel.selected[0] || !sel.selected[2] {
		t.Error("m1、m3 应被预先选中")
	}
	if sel.selected[1] {
		t.Error("m2 不应被选中")
	}
}

// TestNewModuleSelector_PreSelectDefaultSelectedAndRemote 默认选中与远程分支同时存在时都应选中
func TestNewModuleSelector_PreSelectDefaultSelectedAndRemote(t *testing.T) {
	modules := []config.Module{
		{Name: "m1"},
		{Name: "m2"},
		{Name: "m3"},
	}
	defaultSelected := []string{"m1"}
	remoteHasBranch := map[string]bool{"m2": true}
	sel := NewModuleSelector(modules, nil, remoteHasBranch, defaultSelected, "")
	if !sel.selected[0] || !sel.selected[1] || sel.selected[2] {
		t.Error("m1（默认选中）和 m2（远程有分支）应被选中，m3 不应被选中")
	}
}

// TestModuleSelector_SelectedModules_None 未选中任何模块时返回空切片
func TestModuleSelector_SelectedModules_None(t *testing.T) {
	modules := []config.Module{{Name: "m1"}, {Name: "m2"}}
	sel := NewModuleSelector(modules, nil, nil, nil, "") // existing 为空，全部未选
	got := sel.SelectedModules()
	if len(got) != 0 {
		t.Errorf("SelectedModules() 应为空, got %v", got)
	}
}

// TestModuleSelector_SelectedModules_All 全部选中时返回全部模块
func TestModuleSelector_SelectedModules_All(t *testing.T) {
	modules := []config.Module{{Name: "a"}, {Name: "b"}}
	sel := NewModuleSelector(modules, []string{"a", "b"}, nil, nil, "")
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
	sel := NewModuleSelector(modules, nil, nil, nil, "")
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
	sel := NewModuleSelector([]config.Module{{Name: "m1"}}, nil, nil, nil, "")
	_, cmd := sel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("回车应返回 Quit 命令")
	}
}

// TestModuleSelector_Update_Space 空格切换选中状态
func TestModuleSelector_Update_Space(t *testing.T) {
	modules := []config.Module{{Name: "m1"}, {Name: "m2"}}
	sel := NewModuleSelector(modules, nil, nil, nil, "")
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
	sel := NewModuleSelector([]config.Module{{Name: "mod-a"}}, nil, nil, nil, "")
	view := sel.View()
	if view == "" || len(view) < 4 {
		t.Error("View() 应返回非空字符串")
	}
}

func TestWorktreeSelector_NormalDelete(t *testing.T) {
	selector := NewWorktreeSelector([]core.WorktreeEnv{
		{Name: "feature-a"},
		{Name: "feature-b"},
	})

	selector.Update(tea.KeyMsg{Type: tea.KeySpace})
	selector.Update(tea.KeyMsg{Type: tea.KeyDown})
	selector.Update(tea.KeyMsg{Type: tea.KeySpace})
	_, cmd := selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	if cmd == nil {
		t.Fatal("按 d 应确认普通删除并退出选择器")
	}
	if selector.Force() {
		t.Fatal("普通删除不应跳过脏检查")
	}
	selected := selector.SelectedFeatures()
	if len(selected) != 2 || selected[0] != "feature-a" || selected[1] != "feature-b" {
		t.Fatalf("SelectedFeatures() = %v, 期望 [feature-a feature-b]", selected)
	}
}

func TestWorktreeSelector_ForceDelete(t *testing.T) {
	selector := NewWorktreeSelector([]core.WorktreeEnv{{Name: "dirty-feature"}})
	selector.Update(tea.KeyMsg{Type: tea.KeySpace})
	_, cmd := selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})

	if cmd == nil {
		t.Fatal("按 f 应确认强制删除并退出选择器")
	}
	if !selector.Force() {
		t.Fatal("强制删除应跳过脏检查")
	}
}

func TestWorktreeSelector_Quit(t *testing.T) {
	selector := NewWorktreeSelector([]core.WorktreeEnv{{Name: "feature-a"}})
	_, cmd := selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	if cmd == nil {
		t.Fatal("按 q 应退出选择器")
	}
	if !selector.Quitting() {
		t.Fatal("退出选择器后 Quitting() 应为 true")
	}
}

func TestWorktreeSelector_ViewShowsWorktreesAndDeleteKeys(t *testing.T) {
	selector := NewWorktreeSelector([]core.WorktreeEnv{
		{
			Name: "clean-feature",
			Modules: []core.ModuleStatus{
				{Name: "module-a", IsDirty: false},
			},
		},
		{
			Name: "dirty-feature",
			Modules: []core.ModuleStatus{
				{Name: "module-b", IsDirty: true},
			},
		},
	})

	view := selector.View()
	for _, expected := range []string{"clean-feature", "dirty-feature", "clean", "1 dirty", "d 普通删除", "f 强制删除"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("View() 未包含 %q: %s", expected, view)
		}
	}
	if !strings.Contains(view, successStyle.Render("clean")) {
		t.Fatalf("clean 状态应使用绿色样式: %q", view)
	}
	if !strings.Contains(view, errorStyle.Render("1 dirty")) {
		t.Fatalf("dirty 状态应使用红色样式: %q", view)
	}
}

// TestWorktreeSelector_ViewUsesMainListDirtyScope 验证批量页与主列表一样只统计模块脏状态。
func TestWorktreeSelector_ViewUsesMainListDirtyScope(t *testing.T) {
	worktree := core.WorktreeEnv{
		Name:        "feature-main-dirty",
		MainProject: &core.ModuleStatus{Name: "modu", IsDirty: true},
		Modules:     []core.ModuleStatus{{Name: "module-a", IsDirty: false}},
	}
	selector := NewWorktreeSelector([]core.WorktreeEnv{worktree})

	view := selector.View()
	if !strings.Contains(view, successStyle.Render("clean")) {
		t.Fatalf("批量页应按主列表口径显示 clean: %q", view)
	}
	if strings.Contains(view, errorStyle.Render("1 dirty")) {
		t.Fatalf("批量页不应单独统计主项目 dirty: %q", view)
	}
}

func TestApp_ListKey_OpensBatchDeleteSelection(t *testing.T) {
	app := &App{
		state: "list",
		Envs: []core.WorktreeEnv{
			{Name: "feature-a"},
			{Name: "feature-b"},
		},
	}

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})

	if app.state != "batch_delete" {
		t.Fatalf("按 b 后 state = %q，期望 batch_delete", app.state)
	}
	view := app.View()
	if !strings.Contains(view, "feature-a") || !strings.Contains(view, "feature-b") {
		t.Fatalf("批量删除页应展示当前 worktree: %s", view)
	}
}

func TestApp_ListKey_UppercaseDDoesNotOpenBatchDelete(t *testing.T) {
	app := &App{
		state: "list",
		Envs:  []core.WorktreeEnv{{Name: "feature-a"}},
	}

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})

	if app.state != "list" {
		t.Fatalf("大写 D 不应触发批量删除，state=%q", app.state)
	}
}

func TestApp_BatchDelete_NormalModeRequiresConfirmation(t *testing.T) {
	app := &App{
		state: "list",
		Envs: []core.WorktreeEnv{
			{Name: "feature-a"},
			{Name: "feature-b"},
		},
	}

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	app.Update(tea.KeyMsg{Type: tea.KeySpace})
	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	app.Update(tea.KeyMsg{Type: tea.KeySpace})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	if app.state != "confirm_batch_delete" {
		t.Fatalf("按 d 后 state = %q，期望 confirm_batch_delete", app.state)
	}
	view := app.View()
	for _, expected := range []string{"确认批量删除", "feature-a", "feature-b", "按 y 确认"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("批量删除确认页未包含 %q: %s", expected, view)
		}
	}
}

func TestApp_BatchDelete_ForceConfirmationShowsColoredStatus(t *testing.T) {
	app := &App{
		state: "list",
		Envs: []core.WorktreeEnv{
			{Name: "clean-feature"},
			{
				Name: "dirty-feature",
				Modules: []core.ModuleStatus{
					{Name: "module1", IsDirty: true},
				},
			},
		},
	}

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	app.Update(tea.KeyMsg{Type: tea.KeySpace})
	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	app.Update(tea.KeyMsg{Type: tea.KeySpace})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})

	if app.state != "confirm_batch_delete" || !app.batchDeleteForce {
		t.Fatalf("按 f 后应进入批量强制删除确认页: state=%q force=%v", app.state, app.batchDeleteForce)
	}
	view := app.View()
	for _, expected := range []string{
		fmt.Sprintf("- clean-feature (%s)", successStyle.Render("clean")),
		fmt.Sprintf("- dirty-feature (%s)", errorStyle.Render("1 dirty")),
		"跳过脏检查和未推送分支保护",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("批量强制删除确认页未显示带颜色的状态 %q: %s", expected, view)
		}
	}
}

func TestApp_BatchDelete_CancelConfirmationReturnsToSelection(t *testing.T) {
	app := &App{
		state:       "confirm_batch_delete",
		batchDelete: NewWorktreeSelector([]core.WorktreeEnv{{Name: "feature-a"}}),
	}
	app.batchDelete.selected[0] = true

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})

	if app.state != "batch_delete" {
		t.Fatalf("取消确认后 state = %q，期望返回 batch_delete", app.state)
	}
	if got := app.batchDelete.SelectedFeatures(); len(got) != 1 || got[0] != "feature-a" {
		t.Fatalf("取消确认后应保留选择，got %v", got)
	}
}

func TestApp_BatchDelete_ConfirmExecutesNormalDelete(t *testing.T) {
	app, fake, worktreeRoot := newBatchDeleteTestApp(t)

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatal("确认批量删除后应返回执行命令")
	}
	if fake.removeCalls != 0 {
		t.Fatalf("执行命令前不应删除，removeCalls=%d", fake.removeCalls)
	}

	msg := cmd()
	done, ok := msg.(batchDeleteDoneMsg)
	if !ok {
		t.Fatalf("执行结果类型 = %T，期望 batchDeleteDoneMsg", msg)
	}
	_, refreshCmd := app.Update(done)

	if fake.removeCalls != 2 {
		t.Fatalf("只应删除 clean-feature 的模块和主项目，removeCalls=%d", fake.removeCalls)
	}
	if _, err := os.Stat(filepath.Join(worktreeRoot, "dirty-feature")); err != nil {
		t.Fatalf("普通删除应保留 dirty-feature: %v", err)
	}
	if _, err := os.Stat(filepath.Join(worktreeRoot, "clean-feature")); !os.IsNotExist(err) {
		t.Fatalf("clean-feature 应已删除")
	}
	if refreshCmd == nil || app.state != "loading" {
		t.Fatalf("删除完成后应刷新列表: state=%q", app.state)
	}
	expectedMessage := "批量删除完成，成功 1 个\n删除失败：\n- dirty-feature：存在未提交修改"
	if app.message != expectedMessage || !app.messageIsError {
		t.Fatalf("批量删除结果消息异常: message=%q isError=%v", app.message, app.messageIsError)
	}
}

func TestApp_BatchDelete_ConfirmExecutesForceDelete(t *testing.T) {
	app, fake, worktreeRoot := newBatchDeleteTestApp(t)
	app.batchDeleteForce = true

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatal("确认批量强制删除后应返回执行命令")
	}
	done, ok := cmd().(batchDeleteDoneMsg)
	if !ok {
		t.Fatal("批量强制删除应返回 batchDeleteDoneMsg")
	}
	app.Update(done)

	if fake.removeCalls != 4 {
		t.Fatalf("应删除两个 feature 的模块和主项目，removeCalls=%d", fake.removeCalls)
	}
	for _, feature := range []string{"dirty-feature", "clean-feature"} {
		if _, err := os.Stat(filepath.Join(worktreeRoot, feature)); !os.IsNotExist(err) {
			t.Fatalf("%s 应已删除", feature)
		}
	}
	if app.message != "批量删除完成，成功 2 个" || app.messageIsError {
		t.Fatalf("批量强制删除结果消息异常: message=%q isError=%v", app.message, app.messageIsError)
	}
}

func TestFormatBatchDeleteResult_ListsFailureReasonsInSelectionOrder(t *testing.T) {
	done := batchDeleteDoneMsg{
		features:  []string{"feature-z", "feature-a", "feature-ok"},
		succeeded: []string{"feature-ok"},
		failed: map[string]error{
			"feature-z": fmt.Errorf("dirty: %w", errs.ErrDirtyWorktree),
			"feature-a": errors.New("remove worktree failed"),
		},
	}

	expected := "批量删除完成，成功 1 个\n删除失败：\n" +
		"- feature-z：存在未提交修改\n" +
		"- feature-a：remove worktree failed"
	if got := formatBatchDeleteResult(done); got != expected {
		t.Fatalf("formatBatchDeleteResult() = %q，期望 %q", got, expected)
	}
}

// TestApp_BatchDelete_RendersRemovalWarnings 验证后台删除警告通过消息返回并逐行展示。
func TestApp_BatchDelete_RendersRemovalWarnings(t *testing.T) {
	app, fake, _ := newBatchDeleteTestApp(t)
	fake.removeResult = gitproxy.RemoveWorktreeResult{Warnings: []string{"branch mismatch"}}

	done, ok := app.executeBatchDelete([]string{"clean-feature"}, true)().(batchDeleteDoneMsg)
	if !ok {
		t.Fatal("批量删除应返回 batchDeleteDoneMsg")
	}
	got := formatBatchDeleteResult(done)
	want := "批量删除完成，成功 1 个\n警告：\n" +
		"- clean-feature：module module1: branch mismatch\n" +
		"- clean-feature：main project: branch mismatch"
	if got != want {
		t.Fatalf("批量删除警告展示异常: got %q want %q", got, want)
	}
}

// TestApp_BatchDelete_ReportsRemovalFailure 验证底层实际删除失败会进入失败详情而非成功计数。
func TestApp_BatchDelete_ReportsRemovalFailure(t *testing.T) {
	app, fake, _ := newBatchDeleteTestApp(t)
	fake.removeErr = errors.New("simulated removal failure")

	done, ok := app.executeBatchDelete([]string{"clean-feature"}, true)().(batchDeleteDoneMsg)
	if !ok {
		t.Fatal("批量删除应返回 batchDeleteDoneMsg")
	}
	if len(done.succeeded) != 0 || !errors.Is(done.failed["clean-feature"], errs.ErrPartialFailure) {
		t.Fatalf("实际删除失败不应统计成功: succeeded=%v failed=%v", done.succeeded, done.failed)
	}
	result := formatBatchDeleteResult(done)
	if !strings.Contains(result, "- clean-feature：") || !strings.Contains(result, "simulated removal failure") {
		t.Fatalf("结果应展示失败 worktree 和底层原因: %q", result)
	}
}

func newBatchDeleteTestApp(t *testing.T) (*App, *uiFakeGitClient, string) {
	t.Helper()

	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	moduleRepoPath := filepath.Join(workspace, "module1")
	for _, path := range []string{
		workspace,
		moduleRepoPath,
		filepath.Join(worktreeRoot, "dirty-feature", "module1"),
		filepath.Join(worktreeRoot, "clean-feature", "module1"),
	} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	fake := &uiFakeGitClient{
		statusFunc: func(path string) gitproxy.Status {
			switch {
			case strings.Contains(path, "dirty-feature"):
				return gitproxy.Status{Branch: "dirty-feature", IsDirty: true}
			case strings.Contains(path, "clean-feature"):
				return gitproxy.Status{Branch: "clean-feature", IsDirty: false}
			default:
				return gitproxy.Status{Branch: "main"}
			}
		},
		pushStatuses: map[string]gitproxy.BranchPushStatus{
			workspace + "|dirty-feature": {
				Branch: "dirty-feature", RemoteRef: "origin/dirty-feature", IsPushed: false,
			},
			moduleRepoPath + "|dirty-feature": {
				Branch: "dirty-feature", RemoteRef: "origin/dirty-feature", IsPushed: false,
			},
			workspace + "|clean-feature": {
				Branch: "clean-feature", RemoteRef: "origin/clean-feature", IsPushed: true,
			},
			moduleRepoPath + "|clean-feature": {
				Branch: "clean-feature", RemoteRef: "origin/clean-feature", IsPushed: true,
			},
		},
	}
	eng := engine.NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  true,
		Modules:      []config.Module{{Name: "module1"}},
	}, fake)
	selector := NewWorktreeSelector([]core.WorktreeEnv{
		{Name: "dirty-feature"},
		{Name: "clean-feature"},
	})
	selector.selected[0] = true
	selector.selected[1] = true

	return &App{
		Engine:      eng,
		state:       "confirm_batch_delete",
		batchDelete: selector,
	}, fake, worktreeRoot
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

// TestApp_View_UnpushedConfirm 未推送分支确认视图包含风险分支
func TestApp_View_UnpushedConfirm(t *testing.T) {
	app := &App{
		state:   "confirm_unpushed",
		feature: "feat-a",
		unpushedBranches: []engine.UnpushedBranch{
			{Name: "module1", Branch: "feat-a", Reason: "1 local commits not pushed"},
		},
	}
	view := app.renderUnpushedConfirm()
	if !strings.Contains(view, "确认删除未推送分支") || !strings.Contains(view, "module1") {
		t.Fatalf("renderUnpushedConfirm() missing risk details: %q", view)
	}
}

// TestApp_UnpushedConfirm_Cancel 取消未推送确认不执行删除
func TestApp_UnpushedConfirm_Cancel(t *testing.T) {
	app, fake := newDeleteRiskTestApp(t)
	app.state = "confirm_unpushed"
	app.unpushedBranches = []engine.UnpushedBranch{{Name: "module1", Branch: "feat-a"}}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if cmd != nil {
		t.Fatal("取消未推送确认不应返回命令")
	}
	if app.state != "list" {
		t.Fatalf("state = %q, want list", app.state)
	}
	if fake.removeCalls != 0 {
		t.Fatalf("取消后不应删除，removeCalls=%d", fake.removeCalls)
	}
}

// TestApp_DeleteFlow_UnpushedConfirmThenDelete 未推送分支需要二次确认后删除
func TestApp_DeleteFlow_UnpushedConfirmThenDelete(t *testing.T) {
	app, fake := newDeleteRiskTestApp(t)

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd != nil {
		t.Fatal("进入未推送确认不应立即返回删除命令")
	}
	if app.state != "confirm_unpushed" {
		t.Fatalf("state = %q, want confirm_unpushed", app.state)
	}
	if fake.removeCalls != 0 {
		t.Fatalf("二次确认前不应删除，removeCalls=%d", fake.removeCalls)
	}

	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatal("确认删除后应返回刷新列表命令")
	}
	if app.state != "loading" {
		t.Fatalf("state = %q, want loading", app.state)
	}
	if fake.removeCalls != 2 {
		t.Fatalf("期望删除模块和主项目，removeCalls=%d", fake.removeCalls)
	}
}

func TestApp_ForceDeleteConfirmationSkipsDirtyCheck(t *testing.T) {
	app, fake := newDeleteRiskTestApp(t)
	app.deleteForce = true
	app.Engine.Config.StrictDirty = true
	fake.dirty = true
	for key, status := range fake.pushStatuses {
		status.IsPushed = true
		status.AheadCount = 0
		fake.pushStatuses[key] = status
	}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	if cmd == nil {
		t.Fatal("强制删除确认后应删除并刷新列表")
	}
	if app.state != "loading" || fake.removeCalls != 2 {
		t.Fatalf("强制删除未完成: state=%q removeCalls=%d err=%v", app.state, fake.removeCalls, app.err)
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

// TestApp_getSelectedPath_MainProject 选中主项目时返回主项目路径
func TestApp_getSelectedPath_MainProject(t *testing.T) {
	app := &App{
		mainProject: &engine.MainProjectStatus{Name: "main", Path: "/path/to/main"},
		Envs:        []core.WorktreeEnv{},
		selected:    0,
	}
	path, err := app.getSelectedPath()
	if err != nil {
		t.Errorf("getSelectedPath() 不应返回错误, got %v", err)
	}
	if path != "/path/to/main" {
		t.Errorf("getSelectedPath() = %q, 期望 %q", path, "/path/to/main")
	}
}

// TestApp_getSelectedPath_Feature 选中 feature 时返回主项目路径
func TestApp_getSelectedPath_Feature(t *testing.T) {
	app := &App{
		mainProject: &engine.MainProjectStatus{Name: "main"},
		Envs: []core.WorktreeEnv{{
			Name:        "feat-a",
			MainProject: &core.ModuleStatus{Name: "main", Path: "/path/to/main"},
		}},
		selected: 1,
	}
	path, err := app.getSelectedPath()
	if err != nil {
		t.Errorf("getSelectedPath() 不应返回错误, got %v", err)
	}
	if path != "/path/to/main" {
		t.Errorf("getSelectedPath() = %q, 期望 %q", path, "/path/to/main")
	}
}

// TestApp_getSelectedPath_FeatureNoMainProject 选中无主项目的 feature 时返回错误
func TestApp_getSelectedPath_FeatureNoMainProject(t *testing.T) {
	app := &App{
		mainProject: nil,
		Envs: []core.WorktreeEnv{{
			Name:        "feat-a",
			MainProject: nil,
		}},
		selected: 0,
	}
	_, err := app.getSelectedPath()
	if err == nil {
		t.Error("getSelectedPath() 应返回错误")
	}
}

// TestApp_getSelectedPath_NoSelection 未选中任何项时返回错误
func TestApp_getSelectedPath_NoSelection(t *testing.T) {
	app := &App{
		mainProject: nil,
		Envs:        []core.WorktreeEnv{},
		selected:    0,
	}
	_, err := app.getSelectedPath()
	if err == nil {
		t.Error("getSelectedPath() 应返回错误")
	}
}

// TestApp_copyPathAndBack_ClipboardError 剪贴板写入失败时设置错误状态
func TestApp_copyPathAndBack_ClipboardError(t *testing.T) {
	app := &App{
		state:       "menu",
		mainProject: &engine.MainProjectStatus{Name: "main", Path: "/path/to/main"},
		selected:    0,
	}
	// 传入一个不存在的路径，clipboard.WriteAll 会失败
	// 但实际上 clipboard 可能在测试环境不可用，会返回错误
	app.copyPathAndBack()
	// 如果剪贴板不可用，应该进入 error 状态
	if app.state != "error" && app.state != "list" {
		t.Errorf("copyPathAndBack() 后 state 应为 error 或 list, got %q", app.state)
	}
}

func TestApp_renderMenu_ConfiguredAppOpeners(t *testing.T) {
	withAppOpenerHooks(t, map[string]bool{
		"Zed":    true,
		"Cursor": true,
	})

	app := New(&config.Config{
		Workspace:    "/workspace/main",
		WorktreeRoot: "/workspace/worktrees",
		DefaultBase:  "develop",
		AppOpeners: []config.AppOpener{
			{Name: "zed", App: "Zed", Label: "Zed", Shortcut: "z"},
			{Name: "ghost", App: "Ghost", Label: "Ghost", Shortcut: "g"},
			{Name: "cursor", App: "Cursor", Label: "Cursor", Shortcut: "c"},
		},
	})
	app.state = "menu"
	app.mainProject = &engine.MainProjectStatus{Name: "main", Path: "/workspace/main"}
	app.selected = 0
	app.isMainProjectMenu = true

	view := app.renderMenu()
	if !strings.Contains(view, "打开 Zed (z)") {
		t.Fatalf("menu should render installed Zed opener with shortcut, got:\n%s", view)
	}
	if strings.Contains(view, "打开 Ghost") {
		t.Fatalf("menu should hide missing app opener, got:\n%s", view)
	}
	if !strings.Contains(view, "打开 Cursor") {
		t.Fatalf("menu should keep conflict opener selectable, got:\n%s", view)
	}
	if strings.Contains(view, "打开 Cursor (c)") {
		t.Fatalf("menu should hide conflicting shortcut, got:\n%s", view)
	}

	codexIndex := strings.Index(view, "打开 Codex")
	zedIndex := strings.Index(view, "打开 Zed")
	copyIndex := strings.Index(view, "复制路径")
	if !(codexIndex >= 0 && zedIndex > codexIndex && copyIndex > zedIndex) {
		t.Fatalf("custom openers should render after built-in openers and before copy, got:\n%s", view)
	}
}

func TestApp_ForceDeleteMenuStartsForceConfirmation(t *testing.T) {
	app := &App{
		state:    "menu",
		Envs:     []core.WorktreeEnv{{Name: "dirty-feature"}},
		selected: 0,
	}

	var forceDeleteItem *operationMenuItem
	items := app.operationMenuItems()
	for i := range items {
		if items[i].label == "强制删除" {
			forceDeleteItem = &items[i]
			break
		}
	}
	if forceDeleteItem == nil {
		t.Fatal("feature 管理菜单应包含强制删除")
	}
	forceDeleteItem.run()

	if app.state != "confirm" || app.feature != "dirty-feature" || !app.deleteForce {
		t.Fatalf("强制删除确认状态异常: state=%q feature=%q force=%v", app.state, app.feature, app.deleteForce)
	}
	if view := app.renderConfirm(); !strings.Contains(view, "跳过脏检查") {
		t.Fatalf("强制删除确认应提示跳过脏检查: %s", view)
	}
}

func TestApp_ForceDeleteMenuShortcut(t *testing.T) {
	app := &App{
		state:    "menu",
		Envs:     []core.WorktreeEnv{{Name: "dirty-feature"}},
		selected: 0,
	}

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})

	if app.state != "confirm" || !app.deleteForce {
		t.Fatalf("按 f 应进入强制删除确认: state=%q force=%v", app.state, app.deleteForce)
	}
}

func TestApp_handleMenuKey_ConfiguredAppOpenerShortcutAndEnter(t *testing.T) {
	calls := withAppOpenerHooks(t, map[string]bool{"Zed": true})
	app := New(&config.Config{
		Workspace:    "/workspace/main",
		WorktreeRoot: "/workspace/worktrees",
		DefaultBase:  "develop",
		AppOpeners: []config.AppOpener{
			{Name: "zed", App: "Zed", Label: "Zed", Shortcut: "z"},
		},
	})
	app.state = "menu"
	app.mainProject = &engine.MainProjectStatus{Name: "main", Path: "/workspace/main"}
	app.selected = 0
	app.isMainProjectMenu = true

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	if cmd != nil {
		t.Fatal("configured app opener shortcut should not return a tea command")
	}
	assertStartedCommand(t, *calls, 0, "open", "-a", "Zed", "/workspace/main")
	if app.state != "list" {
		t.Fatalf("after opening configured app state = %q, want list", app.state)
	}

	app.state = "menu"
	app.menuSelected = 2
	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("configured app opener enter should not return a tea command")
	}
	assertStartedCommand(t, *calls, 1, "open", "-a", "Zed", "/workspace/main")
	if app.state != "list" {
		t.Fatalf("after entering configured app state = %q, want list", app.state)
	}
}

func TestApp_handleMenuKey_ConfiguredShortcutConflictKeepsBuiltIn(t *testing.T) {
	calls := withAppOpenerHooks(t, map[string]bool{"Cursor": true})
	app := New(&config.Config{
		Workspace:    "/workspace/main",
		WorktreeRoot: "/workspace/worktrees",
		DefaultBase:  "develop",
		AppOpeners: []config.AppOpener{
			{Name: "cursor", App: "Cursor", Label: "Cursor", Shortcut: "c"},
		},
	})
	app.state = "menu"
	app.mainProject = &engine.MainProjectStatus{Name: "main", Path: "/workspace/main"}
	app.selected = 0
	app.isMainProjectMenu = true

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if len(*calls) != 0 {
		t.Fatalf("conflicting c shortcut should trigger built-in copy, not Cursor open, got %+v", *calls)
	}
}

func TestApp_handleMenuKey_ConfiguredAppOpenerMissingPath(t *testing.T) {
	calls := withAppOpenerHooks(t, map[string]bool{"Zed": true})
	app := New(&config.Config{
		Workspace:    "/workspace/main",
		WorktreeRoot: "/workspace/worktrees",
		DefaultBase:  "develop",
		AppOpeners: []config.AppOpener{
			{Name: "zed", App: "Zed", Label: "Zed", Shortcut: "z"},
		},
	})
	app.state = "menu"
	app.Envs = []core.WorktreeEnv{{Name: "feat-a"}}
	app.selected = 0
	app.isMainProjectMenu = false

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	if app.state != "error" {
		t.Fatalf("missing main project path should move to error, got %q", app.state)
	}
	if app.err == nil || !strings.Contains(app.err.Error(), "无法打开") {
		t.Fatalf("missing main project path should show open error, got %v", app.err)
	}
	if len(*calls) != 0 {
		t.Fatalf("missing path should not start app, got %+v", *calls)
	}
}

// TestApp_ListKey_StartCreateAndQuit 保持列表 q 退出，同时新增 n 创建入口
func TestApp_ListKey_StartCreateAndQuit(t *testing.T) {
	app := &App{state: "list"}
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if cmd != nil {
		t.Error("按 n 进入创建输入不应返回命令")
	}
	if app.state != "create_input" {
		t.Fatalf("按 n 后 state = %q，期望 create_input", app.state)
	}

	app = &App{state: "list"}
	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("列表视图按 q 应保持退出行为")
	}
}

type uiStartedCommand struct {
	name string
	args []string
}

func withAppOpenerHooks(t *testing.T, installed map[string]bool) *[]uiStartedCommand {
	t.Helper()

	oldIsAppInstalled := isAppInstalled
	oldStartCommand := startCommand
	calls := []uiStartedCommand{}

	isAppInstalled = func(app string) bool {
		return installed[app]
	}
	startCommand = func(name string, args ...string) error {
		calls = append(calls, uiStartedCommand{name: name, args: append([]string(nil), args...)})
		return nil
	}

	t.Cleanup(func() {
		isAppInstalled = oldIsAppInstalled
		startCommand = oldStartCommand
	})
	return &calls
}

func assertStartedCommand(t *testing.T, calls []uiStartedCommand, index int, name string, args ...string) {
	t.Helper()
	if len(calls) <= index {
		t.Fatalf("expected command call %d, got %d calls: %+v", index, len(calls), calls)
	}
	call := calls[index]
	if call.name != name {
		t.Fatalf("command name = %q, want %q", call.name, name)
	}
	if len(call.args) != len(args) {
		t.Fatalf("command args = %v, want %v", call.args, args)
	}
	for i := range args {
		if call.args[i] != args[i] {
			t.Fatalf("command args = %v, want %v", call.args, args)
		}
	}
}

// TestApp_renderList_IncludesCreateAndBatchDeleteShortcuts 列表帮助展示创建和批量删除快捷键。
func TestApp_renderList_IncludesCreateAndBatchDeleteShortcuts(t *testing.T) {
	app := &App{state: "list"}
	view := app.renderList()
	if !strings.Contains(view, "n 新建 feature") {
		t.Fatalf("renderList() 应包含创建快捷键，got:\n%s", view)
	}
	if !strings.Contains(view, "d 删除  b 批量删除") {
		t.Fatalf("批量删除快捷键应紧跟单个删除快捷键，got:\n%s", view)
	}
}

// TestApp_CreateInput_QIsText 输入状态下 q 是普通文本
func TestApp_CreateInput_QIsText(t *testing.T) {
	app := &App{state: "create_input"}
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd != nil {
		t.Error("输入 q 不应退出或返回命令")
	}
	if app.state != "create_input" {
		t.Fatalf("输入 q 后 state = %q，期望 create_input", app.state)
	}
	if got := string(app.createFeatureInput); got != "q" {
		t.Fatalf("输入内容 = %q，期望 q", got)
	}
}

// TestApp_CreateInput_CursorMovement 光标可左右和首尾移动
func TestApp_CreateInput_CursorMovement(t *testing.T) {
	app := &App{state: "create_input", createFeatureInput: []rune("abcd"), createFeatureCursor: 2}

	app.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if app.createFeatureCursor != 1 {
		t.Fatalf("left 后 cursor = %d，期望 1", app.createFeatureCursor)
	}
	app.Update(tea.KeyMsg{Type: tea.KeyRight})
	if app.createFeatureCursor != 2 {
		t.Fatalf("right 后 cursor = %d，期望 2", app.createFeatureCursor)
	}
	app.Update(tea.KeyMsg{Type: tea.KeyHome})
	if app.createFeatureCursor != 0 {
		t.Fatalf("home 后 cursor = %d，期望 0", app.createFeatureCursor)
	}
	app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if app.createFeatureCursor != 4 {
		t.Fatalf("end 后 cursor = %d，期望 4", app.createFeatureCursor)
	}
}

// TestApp_CreateInput_InsertAndDeleteAtCursor 支持光标位置插入和删除
func TestApp_CreateInput_InsertAndDeleteAtCursor(t *testing.T) {
	app := &App{state: "create_input", createFeatureInput: []rune("abcd"), createFeatureCursor: 2}

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if got := string(app.createFeatureInput); got != "abqcd" {
		t.Fatalf("插入后 input = %q，期望 abqcd", got)
	}
	if app.createFeatureCursor != 3 {
		t.Fatalf("插入后 cursor = %d，期望 3", app.createFeatureCursor)
	}

	app.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if got := string(app.createFeatureInput); got != "abcd" {
		t.Fatalf("backspace 后 input = %q，期望 abcd", got)
	}
	if app.createFeatureCursor != 2 {
		t.Fatalf("backspace 后 cursor = %d，期望 2", app.createFeatureCursor)
	}

	app.Update(tea.KeyMsg{Type: tea.KeyDelete})
	if got := string(app.createFeatureInput); got != "abd" {
		t.Fatalf("delete 后 input = %q，期望 abd", got)
	}
	if app.createFeatureCursor != 2 {
		t.Fatalf("delete 后 cursor = %d，期望 2", app.createFeatureCursor)
	}
}

// TestApp_CreateInput_EmptyNameValidation 空名称停留在输入状态并提示
func TestApp_CreateInput_EmptyNameValidation(t *testing.T) {
	app := &App{state: "create_input", createFeatureInput: []rune("   "), createFeatureCursor: 3}
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("空名称不应返回命令")
	}
	if app.state != "create_input" {
		t.Fatalf("空名称后 state = %q，期望 create_input", app.state)
	}
	if app.createFeatureError == "" {
		t.Fatal("空名称应设置中文校验提示")
	}
}

// TestApp_CreateInput_Cancel Esc/Ctrl+C 取消创建输入
func TestApp_CreateInput_Cancel(t *testing.T) {
	app := &App{state: "create_input", createFeatureInput: []rune("feat"), createFeatureCursor: 4}
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Error("Esc 取消不应返回命令")
	}
	if app.state != "list" {
		t.Fatalf("Esc 后 state = %q，期望 list", app.state)
	}

	app = &App{state: "create_input", createFeatureInput: []rune("feat"), createFeatureCursor: 4}
	app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if app.state != "list" {
		t.Fatalf("Ctrl+C 后 state = %q，期望 list", app.state)
	}
}

// TestApp_CreateInput_ValidNameInitializesSelection 有效名称进入项目选择并应用默认/远程预选
func TestApp_CreateInput_ValidNameInitializesSelection(t *testing.T) {
	fake := &uiFakeGitClient{remoteBranches: map[string]bool{"git@example.com:m2.git|feat-a": true}}
	app := &App{
		Engine: engine.NewWithClient(&config.Config{
			Workspace:              t.TempDir(),
			WorktreeRoot:           t.TempDir(),
			DefaultBase:            "develop",
			DefaultSelectedModules: []string{"m1"},
			Modules: []config.Module{
				{Name: "m1", URL: "git@example.com:m1.git"},
				{Name: "m2", URL: "git@example.com:m2.git"},
				{Name: "m3", URL: "git@example.com:m3.git"},
			},
			Concurrency: 2,
		}, fake),
		state:               "create_input",
		createFeatureInput:  []rune("feat-a"),
		createFeatureCursor: len("feat-a"),
	}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("初始化选择不应返回命令")
	}
	if app.state != "create_modules" {
		t.Fatalf("有效名称后 state = %q，期望 create_modules", app.state)
	}
	if app.createFeature != "feat-a" {
		t.Fatalf("createFeature = %q，期望 feat-a", app.createFeature)
	}
	if app.moduleSelector == nil {
		t.Fatal("moduleSelector 不应为空")
	}
	if !app.moduleSelector.selected[0] || !app.moduleSelector.selected[1] || app.moduleSelector.selected[2] {
		t.Fatalf("期望 m1 默认选中、m2 远程选中、m3 未选中，got %v", app.moduleSelector.selected)
	}
}

// TestApp_CreateModules_Cancel 项目选择取消不调用创建
func TestApp_CreateModules_Cancel(t *testing.T) {
	fake := &uiFakeGitClient{}
	app := &App{
		Engine:         engine.NewWithClient(&config.Config{Workspace: t.TempDir(), WorktreeRoot: t.TempDir()}, fake),
		state:          "create_modules",
		createFeature:  "feat-a",
		moduleSelector: NewModuleSelector([]config.Module{{Name: "m1"}}, nil, nil, nil, ""),
	}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd != nil {
		t.Error("取消选择不应返回命令")
	}
	if app.state != "list" {
		t.Fatalf("取消后 state = %q，期望 list", app.state)
	}
	if len(fake.createWorktreeCalls) != 0 {
		t.Fatalf("取消不应创建 worktree，got %v", fake.createWorktreeCalls)
	}
}

// TestApp_CreateModules_ConfirmCreatesSelectedModules 确认后只创建选中的模块
func TestApp_CreateModules_ConfirmCreatesSelectedModules(t *testing.T) {
	app, fake := newCreateFeatureTestApp(t)
	app.moduleSelector.selected = []bool{true, false}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("确认选择应返回创建命令")
	}
	_ = cmd()

	if len(fake.createWorktreeCalls) != 2 {
		t.Fatalf("期望创建主项目和 m1 两个 worktree，got %d: %v", len(fake.createWorktreeCalls), fake.createWorktreeCalls)
	}
	if fake.createWorktreeCalls[0].branch != "feat-a" || fake.createWorktreeCalls[0].baseBranch != "develop" {
		t.Fatalf("主项目创建参数不正确: %+v", fake.createWorktreeCalls[0])
	}
	if fake.createWorktreeCalls[1].repoPath != filepath.Join(app.Engine.Config.Workspace, "m1") {
		t.Fatalf("期望只创建 m1，got %+v", fake.createWorktreeCalls[1])
	}
}

// TestApp_CreateModules_ConfirmCreatesMainOnly 零模块选择时只创建主项目
func TestApp_CreateModules_ConfirmCreatesMainOnly(t *testing.T) {
	app, fake := newCreateFeatureTestApp(t)
	app.moduleSelector.selected = []bool{false, false}

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("确认零模块选择应返回创建命令")
	}
	_ = cmd()

	if len(fake.createWorktreeCalls) != 1 {
		t.Fatalf("期望仅创建主项目 worktree，got %d: %v", len(fake.createWorktreeCalls), fake.createWorktreeCalls)
	}
}

// TestApp_CreateFeatureDone_SuccessReloadsList 创建成功后展示消息并触发列表刷新
func TestApp_CreateFeatureDone_SuccessReloadsList(t *testing.T) {
	app := &App{
		Engine:              engine.NewWithClient(&config.Config{Workspace: t.TempDir(), WorktreeRoot: t.TempDir()}, &uiFakeGitClient{}),
		state:               "loading",
		createFeature:       "feat-a",
		createFeatureInput:  []rune("feat-a"),
		createFeatureCursor: len("feat-a"),
	}

	_, cmd := app.Update(createFeatureDoneMsg{feature: "feat-a"})
	if app.state != "loading" {
		t.Fatalf("成功后 state = %q，期望 loading 以刷新列表", app.state)
	}
	if app.message != "已创建 feature: feat-a" {
		t.Fatalf("成功消息 = %q", app.message)
	}
	if app.createFeature != "" || len(app.createFeatureInput) != 0 {
		t.Fatalf("成功后应清理创建状态, feature=%q input=%q", app.createFeature, string(app.createFeatureInput))
	}
	if cmd == nil {
		t.Fatal("成功后应返回刷新列表命令")
	}
}

// TestApp_CreateFeatureDone_ErrorShowsError 创建失败后进入错误状态
func TestApp_CreateFeatureDone_ErrorShowsError(t *testing.T) {
	app := &App{state: "loading"}

	_, cmd := app.Update(createFeatureDoneMsg{feature: "feat-a", err: os.ErrInvalid})
	if cmd != nil {
		t.Error("失败后不应返回刷新命令")
	}
	if app.state != "error" {
		t.Fatalf("失败后 state = %q，期望 error", app.state)
	}
	if app.err == nil {
		t.Fatal("失败后应保留错误")
	}
}

func newCreateFeatureTestApp(t *testing.T) (*App, *uiFakeGitClient) {
	t.Helper()

	workspace := t.TempDir()
	worktreeRoot := filepath.Join(t.TempDir(), "worktrees")
	for _, name := range []string{"m1", "m2"} {
		if err := os.MkdirAll(filepath.Join(workspace, name), 0755); err != nil {
			t.Fatalf("创建测试模块目录失败: %v", err)
		}
	}

	fake := &uiFakeGitClient{}
	cfg := &config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		DefaultBase:  "develop",
		Modules: []config.Module{
			{Name: "m1", URL: "git@example.com:m1.git"},
			{Name: "m2", URL: "git@example.com:m2.git"},
		},
		Concurrency: 2,
	}
	app := &App{
		Engine:         engine.NewWithClient(cfg, fake),
		state:          "create_modules",
		createFeature:  "feat-a",
		moduleSelector: NewModuleSelector(cfg.Modules, nil, nil, nil, ""),
	}
	return app, fake
}

func newDeleteRiskTestApp(t *testing.T) (*App, *uiFakeGitClient) {
	t.Helper()

	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "feat-a")
	moduleRepoPath := filepath.Join(workspace, "module1")
	moduleWorktreePath := filepath.Join(featurePath, "module1")
	for _, path := range []string{workspace, moduleRepoPath, featurePath, moduleWorktreePath} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	fake := &uiFakeGitClient{
		pushStatuses: map[string]gitproxy.BranchPushStatus{
			workspace + "|feat-a": {
				Branch:     "feat-a",
				RemoteRef:  "origin/feat-a",
				IsPushed:   false,
				AheadCount: 1,
				Reason:     "1 local commits not pushed",
			},
			moduleRepoPath + "|feat-a": {
				Branch:     "feat-a",
				RemoteRef:  "origin/feat-a",
				IsPushed:   false,
				AheadCount: 1,
				Reason:     "1 local commits not pushed",
			},
		},
	}
	cfg := &config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		Modules: []config.Module{
			{Name: "module1", URL: "git@example.com:module1.git"},
		},
	}
	app := &App{
		Engine:  engine.NewWithClient(cfg, fake),
		state:   "confirm",
		feature: "feat-a",
	}
	return app, fake
}

type uiFakeCreateWorktreeCall struct {
	repoPath     string
	branch       string
	baseBranch   string
	worktreePath string
}

type uiFakeGitClient struct {
	remoteBranches      map[string]bool
	pushStatuses        map[string]gitproxy.BranchPushStatus
	createWorktreeCalls []uiFakeCreateWorktreeCall
	removeCalls         int
	removeResult        gitproxy.RemoveWorktreeResult
	removeErr           error
	dirty               bool
	statusFunc          func(path string) gitproxy.Status
}

func (f *uiFakeGitClient) Clone(ctx context.Context, url, path string) error {
	return nil
}

func (f *uiFakeGitClient) CreateWorktree(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error {
	f.createWorktreeCalls = append(f.createWorktreeCalls, uiFakeCreateWorktreeCall{
		repoPath:     repoPath,
		branch:       branch,
		baseBranch:   baseBranch,
		worktreePath: worktreePath,
	})
	return os.MkdirAll(worktreePath, 0755)
}

func (f *uiFakeGitClient) GetStatus(ctx context.Context, path string) (gitproxy.Status, error) {
	if f.statusFunc != nil {
		return f.statusFunc(path), nil
	}
	return gitproxy.Status{Branch: "feat-a", IsDirty: f.dirty}, nil
}

func (f *uiFakeGitClient) RemoveWorktree(ctx context.Context, path string) error {
	return nil
}

func (f *uiFakeGitClient) RemoveWorktreeAndBranch(ctx context.Context, repoPath, worktreePath, featureDirName string) (gitproxy.RemoveWorktreeResult, error) {
	f.removeCalls++
	return f.removeResult, f.removeErr
}

func (f *uiFakeGitClient) ListWorktrees(ctx context.Context, repoPath string) ([]gitproxy.WorktreeInfo, error) {
	return nil, nil
}

func (f *uiFakeGitClient) Fetch(ctx context.Context, repoPath string) error {
	return nil
}

func (f *uiFakeGitClient) Rebase(ctx context.Context, path string) error {
	return nil
}

func (f *uiFakeGitClient) FetchAndSwitchBranch(ctx context.Context, repoPath, branch string) error {
	return nil
}

func (f *uiFakeGitClient) BranchExists(ctx context.Context, repoPath, branch string) bool {
	return false
}

func (f *uiFakeGitClient) CheckBranchWorktreeStatus(ctx context.Context, repoPath, branch string) (bool, error) {
	return false, nil
}

func (f *uiFakeGitClient) CreateWorktreeFromExistingBranch(ctx context.Context, repoPath, branch, worktreePath string) error {
	return os.MkdirAll(worktreePath, 0755)
}

func (f *uiFakeGitClient) RemoteBranchExists(ctx context.Context, repoURL, branch string) bool {
	if f.remoteBranches == nil {
		return false
	}
	return f.remoteBranches[repoURL+"|"+branch]
}

func (f *uiFakeGitClient) CreateWorktreeFromRemoteBranch(ctx context.Context, repoPath, branch, worktreePath string) error {
	return os.MkdirAll(worktreePath, 0755)
}

func (f *uiFakeGitClient) GetBranchPushStatus(ctx context.Context, repoPath, branch string) (gitproxy.BranchPushStatus, error) {
	if f.pushStatuses != nil {
		if status, ok := f.pushStatuses[repoPath+"|"+branch]; ok {
			return status, nil
		}
	}
	return gitproxy.BranchPushStatus{Branch: branch, RemoteRef: "origin/" + branch, IsPushed: true}, nil
}
