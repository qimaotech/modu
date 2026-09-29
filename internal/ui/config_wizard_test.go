package ui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/qimaotech/modu/internal/config"
)

func TestNewConfigWizard(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	wizard := NewConfigWizard()

	if wizard.step != 0 {
		t.Errorf("expected step 0, got %d", wizard.step)
	}
	if wizard.workspace != "" {
		t.Errorf("expected empty workspace, got %s", wizard.workspace)
	}
	if wizard.worktree != "" {
		t.Errorf("expected empty worktree, got %s", wizard.worktree)
	}
	if wizard.base != "" {
		t.Errorf("expected empty base, got %s", wizard.base)
	}
	if len(wizard.modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(wizard.modules))
	}
	if wizard.inputField != 0 {
		t.Errorf("expected inputField 0, got %d", wizard.inputField)
	}
	if wizard.workspaceDefault != tmpDir {
		t.Errorf("expected workspaceDefault %s, got %s", tmpDir, wizard.workspaceDefault)
	}
	if wizard.quitting {
		t.Error("expected quitting to be false")
	}
	if wizard.saved {
		t.Error("expected saved to be false")
	}
	if wizard.err != nil {
		t.Errorf("expected err to be nil, got %v", wizard.err)
	}
}

func TestConfigWizard_Init(t *testing.T) {
	wizard := NewConfigWizard()
	cmd := wizard.Init()
	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
}

func TestConfigWizard_Update_Quit(t *testing.T) {
	t.Run("ctrl+c退出", func(t *testing.T) {
		w := NewConfigWizard()
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		_, cmd := w.Update(msg)
		if cmd == nil {
			t.Error("expected non-nil cmd for quit")
		}
		if !w.quitting {
			t.Error("expected quitting to be true")
		}
	})

	t.Run("esc退出", func(t *testing.T) {
		w := NewConfigWizard()
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		_, cmd := w.Update(msg)
		if cmd == nil {
			t.Error("expected non-nil cmd for quit")
		}
		if !w.quitting {
			t.Error("expected quitting to be true")
		}
	})
}

func TestConfigWizard_Update_Tab(t *testing.T) {
	wizard := NewConfigWizard()
	msg := tea.KeyMsg{Type: tea.KeyTab}
	_, cmd := wizard.Update(msg)
	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
	if wizard.step != 0 {
		t.Errorf("expected step 0, got %d", wizard.step)
	}
}

func TestConfigWizard_Update_CharacterInput(t *testing.T) {
	t.Run("步骤0输入到workspace", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
		_, cmd := w.Update(msg)
		if cmd != nil {
			t.Errorf("expected nil cmd, got %v", cmd)
		}
		if w.workspace != "a" {
			t.Errorf("expected workspace 'a', got %s", w.workspace)
		}
	})

	t.Run("步骤1输入到worktree", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 1

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
		_, cmd := w.Update(msg)
		if cmd != nil {
			t.Errorf("expected nil cmd, got %v", cmd)
		}
		if w.worktree != "b" {
			t.Errorf("expected worktree 'b', got %s", w.worktree)
		}
	})

	t.Run("步骤2输入到base", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 2

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
		_, cmd := w.Update(msg)
		if cmd != nil {
			t.Errorf("expected nil cmd, got %v", cmd)
		}
		if w.base != "c" {
			t.Errorf("expected base 'c', got %s", w.base)
		}
	})

	t.Run("步骤0粘贴整段路径", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/data/workspace"), Paste: true}
		_, _ = w.Update(msg)

		if w.workspace != "/data/workspace" {
			t.Errorf("expected workspace '/data/workspace', got %s", w.workspace)
		}
	})
}

func TestConfigWizard_Update_Enter(t *testing.T) {
	t.Run("步骤0回车进入步骤1", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 0
		w.workspace = "/path/to/workspace"

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, _ = w.Update(msg)
		if w.step != 1 {
			t.Errorf("expected step 1, got %d", w.step)
		}
	})

	t.Run("步骤0空workspace默认.", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)
		w := NewConfigWizard()
		w.step = 0
		w.workspace = ""

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, _ = w.Update(msg)
		if w.workspace != tmpDir {
			t.Errorf("expected workspace %s, got %s", tmpDir, w.workspace)
		}
		if w.step != 1 {
			t.Errorf("expected step 1, got %d", w.step)
		}
	})

	t.Run("步骤1回车进入步骤2", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 1
		w.worktree = "/path/to/worktrees"

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, _ = w.Update(msg)
		if w.step != 2 {
			t.Errorf("expected step 2, got %d", w.step)
		}
	})

	t.Run("步骤1空worktree默认../worktrees", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 1
		w.worktree = ""

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, _ = w.Update(msg)
		if w.worktree != "../worktrees" {
			t.Errorf("expected worktree '../worktrees', got %s", w.worktree)
		}
		if w.step != 2 {
			t.Errorf("expected step 2, got %d", w.step)
		}
	})

	t.Run("步骤2回车进入步骤3", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 2
		w.base = "main"

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, _ = w.Update(msg)
		if w.step != 3 {
			t.Errorf("expected step 3, got %d", w.step)
		}
	})

	t.Run("步骤2空base默认develop", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 2
		w.base = ""

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, _ = w.Update(msg)
		if w.base != "develop" {
			t.Errorf("expected base 'develop', got %s", w.base)
		}
		if w.step != 3 {
			t.Errorf("expected step 3, got %d", w.step)
		}
	})

	t.Run("步骤3回车执行保存", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 3

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := w.Update(msg)
		if cmd == nil {
			t.Error("expected non-nil cmd for step 3")
		}
	})
}

func TestConfigWizard_Update_Backspace(t *testing.T) {
	t.Run("步骤0退格", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 0
		w.workspace = "test"

		msg := tea.KeyMsg{Type: tea.KeyBackspace}
		_, _ = w.Update(msg)
		if w.workspace != "tes" {
			t.Errorf("expected workspace 'tes', got %s", w.workspace)
		}
	})

	t.Run("步骤0空字符串不退格", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 0
		w.workspace = ""

		msg := tea.KeyMsg{Type: tea.KeyBackspace}
		_, _ = w.Update(msg)
		if w.workspace != "" {
			t.Errorf("expected workspace '', got %s", w.workspace)
		}
	})

	t.Run("步骤1退格", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 1
		w.worktree = "test"

		msg := tea.KeyMsg{Type: tea.KeyBackspace}
		_, _ = w.Update(msg)
		if w.worktree != "tes" {
			t.Errorf("expected worktree 'tes', got %s", w.worktree)
		}
	})

	t.Run("步骤2退格", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 2
		w.base = "test"

		msg := tea.KeyMsg{Type: tea.KeyBackspace}
		_, _ = w.Update(msg)
		if w.base != "tes" {
			t.Errorf("expected base 'tes', got %s", w.base)
		}
	})

	t.Run("步骤0退格支持多字节字符", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 0
		w.workspace = "/tmp/工作区"

		msg := tea.KeyMsg{Type: tea.KeyBackspace}
		_, _ = w.Update(msg)
		if w.workspace != "/tmp/工作" {
			t.Errorf("expected workspace '/tmp/工作', got %s", w.workspace)
		}
	})
}

func TestConfigWizard_Update_ConfigSavedMsg(t *testing.T) {
	t.Run("保存成功", func(t *testing.T) {
		w := NewConfigWizard()
		msg := configSavedMsg{err: nil}
		_, _ = w.Update(msg)
		if !w.saved {
			t.Error("expected saved to be true")
		}
	})

	t.Run("保存失败", func(t *testing.T) {
		w := NewConfigWizard()
		saveErr := &testError{msg: "save failed"}
		msg := configSavedMsg{err: saveErr}
		_, _ = w.Update(msg)
		if w.err != saveErr {
			t.Errorf("expected err %v, got %v", saveErr, w.err)
		}
	})
}

func TestConfigWizard_View(t *testing.T) {
	t.Run("步骤0视图", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 0
		w.workspace = "/test"

		view := w.View()
		if !strings.Contains(view, "步骤 1/3") {
			t.Error("expected view to contain '步骤 1/3'")
		}
		if !strings.Contains(view, "workspace") {
			t.Error("expected view to contain 'workspace'")
		}
		if !strings.Contains(view, "/test") {
			t.Error("expected view to contain '/test'")
		}
	})

	t.Run("步骤0空输入显示提示", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)
		w := NewConfigWizard()
		w.step = 0

		view := w.View()
		if !strings.Contains(view, tmpDir) {
			t.Error("expected view to contain workspace hint")
		}
		if w.workspace != "" {
			t.Errorf("expected workspace to stay empty, got %s", w.workspace)
		}
	})

	t.Run("步骤0空输入显示清理后的当前目录提示", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "nested")
		t.Chdir(tmpDir)

		w := NewConfigWizard()
		w.workspaceDefault = nestedDir

		view := w.View()
		if !strings.Contains(view, nestedDir) {
			t.Errorf("expected view to contain workspace default %s", nestedDir)
		}
	})

	t.Run("步骤1视图", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 1
		w.worktree = "/worktrees"

		view := w.View()
		if !strings.Contains(view, "步骤 2/3") {
			t.Error("expected view to contain '步骤 2/3'")
		}
		if !strings.Contains(view, "Worktree") {
			t.Error("expected view to contain 'Worktree'")
		}
		if !strings.Contains(view, "/worktrees") {
			t.Error("expected view to contain '/worktrees'")
		}
	})

	t.Run("步骤1空输入显示提示", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 1

		view := w.View()
		if !strings.Contains(view, "../worktrees") {
			t.Error("expected view to contain worktree default '../worktrees'")
		}
		if w.worktree != "" {
			t.Errorf("expected worktree to stay empty, got %s", w.worktree)
		}
	})

	t.Run("步骤2视图", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 2
		w.base = "main"

		view := w.View()
		if !strings.Contains(view, "步骤 3/3") {
			t.Error("expected view to contain '步骤 3/3'")
		}
		if !strings.Contains(view, "分支") {
			t.Error("expected view to contain '分支'")
		}
		if !strings.Contains(view, "main") {
			t.Error("expected view to contain 'main'")
		}
	})

	t.Run("步骤2空输入显示提示", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 2

		view := w.View()
		if !strings.Contains(view, "develop") {
			t.Error("expected view to contain base hint")
		}
		if w.base != "" {
			t.Errorf("expected base to stay empty, got %s", w.base)
		}
	})

	t.Run("步骤3确认视图", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 3
		w.workspace = "/workspace"
		w.worktree = "/worktrees"
		w.base = "develop"
		w.modules = []config.Module{{Name: "m1", URL: "url"}}

		view := w.View()
		if !strings.Contains(view, "确认配置") {
			t.Error("expected view to contain '确认配置'")
		}
		if !strings.Contains(view, "/workspace") {
			t.Error("expected view to contain '/workspace'")
		}
		if !strings.Contains(view, "/worktrees") {
			t.Error("expected view to contain '/worktrees'")
		}
		if !strings.Contains(view, "develop") {
			t.Error("expected view to contain 'develop'")
		}
		if !strings.Contains(view, "1") {
			t.Error("expected view to contain '1'")
		}
	})

	t.Run("退出视图", func(t *testing.T) {
		w := NewConfigWizard()
		w.quitting = true

		view := w.View()
		if view != "已退出配置向导\n" {
			t.Errorf("expected '已退出配置向导\\n', got %q", view)
		}
	})

	t.Run("未知状态", func(t *testing.T) {
		w := NewConfigWizard()
		w.step = 99

		view := w.View()
		if !strings.Contains(view, "未知状态") {
			t.Error("expected view to contain '未知状态'")
		}
	})
}

func TestSavedConfigInfo(t *testing.T) {
	info := SavedConfigInfo{
		ConfigPath: "/path/to/.modu.yaml",
		Workspace:  "/workspace",
		Worktree:   "/worktrees",
		Base:       "develop",
	}

	if info.ConfigPath != "/path/to/.modu.yaml" {
		t.Errorf("expected ConfigPath /path/to/.modu.yaml, got %s", info.ConfigPath)
	}
	if info.Workspace != "/workspace" {
		t.Errorf("expected Workspace /workspace, got %s", info.Workspace)
	}
	if info.Worktree != "/worktrees" {
		t.Errorf("expected Worktree /worktrees, got %s", info.Worktree)
	}
	if info.Base != "develop" {
		t.Errorf("expected Base develop, got %s", info.Base)
	}
}

func TestConfigWizard_SavedConfigInfo_Fields(t *testing.T) {
	info := SavedConfigInfo{
		ConfigPath: "/test/config.yaml",
		Workspace:  "/test/workspace",
		Worktree:   "/test/worktree",
		Base:       "main",
	}

	if info.ConfigPath != "/test/config.yaml" {
		t.Errorf("expected ConfigPath /test/config.yaml, got %s", info.ConfigPath)
	}
	if info.Workspace != "/test/workspace" {
		t.Errorf("expected Workspace /test/workspace, got %s", info.Workspace)
	}
	if info.Worktree != "/test/worktree" {
		t.Errorf("expected Worktree /test/worktree, got %s", info.Worktree)
	}
	if info.Base != "main" {
		t.Errorf("expected Base main, got %s", info.Base)
	}
}

// testError 测试用错误类型
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
