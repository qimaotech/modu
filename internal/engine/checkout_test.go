package engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/qimaotech/modu/internal/config"
	"github.com/qimaotech/modu/internal/engine"
	errs "github.com/qimaotech/modu/internal/errors"
)

const checkoutBranch = "feature/checkout"

func TestCheckoutFeature_DiscoversAndClonesRemoteModules(t *testing.T) {
	fixture := newCheckoutFixture(t)
	results, err := fixture.engine.CheckoutFeature(context.Background(), checkoutBranch)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 4 {
		t.Fatalf("expected main and three module results, got %+v", results)
	}
	for _, name := range []string{"project", "backend", "frontend"} {
		worktree := fixture.featurePath
		if name != "project" {
			worktree = filepath.Join(worktree, name)
		}
		if got := checkoutGit(t, worktree, "branch", "--show-current"); got != checkoutBranch {
			t.Fatalf("%s branch = %q", name, got)
		}
		if got := checkoutGit(t, worktree, "rev-parse", "--abbrev-ref", "@{upstream}"); got != "origin/feature/checkout" {
			t.Fatalf("%s upstream = %q", name, got)
		}
		wantCommit := checkoutGit(t, fixture.remotes[name], "rev-parse", "refs/heads/feature/checkout")
		if got := checkoutGit(t, worktree, "rev-parse", "HEAD"); got != wantCommit {
			t.Fatalf("%s did not checkout the published commit", name)
		}
		contents, err := os.ReadFile(filepath.Join(worktree, "requirement.txt"))
		if err != nil || string(contents) != "published requirement\n" {
			t.Fatalf("%s requirement = %q, error = %v", name, contents, err)
		}
	}
	for _, path := range []string{filepath.Join(fixture.config.Workspace, "unrelated"), filepath.Join(fixture.featurePath, "unrelated")} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("module without remote feature must not be created: %s (%v)", path, err)
		}
	}
}

func TestCheckoutFeature_RepeatedCheckoutPreservesWorkAndAddsNewModule(t *testing.T) {
	fixture := newCheckoutFixture(t)
	if _, err := fixture.engine.CheckoutFeature(context.Background(), checkoutBranch); err != nil {
		t.Fatal(err)
	}
	backend := filepath.Join(fixture.featurePath, "backend")
	frontendFile := filepath.Join(fixture.featurePath, "frontend", "requirement.txt")
	checkoutCommit(t, backend, "local.txt", "tester local commit\n")
	localCommit := checkoutGit(t, backend, "rev-parse", "HEAD")
	if err := os.WriteFile(frontendFile, []byte("tester uncommitted changes\n"), 0644); err != nil {
		t.Fatal(err)
	}
	projectCommit := checkoutGit(t, fixture.featurePath, "rev-parse", "HEAD")
	checkoutCommit(t, fixture.developers["project"], "requirement.txt", "new remote version\n")
	checkoutGit(t, fixture.developers["project"], "push", "origin", checkoutBranch)
	checkoutGit(t, fixture.developers["unrelated"], "checkout", "-b", checkoutBranch)
	checkoutCommit(t, fixture.developers["unrelated"], "requirement.txt", "newly included module\n")
	checkoutGit(t, fixture.developers["unrelated"], "push", "origin", checkoutBranch)

	results, err := fixture.engine.CheckoutFeature(context.Background(), checkoutBranch)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results[:3] {
		if result.Status != "skipped" || result.Commit == "" {
			t.Fatalf("existing worktree not reported as preserved: %+v", result)
		}
	}
	if got := checkoutGit(t, backend, "rev-parse", "HEAD"); got != localCommit {
		t.Fatal("repeat checkout changed local commit")
	}
	if got := checkoutGit(t, fixture.featurePath, "rev-parse", "HEAD"); got != projectCommit {
		t.Fatal("repeat checkout updated existing worktree")
	}
	contents, err := os.ReadFile(frontendFile)
	if err != nil || string(contents) != "tester uncommitted changes\n" {
		t.Fatalf("local edit was changed: %q, %v", contents, err)
	}
	contents, err = os.ReadFile(filepath.Join(fixture.featurePath, "unrelated", "requirement.txt"))
	if err != nil || string(contents) != "newly included module\n" {
		t.Fatalf("new module was not checked out: %q, %v", contents, err)
	}
}

func TestCheckoutFeature_ReusesUnoccupiedLocalBranchAtRemoteCommit(t *testing.T) {
	fixture := newCheckoutFixture(t)
	checkoutGit(t, fixture.config.Workspace, "fetch", "origin", "refs/heads/feature/checkout:refs/remotes/origin/feature/checkout")
	checkoutGit(t, fixture.config.Workspace, "branch", "--no-track", checkoutBranch, "origin/feature/checkout")
	if _, err := fixture.engine.CheckoutFeature(context.Background(), checkoutBranch); err != nil {
		t.Fatal(err)
	}
	if got := checkoutGit(t, fixture.featurePath, "rev-parse", "--abbrev-ref", "@{upstream}"); got != "origin/feature/checkout" {
		t.Fatalf("reused branch must track published feature, got %s", got)
	}
}

func TestCheckoutCLI_FeatureOnlyReturnsJSON(t *testing.T) {
	fixture := newCheckoutFixture(t)
	binary := filepath.Join(t.TempDir(), "modu")
	build := exec.Command("go", "build", "-o", binary, "../../cmd/modu")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	configPath := filepath.Join(t.TempDir(), ".modu.yaml")
	if err := config.SaveConfig(fixture.config, configPath); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "checkout", checkoutBranch, "-c", configPath, "-o", "json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("checkout CLI: %v\n%s\n%s", err, out, stderr.String())
	}
	var response struct {
		Success bool   `json:"success"`
		Action  string `json:"action"`
		Feature string `json:"feature"`
		Results []struct {
			Module string `json:"module"`
			Status string `json:"status"`
			Path   string `json:"path"`
			Commit string `json:"commit"`
		} `json:"results"`
	}
	if err := json.Unmarshal(out, &response); err != nil {
		t.Fatalf("stdout must be one JSON document: %v\n%s", err, out)
	}
	if !response.Success || response.Action != "checkout" || response.Feature != checkoutBranch || len(response.Results) != 4 {
		t.Fatalf("unexpected response: %s", out)
	}
	if response.Results[0].Path != fixture.featurePath || response.Results[0].Commit == "" || response.Results[3].Status != "skipped" {
		t.Fatalf("missing checkout details: %s", out)
	}

	cmd = exec.Command(binary, "checkout", "feature/not-published", "-c", configPath, "-o", "json")
	stderr.Reset()
	cmd.Stderr = &stderr
	out, err = cmd.Output()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("missing branch exit = %v", err)
	}
	if err := json.Unmarshal(out, &response); err != nil || response.Success {
		t.Fatalf("missing branch must return failed JSON response: %v\n%s", err, out)
	}

	cmd = exec.Command(binary, "checkout", checkoutBranch, "-c", configPath)
	out, err = cmd.Output()
	if err != nil || !strings.Contains(string(out), "已接手 feature") || !strings.Contains(string(out), "保留本地内容") {
		t.Fatalf("text response missing preserved checkout: %v\n%s", err, out)
	}
}

func TestCheckoutFeature_DiscoveryFailureCreatesNothing(t *testing.T) {
	for _, scenario := range []string{"missing-main-branch", "unavailable-module-remote", "uninitialized-module-directory"} {
		t.Run(scenario, func(t *testing.T) {
			fixture := newCheckoutFixture(t)
			switch scenario {
			case "missing-main-branch":
				checkoutGit(t, fixture.developers["project"], "push", "origin", "--delete", checkoutBranch)
			case "unavailable-module-remote":
				fixture.config.Modules[0].URL = filepath.Join(t.TempDir(), "missing.git")
			case "uninitialized-module-directory":
				if err := os.Mkdir(filepath.Join(fixture.config.Workspace, "backend"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			_, err := fixture.engine.CheckoutFeature(context.Background(), checkoutBranch)
			if err == nil {
				t.Fatal("expected discovery error")
			}
			if scenario == "missing-main-branch" && !errors.Is(err, errs.ErrFeatureNotFound) {
				t.Fatalf("missing branch error = %v", err)
			}
			if scenario == "unavailable-module-remote" && !errors.Is(err, errs.ErrGitExec) {
				t.Fatalf("query failure must not be treated as missing branch: %v", err)
			}
			for _, path := range []string{fixture.featurePath, filepath.Join(fixture.config.Workspace, "frontend")} {
				if _, err := os.Lstat(path); !os.IsNotExist(err) {
					t.Fatalf("discovery failure created %s: %v", path, err)
				}
			}
		})
	}
}

func TestCheckoutFeature_RejectsConflictingLocalState(t *testing.T) {
	for _, scenario := range []string{"different-local-commit", "occupied-branch", "directory-collision", "different-branch", "foreign-worktree"} {
		t.Run(scenario, func(t *testing.T) {
			fixture := newCheckoutFixture(t)
			workspace := fixture.config.Workspace
			expectedCommit := ""
			switch scenario {
			case "different-local-commit":
				checkoutGit(t, workspace, "branch", checkoutBranch, "main")
				expectedCommit = checkoutGit(t, workspace, "rev-parse", "refs/heads/feature/checkout")
			case "occupied-branch":
				checkoutGit(t, workspace, "fetch", "origin", "refs/heads/feature/checkout:refs/remotes/origin/feature/checkout")
				checkoutGit(t, workspace, "worktree", "add", "-b", checkoutBranch, filepath.Join(t.TempDir(), "occupied"), "origin/feature/checkout")
			case "directory-collision":
				if err := os.MkdirAll(fixture.featurePath, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(fixture.featurePath, "notes.txt"), []byte("keep this file\n"), 0644); err != nil {
					t.Fatal(err)
				}
			case "different-branch":
				checkoutGit(t, workspace, "worktree", "add", "-b", "feature-checkout", fixture.featurePath, "main")
			case "foreign-worktree":
				checkoutGit(t, workspace, "clone", "--branch", checkoutBranch, fixture.remotes["backend"], fixture.featurePath)
			}
			if _, err := fixture.engine.CheckoutFeature(context.Background(), checkoutBranch); !errors.Is(err, errs.ErrFeatureExists) {
				t.Fatalf("expected explicit conflict, got %v", err)
			}
			if expectedCommit != "" && checkoutGit(t, workspace, "rev-parse", "refs/heads/feature/checkout") != expectedCommit {
				t.Fatal("local branch was reset")
			}
			if scenario == "directory-collision" {
				contents, err := os.ReadFile(filepath.Join(fixture.featurePath, "notes.txt"))
				if err != nil || string(contents) != "keep this file\n" {
					t.Fatalf("directory content changed: %q, %v", contents, err)
				}
			}
		})
	}
}

func TestCheckoutFeature_UpdateFetchesLaterDeveloperCommits(t *testing.T) {
	fixture := newCheckoutFixture(t)
	// 已初始化模块使用实际 origin；受限 refspec 也必须能拉取需求分支。
	frontendSource := filepath.Join(fixture.config.Workspace, "frontend")
	checkoutGit(t, fixture.config.Workspace, "clone", "--single-branch", "--branch", "main", fixture.remotes["frontend"], frontendSource)
	fixture.config.Modules[1].URL = "unused-config-url"
	if _, err := fixture.engine.CheckoutFeature(context.Background(), checkoutBranch); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"project", "backend", "frontend"} {
		checkoutCommit(t, fixture.developers[name], "requirement.txt", "developer fix version 2\n")
		checkoutGit(t, fixture.developers[name], "push", "origin", checkoutBranch)
	}
	success, failed := fixture.engine.UpdateWorktree(context.Background(), checkoutBranch)
	if success != 3 || len(failed) != 0 {
		t.Fatalf("update: success=%d failed=%v", success, failed)
	}
	for _, path := range []string{fixture.featurePath, filepath.Join(fixture.featurePath, "backend"), filepath.Join(fixture.featurePath, "frontend")} {
		contents, err := os.ReadFile(filepath.Join(path, "requirement.txt"))
		if err != nil || string(contents) != "developer fix version 2\n" {
			t.Fatalf("update did not fetch developer fix: %q, %v", contents, err)
		}
	}
}

func TestCheckoutFeature_CancelledContextCreatesNothing(t *testing.T) {
	fixture := newCheckoutFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fixture.engine.CheckoutFeature(ctx, checkoutBranch); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if _, err := os.Lstat(fixture.featurePath); !os.IsNotExist(err) {
		t.Fatalf("cancelled checkout created worktree: %v", err)
	}
}

func TestCheckoutFeature_UsesRepositoryOriginContext(t *testing.T) {
	fixture := newCheckoutFixture(t)
	checkoutGit(t, fixture.config.Workspace, "remote", "set-url", "origin", "../project.git")
	if _, err := fixture.engine.CheckoutFeature(context.Background(), checkoutBranch); err != nil {
		t.Fatal(err)
	}
	wantCommit := checkoutGit(t, fixture.remotes["project"], "rev-parse", "refs/heads/feature/checkout")
	if got := checkoutGit(t, fixture.featurePath, "rev-parse", "HEAD"); got != wantCommit {
		t.Fatal("checkout did not use workspace's relative origin")
	}
}

type checkoutFixture struct {
	engine      *engine.Engine
	config      *config.Config
	featurePath string
	remotes     map[string]string
	developers  map[string]string
}

func newCheckoutFixture(t *testing.T) *checkoutFixture {
	t.Helper()
	// 隔离 pre-commit 的仓库环境及开发者全局 Git 配置。
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "GIT_") {
			t.Setenv(key, "")
			if err := os.Unsetenv(key); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	t.Setenv("GIT_AUTHOR_NAME", "modu test")
	t.Setenv("GIT_AUTHOR_EMAIL", "modu@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "modu test")
	t.Setenv("GIT_COMMITTER_EMAIL", "modu@example.invalid")
	root := t.TempDir()
	fixture := &checkoutFixture{remotes: make(map[string]string), developers: make(map[string]string)}
	for _, name := range []string{"project", "backend", "frontend", "unrelated"} {
		remote := filepath.Join(root, name+".git")
		seed := filepath.Join(root, "dev-"+name)
		checkoutGit(t, root, "init", "--bare", "--initial-branch=main", remote)
		checkoutGit(t, root, "init", "--initial-branch=main", seed)
		checkoutCommit(t, seed, ".gitignore", "backend/\nfrontend/\nunrelated/\n*.code-workspace\n")
		checkoutGit(t, seed, "remote", "add", "origin", remote)
		checkoutGit(t, seed, "push", "-u", "origin", "main")
		fixture.remotes[name] = remote
		fixture.developers[name] = seed
		if name != "unrelated" {
			checkoutGit(t, seed, "checkout", "-b", checkoutBranch)
			checkoutCommit(t, seed, "requirement.txt", "published requirement\n")
			checkoutGit(t, seed, "push", "-u", "origin", checkoutBranch)
		}
	}
	workspace := filepath.Join(root, "workspace")
	checkoutGit(t, root, "clone", "--single-branch", "--branch", "main", fixture.remotes["project"], workspace)
	// checkout 只依赖 origin，不能误触发其他远端。
	checkoutGit(t, workspace, "remote", "add", "unavailable", filepath.Join(root, "not-a-repository"))
	fixture.config = &config.Config{
		Workspace: workspace, WorktreeRoot: filepath.Join(root, "worktrees"),
		DefaultBase: "nonexistent-base", Concurrency: 2,
		DefaultSelectedModules: []string{"unrelated"},
	}
	for _, name := range []string{"backend", "frontend", "unrelated"} {
		fixture.config.Modules = append(fixture.config.Modules, config.Module{Name: name, URL: fixture.remotes[name]})
	}
	fixture.featurePath = filepath.Join(fixture.config.WorktreeRoot, "feature-checkout")
	fixture.engine = engine.New(fixture.config)
	return fixture
}

func checkoutGit(t *testing.T, directory string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", directory}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func checkoutCommit(t *testing.T, repository, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repository, name), []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	checkoutGit(t, repository, "add", name)
	checkoutGit(t, repository, "commit", "-m", "test fixture update")
}
