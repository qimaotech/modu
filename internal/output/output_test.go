package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/qimaotech/modu/internal/core"
	"github.com/qimaotech/modu/internal/engine"
)

// TestNew_DefaultFormat 测试默认格式
func TestNew_DefaultFormat(t *testing.T) {
	formatter := New("")
	if formatter.format != "text" {
		t.Errorf("expected text format, got %s", formatter.format)
	}
}

// TestNew_TextFormat 测试文本格式
func TestNew_TextFormat(t *testing.T) {
	formatter := New("text")
	if formatter.format != "text" {
		t.Errorf("expected text format, got %s", formatter.format)
	}
}

// TestNew_JsonFormat 测试 JSON 格式
func TestNew_JsonFormat(t *testing.T) {
	formatter := New("json")
	if formatter.format != "json" {
		t.Errorf("expected json format, got %s", formatter.format)
	}
}

// TestFormatCreateResponse_Text_Success 测试文本格式创建成功响应
func TestFormatCreateResponse_Text_Success(t *testing.T) {
	formatter := New("text")
	results := []Result{
		{Module: "module1", Status: "success", Path: "/path/to/module1"},
		{Module: "module2", Status: "success", Path: "/path/to/module2"},
	}

	output := formatter.FormatCreateResponse("feature-test", results, nil)

	if output == "" {
		t.Error("expected non-empty output")
	}
	if !contains(output, "Successfully created feature: feature-test") {
		t.Error("expected success message")
	}
	if !contains(output, "module1") {
		t.Error("expected module1 in output")
	}
}

// TestFormatCreateResponse_Text_Failure 测试文本格式创建失败响应
func TestFormatCreateResponse_Text_Failure(t *testing.T) {
	formatter := New("text")
	results := []Result{
		{Module: "module1", Status: "failed", Error: "some error"},
	}

	output := formatter.FormatCreateResponse("feature-test", results, []error{errTest})

	if !contains(output, "Failed to create feature: feature-test") {
		t.Error("expected failure message")
	}
}

// TestFormatCreateResponse_Json_Success 测试 JSON 格式创建成功响应
func TestFormatCreateResponse_Json_Success(t *testing.T) {
	formatter := New("json")
	results := []Result{
		{Module: "module1", Status: "success", Path: "/path/to/module1"},
	}

	output := formatter.FormatCreateResponse("feature-test", results, nil)

	// JSON MarshalIndent 使用空格格式化
	if !contains(output, `"success"`) || !contains(output, "feature-test") {
		t.Error("expected success and feature in JSON")
	}
}

// TestFormatCreateResponse_Json_Failure 测试 JSON 格式创建失败响应
func TestFormatCreateResponse_Json_Failure(t *testing.T) {
	formatter := New("json")
	results := []Result{}

	output := formatter.FormatCreateResponse("feature-test", results, []error{errTest})

	if !contains(output, `"success"`) || !contains(output, "feature-test") {
		t.Error("expected success and feature in JSON")
	}
}

// TestFormatDeleteResponse_Text_Success 测试文本格式删除成功响应
func TestFormatDeleteResponse_Text_Success(t *testing.T) {
	formatter := New("text")
	backupPath := "/worktrees/.modu/backups/20260515-153012_feature-test.tar.gz"

	output := formatter.FormatDeleteResponse("feature-test", nil, backupPath)

	if !contains(output, "已删除 feature: feature-test") {
		t.Error("expected success message")
	}
	if !contains(output, "备份文件: "+backupPath) {
		t.Error("expected backup path")
	}
}

// TestFormatDeleteResponse_Text_Failure 测试文本格式删除失败响应
func TestFormatDeleteResponse_Text_Failure(t *testing.T) {
	formatter := New("text")
	backupPath := "/worktrees/.modu/backups/20260515-153012_feature-test.tar.gz"

	output := formatter.FormatDeleteResponse("feature-test", []error{errTest}, backupPath)

	if !contains(output, "删除 feature 失败: feature-test") {
		t.Error("expected failure message")
	}
	if contains(output, "备份文件:") {
		t.Error("expected failure output to stay compatible without backup path")
	}
}

// TestFormatDeleteResponse_Json 测试 JSON 格式删除响应
func TestFormatDeleteResponse_Json(t *testing.T) {
	formatter := New("json")
	backupPath := "/worktrees/.modu/backups/20260515-153012_feature-test.tar.gz"

	output := formatter.FormatDeleteResponse("feature-test", nil, backupPath)

	if !contains(output, `"action"`) || !contains(output, "delete") {
		t.Error("expected action in JSON")
	}

	var resp DeleteResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}
	if resp.BackupPath != backupPath {
		t.Fatalf("expected backupPath %q, got %q", backupPath, resp.BackupPath)
	}
}

func TestFormatBackupListResponse_Text(t *testing.T) {
	formatter := New("text")
	backups := []engine.DeleteBackupInfo{{
		ID:        "20260515-153012_feature-test",
		Feature:   "feature-test",
		Path:      "/worktrees/.modu/backups/20260515-153012_feature-test.tar.gz",
		CreatedAt: time.Date(2026, 5, 15, 15, 30, 12, 0, time.Local),
		SizeBytes: 1024,
	}}

	output := formatter.FormatBackupListResponse(backups)

	if !contains(output, "备份列表:") {
		t.Error("expected backup list header")
	}
	if !contains(output, "20260515-153012_feature-test") {
		t.Error("expected backup id")
	}
	if !contains(output, "feature-test") {
		t.Error("expected feature")
	}
	if !contains(output, "路径: /worktrees/.modu/backups/20260515-153012_feature-test.tar.gz") {
		t.Error("expected backup path")
	}
}

func TestFormatBackupListResponse_Json(t *testing.T) {
	formatter := New("json")
	backups := []engine.DeleteBackupInfo{{
		ID:        "20260515-153012_feature-test",
		Feature:   "feature-test",
		Path:      "/worktrees/.modu/backups/20260515-153012_feature-test.tar.gz",
		CreatedAt: time.Date(2026, 5, 15, 15, 30, 12, 0, time.Local),
		SizeBytes: 1024,
	}}

	output := formatter.FormatBackupListResponse(backups)

	var resp BackupListResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}
	if !resp.Success || resp.Action != "backup-list" {
		t.Fatalf("unexpected response metadata: %#v", resp)
	}
	if len(resp.Backups) != 1 || resp.Backups[0].ID != backups[0].ID {
		t.Fatalf("unexpected backups: %#v", resp.Backups)
	}
}

func TestFormatBackupRestoreResponse_Text(t *testing.T) {
	formatter := New("text")
	result := &engine.RestoreDeleteBackupResult{
		Feature:    "feature-test",
		Path:       "/worktrees/feature-test",
		BackupPath: "/worktrees/.modu/backups/20260515-153012_feature-test.tar.gz",
	}

	output := formatter.FormatBackupRestoreResponse(result)

	if !contains(output, "已恢复备份: feature-test") {
		t.Error("expected restore success message")
	}
	if !contains(output, "目标路径: /worktrees/feature-test") {
		t.Error("expected restore path")
	}
	if !contains(output, "备份文件: "+result.BackupPath) {
		t.Error("expected backup path")
	}
}

func TestFormatBackupRestoreResponse_Json(t *testing.T) {
	formatter := New("json")
	result := &engine.RestoreDeleteBackupResult{
		Feature:    "feature-test",
		Path:       "/worktrees/feature-test",
		BackupPath: "/worktrees/.modu/backups/20260515-153012_feature-test.tar.gz",
	}

	output := formatter.FormatBackupRestoreResponse(result)

	var resp BackupRestoreResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}
	if !resp.Success || resp.Action != "backup-restore" {
		t.Fatalf("unexpected response metadata: %#v", resp)
	}
	if resp.Feature != result.Feature || resp.Path != result.Path || resp.BackupPath != result.BackupPath {
		t.Fatalf("unexpected restore response: %#v", resp)
	}
}

// TestFormatError_Text 测试文本格式错误响应
func TestFormatError_Text(t *testing.T) {
	formatter := New("text")

	output := formatter.FormatError("ERR_TEST", "test error message", nil)

	if !contains(output, "Error [ERR_TEST]") {
		t.Error("expected error code in output")
	}
	if !contains(output, "test error message") {
		t.Error("expected error message in output")
	}
}

// TestFormatError_Json 测试 JSON 格式错误响应
func TestFormatError_Json(t *testing.T) {
	formatter := New("json")

	output := formatter.FormatError("ERR_TEST", "test error message", map[string]string{"key": "value"})

	if !contains(output, `"code"`) || !contains(output, "ERR_TEST") {
		t.Error("expected code in JSON")
	}
	if !contains(output, `"message"`) || !contains(output, "test error message") {
		t.Error("expected message in JSON")
	}
}

// TestFormatListResponse_Text 测试文本格式列表响应
func TestFormatListResponse_Text(t *testing.T) {
	formatter := New("text")
	envs := []core.WorktreeEnv{
		{
			Name: "feature-1",
			MainProject: &core.ModuleStatus{
				Name: "main",
				Path: "/path/to/main",
			},
			Modules: []core.ModuleStatus{
				{Name: "module1", Branch: "feature-1", IsDirty: false},
				{Name: "module2", Branch: "feature-1", IsDirty: true},
			},
		},
	}

	output := formatter.FormatListResponse(envs, true)

	if !contains(output, "Features:") {
		t.Error("expected Features header")
	}
	if !contains(output, "feature-1") {
		t.Error("expected feature name in output")
	}
	if !contains(output, "clean") {
		t.Error("expected clean status")
	}
	if !contains(output, "dirty") {
		t.Error("expected dirty status")
	}
}

func TestFormatListResponse_Text_ShowsMainProjectDirty(t *testing.T) {
	formatter := New("text")
	envs := []core.WorktreeEnv{
		{
			Name: "feature-main-dirty",
			MainProject: &core.ModuleStatus{
				Name:    "main",
				Path:    "/path/to/main",
				IsDirty: true,
				Branch:  "feature-main-dirty",
			},
			Modules: []core.ModuleStatus{
				{Name: "module1", Branch: "feature-main-dirty", IsDirty: false},
			},
		},
	}

	output := formatter.FormatListResponse(envs, true)

	if !contains(output, "main: dirty") {
		t.Fatalf("expected main project dirty status, got %q", output)
	}
}

func TestFormatInfoResponse_Text_ShowsMainProjectDirty(t *testing.T) {
	formatter := New("text")
	env := &core.WorktreeEnv{
		Name: "feature-main-dirty",
		MainProject: &core.ModuleStatus{
			Name:    "main",
			Path:    "/path/to/main",
			IsDirty: true,
			Branch:  "feature-main-dirty",
		},
	}

	output := formatter.FormatInfoResponse(env)

	if !contains(output, "Main Project") || !contains(output, "dirty") {
		t.Fatalf("expected main project dirty status in info output, got %q", output)
	}
}

// TestFormatListResponse_Empty 测试空列表响应
func TestFormatListResponse_Empty(t *testing.T) {
	formatter := New("text")

	output := formatter.FormatListResponse([]core.WorktreeEnv{}, false)

	if !contains(output, "Features:") {
		t.Error("expected Features header")
	}
}

// TestFormatListResponse_Json 测试 JSON 格式列表响应
func TestFormatListResponse_Json(t *testing.T) {
	formatter := New("json")
	envs := []core.WorktreeEnv{
		{Name: "feature-1"},
	}

	output := formatter.FormatListResponse(envs, false)

	if !contains(output, `"name"`) || !contains(output, "feature-1") {
		t.Error("expected feature name in JSON")
	}
}

// TestFormatInfoResponse_Text 测试文本格式详情响应
func TestFormatInfoResponse_Text(t *testing.T) {
	formatter := New("text")
	env := &core.WorktreeEnv{
		Name: "feature-1",
		Modules: []core.ModuleStatus{
			{Name: "module1", Branch: "feature-1", IsDirty: false, Path: "/path/to/module1"},
		},
	}

	output := formatter.FormatInfoResponse(env)

	if !contains(output, "Feature: feature-1") {
		t.Error("expected feature name")
	}
	if !contains(output, "module1") {
		t.Error("expected module name")
	}
	if !contains(output, "Branch: feature-1") {
		t.Error("expected branch info")
	}
}

// TestFormatInfoResponse_Json 测试 JSON 格式详情响应
func TestFormatInfoResponse_Json(t *testing.T) {
	formatter := New("json")
	env := &core.WorktreeEnv{
		Name: "feature-1",
		Modules: []core.ModuleStatus{
			{Name: "module1", Branch: "feature-1"},
		},
	}

	output := formatter.FormatInfoResponse(env)

	if !contains(output, `"name"`) || !contains(output, "feature-1") {
		t.Error("expected name in JSON")
	}
}

// TestFormatMainProjectResponseText 测试文本格式主项目响应
func TestFormatMainProjectResponseText(t *testing.T) {
	formatter := New("text")
	modules := []core.ModuleStatus{
		{Name: "module1", Branch: "develop", IsDirty: false},
		{Name: "module2", Branch: "develop", IsDirty: true},
	}

	output := formatter.FormatMainProjectResponse("develop", modules)

	if !contains(output, "Workspace [develop]") {
		t.Error("expected Workspace header")
	}
	if !contains(output, "module1: develop") {
		t.Error("expected module1 in output")
	}
	if !contains(output, "module2: develop") {
		t.Error("expected module2 in output")
	}
}

// TestFormatMainProjectResponseJson 测试 JSON 格式主项目响应
func TestFormatMainProjectResponseJson(t *testing.T) {
	formatter := New("json")
	modules := []core.ModuleStatus{
		{Name: "module1", Branch: "develop"},
		{Name: "module2", Branch: "develop"},
	}

	output := formatter.FormatMainProjectResponse("develop", modules)

	if !contains(output, `"name"`) || !contains(output, "workspace") {
		t.Error("expected name in JSON")
	}
	if !contains(output, `"branch"`) || !contains(output, "develop") {
		t.Error("expected branch in JSON")
	}
	if !contains(output, `"modules"`) {
		t.Error("expected modules in JSON")
	}
}

// contains 是辅助函数，检查字符串是否包含子串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// errTest 是测试用的错误
var errTest = &testError{"test error"}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
