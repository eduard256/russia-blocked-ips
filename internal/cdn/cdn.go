package cdn

import (
	"fmt"
	"net/netip"
	"regexp"

	"github.com/eduard256/russia-blocked-ips/pkg/fetcher"
	"github.com/eduard256/russia-blocked-ips/pkg/output"
	"github.com/eduard256/russia-blocked-ips/pkg/parser"
)

type source struct {
	name   string
	url    string
	format string // "plain", "aws", "google", "fastly", "azure", "oracle", "csv"
}

var staticSources = []source{
	// Cloudflare
	{"Cloudflare/ips-v4", "https://www.cloudflare.com/ips-v4", "plain"},
	{"Cloudflare/ips-v6", "https://www.cloudflare.com/ips-v6", "plain"},

	// Fastly (Reddit, Vimeo, Imgur, GitHub Pages, Twitch CDN)
	{"Fastly/public-ip-list", "https://api.fastly.com/public-ip-list", "fastly"},

	// AWS (Twitch, Slack, Netflix partial)
	{"AWS/ip-ranges.json", "https://ip-ranges.amazonaws.com/ip-ranges.json", "aws"},

	// Google (YouTube, Gmail, Drive, Search, Spotify on GCP)
	{"Google/goog.json", "https://www.gstatic.com/ipranges/goog.json", "google"},
	{"Google/cloud.json", "https://www.gstatic.com/ipranges/cloud.json", "google"},

	// Oracle Cloud
	{"Oracle/public_ip_ranges.json", "https://docs.oracle.com/en-us/iaas/tools/public_ip_ranges.json", "oracle"},

	// DigitalOcean
	{"DigitalOcean/google.csv", "https://www.digitalocean.com/geo/google.csv", "csv"},
}

// Init fetches all CDN/cloud sources including Azure (dynamic URL)
func Init() ([]netip.Prefix, []output.Source) {
	var all []netip.Prefix
	var meta []output.Source

	// Azure -- needs to resolve current download URL
	azureURL := resolveAzureURL()
	allSources := make([]source, len(staticSources))
	copy(allSources, staticSources)
	if azureURL != "" {
		allSources = append(allSources, source{"Azure/ServiceTags", azureURL, "azure"})
	}

	for _, s := range allSources {
		data, err := fetcher.Get(s.url)
		if err != nil {
			fmt.Printf("[cdn] WARN %s: %v\n", s.name, err)
			meta = append(meta, output.Source{Name: s.name, URL: s.url, Status: "error"})
			continue
		}

		var prefixes []netip.Prefix
		switch s.format {
		case "plain":
			prefixes = parser.PlainText(data)
		case "aws":
			prefixes = parser.AWSRanges(data)
		case "google":
			prefixes = parser.GoogleRanges(data)
		case "fastly":
			prefixes = parser.FastlyRanges(data)
		case "azure":
			prefixes = parser.AzureServiceTags(data)
		case "oracle":
			prefixes = parser.OracleCloudRanges(data)
		case "csv":
			prefixes = parser.CSVColumn(data, 0, ',')
		}

		all = append(all, prefixes...)
		meta = append(meta, output.Source{Name: s.name, URL: s.url, Entries: len(prefixes), Status: "ok"})
		fmt.Printf("[cdn] %s: %d entries\n", s.name, len(prefixes))
	}

	return all, meta
}

// internals

var azureRe = regexp.MustCompile(`https://download\.microsoft\.com/download/[^"]+ServiceTags_Public_[^"]+\.json`)

// resolveAzureURL fetches the Azure download page and extracts the current ServiceTags URL
func resolveAzureURL() string {
	data, err := fetcher.Get("https://www.microsoft.com/en-us/download/details.aspx?id=56519")
	if err != nil {
		fmt.Printf("[cdn] WARN Azure URL resolve: %v\n", err)
		return ""
	}
	match := azureRe.Find(data)
	if match == nil {
		fmt.Println("[cdn] WARN Azure: could not find ServiceTags URL")
		return ""
	}
	return string(match)
}
