package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/qimaotech/modu/internal/progress"
)

var progressFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// RenderProgress 同时用于 CLI 动态行和 TUI，仓库数量与 Git 阶段百分比分开展示。
func RenderProgress(snapshot progress.Snapshot, frame, width, maxRows int) string {
	now := time.Now()
	done, failed, running := 0, 0, 0
	for _, item := range snapshot.Items {
		switch item.State {
		case progress.Running:
			running++
		case progress.Succeeded, progress.Skipped, progress.Cancelled:
			done++
		case progress.Failed:
			done++
			failed++
		}
	}
	elapsed := now.Sub(snapshot.Started)
	header := fmt.Sprintf("%s · 完成 %d/%d · 运行中 %d · 失败 %d · 已耗时 %s", snapshot.Action, done, len(snapshot.Items), running, failed, elapsed.Round(time.Second))
	if len(snapshot.Items) == 0 {
		header = fmt.Sprintf("%s %s · 已耗时 %s", progressFrames[frame%len(progressFrames)], snapshot.Action, elapsed.Round(time.Second))
	}
	lines := []string{header}
	// 优先显示仍在运行和失败的仓库，避免大工作区将活动任务挤出终端。
	items := make([]progress.Item, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		if item.State == progress.Running || item.State == progress.Failed {
			items = append(items, item)
		}
	}
	for _, item := range snapshot.Items {
		if item.State != progress.Running && item.State != progress.Failed {
			items = append(items, item)
		}
	}
	for i, item := range items {
		if maxRows > 0 && i >= maxRows {
			lines = append(lines, fmt.Sprintf("  另有 %d 个仓库", len(items)-i))
			break
		}
		icon := "·"
		switch item.State {
		case progress.Running:
			icon = progressFrames[frame%len(progressFrames)]
		case progress.Succeeded:
			icon = "✓"
		case progress.Failed:
			icon = "✗"
		case progress.Cancelled:
			icon = "■"
		case progress.Skipped:
			icon = "–"
		}
		detail := item.Stage
		if item.Detail != "" {
			detail += " · " + item.Detail
		}
		if item.Error != "" {
			detail += " · " + strings.ReplaceAll(item.Error, "\n", " ")
		}
		lines = append(lines, fmt.Sprintf("%s %-16s %s  %.1fs", icon, item.Module, detail, item.Elapsed(now).Seconds()))
	}
	if len(snapshot.Items) == 0 || done < len(snapshot.Items) {
		lines = append(lines, "Ctrl+C 取消")
	}
	if width > 0 {
		for i := range lines {
			lines[i] = ansi.Truncate(lines[i], width, "…")
		}
	}
	return strings.Join(lines, "\n")
}

type Progress struct {
	Tracker     *progress.Tracker
	writer      io.Writer
	interactive bool
	width       int
	rows        int
	mu          sync.Mutex
	last        map[string]string
	lines       int
	stop        chan struct{}
	done        chan struct{}
	closeOnce   sync.Once
}

func NewProgress(action string, writer io.Writer, interactive bool, width, rows int) *Progress {
	p := &Progress{Tracker: progress.New(action), writer: writer, interactive: interactive, width: width, rows: rows, last: make(map[string]string), stop: make(chan struct{}), done: make(chan struct{})}
	fmt.Fprintf(writer, "%s…\n", action)
	p.lines = 1
	if interactive {
		go func() {
			defer close(p.done)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			frame := 0
			for {
				select {
				case <-p.stop:
					p.render(frame)
					return
				case <-ticker.C:
					p.render(frame)
					frame++
				}
			}
		}()
	} else {
		close(p.done)
	}
	return p
}

func (p *Progress) Context(ctx context.Context) context.Context {
	return progress.WithReporter(ctx, func(event progress.Event) {
		p.Tracker.Record(event)
		if p.interactive || event.State == progress.Queued {
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		key := string(event.State) + "/" + event.Stage
		if p.last[event.Module] == key {
			return
		}
		p.last[event.Module] = key
		fmt.Fprintf(p.writer, "[%s] %s\n", event.Module, event.Stage)
		if event.Error != "" {
			fmt.Fprintf(p.writer, "  %s\n", event.Error)
		}
	})
}

func (p *Progress) render(frame int) {
	view := RenderProgress(p.Tracker.Snapshot(), frame, p.width, p.rows)
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(p.writer, "\x1b[%dA\r\x1b[J%s\n", p.lines, view)
	p.lines = strings.Count(view, "\n") + 1
}

func (p *Progress) Close() {
	p.closeOnce.Do(func() { close(p.stop); <-p.done })
}

// FormatOperationResponse 给 init/update 提供单个完整 JSON 文档。
func FormatOperationResponse(action, feature string, snapshot progress.Snapshot, operationErr error) string {
	type result struct {
		Module   string         `json:"module"`
		Status   progress.State `json:"status"`
		Duration float64        `json:"durationSeconds"`
		Error    string         `json:"error,omitempty"`
	}
	response := struct {
		Success bool     `json:"success"`
		Action  string   `json:"action"`
		Feature string   `json:"feature,omitempty"`
		Results []result `json:"results"`
		Error   string   `json:"error,omitempty"`
	}{Success: operationErr == nil, Action: action, Feature: feature, Results: []result{}}
	for _, item := range snapshot.Items {
		response.Results = append(response.Results, result{Module: item.Module, Status: item.State, Duration: item.Elapsed(time.Now()).Seconds(), Error: item.Error})
	}
	if operationErr != nil {
		response.Error = operationErr.Error()
	}
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"success\":false,\"error\":%q}\n", err.Error())
	}
	return string(data) + "\n"
}
