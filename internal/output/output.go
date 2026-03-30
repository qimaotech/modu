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

// MainProjectInfo 主项目信息（用于 list -a 输出）
type MainProjectInfo struct {
	Name    string         `json:"name"`
	Branch  string         `json:"branch"`
	Modules []ModuleStatus `json:"modules"`
}

// ModuleStatus 模块状态
type ModuleStatus struct {
	Name   string `json:"name"`
	Branch string `json:"branch"`
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
// showStatus 控制是否显示模块的 clean/dirty 状态
func (f *Formatter) FormatListResponse(envs []core.WorktreeEnv, showStatus bool) string {
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
			var statusStr string
			if showStatus {
				status := "clean"
				if mod.IsDirty {
					status = "dirty"
				}
				statusStr = fmt.Sprintf(" (%s)", status)
			}
			sb.WriteString(fmt.Sprintf("    - %s: %s%s\n", mod.Name, mod.Branch, statusStr))
		}
	}

	return sb.String()
}

// FormatMainProjectResponse 格式化主项目信息（用于 list -a）
// 返回主项目信息字符串，会显示在 Features 列表之前
func (f *Formatter) FormatMainProjectResponse(mainBranch string, modules []core.ModuleStatus) string {
	if f.format == "json" {
		mods := make([]ModuleStatus, len(modules))
		for i, m := range modules {
			mods[i] = ModuleStatus{
				Name:   m.Name,
				Branch: m.Branch,
			}
		}
		info := MainProjectInfo{
			Name:    "workspace",
			Branch:  mainBranch,
			Modules: mods,
		}
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return `{"error": "failed to marshal response"}`
		}
		return string(data)
	}

	// 文本格式：Workspace [develop]
	//   - pixiu-ad-backend: develop
	//   - pixiu-frontend: develop
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Workspace [%s]\n", mainBranch))
	for _, mod := range modules {
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", mod.Name, mod.Branch))
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

	// 文本格式（带 emoji 美化）
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔖 Feature: %s\n", env.Name))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("📦 Modules:\n")
	for i, mod := range env.Modules {
		status := "✅ clean"
		if mod.IsDirty {
			status = "🔴 dirty"
		}
		prefix := "├─"
		indent := "│   "
		if i == len(env.Modules)-1 {
			prefix = "└─"
			indent = "    "
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", prefix, mod.Name))
		sb.WriteString(fmt.Sprintf("  %s🌿 Branch: %s  %s\n", indent, mod.Branch, status))
		sb.WriteString(fmt.Sprintf("  %s📁 Path: %s\n", indent, mod.Path))
	}

	return sb.String()
}
