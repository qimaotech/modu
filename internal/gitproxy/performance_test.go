package gitproxy

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/qimaotech/modu/internal/progress"
)

func TestClone_PartialClonePreservesHistoryAndCheckout(t *testing.T) {
	source := setupPushStatusRepo(t, "main")
	writeCommit(t, source, "README.md", "second version", "second commit")
	runGit(t, source, "config", "uploadpack.allowFilter", "true")
	destination := filepath.Join(t.TempDir(), "clone")
	sourceURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(source)}).String()
	if err := New().Clone(context.Background(), sourceURL, destination, CloneOptions{Filter: "blob:none"}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(destination, "README.md"))
	if err != nil || string(content) != "second version" {
		t.Fatalf("checkout content=%q error=%v", content, err)
	}
	for _, check := range []struct {
		args []string
		want string
	}{
		{[]string{"rev-list", "--count", "HEAD"}, "2"},
		{[]string{"config", "--get", "remote.origin.partialclonefilter"}, "blob:none"},
		{[]string{"show", "HEAD~1:README.md"}, "initial"},
	} {
		args := append([]string{"-C", destination}, check.args...)
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil || strings.TrimSpace(string(out)) != check.want {
			t.Fatalf("git %v: %q, %v", check.args, out, err)
		}
	}
}

func TestUpdate_IgnoresUnavailableUnrelatedRemote(t *testing.T) {
	repo := setupPushStatusRepo(t, "main")
	runGit(t, repo, "remote", "add", "unused", filepath.Join(t.TempDir(), "missing.git"))
	client := New()
	if err := client.FetchAndSwitchBranch(context.Background(), repo, "main"); err != nil {
		t.Fatal(err)
	}
	if err := client.Rebase(context.Background(), repo); err != nil {
		t.Fatal(err)
	}
}

func TestGetStatus_OneGitInvocationPreservesFileNames(t *testing.T) {
	repo := setupPushStatusRepo(t, "main")
	name := "space and\nnewline.txt"
	if err := os.WriteFile(filepath.Join(repo, name), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("modified"), 0644); err != nil {
		t.Fatal(err)
	}
	trace := filepath.Join(t.TempDir(), "trace.jsonl")
	t.Setenv("GIT_TRACE2_EVENT", trace)
	status, err := New().GetStatus(context.Background(), repo)
	if err != nil {
		t.Fatal(err)
	}
	if !status.IsDirty || status.Branch != "main" || len(status.Files) != 2 {
		t.Fatalf("unexpected status: %+v", status)
	}
	found := false
	for _, file := range status.Files {
		if file.Name == name {
			found = true
		}
	}
	if !found {
		t.Fatalf("filename was not preserved: %+v", status.Files)
	}
	data, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	starts := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var event struct {
			Event string `json:"event"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if event.Event == "start" {
			starts++
		}
	}
	if starts != 1 {
		t.Fatalf("status launched %d Git processes, want one", starts)
	}
}

func TestClone_CancellationReportsProgressAndCleansTemporaryDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture uses a POSIX shell")
	}
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nprintf 'Receiving objects: 10%%\\r' >&2\nexec sleep 30\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	received := make(chan struct{}, 1)
	ctx = progress.WithReporter(progress.ForModule(ctx, "repo"), func(event progress.Event) {
		if strings.Contains(event.Detail, "10%") {
			select {
			case received <- struct{}{}:
			default:
			}
		}
	})
	destination := filepath.Join(root, "repo")
	completed := make(chan error, 1)
	go func() { completed <- New().Clone(ctx, "local-fixture", destination, CloneOptions{}) }()
	select {
	case <-received:
	case <-time.After(5 * time.Second):
		t.Fatal("clone did not stream progress while blocked")
	}
	cancel()
	select {
	case err := <-completed:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v, want cancellation", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("clone did not stop")
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("unfinished destination remains: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".modu-clone-") {
			t.Fatal("temporary clone was not cleaned")
		}
	}
}

func TestRemoveWorktree_CancelledContextPreservesFiles(t *testing.T) {
	for _, withBranch := range []bool{false, true} {
		t.Run(map[bool]string{false: "worktree", true: "worktree_and_branch"}[withBranch], func(t *testing.T) {
			path := t.TempDir()
			file := filepath.Join(path, "keep.txt")
			if err := os.WriteFile(file, []byte("keep"), 0644); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			client := New()
			var err error
			if withBranch {
				err = client.RemoveWorktreeAndBranch(ctx, path, path, "feature")
			} else {
				err = client.RemoveWorktree(ctx, path)
			}
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("got %v", err)
			}
			if _, err := os.Stat(file); err != nil {
				t.Fatalf("cancelled operation removed files: %v", err)
			}
		})
	}
}
