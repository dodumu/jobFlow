package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"jobFlow/database"
	"jobFlow/models"
	"log"
	"time"
)

const (
	pollInterval = 2 * time.Second
	jobTimeout   = 2 * time.Minute
)

type Output struct {
	Result        string
	Model         string
	PromptVersion string
}

type Handler func(ctx context.Context, job models.AIJob) (Output, error)

type Worker struct {
	handlers map[string]Handler
}

func New() *Worker {
	return &Worker{handlers: make(map[string]Handler)}
}

func (w *Worker) Register(jobType string, handler Handler) {
	w.handlers[jobType] = handler
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Println("ai worker started")

	for {
		for ctx.Err() == nil && w.runOne(ctx) {

		}

		select {
		case <-ctx.Done():
			log.Println("ai worker stopped")
			return
		case <-ticker.C:

		}
	}
}

func (w *Worker) runOne(ctx context.Context) bool {
	job, err := database.ClaimNextAIJob()
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	if err != nil {
		log.Printf("ai worker: claiming job: %v", err)
		return false
	}

	w.process(ctx, job)
	return true
}

func (w *Worker) process(ctx context.Context, job models.AIJob) {
	handler, ok := w.handlers[job.Type]
	if !ok {
		w.fail(job, fmt.Sprintf("no handler for job type %q", job.Type))
		return
	}

	jobCtx, cancel := context.WithTimeout(ctx, jobTimeout)
	defer cancel()

	out, err := safeRun(jobCtx, handler, job)
	if err != nil {
		w.fail(job, err.Error())
		return
	}

	if err := database.CompleteAIJob(job.ID, out.Result, out.Model, out.PromptVersion); err != nil {
		log.Printf("ai worker: %v", err)
	}
}

func (w *Worker) fail(job models.AIJob, message string) {
	log.Printf("ai worker: job %d (%s) failed: %s", job.ID, job.Type, message)

	if err := database.FailAIJob(job.ID, message); err != nil {
		log.Printf("ai worker: %v", err)
	}
}

// safeRun calls the handler, but turns a panic into an ordinary error so one
// bad job cannot crash the whole program.
func safeRun(ctx context.Context, handler Handler, job models.AIJob) (out Output, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panicked: %v", r)
		}
	}()

	return handler(ctx, job)
}

// PingHandler is a harmless handler for testing the queue end to end.
func PingHandler(ctx context.Context, job models.AIJob) (Output, error) {
	select {
	case <-time.After(time.Second):
	case <-ctx.Done():
		return Output{}, ctx.Err()
	}

	return Output{Result: `{"pong": true}`}, nil
}
