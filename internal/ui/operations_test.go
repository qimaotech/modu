package ui

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/qimaotech/modu/internal/config"
	"github.com/qimaotech/modu/internal/core"
	"github.com/qimaotech/modu/internal/engine"
	"github.com/qimaotech/modu/internal/gitproxy"
)

type blockedRemoteClient struct {
	gitproxy.GitClient
	started chan struct{}
}

func (client *blockedRemoteClient) RemoteBranchExists(ctx context.Context, url, branch string) bool {
	close(client.started)
	<-ctx.Done()
	return false
}

func TestApp_RemoteQueryAnimatesAndCancels(t *testing.T) {
	client := &blockedRemoteClient{GitClient: &uiFakeGitClient{}, started: make(chan struct{})}
	app := &App{Engine: engine.NewWithClient(&config.Config{Modules: []config.Module{{Name: "module", URL: "fixture"}}, Concurrency: 1}, client), state: "create_input", createFeatureInput: []rune("feature-a"), statusesPending: true}
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil || app.state != "loading" {
		t.Fatal("remote query should run in the background")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatal("expected background command and animation tick")
	}
	completed := make(chan tea.Msg, 1)
	go func() { completed <- batch[0]() }()
	select {
	case <-client.started:
	case <-time.After(5 * time.Second):
		t.Fatal("query did not start")
	}
	before := app.View()
	_, tick := app.Update(operationTickMsg(app.operationID))
	if tick == nil || app.View() == before {
		t.Fatal("animation did not advance during blocked query")
	}
	app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	select {
	case result := <-completed:
		_, refresh := app.Update(result)
		if refresh == nil {
			t.Fatal("interrupted initial status refresh was not resumed")
		}
		app.Update(refresh())
	case <-time.After(5 * time.Second):
		t.Fatal("Ctrl+C did not cancel query")
	}
	if app.state != "list" || app.message != "操作已取消" {
		t.Fatalf("state=%s message=%s", app.state, app.message)
	}
	if app.statusesPending || app.statusLoading {
		t.Fatal("status refresh remained pending after cancellation")
	}
}

type partialModuleClient struct {
	gitproxy.GitClient
	started chan struct{}
}

func (client *partialModuleClient) CreateWorktree(ctx context.Context, repoPath, branch, base, path string) error {
	if filepath.Base(path) == "first" {
		return os.MkdirAll(path, 0755)
	}
	close(client.started)
	<-ctx.Done()
	return ctx.Err()
}

func TestApp_ModuleCancellationRefreshesCompletedChanges(t *testing.T) {
	root := t.TempDir()
	featurePath := filepath.Join(root, "features", "feature")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatal(err)
	}
	modules := []config.Module{{Name: "first"}, {Name: "second"}}
	client := &partialModuleClient{GitClient: &uiFakeGitClient{}, started: make(chan struct{})}
	cfg := &config.Config{Workspace: filepath.Join(root, "workspace"), WorktreeRoot: filepath.Join(root, "features"), Modules: modules}
	for _, module := range modules {
		if err := os.MkdirAll(filepath.Join(cfg.Workspace, module.Name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	app := &App{Engine: engine.NewWithClient(cfg, client), state: "modules", modulesFeature: "feature", Envs: []core.WorktreeEnv{{Name: "feature", DirName: "feature"}}, moduleSelector: NewModuleSelector(modules, nil, nil, []string{"first", "second"}, "modules")}
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	batch := cmd().(tea.BatchMsg)
	completed := make(chan tea.Msg, 1)
	go func() { completed <- batch[0]() }()
	select {
	case <-client.started:
	case <-time.After(5 * time.Second):
		t.Fatal("second module did not start")
	}
	app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	select {
	case message := <-completed:
		_, refresh := app.Update(message)
		if refresh == nil {
			t.Fatal("cancelled module change did not refresh the environment")
		}
		app.Update(refresh())
	case <-time.After(5 * time.Second):
		t.Fatal("module operation did not cancel")
	}
	if len(app.Envs[0].Modules) != 1 || app.Envs[0].Modules[0].Name != "first" {
		t.Fatalf("completed module missing from list: %+v", app.Envs[0])
	}
}
