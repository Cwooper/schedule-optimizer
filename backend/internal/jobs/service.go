package jobs

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job represents a schedulable task with its own timing logic.
type Job interface {
	Name() string
	ShouldRun(now time.Time) bool
	Run(ctx context.Context, now time.Time) error
}

// Service manages and runs registered jobs on a check interval.
// Jobs are checked and run sequentially in registration order.
// This means jobs do not need internal synchronization for their own state,
// but they should not block for extended periods.
type Service struct {
	mu            sync.Mutex
	jobs          []Job
	checkInterval time.Duration
	started       bool
	stopCh        chan struct{}
}

// NewService creates a job service that checks jobs at the given interval.
func NewService(checkInterval time.Duration) *Service {
	if checkInterval <= 0 {
		checkInterval = time.Minute
	}
	return &Service{
		jobs:          make([]Job, 0),
		checkInterval: checkInterval,
		stopCh:        make(chan struct{}),
	}
}

// Register adds a job to the service. Must be called before Start.
func (s *Service) Register(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		slog.Warn("Cannot register job after service started", "job", job.Name())
		return
	}
	s.jobs = append(s.jobs, job)
	slog.Info("Registered job", "job", job.Name())
}

// Start begins the job checking loop. Blocks until Stop is called.
func (s *Service) Start(ctx context.Context) {
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()

	slog.Info("Jobs service starting", "check_interval", s.checkInterval, "job_count", len(s.jobs))

	s.checkJobs(ctx, time.Now())

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Jobs service cancelled")
			return
		case <-s.stopCh:
			slog.Info("Jobs service stopped")
			return
		case now := <-ticker.C:
			s.checkJobs(ctx, now)
		}
	}
}

// Stop signals the service to shut down.
func (s *Service) Stop() {
	close(s.stopCh)
}

func (s *Service) checkJobs(ctx context.Context, now time.Time) {
	for _, job := range s.jobs {
		if job.ShouldRun(now) {
			slog.Info("Running job", "job", job.Name())
			start := time.Now()
			if err := job.Run(ctx, now); err != nil {
				slog.Error("Job failed", "job", job.Name(), "error", err, "duration", time.Since(start))
			} else {
				slog.Info("Job completed", "job", job.Name(), "duration", time.Since(start))
			}
		}
	}
}
