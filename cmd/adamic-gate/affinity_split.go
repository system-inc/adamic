package main

import "fmt"

// Whole top-level tests retain their original mutation checks and fixture
// initialization. Each declared half runs in one process, so layoutOnce builds
// once per half. The halves never share a shard or unrelated package work.
func planningAffinities() []affinity {
	var out []affinity
	for _, group := range knownAffinities() {
		if group.Fixture != "layoutOnce" {
			out = append(out, group)
			continue
		}
		for _, half := range layoutHalves() {
			out = append(out, affinity{Package: group.Package, Fixture: group.Fixture, Tests: half.Tests, Split: half.Split, SplitCount: 2})
		}
	}
	return out
}

func layoutHalves() []affinity {
	return []affinity{
		{Split: "half-a", Tests: []string{"TestMarkdownCodeBlockLayout", "TestMarkdownLeafComposition", "TestMarkdownListLayout", "TestMarkdownWhitespaceLayout"}},
		{Split: "half-b", Tests: []string{"TestMarkdownHTMLBlockLayout", "TestMarkdownQuoteLayout", "TestMarkdownRootLayout", "TestMarkdownStructureLayout", "TestMarkdownTableLayout"}},
	}
}

func validateDeclaredSplits(p plan) error {
	var halves []affinity
	for _, group := range p.Affinity {
		if group.Split == "" {
			if group.SplitCount != 0 {
				return fmt.Errorf("affinity %s missing split name", group.key())
			}
			continue
		}
		if group.Fixture != "layoutOnce" || group.Package != knownAffinities()[1].Package || group.SplitCount != 2 {
			return fmt.Errorf("unknown declared affinity split %s", group.key())
		}
		if group.Shard < 0 || group.Shard >= p.Count {
			return fmt.Errorf("layout half %s has invalid shard", group.Split)
		}
		halves = append(halves, group)
	}
	if len(halves) == 0 {
		// Historical plans may carry one whole fixture, but must not hide a split
		// by dropping the declaration from two partial affinity records.
		layouts := []affinity{}
		for _, group := range p.Affinity {
			if group.Fixture == "layoutOnce" {
				layouts = append(layouts, group)
			}
		}
		if len(layouts) > 1 {
			return fmt.Errorf("layout affinity split lacks declaration")
		}
		if len(layouts) == 1 {
			expected := knownAffinities()[1]
			if len(layouts[0].Tests) != len(expected.Tests) {
				return fmt.Errorf("undeclared partial layout affinity")
			}
			members := map[string]bool{}
			for _, name := range layouts[0].Tests {
				members[name] = true
			}
			for _, name := range expected.Tests {
				if !members[name] {
					return fmt.Errorf("undeclared layout affinity missing %s", name)
				}
			}
		}
		return nil // Historical unsplit plans remain readable.
	}
	if len(halves) != 2 {
		return fmt.Errorf("layout affinity requires exactly two declared halves")
	}
	if halves[0].Shard == halves[1].Shard {
		return fmt.Errorf("layout affinity halves share a shard")
	}
	names := map[string]bool{}
	for _, expected := range layoutHalves() {
		found := false
		for _, half := range halves {
			if half.Split != expected.Split {
				continue
			}
			if found {
				return fmt.Errorf("duplicate layout half %s", half.Split)
			}
			found = true
			if len(half.Tests) != len(expected.Tests) {
				return fmt.Errorf("layout half %s coverage differs", half.Split)
			}
			members := map[string]bool{}
			for _, name := range half.Tests {
				if members[name] || names[name] {
					return fmt.Errorf("duplicate layout split member %s", name)
				}
				members[name] = true
				names[name] = true
			}
			for _, name := range expected.Tests {
				if !members[name] {
					return fmt.Errorf("layout half %s missing %s", half.Split, name)
				}
			}
			for _, u := range p.Units {
				if u.Shard == half.Shard && (u.Package != half.Package || !members[u.Test]) {
					return fmt.Errorf("layout half %s shard contains unrelated unit %s", half.Split, u.key())
				}
			}
		}
		if !found {
			return fmt.Errorf("missing declared layout half %s", expected.Split)
		}
	}
	return nil
}

func splitSuffix(group affinity) string {
	if group.Split == "" {
		return ""
	}
	return "/" + group.Split
}

// Select all existing and future subtests of a half's assigned parents. Keeping
// parents whole avoids duplicating their unguarded policy work across halves.
func affinityRunPath(p plan, u unit) string {
	for _, g := range p.Affinity {
		if g.Split == "" || g.Package != u.Package {
			continue
		}
		for _, name := range g.Tests {
			if name == u.Test {
				return exactPath(name) + "/.*"
			}
		}
	}
	return exactPath(u.Test)
}
