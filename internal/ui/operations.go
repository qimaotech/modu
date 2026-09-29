package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/qimaotech/modu/internal/core"
	"github.com/qimaotech/modu/internal/engine"
	"github.com/qimaotech/modu/internal/progress"
)

type operationTickMsg uint64
type operationResultMsg struct {
	id        uint64
	message   tea.Msg
	cancelled bool
}

func operationTick(id uint64) tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return operationTickMsg(id) })
}

func (m *App) rootContext() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

func (m *App) stopStatusRefresh() {
	if m.statusCancel != nil {
		m.statusCancel()
		m.statusCancel = nil
	}
	m.statusID++
	m.statusLoading = false
}

func (m *App) startOperation(action string, work func(context.Context) tea.Msg) tea.Cmd {
	m.stopStatusRefresh()
	if m.operationCancel != nil {
		m.operationCancel()
	}
	ctx, cancel := context.WithCancel(m.rootContext())
	m.operationCancel = cancel
	m.operationID++
	id := m.operationID
	m.cancelling = false
	m.progress = progress.New(action)
	m.progressFrame = 0
	m.state = "loading"
	ctx = progress.WithReporter(ctx, m.progress.Record)
	return tea.Batch(func() tea.Msg {
		message := work(ctx)
		return operationResultMsg{id: id, message: message, cancelled: ctx.Err() != nil}
	}, operationTick(id))
}

func (m *App) loadEnvsCommand() tea.Cmd {
	return m.startOperation("正在载入工作区", func(ctx context.Context) tea.Msg {
		summary, err := m.Engine.ListWorkspaceSummaries(ctx)
		if err != nil {
			return errorMsg{err}
		}
		return loadedMsg{envs: summary.Envs, mainProject: summary.MainProject, summary: true}
	})
}

type statusRefreshedMsg struct {
	id      uint64
	feature string
	all     bool
	envs    []core.WorktreeEnv
	main    *engine.MainProjectStatus
	err     error
}

func (m *App) refreshStatuses(feature string, all bool) tea.Cmd {
	m.stopStatusRefresh()
	id := m.statusID
	ctx, cancel := context.WithCancel(m.rootContext())
	m.statusCancel = cancel
	m.statusLoading = true
	envs := append([]core.WorktreeEnv(nil), m.Envs...)
	return func() tea.Msg {
		defer cancel()
		message := statusRefreshedMsg{id: id, feature: feature, all: all}
		if all {
			message.envs, message.err = m.Engine.RefreshWorktreeStatuses(ctx, envs)
			if message.err == nil {
				message.main, message.err = m.Engine.GetMainProject(ctx)
			}
		} else if feature == "" {
			message.main, message.err = m.Engine.GetMainProject(ctx)
		} else {
			env, err := m.Engine.GetWorktreeInfo(ctx, feature)
			message.err = err
			if env != nil {
				message.envs = []core.WorktreeEnv{*env}
			}
		}
		return message
	}
}

func (m *App) refreshAffected(feature string) tea.Cmd {
	return m.refreshStatuses(feature, false)
}

func (m *App) resumePendingStatuses() tea.Cmd {
	if m.state == "list" && m.statusesPending && !m.statusLoading {
		return m.refreshStatuses("", true)
	}
	return nil
}
