package jobs

import (
	"fmt"
	"log"
	"time"

	"asset-diary/services"

	"github.com/robfig/cron/v3"
)

type RecordDailyTotalAssetValueJob struct {
	service services.DailyTotalAssetValueServiceInterface
}

func NewRecordDailyTotalAssetValueJob(service services.DailyTotalAssetValueServiceInterface) *RecordDailyTotalAssetValueJob {
	return &RecordDailyTotalAssetValueJob{
		service: service,
	}
}

func (j *RecordDailyTotalAssetValueJob) run() {
	log.Println("Scheduled daily asset recording job starting...")
	if err := j.service.RecordDailyTotalAssetValue(); err != nil {
		log.Printf("Error recording daily assets: %v", err)
	} else {
		log.Println("Scheduled daily asset recording completed successfully")
	}
}

// The job will run daily at 22:05 UTC (06:05 CST)
func (j *RecordDailyTotalAssetValueJob) Schedule() (*cron.Cron, error) {
	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("5 22 * * *", j.run)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule daily asset job: %v", err)
	}

	c.Start()
	log.Println("Scheduled daily asset job started")

	return c, nil
}

func (j *RecordDailyTotalAssetValueJob) Stop(c *cron.Cron) {
	c.Stop()
	log.Println("Scheduled daily asset job stopped")
}
