package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Job interface {
	Name() string
	Run(ctx context.Context) error
	Interval() time.Duration
}

type Runner struct {
	jobs []Job
	log  *slog.Logger
}

func NewRunner(log *slog.Logger) *Runner {
	return &Runner{log: log}
}

func (r *Runner) Add(job Job) {
	r.jobs = append(r.jobs, job)
}

func (r *Runner) Start(ctx context.Context) {
	for _, job := range r.jobs {
		j := job
		go r.runLoop(ctx, j)
	}
}

func (r *Runner) runLoop(ctx context.Context, job Job) {
	r.log.Info(fmt.Sprintf("starting worker: %s (interval: %s)", job.Name(), job.Interval()))

	ticker := time.NewTicker(job.Interval())
	defer ticker.Stop()

	if err := job.Run(ctx); err != nil {
		r.log.Error(fmt.Sprintf("worker %s failed on first run", job.Name()), "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			r.log.Info(fmt.Sprintf("worker %s stopped", job.Name()))
			return
		case <-ticker.C:
			if err := job.Run(ctx); err != nil {
				r.log.Error(fmt.Sprintf("worker %s failed", job.Name()), "error", err)
			}
		}
	}
}
