package main

import (
	"fmt"
	"sort"
)

func sortPackages(p plan, index int, packages []string) {
	sort.Slice(packages, func(i, j int) bool {
		a, b := p.PackageSeconds[packagePredictionKey(index, packages[i])], p.PackageSeconds[packagePredictionKey(index, packages[j])]
		if a != b {
			return a > b
		}
		return packages[i] < packages[j]
	})
}

func packagePredictionKey(index int, pkg string) string { return fmt.Sprintf("%d::%s", index, pkg) }

// Reuse the same affinity and parent-residual accounting as shard predictions.
func packagePredictions(p plan, weights map[string]float64) map[string]float64 {
	packages := map[string]bool{}
	for _, u := range p.Units {
		packages[u.Package] = true
	}
	for _, c := range p.Complements {
		packages[c.Package] = true
	}
	out := map[string]float64{}
	for pkg := range packages {
		part := plan{Count: p.Count}
		for _, u := range p.Units {
			if u.Package == pkg {
				part.Units = append(part.Units, u)
			}
		}
		for _, g := range p.Affinity {
			if g.Package == pkg {
				part.Affinity = append(part.Affinity, g)
			}
		}
		for _, c := range p.Complements {
			if c.Package == pkg {
				part.Complements = append(part.Complements, c)
			}
		}
		for _, r := range predictions(part, weights) {
			if _, assigned := assignedPackages(part, r.Shard)[pkg]; assigned {
				out[packagePredictionKey(r.Shard, pkg)] = r.Seconds
			}
		}
	}
	return out
}
