package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qimaotech/modu/internal/core"
)

// Result 单个模块的操作结果
type Result struct {
	Module string `json:"module"`
	Status string `json:"status"`
	Path   string `json:"path,omitempty"`
	Error  string `json:"error,omitempty"`
}

// CreateResponse 创建操作的响应
type CreateResponse struct {
	Success bool     `json:"success"`
	Action  string   `json:"action"`
	Feature string   `json:"feature"`
	Results []Result `json:"results"`
	Errors  []string `json:"errors"`
}

// DeleteResponse 删除操作的响应
type DeleteResponse struct {
	Success bool     `json:"success"`
	Action  string   `json:"action"`
	Feature string   `json:"feature"`
	Results []Result `json:"results"`
	Errors  []string `json:"errors"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Formatter 输出格式化器
type Formatter struct {
	format string // "text" or "json"
}

// New 创建格式化器
func New(format string) *Formatter {
	if format == "" {
		format = "text"
	}
	return &Formatter{format: format}
}

// FormatCreateResponse 格式化创建响应
func (f *Formatter) FormatCreateResponse(feature string, results []Result, errs []error) string {
	if f.format == "json" {
		resp := CreateResponse{
			Success: len(errs) == 0,
			Action:  "create",
			Feature: feature,
			Results: results,
		}
		for _, e := range errs {
			resp.Errors = append(resp.Errors, e.Error())
		}
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return `{"error": "failed to marshal response"}`
		}
		return string(data)
	}

	// 文本格式
	var sb strings.Builder
	if len(errs) == 0 {
		sb.WriteString(fmt.Sprintf("✓ Successfully created feature: %s\n", feature))
	} else {
		sb.WriteString(fmt.Sprintf("✗ Failed to create feature: %s\n", feature))
	}

	for _, r := range results {
		if r.Status == "success" {
			sb.WriteString(fmt.Sprintf("  ✓ %s: %s\n", r.Module, r.Path))
		} else {
			sb.WriteString(fmt.Sprintf("  ✗ %s: %s\n", r.Module, r.Error))
		}
	}

	if len(errs) > 0 {
		sb.WriteString("\nErrors:\n")
		for _, e := range errs {
			sb.WriteString(fmt.Sprintf("  - %s\n", e.Error()))
		}
	}

	return sb.String()
}

// FormatDeleteResponse 格式化删除响应
func (f *Formatter) FormatDeleteResponse(feature string, errs []error) string {
	if f.format == "json" {
		resp := DeleteResponse{
			Success: len(errs) == 0,
			Action:  "delete",
			Feature: feature,
		}
		for _, e := range errs {
			resp.Errors = append(resp.Errors, e.Error())
		}
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return `{"error": "failed to marshal response"}`
		}
		return string(data)
	}

	// 文本格式
	var sb strings.Builder
	if len(errs) == 0 {
		sb.WriteString(fmt.Sprintf("✓ 已删除 feature: %s\n", feature))
	} else {
		sb.WriteString(fmt.Sprintf("✗ 删除 feature 失败: %s\n", feature))
		for _, e := range errs {
			sb.WriteString(fmt.Sprintf("  错误: %s\n", e.Error()))
		}
	}

	return sb.String()
}

// FormatError 格式化错误响应
func (f *Formatter) FormatError(code, message string, data interface{}) string {
	if f.format == "json" {
		resp := ErrorResponse{
			Code:    code,
			Message: message,
			Data:    data,
		}
		jsonData, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return `{"error": "failed to marshal response"}`
		}
		return string(jsonData)
	}

	// 文本格式
	return fmt.Sprintf("Error [%s]: %s\n", code, message)
}

// FormatListResponse 格式化列表响应
func (f *Formatter) FormatListResponse(envs []core.WorktreeEnv) string {
	if f.format == "json" {
		data, err := json.MarshalIndent(envs, "", "  ")
		if err != nil {
			return `{"error": "failed to marshal response"}`
		}
		return string(data)
	}

	// 文本格式
	var sb strings.Builder
	sb.WriteString("Features:\n")
	for _, env := range envs {
		sb.WriteString(fmt.Sprintf("  - %s[%s]\n", env.Name, env.MainProject.Path))

		// 输出模块
		for _, mod := range env.Modules {
			status := "clean"
			if mod.IsDirty {
				status = "dirty"
			}
			sb.WriteString(fmt.Sprintf("    - %s: %s (%s)\n", mod.Name, mod.Branch, status))
		}
	}

	return sb.String()
}

// FormatInfoResponse 格式化详情响应
func (f *Formatter) FormatInfoResponse(env *core.WorktreeEnv) string {
	if f.format == "json" {
		data, err := json.MarshalIndent(env, "", "  ")
		if err != nil {
			return `{"error": "failed to marshal response"}`
		}
		return string(data)
	}

	// 文本格式
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Feature: %s\n", env.Name))
	sb.WriteString("Modules:\n")
	for _, mod := range env.Modules {
		status := "clean"
		if mod.IsDirty {
			status = "dirty"
		}
		sb.WriteString(fmt.Sprintf("  - %s\n", mod.Name))
		sb.WriteString(fmt.Sprintf("    Branch: %s\n", mod.Branch))
		sb.WriteString(fmt.Sprintf("    Status: %s\n", status))
		sb.WriteString(fmt.Sprintf("    Path: %s\n", mod.Path))
	}

	return sb.String()
}
