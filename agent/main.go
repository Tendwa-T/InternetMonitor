package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const Version = "1.1.0"

type Status struct {
	Online            bool      `json:"online"`
	LatencyMs         float64   `json:"latency_ms"`
	Timestamp         time.Time `json:"timestamp"`
	DownloadSpeedMbps float64   `json:"download_speed_mbps"`
	Version           string    `json:"version"`
}

type Config struct {
	LatencyThresholdMs float64 `json:"latencyThresholdMs"`
	DegradedSpeedMbps  float64 `json:"degradedSpeedMbps"`
}

// Shared variables for download speed tracking and concurrency control
var (
	speedMutex          sync.RWMutex
	lastSpeed           float64
	lastRecordedSpeed   float64 // Track last recorded speed to avoid consecutive zeros
	guardMutex          sync.Mutex
	measuring           bool      // Flag to prevent concurrent measurements
	lastMeasurementTime time.Time // Track last measurement time for cooldown
	measurementCooldown = 30 * time.Second
	httpClient          = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        2,
			MaxIdleConnsPerHost: 1,
		},
	}
)

func loadConfig() Config {
	path := os.ExpandEnv("$HOME/Library/Application Support/InternetMonitor/config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{LatencyThresholdMs: 150, DegradedSpeedMbps: 10}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("Failed to parse config.json, using defaults: %v", err)
		return Config{LatencyThresholdMs: 150, DegradedSpeedMbps: 10}
	}
	return cfg
}

func checkConnectivity() Status {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", "1.1.1.1:53", 2*time.Second)
	if err != nil {
		return Status{
			Online:    false,
			Timestamp: time.Now(),
			Version:   Version,
		}
	}
	defer conn.Close()
	latency := time.Since(start).Seconds() * 1000
	return Status{
		Online:    true,
		LatencyMs: latency,
		Timestamp: time.Now(),
		Version:   Version,
	}
}

// Asynchronous download speed measurement
func measureDownloadSpeedAsync(db *sql.DB) {
	// Check if measurement already in progress or cooldown not elapsed
	guardMutex.Lock()
	if measuring || time.Since(lastMeasurementTime) < measurementCooldown {
		guardMutex.Unlock()
		return
	}
	measuring = true
	guardMutex.Unlock()

	go func() {
		defer func() {
			guardMutex.Lock()
			measuring = false
			lastMeasurementTime = time.Now()
			guardMutex.Unlock()
		}()

		log.Println("Starting download speed measurement")
		start := time.Now()
		url := "https://speed.cloudflare.com/__down?bytes=5000000" // 5 MB test file
		resp, err := httpClient.Get(url)
		if err != nil {
			log.Printf("Download failed: %v", err)
			updateLastSpeed(0, db)
			return
		}
		defer resp.Body.Close()
		n, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			log.Printf("Failed to read response: %v", err)
			updateLastSpeed(0, db)
			return
		}
		duration := time.Since(start).Seconds()
		if duration == 0 {
			updateLastSpeed(0, db)
			return
		}
		mbps := float64(n*8) / duration / 1_000_000
		log.Printf("Download speed: %.2f Mbps (%d bytes in %.2f sec)", mbps, n, duration)
		updateLastSpeed(mbps, db)
	}()
}

// Update lastSpeed and insert history
func updateLastSpeed(speed float64, db *sql.DB) {
	speedMutex.Lock()
	lastSpeed = speed
	// Skip DB write if speed is 0 and last recorded speed was also 0 (avoid consecutive zeros)
	shouldWrite := speed != 0 || lastRecordedSpeed != 0
	if shouldWrite {
		lastRecordedSpeed = speed
	}
	speedMutex.Unlock()

	if !shouldWrite {
		log.Printf("Skipping DB write: consecutive zero speed measurement")
		return
	}

	stmt, err := db.Prepare("INSERT INTO history(ts, online, latency, speed, version) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("Failed to prepare DB statement: %v", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), true, 0, speed, Version)
	if err != nil {
		log.Printf("Failed to insert history: %v", err)
	}
}

func main() {

	dbPath := os.ExpandEnv("$HOME/Library/Application Support/InternetMonitor/history.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS history(
		ts DATETIME,
		online BOOLEAN,
		latency REAL,
		speed REAL,
		version TEXT
	)`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := checkConnectivity()

		// Always return the latest known speed without waiting
		speedMutex.RLock()
		status.DownloadSpeedMbps = lastSpeed
		speedMutex.RUnlock()

		status.Timestamp = time.Now()
		status.Version = Version

		// Immediately return status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)

		// Trigger new speed measurement if online
		if status.Online {
			measureDownloadSpeedAsync(db)
		}
	})

	log.Println("Agent Running on http://127.0.0.1:8787")
	log.Fatal(http.ListenAndServe("127.0.0.1:8787", nil))
}
