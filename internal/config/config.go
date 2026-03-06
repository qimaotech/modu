package config

import (
	"errors"
	"fmt"
	errs "codeup.aliyun.com/qimao/public/devops/modu/internal/errors"
	"os"
	"path/filepath"

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

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 5
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

// DefaultConfig 返回默认配置模板
func DefaultConfig() *Config {
	return &Config{
		Workspace:    "./workspace",
		WorktreeRoot: "./worktrees",
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

	return nil
}
