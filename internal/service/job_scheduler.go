package service

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/datacollector/internal/repository"
)

// JobScheduler handles automatic job execution
type JobScheduler struct {
	jobRepo    *repository.JobRepository
	jobExecutor *JobExecutor
	ticker     *time.Ticker
	stopChan   chan bool
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(jobRepo *repository.JobRepository, jobExecutor *JobExecutor) *JobScheduler {
	return &JobScheduler{
		jobRepo:    jobRepo,
		jobExecutor: jobExecutor,
		stopChan:   make(chan bool),
	}
}

// Start begins the scheduler loop
func (s *JobScheduler) Start() {
	log.Println("Job scheduler started - checking for jobs every 30 seconds")

	// Check immediately on startup
	go s.checkAndExecuteJobs()

	// Then check every 30 seconds
	s.ticker = time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkAndExecuteJobs()
			case <-s.stopChan:
				log.Println("Job scheduler stopped")
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *JobScheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.stopChan <- true
}

// checkAndExecuteJobs finds and executes jobs that are ready to run
func (s *JobScheduler) checkAndExecuteJobs() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Find all runnable jobs
	jobs, err := s.jobRepo.FindRunnableJobs(ctx)
	if err != nil {
		log.Printf("Failed to find runnable jobs: %v", err)
		return
	}

	if len(jobs) == 0 {
		return
	}

	log.Printf("Found %d job(s) ready to execute", len(jobs))

	// Execute each job in a goroutine
	for _, job := range jobs {
		go func(jobID string) {
			execCtx, execCancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer execCancel()

			log.Printf("Executing job %s", jobID)

			result, err := s.jobExecutor.ExecuteJob(execCtx, jobID)
			if err != nil {
				log.Printf("Job %s execution error: %v", jobID, err)
				return
			}

			if result.Success {
				log.Printf("Job %s executed successfully - fetched %d records in %dms",
					jobID, result.RecordsFetched, result.ExecutionTimeMs)
			} else {
				log.Printf("Job %s execution failed: %s", jobID, result.Message)
			}
		}(job.ID.Hex())
	}
}
