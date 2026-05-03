// Package scheduler 提供定时和手动触发能力。
package scheduler

import (
	"context"
	"time"
)

// Job 表示一个可被调度执行的任务。
type Job func(context.Context) error

// Scheduler 使用标准库定时器周期执行任务。
type Scheduler struct {
	interval time.Duration
	job      Job
}

// New 创建定时调度器。
func New(interval time.Duration, job Job) *Scheduler {
	return &Scheduler{
		interval: interval,
		job:      job,
	}
}

// RunOnce 手动触发一次任务。
func (s *Scheduler) RunOnce(ctx context.Context) error {
	return s.job(ctx)
}

// Start 按固定间隔持续执行任务，直到上下文取消。
func (s *Scheduler) Start(ctx context.Context, onError func(error)) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.job(ctx); err != nil && onError != nil {
				onError(err)
			}
		}
	}
}
