package jobs

import (
	"log"
	"time"

	"asset-diary/services"

	"github.com/robfig/cron/v3"
)

type RecordDailyTotalAssetValueJob struct {
	service services.DailyTotalAssetValueServiceInterface
}

func NewRecordDailyTotalAssetValueJob(service services.DailyTotalAssetValueServiceInterface) *RecordDailyTotalAssetValueJob {
	return &RecordDailyTotalAssetValueJob{service: service}
}

func (j *RecordDailyTotalAssetValueJob) Run() {
	log.Println("Starting daily asset recording job...")
	if err := j.service.RecordDailyTotalAssetValue(); err != nil {
		log.Printf("Error recording daily assets: %v\n", err)
	} else {
		log.Println("Daily asset recording completed successfully")
	}
}

// The job will run daily at 5:00 AM in Asia/Taipei (UTC+8) timezone
func (j *RecordDailyTotalAssetValueJob) Schedule() *cron.Cron {
	tz, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}

	c := cron.New(cron.WithLocation(tz))

	_, err = c.AddFunc("0 5 * * *", j.Run)
	if err != nil {
		log.Fatalf("Failed to schedule daily asset job: %v", err)
	}

	c.Start()
	log.Println("Daily asset job started")

	return c
}
