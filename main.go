package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
)

func diskUsagePercent(path string) (int, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free
	percent := int((float64(used) / float64(total)) * 100)
	return percent, nil
}

func getExternalIP() string {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unknown"
	}
	return string(ip)
}

func sendMailgunAlert(domain, apiKey, recipient, hostname, ip string, usage int) error {
	apiURL := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", domain)
	data := url.Values{}
	data.Set("from", "alert@"+domain)
	data.Set("to", recipient)
	data.Set("subject", fmt.Sprintf("Disk Usage Alert on %s (%s)", hostname, ip))
	data.Set("text", fmt.Sprintf("Warning: /dev/sda1 usage is at %d%%", usage))

	req, _ := http.NewRequest("POST", apiURL, nil)
	req.SetBasicAuth("api", apiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = data.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Mailgun error: %s", resp.Status)
	}
	return nil
}

func main() {
	_ = godotenv.Load()
	mailgunDomain := os.Getenv("MAILGUN_DOMAIN")
	mailgunAPIKey := os.Getenv("MAILGUN_API_KEY")
	mountPoint := os.Getenv("MOUNT_POINT")
	thresholdStr := os.Getenv("USAGE_THRESHOLD")
	recipient := os.Getenv("ALERT_RECIPIENT")

	if mailgunDomain == "" || mailgunAPIKey == "" || mountPoint == "" || thresholdStr == "" || recipient == "" {
		log.Fatal("Required environment variables not set in .env")
	}

	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		log.Fatalf("Invalid USAGE_THRESHOLD: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	ip := getExternalIP()

	usage, err := diskUsagePercent(mountPoint)
	if err != nil {
		log.Fatalf("Failed to get disk usage: %v", err)
	}
	if usage > threshold {
		if err := sendMailgunAlert(mailgunDomain, mailgunAPIKey, recipient, hostname, ip, usage); err != nil {
			log.Printf("Failed to send alert: %v", err)
			os.Exit(1)
		}
		log.Printf("Alert sent: usage at %d%%", usage)
	} else {
		log.Printf("Usage OK: %d%%", usage)
	}
}
