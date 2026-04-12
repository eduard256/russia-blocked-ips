package output

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"strings"
	"time"
)

// Source describes one data source in the manifest
type Source struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Entries int    `json:"entries"`
	Status  string `json:"status"`
}

// Manifest is the main metadata file for client programs
type Manifest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	UpdatedAt   string   `json:"updated_at"`
	TotalCIDRs  int      `json:"total_cidrs"`
	SHA256      string   `json:"sha256"`
	Sources     []Source `json:"sources"`
}

// WriteIPFile writes all prefixes to ip.txt, one CIDR per line
func WriteIPFile(path string, prefixes []netip.Prefix) error {
	var b strings.Builder
	for _, p := range prefixes {
		b.WriteString(p.String())
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

// WriteManifest generates manifest.json with sha256 of ip.txt
func WriteManifest(path string, ipFile string, sources []Source) error {
	data, err := os.ReadFile(ipFile)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(data)

	// count non-empty lines
	total := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			total++
		}
	}

	m := Manifest{
		Name:        "Russia Blocked IPs",
		Description: "Aggregated list of all IP addresses blocked in Russia and by sanctions",
		Author:      "eduard256",
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		TotalCIDRs:  total,
		SHA256:      fmt.Sprintf("%x", hash),
		Sources:     sources,
	}

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, 0644)
}
