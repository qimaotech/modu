// Package progress 传递仓库操作进度，不依赖具体终端或 UI。
package progress

import (
	"context"
	"errors"
	"sync"
	"time"
)

type State string

const (
	Queued    State = "queued"
	Running   State = "running"
	Succeeded State = "success"
	Failed    State = "failed"
	Skipped   State = "skipped"
	Cancelled State = "cancelled"
)

type Event struct {
	Module string
	Stage  string
	Detail string
	State  State
	Error  string
}

type reporterKey struct{}
type moduleKey struct{}

func WithReporter(ctx context.Context, reporter func(Event)) context.Context {
	return context.WithValue(ctx, reporterKey{}, reporter)
}

func ForModule(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, moduleKey{}, name)
}

func Emit(ctx context.Context, event Event) {
	reporter, _ := ctx.Value(reporterKey{}).(func(Event))
	if reporter == nil {
		return
	}
	if event.Module == "" {
		event.Module, _ = ctx.Value(moduleKey{}).(string)
	}
	if event.Module != "" {
		reporter(event)
	}
}

func Finish(ctx context.Context, err error) {
	event := Event{State: Succeeded, Stage: "完成"}
	if err != nil {
		event.State, event.Stage, event.Error = Failed, "失败", err.Error()
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			event.State, event.Stage = Cancelled, "已取消"
		}
	}
	Emit(ctx, event)
}

type Item struct {
	Event
	Started  time.Time
	Finished time.Time
}

func (item Item) Elapsed(now time.Time) time.Duration {
	if item.Started.IsZero() {
		return 0
	}
	if !item.Finished.IsZero() {
		now = item.Finished
	}
	return now.Sub(item.Started)
}

type Snapshot struct {
	Action  string
	Started time.Time
	Items   []Item
}

// Tracker 可供 Git 工作协程写入，终端/UI 通过快照读取。
type Tracker struct {
	mu      sync.Mutex
	action  string
	started time.Time
	names   []string
	items   map[string]Item
}

func New(action string) *Tracker {
	return &Tracker{action: action, started: time.Now(), items: make(map[string]Item)}
}

func (tracker *Tracker) Record(event Event) {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	item, exists := tracker.items[event.Module]
	if !exists {
		tracker.names = append(tracker.names, event.Module)
	}
	if item.Started.IsZero() && event.State == Running {
		item.Started = time.Now()
	}
	if event.State != Queued && event.State != Running {
		item.Finished = time.Now()
	}
	item.Event = event
	tracker.items[event.Module] = item
}

func (tracker *Tracker) Snapshot() Snapshot {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	snapshot := Snapshot{Action: tracker.action, Started: tracker.started, Items: make([]Item, 0, len(tracker.names))}
	for _, name := range tracker.names {
		snapshot.Items = append(snapshot.Items, tracker.items[name])
	}
	return snapshot
}
