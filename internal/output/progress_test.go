package output

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/qimaotech/modu/internal/progress"
)

func TestProgress_NonTerminalLogsStagesWithoutPercentageSpam(t *testing.T) {
	var logs bytes.Buffer
	display := NewProgress("正在更新", &logs, false, 80, 8)
	ctx := progress.ForModule(display.Context(context.Background()), "backend")
	for range 100 {
		progress.Emit(ctx, progress.Event{State: progress.Running, Stage: "同步 origin", Detail: "Receiving objects: 10%"})
	}
	progress.Emit(ctx, progress.Event{State: progress.Running, Stage: "rebase"})
	failure := errors.New("conflict in config.go")
	progress.Finish(ctx, failure)
	display.Close()
	text := logs.String()
	if strings.Contains(text, "\x1b") || strings.Contains(text, "10%") || strings.Count(text, "同步 origin") != 1 || !strings.Contains(text, failure.Error()) {
		t.Fatalf("unexpected stage log: %q", text)
	}
	var response struct {
		Success bool                                     `json:"success"`
		Results []struct{ Module, Status, Error string } `json:"results"`
	}
	if err := json.Unmarshal([]byte(FormatOperationResponse("update", "feature/a", display.Tracker.Snapshot(), failure)), &response); err != nil {
		t.Fatal(err)
	}
	if response.Success || len(response.Results) != 1 || response.Results[0].Status != "failed" || response.Results[0].Error != failure.Error() {
		t.Fatalf("failure details lost in JSON: %+v", response)
	}
}

func TestProgress_TerminalClosesAnimationBeforeFinalResult(t *testing.T) {
	var terminal bytes.Buffer
	display := NewProgress("正在初始化", &terminal, true, 80, 8)
	ctx := progress.ForModule(display.Context(context.Background()), "frontend")
	progress.Emit(ctx, progress.Event{State: progress.Running, Stage: "克隆"})
	progress.Finish(ctx, nil)
	display.Close()
	display.Close()
	terminal.WriteString("final result\n")
	text := terminal.String()
	if !strings.Contains(text, "完成 1/1") || !strings.Contains(text, "frontend") || strings.Contains(text, "Ctrl+C") || !strings.HasSuffix(text, "final result\n") {
		t.Fatalf("unexpected completed terminal: %q", text)
	}
}
