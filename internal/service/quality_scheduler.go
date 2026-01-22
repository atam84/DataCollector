package service

import (
	"context"
	"log"
	"time"
)

// QualityScheduler handles automatic quality checks
type QualityScheduler struct {
	qualityService *QualityService
	ticker         *time.Ticker
	stopChan       chan bool
	interval       time.Duration
}

// NewQualityScheduler creates a new quality scheduler
func NewQualityScheduler(qualityService *QualityService, interval time.Duration) *QualityScheduler {
	if interval == 0 {
		interval = 1 * time.Hour // Default to 1 hour
	}

	return &QualityScheduler{
		qualityService: qualityService,
		interval:       interval,
		stopChan:       make(chan bool),
	}
}

// Start begins the scheduler loop
func (s *QualityScheduler) Start() {
	log.Printf("Quality scheduler started - checking quality every %s", s.interval)

	// Run initial check after a short delay (don't block startup)
	go func() {
		time.Sleep(30 * time.Second) // Wait 30 seconds before first check
		s.runQualityCheck()
	}()

	// Then check at the specified interval
	s.ticker = time.NewTicker(s.interval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.runQualityCheck()
			case <-s.stopChan:
				log.Println("Quality scheduler stopped")
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *QualityScheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
}

// runQualityCheck runs a scheduled quality check
func (s *QualityScheduler) runQualityCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	log.Println("[QUALITY_SCHEDULER] Starting scheduled quality check")

	err := s.qualityService.RunScheduledCheck(ctx)
	if err != nil {
		log.Printf("[QUALITY_SCHEDULER] Failed to run scheduled check: %v", err)
	}
}
