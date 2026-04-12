package asn

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/eduard256/russia-blocked-ips/pkg/fetcher"
	"github.com/eduard256/russia-blocked-ips/pkg/output"
	"github.com/eduard256/russia-blocked-ips/pkg/parser"
)

// all ASN numbers to fetch via RIPE API
// sourced from b4geoip config + our research
var asnList = []struct {
	asn  int
	name string
}{
	// AI services
	{396982, "OpenAI/ChatGPT"},
	{401307, "Anthropic/Claude"},

	// Social media
	{32934, "Meta/Instagram/Facebook"},
	{13414, "Twitter/X"},
	{36459, "Discord"},

	// Video/streaming
	{2906, "Netflix"},
	{46489, "Twitch"},
	{36040, "YouTube CDN"},

	// Communication
	{62041, "Telegram (primary)"},
	{211157, "Telegram (secondary)"},
	{30103, "Zoom"},
	{8403, "Spotify"},
	{14413, "LinkedIn"},

	// Search/cloud
	{15169, "Google"},
	{8075, "Microsoft"},

	// CDN
	{20940, "Akamai"},
	{54113, "Fastly"},
	{13335, "Cloudflare"},

	// Hosting
	{24940, "Hetzner"},
	{14061, "DigitalOcean"},
	{16509, "Amazon AWS (primary)"},
	{14618, "Amazon AWS (secondary)"},
	{31898, "Oracle Cloud"},

	// from b4geoip extended list
	{174, "Cogent Communications"},
	{714, "Apple"},
	{2527, "Sony Interactive Entertainment"},
	{6142, "Valve/Steam (primary)"},
	{6185, "Apple (CDN)"},
	{6507, "Riot Games"},
	{8361, "Valve/Steam (secondary)"},
	{8849, "Mismatch"},
	{10747, "Twitch (secondary)"},
	{11278, "Riot Games (secondary)"},
	{11281, "Roblox (primary)"},
	{11795, "Epic Games"},
	{12222, "Naver/LINE"},
	{12876, "Scaleway/Online.net"},
	{13720, "ImageShack"},
	{15224, "Valve/Steam (EU)"},
	{16276, "OVH"},
	{16625, "Akamai (secondary)"},
	{17204, "Valve/Steam (relay)"},
	{19541, "Riot Games (valorant)"},
	{20054, "Valve/Steam (CDN)"},
	{20473, "Vultr"},
	{21342, "Akamai (CDN2)"},
	{22634, "Tumblr"},
	{22697, "Roblox (secondary)"},
	{24319, "Valve/Steam (chat)"},
	{25562, "Riot Games (EU)"},
	{26008, "Nintendo"},
	{29447, "Valve/Steam (EU2)"},
	{31108, "Valve/Steam (US)"},
	{31223, "Roblox (EU)"},
	{32590, "Valve/Steam (relay2)"},
	{32787, "Akamai (CDN3)"},
	{33353, "Roblox (US)"},
	{33572, "Valve/Steam (HK)"},
	{33905, "Akamai (CDN4)"},
	{34164, "Valve/Steam (EU3)"},
	{35540, "Valve/Steam (JP)"},
	{35834, "Valve/Steam (AU)"},
	{35994, "Akamai (CDN5)"},
	{36183, "Akamai (CDN6)"},
	{37153, "Valve/Steam (ZA)"},
	{38118, "Valve/Steam (SG)"},
	{42708, "Portlane"},
	{44907, "Valve/Steam (CL)"},
	{46407, "Valve/Steam (PL)"},
	{46555, "Valve/Steam (BR)"},
	{49544, "Valve/Steam (BR2)"},
	{49846, "Valve/Steam (IN)"},
	{51167, "Contabo"},
	{53667, "Ponynet/FranTech"},
	{54253, "Valve/Steam (KR)"},
	{54265, "Roblox (CDN)"},
	{55743, "Valve/Steam (AR)"},
	{56630, "Valve/Steam (PE)"},
	{57976, "Blizzard Entertainment"},
	{58061, "Valve/Steam (EG)"},
	{59930, "Valve/Steam (AE)"},
	{60068, "CDN77/Datacamp"},
	{60220, "Valve/Steam (RU partner)"},
	{62014, "Valve/Steam (TW)"},
	{62567, "DigitalOcean (secondary)"},
	{63023, "Valve/Steam (MY)"},
	{63949, "Akamai (Linode)"},
	{139808, "Valve/Steam (CN)"},
	{141995, "Valve/Steam (IL)"},
	{199524, "Gcore"},
	{200005, "Valve/Steam (FI)"},
	{202023, "Valve/Steam (TR)"},
	{202422, "Valve/Steam (PH)"},
	{203663, "Valve/Steam (UK)"},
	{210366, "Valve/Steam (IT)"},
	{210492, "Valve/Steam (SE)"},
	{210644, "Valve/Steam (ES)"},
	{212238, "Cloudflare (secondary)"},
	{212317, "Valve/Steam (FR)"},
	{213120, "Akamai (CDN7)"},
	{213230, "Valve/Steam (CZ)"},
	{215859, "Valve/Steam (ID)"},
	{216183, "Valve/Steam (TH)"},
	{216246, "Valve/Steam (VN)"},
	{393234, "Valve/Steam (PoP1)"},
	{393349, "Valve/Steam (PoP2)"},
	{393560, "Valve/Steam (PoP3)"},
	{394073, "Valve/Steam (PoP4)"},
	{397077, "Valve/Steam (PoP5)"},
	{399358, "Valve/Steam (PoP6)"},
}

const ripeURL = "https://stat.ripe.net/data/announced-prefixes/data.json?resource=AS%d"

// Init fetches all ASN prefixes via RIPE API concurrently
func Init() ([]netip.Prefix, []output.Source) {
	type result struct {
		prefixes []netip.Prefix
		source   output.Source
	}

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results []result
		sem     = make(chan struct{}, 10) // limit concurrency to 10
	)

	for _, a := range asnList {
		wg.Add(1)
		go func(asn int, name string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			url := fmt.Sprintf(ripeURL, asn)
			data, err := fetcher.Get(url)
			if err != nil {
				fmt.Printf("[asn] WARN AS%d (%s): %v\n", asn, name, err)
				mu.Lock()
				results = append(results, result{
					source: output.Source{
						Name:   fmt.Sprintf("AS%d (%s)", asn, name),
						URL:    url,
						Status: "error",
					},
				})
				mu.Unlock()
				return
			}

			prefixes := parser.RIPEPrefixes(data)
			src := output.Source{
				Name:    fmt.Sprintf("AS%d (%s)", asn, name),
				URL:     url,
				Entries: len(prefixes),
				Status:  "ok",
			}

			mu.Lock()
			results = append(results, result{prefixes: prefixes, source: src})
			mu.Unlock()

			fmt.Printf("[asn] AS%d (%s): %d prefixes\n", asn, name, len(prefixes))
		}(a.asn, a.name)
	}

	wg.Wait()

	var all []netip.Prefix
	var meta []output.Source
	for _, r := range results {
		all = append(all, r.prefixes...)
		meta = append(meta, r.source)
	}

	return all, meta
}
