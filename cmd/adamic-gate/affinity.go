package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"
)

// Tests remain separate coverage identities, but a shared fixture is one packing unit.
type affinity struct {
	Package, Fixture string
	WholePackage     bool
	SetupSeconds     float64
	SetupMeasurement string
	Tests            []string
	Shard            int
	Seconds          float64
	Unknown          int
}

func (a affinity) key() string { return a.Package + "::@" + a.Fixture }

func knownAffinities() []affinity {
	return []affinity{
		{Package: "github.com/system-inc/adamic/stage1/cohere/markdownblocks", Fixture: "layoutOnce", SetupSeconds: 448.889, SetupMeasurement: "33935158 joint rerun: cont to original-oracle agreement; fixture plus controls, uncached", Tests: []string{
			"TestMarkdownCodeBlockLayout", "TestMarkdownHTMLBlockLayout", "TestMarkdownLeafComposition", "TestMarkdownListLayout", "TestMarkdownQuoteLayout", "TestMarkdownRootLayout", "TestMarkdownStructureLayout", "TestMarkdownTableLayout", "TestMarkdownWhitespaceLayout",
		}},
		{Package: "github.com/system-inc/adamic/stage1/cohere/markdownblocks", Fixture: "formatterOnce", SetupSeconds: 206.546467590, SetupMeasurement: "1abec4c7: first Go formatter build on prepared box; formatter dependencies cold (warm repeats 7.789 and 5.701 s)", Tests: []string{
			"TestMarkdownSourceDecoding", "TestMarkdownTextSplitting", "TestMarkdownUnicodeWidths",
		}},
	}
}

type packingUnit struct {
	key             string
	indices         []int
	seconds         float64
	known, required bool
	affinity        int
}

func assignUnits(p *plan, weights map[string]float64) error {
	indices := map[string]int{}
	for i := range p.Units {
		u := &p.Units[i]
		if _, exists := indices[u.key()]; exists {
			return fmt.Errorf("duplicate planned unit %s", u.key())
		}
		indices[u.key()] = i
		if seconds, ok := weights[u.key()]; ok {
			if seconds < 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
				return fmt.Errorf("invalid timing %s", u.key())
			}
			u.Seconds = seconds
		}
	}
	grouped := map[int]bool{}
	var items []packingUnit
	p.Affinity = nil
	for _, group := range knownAffinities() {
		if group.WholePackage {
			for _, u := range p.Units {
				if u.Package == group.Package {
					group.Tests = append(group.Tests, u.Test)
				}
			}
			sort.Strings(group.Tests)
		}
		item := packingUnit{key: group.key(), known: true, affinity: len(p.Affinity)}
		for _, test := range group.Tests {
			i, found := indices[group.Package+"::"+test]
			if !found {
				continue
			}
			if grouped[i] {
				return fmt.Errorf("overlapping affinity %s", group.key())
			}
			grouped[i] = true
			item.indices = append(item.indices, i)
			item.seconds += p.Units[i].Seconds
			item.required = item.required || p.Units[i].WASI || len(p.Units[i].RequiredEnvironment) > 0
			if _, ok := weights[p.Units[i].key()]; !ok {
				group.Unknown++
				item.known = false
			}
		}
		if len(item.indices) == 0 {
			continue
		}
		if len(item.indices) != len(group.Tests) {
			return fmt.Errorf("affinity %s has missing tests", group.key())
		}
		// Calibration can record a joint span, or reconstruct one with shared setup once.
		// Without it, summing observed members is a conservative upper estimate.
		if seconds, ok := weights[group.key()]; ok {
			if seconds < 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
				return fmt.Errorf("invalid timing %s", group.key())
			}
			item.seconds = seconds
			item.known = true
		}
		group.Seconds = item.seconds
		p.Affinity = append(p.Affinity, group)
		items = append(items, item)
	}
	for i, u := range p.Units {
		if grouped[i] || u.key() == archiveUnit {
			continue
		}
		_, known := weights[u.key()]
		items = append(items, packingUnit{key: u.key(), indices: []int{i}, seconds: u.Seconds, known: known, required: u.WASI || len(u.RequiredEnvironment) > 0, affinity: -1})
	}
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.known != b.known {
			return a.known
		}
		if a.seconds != b.seconds {
			return a.seconds > b.seconds
		}
		return a.key < b.key
	})
	ordinary := p.Count
	if p.Environment != nil && ordinary > 1 {
		ordinary--
	}
	loads := make([]float64, p.Count)
	for _, item := range items {
		best := 0
		if item.required {
			best = p.Count - 1
		} else if !item.known {
			hash := sha256.Sum256([]byte(item.key))
			best = int(binary.BigEndian.Uint64(hash[:8]) % uint64(ordinary))
		} else {
			for j := 1; j < ordinary; j++ {
				if loads[j] < loads[best] {
					best = j
				}
			}
		}
		for _, i := range item.indices {
			p.Units[i].Shard = best
		}
		loads[best] += item.seconds
		if item.affinity >= 0 {
			p.Affinity[item.affinity].Shard = best
		}
	}
	return validateAffinity(*p)
}

func validateAffinity(p plan) error {
	for _, group := range p.Affinity {
		for _, c := range p.Complements {
			if c.Package != group.Package {
				continue
			}
			for _, test := range group.Tests {
				if test == c.Parent || strings.HasPrefix(test, c.Parent+"/") {
					if c.Shard != group.Shard {
						return fmt.Errorf("affinity %s complement split: %s", group.key(), c.Parent)
					}
				}
			}
		}
		for _, test := range group.Tests {
			found := false
			for _, u := range p.Units {
				if u.Package == group.Package && u.Test == test {
					found = true
					if u.Shard != group.Shard {
						return fmt.Errorf("affinity %s split: %s in shard %d, want %d", group.key(), test, u.Shard, group.Shard)
					}
				}
			}
			if !found {
				return fmt.Errorf("affinity %s missing %s", group.key(), test)
			}
		}
	}
	return nil
}
