package main

import (
	"strings"
	"testing"
)

func splitFixturePlan(t *testing.T) (plan, map[string]float64) {
	t.Helper()
	p := plan{Count: 15}
	w := map[string]float64{}
	for _, g := range planningAffinities() {
		if g.WholePackage {
			g.Tests = []string{"TestOracle"}
		}
		for _, name := range g.Tests {
			u := unit{Package: g.Package, Test: name}
			p.Units = append(p.Units, u)
			w[u.key()] = 5
		}
		w[g.key()] = 100
	}
	if err := assignUnits(&p, w); err != nil {
		t.Fatal(err)
	}
	return p, w
}

func TestDeclaredLayoutSplitCoverageAndIsolation(t *testing.T) {
	p, w := splitFixturePlan(t)
	counts := map[string]int{}
	for _, g := range p.Affinity {
		if g.Split == "" {
			continue
		}
		if g.SplitCount != 2 {
			t.Fatal("lost split declaration")
		}
		counts[g.Split] = len(g.Tests)
		if len(assignedPackages(p, g.Shard)[g.Package]) != len(g.Tests) {
			t.Fatal("half lost package assignment")
		}
		selectors := selections(p, g.Shard, g.Package)
		if len(selectors) != 1 {
			t.Fatal("half needs one process", selectors)
		}
		for _, name := range g.Tests {
			if !strings.Contains(selectors[0].Run, exactPath(name)) {
				t.Fatal("half lost selector", name, selectors)
			}
		}
		for _, other := range p.Affinity {
			if other.Split == "" || other.Split == g.Split {
				continue
			}
			for _, name := range other.Tests {
				if strings.Contains(selectors[0].Run, exactPath(name)) {
					t.Fatal("half runs other half's mutations", name)
				}
			}
		}
	}
	if counts["half-a"] != 4 || counts["half-b"] != 5 {
		t.Fatal("lost layout coverage", counts)
	}
	total := 0.0
	for _, prediction := range predictions(p, w) {
		total += prediction.Seconds
	}
	if total != 400 {
		t.Fatal("did not charge one fixture per half", total)
	}
	for _, u := range p.Units {
		if u.Test == "TestMarkdownWhitespaceLayout" {
			if u.Shard > 1 {
				t.Fatal("half not dedicated")
			}
		}
	}
}

func TestDeclaredLayoutSplitRefusesCorruption(t *testing.T) {
	for _, mutation := range []string{"shared-shard", "missing-half", "lost-test", "extra-test", "unrelated-work", "wrong-count", "invalid-shard", "undeclared-split"} {
		t.Run(mutation, func(t *testing.T) {
			p, _ := splitFixturePlan(t)
			a, b := 1, 2
			switch mutation {
			case "shared-shard":
				p.Affinity[b].Shard = p.Affinity[a].Shard
			case "missing-half":
				p.Affinity = append(p.Affinity[:b], p.Affinity[b+1:]...)
			case "lost-test":
				p.Affinity[a].Tests = p.Affinity[a].Tests[1:]
			case "extra-test":
				p.Affinity[a].Tests = append(p.Affinity[a].Tests, p.Affinity[b].Tests[0])
			case "unrelated-work":
				p.Units = append(p.Units, unit{Package: "unrelated", Test: "TestOther", Shard: p.Affinity[a].Shard})
			case "undeclared-split":
				p.Affinity[a].Split, p.Affinity[b].Split = "", ""
				p.Affinity[a].SplitCount, p.Affinity[b].SplitCount = 0, 0
			case "invalid-shard":
				p.Affinity[a].Shard = p.Count
			case "wrong-count":
				p.Affinity[a].SplitCount = 3
			}
			if validateAffinity(p) == nil {
				t.Fatal("corrupt declared split accepted", mutation)
			}
		})
	}
}

func TestDedicatedLayoutShardsStayReservedBeforeHeavyOrdinaryItems(t *testing.T) {
	p, w := splitFixturePlan(t)
	for _, g := range p.Affinity {
		if g.Split == "" {
			w[g.key()] = 10000
		}
	}
	if err := assignUnits(&p, w); err != nil {
		t.Fatal(err)
	}
	for _, u := range p.Units {
		if u.Shard < 2 && u.Package != "github.com/system-inc/adamic/stage1/cohere/markdownblocks" {
			t.Fatal("ordinary item occupied dedicated half", u)
		}
	}
}
