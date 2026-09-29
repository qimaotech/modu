package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qimaotech/modu/internal/config"
)

func TestUpdateCLI_CurrentDirectory(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "modu")
	build := newUpdateTestCommand(t, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("构建 CLI: %v\n%s", err, output)
	}

	for _, scenario := range []struct {
		name                string
		workspaceInsideRoot bool
		location            string
		featureArg          string
		wantFeature         string
	}{
		{"feature根目录", false, "feature", "", "feature-order-export"},
		{"模块子目录推断feature", false, "feature-module", "", "feature-order-export"},
		{"通过符号链接访问feature", false, "symlink-feature", "", "feature-order-export"},
		{"原始主项目", false, "workspace", "", ""},
		{"原始模块子目录", false, "workspace-module", "", ""},
		{"worktreeRoot本身", false, "worktree-root", "", ""},
		{"工作区之外", false, "outside", "", ""},
		{"主项目位于worktreeRoot之下", true, "workspace", "", ""},
		{"原始模块位于worktreeRoot之下", true, "workspace-module", "", ""},
		{"显式参数优先于当前feature", false, "other-feature", "feature/order-export", "feature/order-export"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			workspace := newUpdateTestWorkspace(t, scenario.workspaceInsideRoot)
			cwd := filepath.Join(workspace.featurePath, "backend", "src")
			switch scenario.location {
			case "feature":
				cwd = workspace.featurePath
			case "workspace":
				cwd = workspace.mainPath
			case "workspace-module":
				cwd = filepath.Join(workspace.mainPath, "backend", "src")
			case "worktree-root":
				cwd = filepath.Dir(workspace.featurePath)
			case "outside":
				cwd = t.TempDir()
			case "symlink-feature":
				link := filepath.Join(t.TempDir(), "feature-link")
				if err := os.Symlink(workspace.featurePath, link); err != nil {
					t.Fatal(err)
				}
				cwd = filepath.Join(link, "backend", "src")
			case "other-feature":
				cwd = filepath.Join(filepath.Dir(workspace.featurePath), "feature-other")
				runUpdateTestGit(t, workspace.mainPath, "worktree", "add", "-b", "feature/other", cwd, "develop")
			}
			args := []string{"update", "-c", workspace.configPath, "-o", "json"}
			if scenario.featureArg != "" {
				args = append(args, scenario.featureArg)
			}
			cmd := newUpdateTestCommand(t, binary, args...)
			cmd.Dir = cwd
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("更新代码: %v\nstdout: %s\nstderr: %s", err, output, stderr.String())
			}
			var response struct {
				Success bool   `json:"success"`
				Feature string `json:"feature"`
			}
			if err := json.Unmarshal(output, &response); err != nil {
				t.Fatalf("解析 JSON: %v\n%s", err, output)
			}
			if !response.Success || response.Feature != scenario.wantFeature {
				t.Fatalf("期望 feature=%q，实际输出: %s", scenario.wantFeature, output)
			}
			if scenario.location == "other-feature" {
				assertUpdateTestFile(t, cwd, "initial\n")
			}

			featureContent, mainContent, mainBranch := "feature updated\n", "initial\n", "workspace-local"
			if scenario.wantFeature == "" {
				featureContent, mainContent, mainBranch = "initial\n", "workspace updated\n", "develop"
			}
			for _, repository := range []string{workspace.featurePath, filepath.Join(workspace.featurePath, "backend")} {
				assertUpdateTestFile(t, repository, featureContent)
				if branch := runUpdateTestGit(t, repository, "branch", "--show-current"); branch != "feature/order-export" {
					t.Errorf("feature 分支被切换为 %q", branch)
				}
			}
			for _, repository := range []string{workspace.mainPath, filepath.Join(workspace.mainPath, "backend")} {
				assertUpdateTestFile(t, repository, mainContent)
				if branch := runUpdateTestGit(t, repository, "branch", "--show-current"); branch != mainBranch {
					t.Errorf("原始 workspace 分支为 %q，期望 %q", branch, mainBranch)
				}
			}
		})
	}
}

type updateTestWorkspace struct {
	mainPath    string
	featurePath string
	configPath  string
}

func newUpdateTestWorkspace(t *testing.T, workspaceInsideRoot bool) updateTestWorkspace {
	t.Helper()
	root := t.TempDir()
	origin := filepath.Join(root, "origin")
	if err := os.MkdirAll(origin, 0755); err != nil {
		t.Fatal(err)
	}
	runUpdateTestGit(t, origin, "init", "-b", "develop")
	writeUpdateTestFile(t, filepath.Join(origin, ".gitignore"), "backend/\n")
	writeUpdateTestFile(t, filepath.Join(origin, "src", "version.txt"), "initial\n")
	runUpdateTestGit(t, origin, "add", ".")
	runUpdateTestGit(t, origin, "commit", "-m", "initial")
	runUpdateTestGit(t, origin, "branch", "feature/order-export")

	worktreeRoot := filepath.Join(root, "worktrees")
	if workspaceInsideRoot {
		worktreeRoot = root
	}
	workspace := updateTestWorkspace{
		mainPath:    filepath.Join(root, "main"),
		featurePath: filepath.Join(worktreeRoot, "feature-order-export"),
		configPath:  filepath.Join(root, ".modu.yaml"),
	}
	for _, module := range []string{"", "backend"} {
		repository := filepath.Join(workspace.mainPath, module)
		runUpdateTestGit(t, root, "clone", origin, repository)
		runUpdateTestGit(t, repository, "checkout", "-b", "workspace-local")
		runUpdateTestGit(t, repository, "worktree", "add", "-b", "feature/order-export",
			filepath.Join(workspace.featurePath, module), "origin/feature/order-export")
	}
	if err := config.SaveConfig(&config.Config{
		Workspace:    workspace.mainPath,
		WorktreeRoot: worktreeRoot,
		DefaultBase:  "develop",
		Modules:      []config.Module{{Name: "backend", URL: origin}},
	}, workspace.configPath); err != nil {
		t.Fatal(err)
	}

	// 两个远端分支分别推进，确保测试能区分更新 feature 和更新原始 workspace。
	for _, revision := range []struct{ branch, content string }{
		{"feature/order-export", "feature updated\n"},
		{"develop", "workspace updated\n"},
	} {
		runUpdateTestGit(t, origin, "checkout", revision.branch)
		writeUpdateTestFile(t, filepath.Join(origin, "src", "version.txt"), revision.content)
		runUpdateTestGit(t, origin, "add", "src/version.txt")
		runUpdateTestGit(t, origin, "commit", "-m", "remote update")
	}
	return workspace
}

func runUpdateTestGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	gitArgs := append([]string{
		"-c", "user.name=modu-test",
		"-c", "user.email=modu-test@localhost",
		"-c", "commit.gpgsign=false",
		"-c", "core.hooksPath=/dev/null",
	}, args...)
	cmd := newUpdateTestCommand(t, "git", gitArgs...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

func newUpdateTestCommand(t *testing.T, name string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), name, args...)
	// Git 钩子导出的仓库环境变量不能影响测试使用的临时仓库。
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GIT_") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	return cmd
}

func writeUpdateTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertUpdateTestFile(t *testing.T, repository, want string) {
	t.Helper()
	path := filepath.Join(repository, "src", "version.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != want {
		t.Errorf("%s 内容为 %q，期望 %q", path, content, want)
	}
}
