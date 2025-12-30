package services

import (
	"log"

	"github.com/navaneethk-000/url-shortener-backend/internal/models"
	"github.com/navaneethk-000/url-shortener-backend/internal/repository"
)

type AnalyticsJob struct {
	Click *models.Click
}

type AnalyticsWorker struct {
	JobQueue  chan AnalyticsJob
	ClickRepo *repository.ClickRepository
}

var WorkerInstance *AnalyticsWorker

func InitWorker(clickRepo *repository.ClickRepository, bufferSize int, workers int) {
	WorkerInstance = &AnalyticsWorker{
		JobQueue:  make(chan AnalyticsJob, bufferSize),
		ClickRepo: clickRepo,
	}

	for i := 0; i < workers; i++ {
		go WorkerInstance.process()
	}
	log.Printf("Analytics Worker Pool started with %d workers", workers)
}

func (w *AnalyticsWorker) Push(click *models.Click) {
	select {
	case w.JobQueue <- AnalyticsJob{Click: click}:
	default:
		log.Println("Analytics Queue full! Dropping event.")
	}
}

func (w *AnalyticsWorker) process() {
	for job := range w.JobQueue {
		// Get Location
		country, city, err := GetLocation(job.Click.IPAddress)

		if err != nil {
			log.Printf("GeoAPI Error: %v", err)
			job.Click.Country = "Unknown"
			job.Click.City = "Unknown"
		} else {
			job.Click.Country = country
			job.Click.City = city
			log.Printf("Worker: Assigning Location -> %s, %s", job.Click.City, job.Click.Country)
		}

		// Save to DB
		err = w.ClickRepo.SaveClick(job.Click)
		if err != nil {
			log.Printf(" DB Save Failed: %v", err)
		} else {
			// Verify what was saved
			log.Printf("Saved Click ID: %d | IP: %s | Country: %s", job.Click.ID, job.Click.IPAddress, job.Click.Country)
		}
	}
}
