package jobs

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type mockJob struct {
	name      string
	shouldRun bool
	runCount  atomic.Int32
	runErr    error
}

func (m *mockJob) Name() string              { return m.name }
func (m *mockJob) ShouldRun(now time.Time) bool { return m.shouldRun }
func (m *mockJob) Run(ctx context.Context, now time.Time) error {
	m.runCount.Add(1)
	return m.runErr
}

func TestService_Register(t *testing.T) {
	svc := NewService(time.Minute)

	job1 := &mockJob{name: "job1"}
	job2 := &mockJob{name: "job2"}

	svc.Register(job1)
	svc.Register(job2)

	if len(svc.jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(svc.jobs))
	}
	if svc.jobs[0].Name() != "job1" {
		t.Errorf("expected first job to be job1, got %s", svc.jobs[0].Name())
	}
	if svc.jobs[1].Name() != "job2" {
		t.Errorf("expected second job to be job2, got %s", svc.jobs[1].Name())
	}
}

func TestService_RegisterAfterStart(t *testing.T) {
	svc := NewService(time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	go svc.Start(ctx)
	time.Sleep(10 * time.Millisecond) // Let Start() begin

	job := &mockJob{name: "late-job"}
	svc.Register(job)

	if len(svc.jobs) != 0 {
		t.Error("should not be able to register job after start")
	}

	cancel()
}

func TestService_CheckJobs(t *testing.T) {
	svc := NewService(time.Minute)

	runningJob := &mockJob{name: "running", shouldRun: true}
	skippedJob := &mockJob{name: "skipped", shouldRun: false}

	svc.Register(runningJob)
	svc.Register(skippedJob)

	now := time.Now()
	svc.checkJobs(context.Background(), now)

	if runningJob.runCount.Load() != 1 {
		t.Errorf("running job should have run once, ran %d times", runningJob.runCount.Load())
	}
	if skippedJob.runCount.Load() != 0 {
		t.Errorf("skipped job should not have run, ran %d times", skippedJob.runCount.Load())
	}
}

func TestService_Stop(t *testing.T) {
	svc := NewService(100 * time.Millisecond)

	ctx := context.Background()
	done := make(chan struct{})

	go func() {
		svc.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	svc.Stop()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("service did not stop within timeout")
	}
}

func TestNewService_DefaultInterval(t *testing.T) {
	svc := NewService(0)
	if svc.checkInterval != time.Minute {
		t.Errorf("expected default interval of 1m, got %v", svc.checkInterval)
	}

	svc2 := NewService(-5 * time.Second)
	if svc2.checkInterval != time.Minute {
		t.Errorf("expected default interval of 1m for negative input, got %v", svc2.checkInterval)
	}
}
