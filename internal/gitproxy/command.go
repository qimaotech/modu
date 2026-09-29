package gitproxy

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/qimaotech/modu/internal/progress"
)

const commandOutputLimit = 64 * 1024

// commandOutput 持续转发 Git 的 CR/LF 进度，并仅保留有界的错误上下文。
type commandOutput struct {
	mu      sync.Mutex
	ctx     context.Context
	stage   string
	tail    []byte
	pending []byte
}

func (output *commandOutput) Write(data []byte) (int, error) {
	output.mu.Lock()
	defer output.mu.Unlock()
	output.tail = append(output.tail, data...)
	if len(output.tail) > commandOutputLimit {
		output.tail = append([]byte(nil), output.tail[len(output.tail)-commandOutputLimit:]...)
	}
	for _, value := range data {
		if value == '\r' || value == '\n' {
			if len(output.pending) > 0 {
				detail := strings.TrimSpace(strings.Map(func(char rune) rune {
					if unicode.IsControl(char) {
						return -1
					}
					return char
				}, string(output.pending)))
				progress.Emit(output.ctx, progress.Event{State: progress.Running, Stage: output.stage, Detail: detail})
				output.pending = output.pending[:0]
			}
		} else if len(output.pending) < 4096 {
			output.pending = append(output.pending, value)
		}
	}
	return len(data), nil
}

func gitCommand(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.WaitDelay = 2 * time.Second
	configureCancellation(cmd)
	return cmd
}

func runProgress(ctx context.Context, stage string, args ...string) ([]byte, error) {
	progress.Emit(ctx, progress.Event{State: progress.Running, Stage: stage})
	cmd := gitCommand(ctx, args...)
	output := &commandOutput{ctx: ctx, stage: stage}
	cmd.Stdout, cmd.Stderr = output, output
	err := cmd.Run()
	return output.tail, err
}
