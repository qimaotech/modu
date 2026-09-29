package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/qimaotech/modu/internal/core"
	errs "github.com/qimaotech/modu/internal/errors"
	"github.com/qimaotech/modu/internal/gitproxy"
	"github.com/qimaotech/modu/internal/progress"
)

type checkoutRepository struct {
	name    string
	source  string
	target  string
	url     string
	missing bool
	matched bool
}

// CheckoutFeature 先完整发现远程分支，再接手主项目及命中模块；失败时保留已完成项供重试。
func (e *Engine) CheckoutFeature(ctx context.Context, feature string) ([]core.CheckoutResult, error) {
	if err := gitproxy.ValidateBranchName(ctx, feature); err != nil {
		return nil, err
	}
	featurePath := filepath.Join(e.Config.WorktreeRoot, featureToDirName(feature))
	repositories := []checkoutRepository{{name: e.mainRepositoryName(), source: e.Config.Workspace, target: featurePath}}
	seen := make(map[string]bool)
	for _, module := range e.Config.Modules {
		if module.Name == "" || module.Name == "." || module.Name == ".." || module.Name == ".git" || filepath.Base(module.Name) != module.Name || seen[module.Name] {
			return nil, fmt.Errorf("模块名无效或重复 %q: %w", module.Name, errs.ErrConfigInvalid)
		}
		seen[module.Name] = true
		repositories = append(repositories, checkoutRepository{name: module.Name, source: filepath.Join(e.Config.Workspace, module.Name), target: filepath.Join(featurePath, module.Name), url: module.URL})
	}
	results := make([]core.CheckoutResult, len(repositories))
	queryJobs := make([]repositoryJob, 0, len(repositories))
	for index := range repositories {
		repository := &repositories[index]
		results[index] = core.CheckoutResult{Module: repository.name, Status: string(progress.Skipped), Message: "未执行"}
		queryJobs = append(queryJobs, repositoryJob{name: repository.name, stage: "查询远程分支", run: func(ctx context.Context) error {
			queryPath := repository.source
			_, err := os.Lstat(repository.source)
			if os.IsNotExist(err) && index != 0 {
				repository.missing = true
				queryPath = ""
			} else if err != nil {
				return err
			}
			repository.matched, err = e.GitProxy.QueryRemoteBranch(ctx, queryPath, repository.url, feature)
			if err != nil {
				return err
			}
			if !repository.matched {
				if index == 0 {
					return fmt.Errorf("主项目 origin 不存在分支 %s: %w", feature, errs.ErrFeatureNotFound)
				}
				results[index].Message = "远程无同名分支"
			}
			return nil
		}})
	}
	_, failed := e.runRepositories(ctx, queryJobs)
	if err := checkoutFailures(results, failed); err != nil {
		return results, err
	}

	createJob := func(index int) repositoryJob {
		repository := &repositories[index]
		return repositoryJob{name: repository.name, stage: "接手 worktree", run: func(ctx context.Context) error {
			if repository.missing {
				if err := e.GitProxy.Clone(ctx, repository.url, repository.source, gitproxy.CloneOptions{}); err != nil {
					return err
				}
			}
			checkout, err := e.GitProxy.CheckoutWorktree(ctx, repository.source, feature, repository.target)
			if err != nil {
				return err
			}
			results[index] = core.CheckoutResult{Module: repository.name, Status: string(progress.Succeeded), Path: repository.target, Branch: feature, Commit: checkout.Commit, Message: "已创建"}
			if checkout.Existing {
				results[index].Status = string(progress.Skipped)
				results[index].Message = "已存在，保留本地内容；同步请使用 modu update"
			}
			return nil
		}}
	}
	// 主项目必须先创建，模块 worktree 放在其目录下。
	_, failed = e.runRepositories(ctx, []repositoryJob{createJob(0)})
	if err := checkoutFailures(results, failed); err != nil {
		return results, err
	}
	moduleJobs := make([]repositoryJob, 0, len(repositories)-1)
	for index := 1; index < len(repositories); index++ {
		if repositories[index].matched {
			moduleJobs = append(moduleJobs, createJob(index))
		}
	}
	_, failed = e.runRepositories(ctx, moduleJobs)
	return results, checkoutFailures(results, failed)
}

func checkoutFailures(results []core.CheckoutResult, failed map[string]error) error {
	var failures []error
	for index := range results {
		if err := failed[results[index].Module]; err != nil {
			results[index].Status = string(progress.Failed)
			results[index].Error = err.Error()
			results[index].Message = ""
			failures = append(failures, fmt.Errorf("%s: %w", results[index].Module, err))
		}
	}
	return errors.Join(failures...)
}
