package community

import (
	"fmt"
	"net/netip"

	"github.com/eduard256/russia-blocked-ips/pkg/fetcher"
	"github.com/eduard256/russia-blocked-ips/pkg/output"
	"github.com/eduard256/russia-blocked-ips/pkg/parser"
)

// community.antifilter.download is dead (404)
// these are other community-maintained sources
var sources = []struct {
	name string
	url  string
}{
	// V3nilla -- combined ipset for bypass in Russia
	// already included in services, skipping to avoid double-counting
	// {"V3nilla/ipset-all.txt", "https://raw.githubusercontent.com/V3nilla/IPSets-For-Bypass-in-Russia/refs/heads/main/ipset-all.txt"},

	// escapingworm -- Russian mobile operator whitelist CIDRs
	// NOTE: these are RU-internal IPs that should NOT go through VPN
	// included for reference but client programs should use this as exclusion list
	{"escapingworm/ru-whitelist-cidr.txt", "https://raw.githubusercontent.com/escapingworm/russia-whitelist/refs/heads/main/ru-whitelist-cidr.txt"},
}

// Init fetches community sources
func Init() ([]netip.Prefix, []output.Source) {
	var all []netip.Prefix
	var meta []output.Source

	for _, s := range sources {
		data, err := fetcher.Get(s.url)
		if err != nil {
			fmt.Printf("[community] WARN %s: %v\n", s.name, err)
			meta = append(meta, output.Source{Name: s.name, URL: s.url, Status: "error"})
			continue
		}

		prefixes := parser.PlainText(data)

		// NOTE: escapingworm whitelist is NOT added to the main IP list
		// it's fetched for manifest metadata only -- client programs decide how to use it
		meta = append(meta, output.Source{Name: s.name, URL: s.url, Entries: len(prefixes), Status: "ok"})
		fmt.Printf("[community] %s: %d entries (whitelist, not added to blocklist)\n", s.name, len(prefixes))
	}

	return all, meta
}
