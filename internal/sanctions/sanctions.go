package sanctions

import (
	"fmt"
	"net/netip"

	"github.com/eduard256/russia-blocked-ips/pkg/fetcher"
	"github.com/eduard256/russia-blocked-ips/pkg/output"
	"github.com/eduard256/russia-blocked-ips/pkg/parser"
)

var sources = []struct {
	name string
	url  string
}{
	// Spamhaus DROP -- hijacked/criminal subnets
	{"Spamhaus/DROP", "https://www.spamhaus.org/drop/drop.txt"},

	// Tor exit nodes
	{"Tor/exit-nodes", "https://check.torproject.org/torbulkexitlist"},
}

// Init fetches sanctions-related and Tor sources
func Init() ([]netip.Prefix, []output.Source) {
	var all []netip.Prefix
	var meta []output.Source

	for _, s := range sources {
		data, err := fetcher.Get(s.url)
		if err != nil {
			fmt.Printf("[sanctions] WARN %s: %v\n", s.name, err)
			meta = append(meta, output.Source{Name: s.name, URL: s.url, Status: "error"})
			continue
		}

		prefixes := parser.PlainText(data)
		all = append(all, prefixes...)
		meta = append(meta, output.Source{Name: s.name, URL: s.url, Entries: len(prefixes), Status: "ok"})
		fmt.Printf("[sanctions] %s: %d entries\n", s.name, len(prefixes))
	}

	return all, meta
}
