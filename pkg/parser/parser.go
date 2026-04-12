package parser

import (
	"encoding/csv"
	"encoding/json"
	"net/netip"
	"strings"
)

// PlainText parses one IP or CIDR per line, skips comments and empty lines
func PlainText(data []byte) []netip.Prefix {
	var result []netip.Prefix
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' || line[0] == ';' {
			continue
		}
		// strip inline comments: "1.2.3.0/24 ; SBL123"
		if i := strings.IndexAny(line, ";#"); i > 0 {
			line = strings.TrimSpace(line[:i])
		}
		if p := parsePrefix(line); p.IsValid() {
			result = append(result, p)
		}
	}
	return result
}

// Mixed parses iamwildtuna-style format: comments with //, IPs comma-separated, CIDR, empty lines
func Mixed(data []byte) []netip.Prefix {
	var result []netip.Prefix
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		// skip section headers (lines without dots are not IPs)
		if !strings.Contains(line, ".") && !strings.Contains(line, ":") {
			continue
		}
		// handle comma-separated IPs on one line
		for _, part := range strings.Split(line, ",") {
			part = strings.TrimSpace(part)
			if p := parsePrefix(part); p.IsValid() {
				result = append(result, p)
			}
		}
	}
	return result
}

// CSVColumn parses CSV and extracts IPs/CIDRs from column at given index
func CSVColumn(data []byte, col int, sep rune) []netip.Prefix {
	r := csv.NewReader(strings.NewReader(string(data)))
	r.Comma = sep
	r.LazyQuotes = true
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return nil
	}

	var result []netip.Prefix
	for _, rec := range records {
		if col >= len(rec) {
			continue
		}
		field := strings.TrimSpace(rec[col])
		if p := parsePrefix(field); p.IsValid() {
			result = append(result, p)
		}
	}
	return result
}

// AWSRanges parses AWS ip-ranges.json format
func AWSRanges(data []byte) []netip.Prefix {
	var doc struct {
		Prefixes []struct {
			Prefix string `json:"ip_prefix"`
		} `json:"prefixes"`
		V6 []struct {
			Prefix string `json:"ipv6_prefix"`
		} `json:"ipv6_prefixes"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, p := range doc.Prefixes {
		if pfx := parsePrefix(p.Prefix); pfx.IsValid() {
			result = append(result, pfx)
		}
	}
	for _, p := range doc.V6 {
		if pfx := parsePrefix(p.Prefix); pfx.IsValid() {
			result = append(result, pfx)
		}
	}
	return result
}

// GoogleRanges parses Google goog.json / cloud.json format
func GoogleRanges(data []byte) []netip.Prefix {
	var doc struct {
		Prefixes []struct {
			V4 string `json:"ipv4Prefix"`
			V6 string `json:"ipv6Prefix"`
		} `json:"prefixes"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, p := range doc.Prefixes {
		if p.V4 != "" {
			if pfx := parsePrefix(p.V4); pfx.IsValid() {
				result = append(result, pfx)
			}
		}
		if p.V6 != "" {
			if pfx := parsePrefix(p.V6); pfx.IsValid() {
				result = append(result, pfx)
			}
		}
	}
	return result
}

// FastlyRanges parses Fastly public-ip-list JSON
func FastlyRanges(data []byte) []netip.Prefix {
	var doc struct {
		Addresses   []string `json:"addresses"`
		IPv6        []string `json:"ipv6_addresses"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, s := range append(doc.Addresses, doc.IPv6...) {
		if p := parsePrefix(s); p.IsValid() {
			result = append(result, p)
		}
	}
	return result
}

// AzureServiceTags parses Azure ServiceTags JSON
func AzureServiceTags(data []byte) []netip.Prefix {
	var doc struct {
		Values []struct {
			Properties struct {
				Prefixes []string `json:"addressPrefixes"`
			} `json:"properties"`
		} `json:"values"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, v := range doc.Values {
		for _, s := range v.Properties.Prefixes {
			if p := parsePrefix(s); p.IsValid() {
				result = append(result, p)
			}
		}
	}
	return result
}

// Office365Endpoints parses Office 365 endpoints JSON
func Office365Endpoints(data []byte) []netip.Prefix {
	var doc []struct {
		IPs []string `json:"ips"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, ep := range doc {
		for _, s := range ep.IPs {
			if p := parsePrefix(s); p.IsValid() {
				result = append(result, p)
			}
		}
	}
	return result
}

// OracleCloudRanges parses Oracle Cloud public_ip_ranges.json
func OracleCloudRanges(data []byte) []netip.Prefix {
	var doc struct {
		Regions []struct {
			CIDRs []struct {
				CIDR string `json:"cidr"`
			} `json:"cidrs"`
		} `json:"regions"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, r := range doc.Regions {
		for _, c := range r.CIDRs {
			if p := parsePrefix(c.CIDR); p.IsValid() {
				result = append(result, p)
			}
		}
	}
	return result
}

// GitHubMeta parses api.github.com/meta JSON -- all CIDR arrays
func GitHubMeta(data []byte) []netip.Prefix {
	var doc map[string]json.RawMessage
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, raw := range doc {
		var arr []string
		if json.Unmarshal(raw, &arr) != nil {
			continue
		}
		for _, s := range arr {
			if p := parsePrefix(s); p.IsValid() {
				result = append(result, p)
			}
		}
	}
	return result
}

// RIPEPrefixes parses RIPE stat announced-prefixes JSON
func RIPEPrefixes(data []byte) []netip.Prefix {
	var doc struct {
		Data struct {
			Prefixes []struct {
				Prefix string `json:"prefix"`
			} `json:"prefixes"`
		} `json:"data"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, p := range doc.Data.Prefixes {
		if pfx := parsePrefix(p.Prefix); pfx.IsValid() {
			result = append(result, pfx)
		}
	}
	return result
}

// RIPECountryPrefixes parses RIPE country-resource-list JSON for IPv4 prefixes
func RIPECountryPrefixes(data []byte) []netip.Prefix {
	var doc struct {
		Data struct {
			Resources struct {
				IPv4 []string `json:"ipv4"`
			} `json:"resources"`
		} `json:"data"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, s := range doc.Data.Resources.IPv4 {
		if p := parsePrefix(s); p.IsValid() {
			result = append(result, p)
		}
	}
	return result
}

// SpamhausDROP parses Spamhaus DROP text format: "CIDR ; SBLxxx"
func SpamhausDROP(data []byte) []netip.Prefix {
	return PlainText(data) // same format -- lines with ; comments
}

// RublacklistJSON parses rublacklist API JSON -- array of IP strings
func RublacklistJSON(data []byte) []netip.Prefix {
	var ips []string
	if json.Unmarshal(data, &ips) != nil {
		return nil
	}

	var result []netip.Prefix
	for _, s := range ips {
		if p := parsePrefix(strings.TrimSpace(s)); p.IsValid() {
			result = append(result, p)
		}
	}
	return result
}

// internals

// parsePrefix tries CIDR first, then single IP (adds /32 or /128)
func parsePrefix(s string) netip.Prefix {
	s = strings.TrimSpace(s)
	if s == "" {
		return netip.Prefix{}
	}
	if p, err := netip.ParsePrefix(s); err == nil {
		return p
	}
	if addr, err := netip.ParseAddr(s); err == nil {
		return netip.PrefixFrom(addr, addr.BitLen())
	}
	return netip.Prefix{}
}
