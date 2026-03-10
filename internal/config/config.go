package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	errs "github.com/qimaotech/modu/internal/errors"
	"gopkg.in/yaml.v3"
)

// Config modu 配置文件结构
type Config struct {
	Workspace    string   `yaml:"workspace"`          // 裸仓库/主仓库所在目录
	WorktreeRoot string   `yaml:"worktree-root"`      // 特性分支代码存放目录
	DefaultBase  string   `yaml:"default-base"`       // 默认基准分支 (如 develop)
	Concurrency  int      `yaml:"concurrency"`        // 并发数，默认 5
	AutoFetch    bool     `yaml:"auto-fetch"`         // 操作前自动 fetch
	StrictDirty  bool     `yaml:"strict-dirty-check"` // 删除前强制脏检查
	Modules      []Module `yaml:"modules"`            // 模块列表
}

// IsConfigNotFoundError 检查是否为配置文件不存在错误
func IsConfigNotFoundError(err error) bool {
	return errors.Is(err, ErrConfigNotFound)
}

// IsConfigValidationError 检查是否为配置验证错误
func IsConfigValidationError(err error) bool {
	return errors.Is(err, errs.ErrConfigInvalid)
}

// ErrConfigNotFound 配置文件不存在错误
var ErrConfigNotFound = errors.New("config file not found")

// Module 模块配置
type Module struct {
	Name       string `yaml:"name"`                  // 模块名称
	URL        string `yaml:"url"`                   // 仓库 URL
	BaseBranch string `yaml:"base-branch,omitempty"` // 可选，覆盖全局设置
}

// LoadConfig 加载并校验配置文件
func LoadConfig(path string) (*Config, error) {
	cfg, err := loadConfigImpl(path)
	if err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}
	// 设置默认值
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 5
	}
	return cfg, nil
}

// LoadConfigForScan 加载配置文件用于 scan 命令，跳过模块验证
func LoadConfigForScan(path string) (*Config, error) {
	cfg, err := loadConfigImpl(path)
	if err != nil {
		return nil, err
	}
	// 只验证基础字段，不验证模块
	if err := validateBasic(cfg); err != nil {
		return nil, err
	}
	// 设置默认值
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 5
	}
	return cfg, nil
}

// loadConfigImpl 加载配置文件的内部实现
func loadConfigImpl(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, absPath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	// 将 workspace 和 worktree-root 转换为绝对路径（相对于配置文件所在目录）
	configDir := filepath.Dir(absPath)
	if cfg.Workspace != "" && !filepath.IsAbs(cfg.Workspace) {
		cfg.Workspace = filepath.Join(configDir, cfg.Workspace)
	}
	if cfg.WorktreeRoot != "" && !filepath.IsAbs(cfg.WorktreeRoot) {
		cfg.WorktreeRoot = filepath.Join(configDir, cfg.WorktreeRoot)
	}

	return &cfg, nil
}

// validate 校验配置必填字段
func validate(cfg *Config) error {
	var validationErrs []error
	if cfg.Workspace == "" {
		validationErrs = append(validationErrs, fmt.Errorf("%w: workspace is required", errs.ErrConfigInvalid))
	}
	if cfg.WorktreeRoot == "" {
		validationErrs = append(validationErrs, fmt.Errorf("%w: worktree-root is required", errs.ErrConfigInvalid))
	}
	if cfg.DefaultBase == "" {
		validationErrs = append(validationErrs, fmt.Errorf("%w: default-base is required", errs.ErrConfigInvalid))
	}
	if len(cfg.Modules) == 0 {
		validationErrs = append(validationErrs, fmt.Errorf("%w: at least one module is required", errs.ErrConfigInvalid))
	}
	return errors.Join(validationErrs...)
}

// validateBasic 校验配置基础必填字段（不验证模块）
func validateBasic(cfg *Config) error {
	var validationErrs []error
	if cfg.Workspace == "" {
		validationErrs = append(validationErrs, fmt.Errorf("%w: workspace is required", errs.ErrConfigInvalid))
	}
	if cfg.WorktreeRoot == "" {
		validationErrs = append(validationErrs, fmt.Errorf("%w: worktree-root is required", errs.ErrConfigInvalid))
	}
	if cfg.DefaultBase == "" {
		validationErrs = append(validationErrs, fmt.Errorf("%w: default-base is required", errs.ErrConfigInvalid))
	}
	return errors.Join(validationErrs...)
}

// DefaultConfig 返回默认配置模板
func DefaultConfig() *Config {
	return &Config{
		Workspace:    ".",
		WorktreeRoot: "../worktrees",
		DefaultBase:  "develop",
		Concurrency:  5,
		AutoFetch:    true,
		StrictDirty:  true,
		Modules:      []Module{},
	}
}

// SaveConfig 保存配置到文件
func SaveConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	if err := os.WriteFile(absPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 自动创建 WorktreeRoot 目录（相对于配置文件所在目录）
	if cfg.WorktreeRoot != "" {
		configDir := filepath.Dir(absPath)
		worktreeRootAbs := cfg.WorktreeRoot

		// 处理相对路径（相对于配置文件目录）
		if !filepath.IsAbs(worktreeRootAbs) {
			// 如果是 ../ 开头的相对路径，相对于配置文件目录解析
			worktreeRootAbs = filepath.Join(configDir, worktreeRootAbs)
		}

		// 检查目录是否可以创建（不阻塞保存配置）
		if err := os.MkdirAll(worktreeRootAbs, 0755); err != nil {
			// 如果创建失败，只打印警告，不阻止保存配置
			fmt.Fprintf(os.Stderr, "警告: 无法创建 worktree 目录 %s: %v\n", worktreeRootAbs, err)
		}
	}

	return nil
}

// ScanWorkspace 扫描 workspace 目录，返回所有 git 仓库模块
func ScanWorkspace(ctx context.Context, workspacePath string) ([]Module, error) {
	absPath, err := filepath.Abs(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	// 检查目录是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace directory does not exist: %s", absPath)
	}

	// 读取目录
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace directory: %w", err)
	}

	// 预分配模块切片
	modules := make([]Module, 0, len(entries))
	for _, entry := range entries {
		// 只处理目录
		if !entry.IsDir() {
			continue
		}

		modulePath := filepath.Join(absPath, entry.Name())

		// 检查是否为 git 仓库（存在 .git 目录或 .git 文件）
		gitDir := filepath.Join(modulePath, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			continue
		}

		// 读取 git config 获取 origin URL
		gitConfigPath := filepath.Join(gitDir, "config")
		url, err := readGitRemoteURL(gitConfigPath)
		if err != nil {
			// 跳过无法读取 URL 的仓库
			continue
		}

		modules = append(modules, Module{
			Name: entry.Name(),
			URL:  url,
		})
	}

	return modules, nil
}

// readGitRemoteURL 读取 .git/config 文件获取 origin remote URL
func readGitRemoteURL(gitConfigPath string) (string, error) {
	data, err := os.ReadFile(gitConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read git config: %w", err)
	}

	// 简单解析 [remote "origin"] 下的 url
	content := string(data)
	lines := strings.Split(content, "\n")

	inOrigin := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检测 [remote "origin"]
		if strings.HasPrefix(line, "[") {
			inOrigin = strings.Contains(line, `remote "origin"`)
			continue
		}

		// 在 origin 块中查找 url
		if inOrigin && strings.HasPrefix(line, "url = ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "url = ")), nil
		}
	}

	return "", fmt.Errorf("remote origin not found")
}

// UpdateGitignore 更新主项目的 .gitignore，添加模块目录
func UpdateGitignore(workspacePath string, modules []Module) error {
	gitignorePath := filepath.Join(workspacePath, ".gitignore")
	var existingEntries []string

	// 读取现有的 .gitignore
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existingEntries = strings.Split(strings.TrimSpace(string(data)), "\n")
	}

	// 收集需要添加的模块名
	entriesToAdd := make(map[string]bool)
	for _, m := range modules {
		entriesToAdd[m.Name] = true
	}

	// 检查是否需要添加新条目
	needsUpdate := false
	for name := range entriesToAdd {
		found := false
		for _, entry := range existingEntries {
			entry = strings.TrimSpace(entry)
			if entry == name || entry == name+"/" {
				found = true
				break
			}
		}
		if !found {
			needsUpdate = true
			break
		}
	}

	if !needsUpdate {
		return nil
	}

	// 追加新条目
	var newContent strings.Builder
	if _, err := os.Stat(gitignorePath); err == nil {
		// 文件存在，读取内容
		data, _ := os.ReadFile(gitignorePath)
		newContent.WriteString(strings.TrimSpace(string(data)))
	}

	// 添加模块目录
	newContent.WriteString("\n\n# modu modules\n")
	for name := range entriesToAdd {
		newContent.WriteString(name + "/\n")
	}

	if err := os.WriteFile(gitignorePath, []byte(newContent.String()), 0600); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}
	return nil
}
