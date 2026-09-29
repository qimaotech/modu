package engine

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/qimaotech/modu/internal/progress"
	"golang.org/x/sync/errgroup"
)

type repositoryJob struct {
	name  string
	stage string
	skip  bool
	run   func(context.Context) error
}

func (e *Engine) concurrency() int {
	if e.Config.Concurrency <= 0 {
		return 5
	}
	return e.Config.Concurrency
}

// 主仓库目录名可能与模块同名，汇总和进度中的名称必须可区分。
func (e *Engine) mainRepositoryName() string {
	name := filepath.Base(e.Config.Workspace)
	moduleNames := e.configuredModuleNames()
	for moduleNames[name] {
		name += "（主项目）"
	}
	return name
}

// runRepositories 保留每个仓库的独立结果；取消后不再启动排队中的 Git 操作。
func (e *Engine) runRepositories(ctx context.Context, jobs []repositoryJob) (int, map[string]error) {
	failed := make(map[string]error)
	success := 0
	var mu sync.Mutex
	var group errgroup.Group
	group.SetLimit(e.concurrency())
	for _, job := range jobs {
		progress.Emit(ctx, progress.Event{Module: job.name, Stage: "排队", State: progress.Queued})
	}
	for _, job := range jobs {
		group.Go(func() error {
			moduleCtx := progress.ForModule(ctx, job.name)
			err := ctx.Err()
			if err == nil {
				progress.Emit(moduleCtx, progress.Event{Stage: job.stage, State: progress.Running})
				err = job.run(moduleCtx)
			}
			if err == nil && job.skip {
				progress.Emit(moduleCtx, progress.Event{State: progress.Skipped, Stage: "已存在，跳过"})
			} else {
				progress.Finish(moduleCtx, err)
			}
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failed[job.name] = err
			} else {
				success++
			}
			return nil
		})
	}
	_ = group.Wait()
	return success, failed
}
