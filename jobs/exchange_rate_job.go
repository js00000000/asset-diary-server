package jobs

import (
	"log"
	"time"

	"asset-diary/services"

	"github.com/robfig/cron/v3"
)

type ExchangeRateJob struct {
	service services.ExchangeRateServiceInterface
}

func NewExchangeRateJob(service services.ExchangeRateServiceInterface) *ExchangeRateJob {
	return &ExchangeRateJob{service: service}
}

func (j *ExchangeRateJob) Run() {
	log.Println("Starting exchange rate update job...")
	if err := j.service.FetchAndStoreRates(); err != nil {
		log.Printf("Error updating exchange rates: %v\n", err)
	} else {
		log.Println("Exchange rate update completed successfully")
	}
}

func (j *ExchangeRateJob) Schedule() *cron.Cron {
	// Run daily at 00:05 UTC (08:05 CST)
	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("8 0 * * *", j.Run)
	if err != nil {
		log.Fatalf("Failed to schedule exchange rate job: %v", err)
	}

	// Run once immediately on startup
	go j.Run()

	return c
}
