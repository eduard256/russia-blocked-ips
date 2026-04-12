package aggregate

import (
	"cmp"
	"net/netip"
	"slices"
)

// Merge deduplicates, absorbs contained prefixes, and merges adjacent CIDRs
func Merge(prefixes []netip.Prefix) []netip.Prefix {
	if len(prefixes) == 0 {
		return nil
	}

	// normalize: ensure all prefixes are masked (no host bits set)
	for i, p := range prefixes {
		prefixes[i] = p.Masked()
	}

	// sort by address, then by prefix length (shorter first)
	slices.SortFunc(prefixes, func(a, b netip.Prefix) int {
		if c := a.Addr().Compare(b.Addr()); c != 0 {
			return c
		}
		return cmp.Compare(a.Bits(), b.Bits())
	})

	// pass 1: remove duplicates and absorb contained prefixes
	result := make([]netip.Prefix, 0, len(prefixes)/2)
	for _, p := range prefixes {
		if !p.IsValid() {
			continue
		}
		// skip if contained in the last added prefix
		if len(result) > 0 && result[len(result)-1].Contains(p.Addr()) && result[len(result)-1].Bits() <= p.Bits() {
			continue
		}
		result = append(result, p)
	}

	// pass 2: merge adjacent siblings repeatedly until stable
	for {
		merged := mergePass(result)
		if len(merged) == len(result) {
			break
		}
		result = merged
	}

	return result
}

// mergePass tries to merge adjacent CIDR pairs into their parent
func mergePass(prefixes []netip.Prefix) []netip.Prefix {
	if len(prefixes) < 2 {
		return prefixes
	}

	result := make([]netip.Prefix, 0, len(prefixes))
	i := 0
	for i < len(prefixes) {
		if i+1 < len(prefixes) {
			if merged, ok := mergePair(prefixes[i], prefixes[i+1]); ok {
				result = append(result, merged)
				i += 2
				continue
			}
		}
		result = append(result, prefixes[i])
		i++
	}
	return result
}

// mergePair merges two adjacent prefixes of same length into parent prefix
// ex: 10.0.0.0/25 + 10.0.0.128/25 = 10.0.0.0/24
func mergePair(a, b netip.Prefix) (netip.Prefix, bool) {
	if a.Bits() != b.Bits() || a.Bits() == 0 {
		return netip.Prefix{}, false
	}
	if a.Addr().Is4() != b.Addr().Is4() {
		return netip.Prefix{}, false
	}

	parent, err := a.Addr().Prefix(a.Bits() - 1)
	if err != nil {
		return netip.Prefix{}, false
	}

	if parent.Contains(a.Addr()) && parent.Contains(b.Addr()) {
		return parent, true
	}
	return netip.Prefix{}, false
}
