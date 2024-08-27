package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServiceStatus struct {
	Name   string
	Status int
}

type Config struct {
	Services []string `json:"services"`
	Interval int      `json:"interval"`
}

var (
	registry      = prometheus.NewRegistry()
	serviceHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_health_status",
			Help: "Health status of systemd services (1 = healthy, 0 = unhealthy)",
		},
		[]string{"service_name"},
	)
)

func init() {
	registry.MustRegister(serviceHealth)
}

func LoadConfig(configPath string) Config {
	// ----------------------------
	// Read JSON Config on startup
	// ----------------------------
	jsonFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}

	return config
}

func checkServiceStatus(serviceName string, ch chan<- ServiceStatus, wg *sync.WaitGroup) {
	defer wg.Done()
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	status := 0
	if err == nil && string(output) == "active\n" {
		status = 1
	}
	ch <- ServiceStatus{Name: serviceName, Status: status}
}

func main() {
	config := LoadConfig("config.json")
	services := config.Services

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
		log.Println("Starting server on 127.0.0.1:2112")
		log.Fatal(http.ListenAndServe("127.0.0.1:2112", nil))
	}()

	for {
		var wg sync.WaitGroup
		ch := make(chan ServiceStatus, len(services))

		for _, service := range services {
			wg.Add(1)
			go checkServiceStatus(service, ch, &wg)
		}

		// Wait for all goroutines to finish
		wg.Wait()
		close(ch)

		for result := range ch {
			serviceHealth.WithLabelValues(result.Name).Set(float64(result.Status))
			log.Printf("%s %d\n", result.Name, result.Status)
		}
		log.Println("Waiting 10 seconds to refresh service status")

		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}
