package main

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/eduard256/russia-blocked-ips/internal/asn"
	"github.com/eduard256/russia-blocked-ips/internal/cdn"
	"github.com/eduard256/russia-blocked-ips/internal/community"
	"github.com/eduard256/russia-blocked-ips/internal/rkn"
	"github.com/eduard256/russia-blocked-ips/internal/sanctions"
	"github.com/eduard256/russia-blocked-ips/internal/services"
	"github.com/eduard256/russia-blocked-ips/pkg/aggregate"
	"github.com/eduard256/russia-blocked-ips/pkg/fetcher"
	"github.com/eduard256/russia-blocked-ips/pkg/output"
	"github.com/eduard256/russia-blocked-ips/pkg/parser"
)

type module struct {
	name string
	init func() ([]netip.Prefix, []output.Source)
}

var modules = []module{
	{"rkn", rkn.Init},
	{"services", services.Init},
	{"cdn", cdn.Init},
	{"asn", asn.Init},
	{"sanctions", sanctions.Init},
	{"community", community.Init},
}

func main() {
	start := time.Now()
	fmt.Println("=== russia-blocked-ips updater ===")
	fmt.Println()

	var allPrefixes []netip.Prefix
	var allSources []output.Source

	for _, m := range modules {
		fmt.Printf("--- [%s] fetching ---\n", m.name)
		prefixes, sources := m.init()
		allPrefixes = append(allPrefixes, prefixes...)
		allSources = append(allSources, sources...)
		fmt.Printf("--- [%s] done: %d prefixes ---\n\n", m.name, len(prefixes))
	}

	fmt.Printf("total raw prefixes: %d\n", len(allPrefixes))
	fmt.Println("aggregating...")

	merged := aggregate.Merge(allPrefixes)
	fmt.Printf("after dedup/merge: %d prefixes\n", len(merged))

	// exclude Russian IP ranges -- no point routing them through VPN
	fmt.Println("fetching RU IP ranges from RIPE...")
	ruPrefixes := fetchRUPrefixes()
	if len(ruPrefixes) > 0 {
		before := len(merged)
		merged = aggregate.Exclude(merged, ruPrefixes)
		fmt.Printf("excluded %d RU prefixes, %d remaining\n", before-len(merged), len(merged))
	}

	// write ip.txt
	if err := output.WriteIPFile("ip.txt", merged); err != nil {
		fmt.Printf("FATAL: write ip.txt: %v\n", err)
		return
	}
	fmt.Println("wrote ip.txt")

	// write manifest.json
	if err := output.WriteManifest("manifest.json", "ip.txt", allSources); err != nil {
		fmt.Printf("FATAL: write manifest.json: %v\n", err)
		return
	}
	fmt.Println("wrote manifest.json")

	fmt.Printf("\ndone in %s\n", time.Since(start).Round(time.Second))
}

const ripeRUURL = "https://stat.ripe.net/data/country-resource-list/data.json?resource=RU&v4_format=prefix"

func fetchRUPrefixes() []netip.Prefix {
	data, err := fetcher.Get(ripeRUURL)
	if err != nil {
		fmt.Printf("WARN: could not fetch RU prefixes: %v\n", err)
		return nil
	}
	prefixes := parser.RIPECountryPrefixes(data)
	fmt.Printf("loaded %d RU prefixes\n", len(prefixes))
	return prefixes
}
