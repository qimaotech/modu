package output

import (
	"strings"
	"testing"

	"codeup.aliyun.com/qimao/public/devops/modu/internal/core"
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

	output := formatter.FormatDeleteResponse("feature-test", nil)

	if !contains(output, "Successfully deleted feature: feature-test") {
		t.Error("expected success message")
	}
}

// TestFormatDeleteResponse_Text_Failure 测试文本格式删除失败响应
func TestFormatDeleteResponse_Text_Failure(t *testing.T) {
	formatter := New("text")

	output := formatter.FormatDeleteResponse("feature-test", []error{errTest})

	if !contains(output, "Failed to delete feature: feature-test") {
		t.Error("expected failure message")
	}
}

// TestFormatDeleteResponse_Json 测试 JSON 格式删除响应
func TestFormatDeleteResponse_Json(t *testing.T) {
	formatter := New("json")

	output := formatter.FormatDeleteResponse("feature-test", nil)

	if !contains(output, `"action"`) || !contains(output, "delete") {
		t.Error("expected action in JSON")
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
			Modules: []core.ModuleStatus{
				{Name: "module1", Branch: "feature-1", IsDirty: false},
				{Name: "module2", Branch: "feature-1", IsDirty: true},
			},
		},
	}

	output := formatter.FormatListResponse(envs)

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

// TestFormatListResponse_Empty 测试空列表响应
func TestFormatListResponse_Empty(t *testing.T) {
	formatter := New("text")

	output := formatter.FormatListResponse([]core.WorktreeEnv{})

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

	output := formatter.FormatListResponse(envs)

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
