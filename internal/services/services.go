package services

import (
	"fmt"
	"net/netip"

	"github.com/eduard256/russia-blocked-ips/pkg/fetcher"
	"github.com/eduard256/russia-blocked-ips/pkg/output"
	"github.com/eduard256/russia-blocked-ips/pkg/parser"
)

var sources = []struct {
	name   string
	url    string
	format string // "plain", "mixed", "github", "csv", "office365"
}{
	// Telegram -- official + extended
	{"Telegram/cidr.txt", "https://core.telegram.org/resources/cidr.txt", "plain"},

	// GitHub -- all service IPs
	{"GitHub/meta", "https://api.github.com/meta", "github"},

	// Zoom -- meetings, phone, general
	{"Zoom/Zoom.txt", "https://assets.zoom.us/docs/ipranges/Zoom.txt", "plain"},
	{"Zoom/ZoomMeetings.txt", "https://assets.zoom.us/docs/ipranges/ZoomMeetings.txt", "plain"},
	{"Zoom/ZoomPhone.txt", "https://assets.zoom.us/docs/ipranges/ZoomPhone.txt", "plain"},

	// Apple iCloud Private Relay / FaceTime / iMessage
	{"Apple/icloud-egress.csv", "https://mask-api.icloud.com/egress-ip-ranges.csv", "csv"},

	// Discord IPs from Re-filter
	{"Re-filter/discord_ips.lst", "https://raw.githubusercontent.com/1andrevich/Re-filter-lists/refs/heads/main/discord_ips.lst", "plain"},

	// Re-filter community IPs
	{"Re-filter/community_ips.lst", "https://raw.githubusercontent.com/1andrevich/Re-filter-lists/refs/heads/main/community_ips.lst", "plain"},

	// iamwildtuna -- community maintained service IPs (Meta, YouTube, ChatGPT, Discord, etc.)
	{"iamwildtuna/gist", "https://gist.githubusercontent.com/iamwildtuna/7772b7c84a11bf6e1385f23096a73a15/raw/gistfile2.txt", "mixed"},

	// itdoginfo/allow-domains -- curated subnets per service
	{"itdoginfo/Meta.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/Meta.lst", "plain"},
	{"itdoginfo/telegram.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/telegram.lst", "plain"},
	{"itdoginfo/Discord.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/Discord.lst", "plain"},
	{"itdoginfo/twitter.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/twitter.lst", "plain"},
	{"itdoginfo/cloudflare.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/cloudflare.lst", "plain"},
	{"itdoginfo/cloudfront.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/cloudfront.lst", "plain"},
	{"itdoginfo/digitalocean.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/digitalocean.lst", "plain"},
	{"itdoginfo/hetzner.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/hetzner.lst", "plain"},
	{"itdoginfo/ovh.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/ovh.lst", "plain"},
	{"itdoginfo/google_meet.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/google_meet.lst", "plain"},
	{"itdoginfo/roblox.lst", "https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Subnets/IPv4/roblox.lst", "plain"},

	// V3nilla -- combined ipset for bypass
	{"V3nilla/ipset-all.txt", "https://raw.githubusercontent.com/V3nilla/IPSets-For-Bypass-in-Russia/refs/heads/main/ipset-all.txt", "plain"},

	// Office 365 / Teams / Outlook
	{"Microsoft/Office365", "https://endpoints.office.com/endpoints/worldwide?clientrequestid=b10c5ed1-bad1-445f-b386-b919946339a7", "office365"},
}

// Init fetches all service-specific sources
func Init() ([]netip.Prefix, []output.Source) {
	var all []netip.Prefix
	var meta []output.Source

	for _, s := range sources {
		data, err := fetcher.Get(s.url)
		if err != nil {
			fmt.Printf("[services] WARN %s: %v\n", s.name, err)
			meta = append(meta, output.Source{Name: s.name, URL: s.url, Status: "error"})
			continue
		}

		var prefixes []netip.Prefix
		switch s.format {
		case "plain":
			prefixes = parser.PlainText(data)
		case "mixed":
			prefixes = parser.Mixed(data)
		case "github":
			prefixes = parser.GitHubMeta(data)
		case "csv":
			prefixes = parser.CSVColumn(data, 0, ',')
		case "office365":
			prefixes = parser.Office365Endpoints(data)
		}

		all = append(all, prefixes...)
		meta = append(meta, output.Source{Name: s.name, URL: s.url, Entries: len(prefixes), Status: "ok"})
		fmt.Printf("[services] %s: %d entries\n", s.name, len(prefixes))
	}

	return all, meta
}
