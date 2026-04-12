package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	manifestURL = "https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/manifest.json"
	ipURL       = "https://raw.githubusercontent.com/eduard256/russia-blocked-ips/main/ip.txt"
	userAgent   = "rbi-client/1.0"
)

type manifest struct {
	UpdatedAt  string `json:"updated_at"`
	TotalCIDRs int    `json:"total_cidrs"`
	SHA256     string `json:"sha256"`
}

func main() {
	output := flag.String("output", "./ip.txt", "path to save ip.txt")
	interval := flag.Duration("interval", 5*time.Minute, "check interval")
	onUpdate := flag.String("on-update", "", "shell command to run after update")
	flag.Parse()

	fmt.Println("rbi-client started")
	fmt.Printf("  output:   %s\n", *output)
	fmt.Printf("  interval: %s\n", *interval)
	if *onUpdate != "" {
		fmt.Printf("  on-update: %s\n", *onUpdate)
	}
	fmt.Println()

	var lastSHA string

	// try to read existing file hash
	if data, err := os.ReadFile(*output); err == nil {
		h := sha256.Sum256(data)
		lastSHA = fmt.Sprintf("%x", h)
		fmt.Printf("[client] existing file hash: %s\n", lastSHA[:16]+"...")
	}

	// first check immediately
	lastSHA = check(lastSHA, *output, *onUpdate)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			lastSHA = check(lastSHA, *output, *onUpdate)
		case s := <-sig:
			fmt.Printf("\n[client] received %s, shutting down\n", s)
			return
		}
	}
}

func check(lastSHA, output, onUpdate string) string {
	fmt.Printf("[client] checking manifest... ")

	data, err := fetch(manifestURL)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return lastSHA
	}

	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Printf("parse error: %v\n", err)
		return lastSHA
	}

	if m.SHA256 == lastSHA {
		fmt.Printf("up to date (%d CIDRs, %s)\n", m.TotalCIDRs, m.UpdatedAt)
		return lastSHA
	}

	fmt.Printf("update available! %d CIDRs, %s\n", m.TotalCIDRs, m.UpdatedAt)

	// download ip.txt
	fmt.Print("[client] downloading ip.txt... ")
	ipData, err := fetch(ipURL)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return lastSHA
	}

	// verify hash
	h := sha256.Sum256(ipData)
	got := fmt.Sprintf("%x", h)
	if got != m.SHA256 {
		fmt.Printf("hash mismatch! expected %s, got %s\n", m.SHA256[:16]+"...", got[:16]+"...")
		return lastSHA
	}

	// save file
	if err := os.WriteFile(output, ipData, 0644); err != nil {
		fmt.Printf("write error: %v\n", err)
		return lastSHA
	}

	lines := strings.Count(string(ipData), "\n")
	fmt.Printf("saved %s (%d lines, %d KB)\n", output, lines, len(ipData)/1024)

	// run on-update command
	if onUpdate != "" {
		fmt.Printf("[client] running: %s\n", onUpdate)
		cmd := exec.Command("sh", "-c", onUpdate)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("[client] on-update error: %v\n", err)
		} else {
			fmt.Println("[client] on-update: ok")
		}
	}

	return got
}

func fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
