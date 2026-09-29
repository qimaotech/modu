package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qimaotech/modu/internal/config"
	"github.com/qimaotech/modu/internal/core"
	"github.com/qimaotech/modu/internal/gitproxy"
	"github.com/qimaotech/modu/internal/progress"
)

func TestUpdate_MultipleFailures(t *testing.T) {
	for _, feature := range []string{"", "feature-a"} {
		t.Run(feature, func(t *testing.T) {
			workspace := filepath.Join(t.TempDir(), "workspace")
			worktreeRoot := t.TempDir()
			cfg := &config.Config{Workspace: workspace, WorktreeRoot: worktreeRoot, DefaultBase: "main", Concurrency: 4}
			for _, name := range []string{"one", "two", "three"} {
				cfg.Modules = append(cfg.Modules, config.Module{Name: name})
				for _, root := range []string{workspace, filepath.Join(worktreeRoot, "feature-a")} {
					if err := os.MkdirAll(filepath.Join(root, name), 0755); err != nil {
						t.Fatal(err)
					}
				}
			}
			var started sync.WaitGroup
			started.Add(4)
			failure := errors.New("remote unavailable")
			fail := func(context.Context, string) error {
				started.Done()
				started.Wait()
				return failure
			}
			client := &MockGitClient{RebaseFunc: fail, FetchAndSwitchBranchFunc: func(ctx context.Context, path, branch string) error {
				return fail(ctx, path)
			}}
			eng := NewWithClient(cfg, client)
			var success int
			var failed map[string]error
			if feature == "" {
				success, failed = eng.UpdateMainProject(context.Background())
			} else {
				success, failed = eng.UpdateWorktree(context.Background(), feature)
			}
			if success != 0 || len(failed) != 4 {
				t.Fatalf("success=%d failures=%v", success, failed)
			}
			for _, name := range []string{"workspace", "one", "two", "three"} {
				if !errors.Is(failed[name], failure) {
					t.Errorf("%s: got %v, want remote failure", name, failed[name])
				}
			}
		})
	}
}

func TestUpdate_MainAndModuleWithSameNameKeepBothResults(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "repo")
	worktreeRoot := t.TempDir()
	for _, path := range []string{filepath.Join(workspace, "repo"), filepath.Join(worktreeRoot, "feature", "repo")} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, feature := range []string{"", "feature"} {
		for _, fail := range []bool{false, true} {
			failure := errors.New("network unavailable")
			update := func(context.Context, string) error {
				if fail {
					return failure
				}
				return nil
			}
			client := &MockGitClient{RebaseFunc: update, FetchAndSwitchBranchFunc: func(ctx context.Context, path, branch string) error { return update(ctx, path) }}
			eng := NewWithClient(&config.Config{Workspace: workspace, WorktreeRoot: worktreeRoot, Concurrency: 2, Modules: []config.Module{{Name: "repo"}}}, client)
			tracker := progress.New("更新")
			ctx := progress.WithReporter(context.Background(), tracker.Record)
			var count int
			var failed map[string]error
			if feature == "" {
				count, failed = eng.UpdateMainProject(ctx)
			} else {
				count, failed = eng.UpdateWorktree(ctx, feature)
			}
			if len(tracker.Snapshot().Items) != 2 || count+len(failed) != 2 {
				t.Fatalf("same-name repositories lost: success=%d failures=%v progress=%+v", count, failed, tracker.Snapshot())
			}
			if fail && len(failed) != 2 {
				t.Fatalf("lost failure: %v", failed)
			}
		}
	}
}

func TestUpdate_CancelStopsQueuedRepositories(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspace")
	cfg := &config.Config{Workspace: workspace, Concurrency: 1, DefaultBase: "main"}
	for _, name := range []string{"one", "two"} {
		cfg.Modules = append(cfg.Modules, config.Module{Name: name})
		if err := os.MkdirAll(filepath.Join(workspace, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	var calls atomic.Int32
	client := &MockGitClient{FetchAndSwitchBranchFunc: func(ctx context.Context, path, branch string) error {
		calls.Add(1)
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}}
	tracker := progress.New("更新")
	ctx = progress.WithReporter(ctx, tracker.Record)
	completed := make(chan map[string]error, 1)
	go func() { _, failed := NewWithClient(cfg, client).UpdateMainProject(ctx); completed <- failed }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("update did not start")
	}
	snapshot := tracker.Snapshot()
	if len(snapshot.Items) != 3 || snapshot.Items[0].State != progress.Running {
		t.Fatalf("progress missing during blocked update: %+v", snapshot)
	}
	cancel()
	select {
	case failed := <-completed:
		if len(failed) != 3 || calls.Load() != 1 {
			t.Fatalf("failures=%v git calls=%d", failed, calls.Load())
		}
		for _, err := range failed {
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("lost cancellation: %v", err)
			}
		}
	case <-time.After(5 * time.Second):
		t.Fatal("update did not stop")
	}
}

func TestListWorktreeSummaries_DoesNotScanFileStatus(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspace")
	worktreeRoot := t.TempDir()
	for _, feature := range []string{"alpha", "beta"} {
		if err := os.MkdirAll(filepath.Join(worktreeRoot, feature, "module"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	var registryCalls, statusCalls atomic.Int32
	client := &MockGitClient{
		ListWorktreesFunc: func(ctx context.Context, path string) ([]gitproxy.WorktreeInfo, error) {
			registryCalls.Add(1)
			suffix := ""
			if filepath.Base(path) == "module" {
				suffix = "module"
			}
			return []gitproxy.WorktreeInfo{{Path: path, Branch: "main"}, {Path: filepath.Join(worktreeRoot, "alpha", suffix), Branch: "feature/alpha"}, {Path: filepath.Join(worktreeRoot, "beta", suffix), Branch: "feature/beta"}}, nil
		},
		GetStatusFunc: func(context.Context, string) (gitproxy.Status, error) {
			statusCalls.Add(1)
			return gitproxy.Status{}, errors.New("expensive status must not run")
		},
	}
	eng := NewWithClient(&config.Config{Workspace: workspace, WorktreeRoot: worktreeRoot, Modules: []config.Module{{Name: "module"}}, Concurrency: 2}, client)
	summary, err := eng.ListWorkspaceSummaries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	envs := summary.Envs
	if summary.MainProject == nil || summary.MainProject.Branch != "main" || len(summary.Modules) != 1 || summary.Modules[0].Branch != "main" {
		t.Fatalf("workspace metadata missing: %+v", summary)
	}
	if len(envs) != 2 || envs[0].Branch != "feature/alpha" || len(envs[1].Modules) != 1 || envs[1].Modules[0].Branch != "feature/beta" {
		t.Fatalf("wrong worktree metadata: %+v", envs)
	}
	if registryCalls.Load() != 2 || statusCalls.Load() != 0 {
		t.Fatalf("registry=%d status=%d", registryCalls.Load(), statusCalls.Load())
	}
}

func TestRefreshWorktreeStatuses_BoundedParallelismAndSnapshotIsolation(t *testing.T) {
	var active, maximum atomic.Int32
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	client := &MockGitClient{GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
		current := active.Add(1)
		defer active.Add(-1)
		for old := maximum.Load(); current > old; old = maximum.Load() {
			if maximum.CompareAndSwap(old, current) {
				break
			}
		}
		select {
		case started <- struct{}{}:
		default:
		}
		<-release
		return gitproxy.Status{Branch: "feature/new", IsDirty: true}, nil
	}}
	eng := NewWithClient(&config.Config{Concurrency: 2}, client)
	original := []core.WorktreeEnv{{MainProject: &core.ModuleStatus{Path: "main", Branch: "old"}, Modules: []core.ModuleStatus{{Path: "one"}, {Path: "two"}}}}
	completed := make(chan []core.WorktreeEnv, 1)
	go func() { result, _ := eng.RefreshWorktreeStatuses(context.Background(), original); completed <- result }()
	for range 2 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			close(release)
			t.Fatal("status queries did not run concurrently")
		}
	}
	close(release)
	result := <-completed
	if maximum.Load() != 2 || !result[0].MainProject.IsDirty || result[0].Modules[1].Branch != "feature/new" {
		t.Fatalf("parallel refresh result=%+v maximum=%d", result, maximum.Load())
	}
	if original[0].MainProject.Branch != "old" || original[0].Modules[1].Branch != "" || original[0].MainProject.IsDirty {
		t.Fatal("background refresh changed the displayed snapshot")
	}
}

func TestDeleteWorktree_CancelledContextPreservesEnvironment(t *testing.T) {
	eng, _, removed := newDeletePreflightTestEngine(t, gitproxy.BranchPushStatus{IsPushed: true})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := eng.DeleteWorktree(ctx, "feat-a", false); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if *removed != 0 {
		t.Fatalf("removed %d worktrees after cancellation", *removed)
	}
	if _, err := os.Stat(filepath.Join(eng.Config.WorktreeRoot, "feat-a", "module1")); err != nil {
		t.Fatal(err)
	}
}

func TestCreateWorktree_FailurePreservesExistingEnvironment(t *testing.T) {
	root := t.TempDir()
	featurePath := filepath.Join(root, "feature-a")
	existing := filepath.Join(featurePath, "existing")
	if err := os.MkdirAll(existing, 0755); err != nil {
		t.Fatal(err)
	}
	var removed atomic.Int32
	client := &MockGitClient{
		BranchExistsFunc:            func(context.Context, string, string) bool { return false },
		CreateWorktreeFunc:          func(context.Context, string, string, string, string) error { return errors.New("checkout failed") },
		RemoveWorktreeAndBranchFunc: func(context.Context, string, string, string) error { removed.Add(1); return nil },
	}
	eng := NewWithClient(&config.Config{Workspace: t.TempDir(), WorktreeRoot: root, Concurrency: 2, Modules: []config.Module{{Name: "existing"}, {Name: "new"}}}, client)
	if err := eng.CreateWorktree(context.Background(), "feature-a", "main"); err == nil {
		t.Fatal("expected checkout failure")
	}
	if removed.Load() != 0 {
		t.Fatalf("rollback touched %d preexisting worktrees", removed.Load())
	}
	if _, err := os.Stat(existing); err != nil {
		t.Fatalf("existing worktree removed: %v", err)
	}
}
