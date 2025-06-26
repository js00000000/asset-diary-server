package jobs

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
)

type HealthCheckJob struct {
	serverURL          string
	dailyAssetValueJob *RecordDailyTotalAssetValueJob
}

func NewHealthCheckJob(serverURL string, dailyAssetValueJob *RecordDailyTotalAssetValueJob) *HealthCheckJob {
	// Ensure the server URL doesn't end with a slash
	if len(serverURL) > 0 && serverURL[len(serverURL)-1] == '/' {
		serverURL = serverURL[:len(serverURL)-1]
	}
	return &HealthCheckJob{
		serverURL:          serverURL,
		dailyAssetValueJob: dailyAssetValueJob,
	}
}

func (j *HealthCheckJob) run() {
	url := j.serverURL + "/api/healthz"
	log.Printf("Performing health check: %s", url)

	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	// Call the health check endpoint
	startTime := time.Now()
	resp, err := client.Get(url)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("Health check failed after %v: %v\n", duration, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Health check successful (status: %d, duration: %v)\n", resp.StatusCode, duration)
	} else {
		log.Printf("Health check failed (status: %d, duration: %v)\n", resp.StatusCode, duration)
	}
}

func (j *HealthCheckJob) Schedule() (*cron.Cron, error) {
	c := cron.New(cron.WithSeconds())

	// Schedule to run every 10 minutes
	_, err := c.AddFunc("0 */10 * * * *", j.run)

	if err != nil {
		return nil, fmt.Errorf("failed to schedule health check job: %v", err)
	}

	log.Println("Running initial health check...")
	go j.run()

	c.Start()
	log.Println("Scheduled health check job started")

	return c, nil
}

func (j *HealthCheckJob) Stop(c *cron.Cron) {
	c.Stop()
	log.Println("Scheduled health check job stopped")
}
