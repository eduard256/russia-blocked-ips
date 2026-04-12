package rkn

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
	format string // "plain", "csv", "json", "mixed"
}{
	// antifilter.download -- primary
	{"antifilter.download/ip.lst", "https://antifilter.download/list/ip.lst", "plain"},
	{"antifilter.download/subnet.lst", "https://antifilter.download/list/subnet.lst", "plain"},
	{"antifilter.download/ipsum.lst", "https://antifilter.download/list/ipsum.lst", "plain"},
	{"antifilter.download/allyouneed.lst", "https://antifilter.download/list/allyouneed.lst", "plain"},

	// antifilter.network -- mirrors
	{"antifilter.network/ip.lst", "https://antifilter.network/download/ip.lst", "plain"},
	{"antifilter.network/ipsum.lst", "https://antifilter.network/download/ipsum.lst", "plain"},
	{"antifilter.network/subnet.lst", "https://antifilter.network/download/subnet.lst", "plain"},

	// zapret-info -- original RKN dump
	{"zapret-info/dump-00.csv", "https://raw.githubusercontent.com/zapret-info/z-i/master/dump-00.csv", "csv"},

	// rublacklist -- stale since 2022 but still has data
	{"rublacklist/ips.json", "https://reestr.rublacklist.net/api/v2/ips/json/", "json"},

	// bol-van/rulist -- aggregated RKN dump
	{"bol-van/rulist/reestr_ipban4.txt", "https://raw.githubusercontent.com/bol-van/rulist/refs/heads/main/reestr_ipban4.txt", "plain"},
	{"bol-van/rulist/reestr_smart4.txt", "https://raw.githubusercontent.com/bol-van/rulist/refs/heads/main/reestr_smart4.txt", "plain"},

	// Re-filter-lists -- RKN + OONI
	{"Re-filter/ipsum.lst", "https://raw.githubusercontent.com/1andrevich/Re-filter-lists/refs/heads/main/ipsum.lst", "plain"},
}

// Init fetches all RKN sources and returns merged prefixes + source metadata
func Init() ([]netip.Prefix, []output.Source) {
	var all []netip.Prefix
	var meta []output.Source

	for _, s := range sources {
		data, err := fetcher.Get(s.url)
		if err != nil {
			fmt.Printf("[rkn] WARN %s: %v\n", s.name, err)
			meta = append(meta, output.Source{Name: s.name, URL: s.url, Status: "error"})
			continue
		}

		var prefixes []netip.Prefix
		switch s.format {
		case "plain":
			prefixes = parser.PlainText(data)
		case "csv":
			prefixes = parser.CSVColumn(data, 0, ';')
		case "json":
			prefixes = parser.RublacklistJSON(data)
		}

		all = append(all, prefixes...)
		meta = append(meta, output.Source{Name: s.name, URL: s.url, Entries: len(prefixes), Status: "ok"})
		fmt.Printf("[rkn] %s: %d entries\n", s.name, len(prefixes))
	}

	return all, meta
}
