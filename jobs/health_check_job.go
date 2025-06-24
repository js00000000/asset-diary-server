package jobs

import (
	"log"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
)

type HealthCheckJob struct {
	serverURL string
}

func NewHealthCheckJob(serverURL string) *HealthCheckJob {
	// Ensure the server URL doesn't end with a slash
	if len(serverURL) > 0 && serverURL[len(serverURL)-1] == '/' {
		serverURL = serverURL[:len(serverURL)-1]
	}
	return &HealthCheckJob{
		serverURL: serverURL,
	}
}

func (j *HealthCheckJob) Run() {
	url := j.serverURL + "/healthz"
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

func (j *HealthCheckJob) Schedule() *cron.Cron {
	// Create a new cron instance with seconds precision
	c := cron.New(cron.WithSeconds())

	// Schedule to run every 10 minutes
	schedule := "0 */10 * * * *" // Every 10 minutes
	_, err := c.AddFunc(schedule, func() {
		log.Println("Scheduled health check starting...")
		j.Run()
	})

	if err != nil {
		log.Fatalf("Failed to schedule health check job: %v", err)
	}

	log.Printf("Health check job scheduled to run every 10 minutes")

	// Run once immediately on startup
	log.Println("Running initial health check...")
	go j.Run()

	c.Start()
	log.Println("Health check job started")

	return c
}
