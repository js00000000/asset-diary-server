package jobs

import (
	"fmt"
	"log"
	"time"

	"asset-diary/services"

	"github.com/robfig/cron/v3"
)

type ExchangeRateJob struct {
	service            services.ExchangeRateServiceInterface
	dailyAssetValueJob *RecordDailyTotalAssetValueJob
}

func NewExchangeRateJob(service services.ExchangeRateServiceInterface, dailyAssetValueJob *RecordDailyTotalAssetValueJob) *ExchangeRateJob {
	return &ExchangeRateJob{
		service:            service,
		dailyAssetValueJob: dailyAssetValueJob,
	}
}

func (j *ExchangeRateJob) run() {
	log.Println("Scheduled exchange rate update starting...")
	if err := j.service.FetchAndStoreRates(); err != nil {
		log.Printf("Error updating exchange rates: %v\n", err)
	} else {
		log.Println("Scheduled exchange rate update completed successfully")
	}
}

func (j *ExchangeRateJob) Schedule() (*cron.Cron, error) {
	// Run daily at 00:05 UTC (08:05 CST)
	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("5 0 * * *", j.run)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule exchange rate job: %v", err)
	}

	// Run once immediately on startup
	log.Println("Running initial exchange rate update...")
	go j.run()

	c.Start()
	log.Println("Scheduled exchange rate job started")

	return c, nil
}

func (j *ExchangeRateJob) Stop(c *cron.Cron) {
	c.Stop()
	log.Println("Scheduled exchange rate job stopped")
}
