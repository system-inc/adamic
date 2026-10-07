package main

import "testing"

func TestKnownAffinitiesStayWhole(t *testing.T) {
	p := plan{Count: 15}
	weights := map[string]float64{}
	groups := knownAffinities()
	if len(groups) != 2 || len(groups[0].Tests) != 9 || len(groups[1].Tests) != 3 || groups[0].SetupSeconds <= 0 || groups[1].SetupSeconds <= 0 {
		t.Fatal("lost fixture inventory", groups)
	}
	for _, g := range groups {
		if g.WholePackage {
			g.Tests = []string{"TestParent/a", "TestParent/b", "TestOther"}
		}
		for i, name := range g.Tests {
			u := unit{Package: g.Package, Test: name}
			p.Units = append(p.Units, u)
			weights[u.key()] = float64(100 + i)
		}
		weights[g.key()] = 123
	}
	if err := assignUnits(&p, weights); err != nil {
		t.Fatal(err)
	}
	if len(p.Affinity) != 2 {
		t.Fatal("missing affinity in plan")
	}
	for _, g := range groups {
		shard := -1
		for _, u := range p.Units {
			if u.Package == g.Package {
				for _, name := range g.Tests {
					if u.Test == name {
						if shard < 0 {
							shard = u.Shard
						}
						if shard != u.Shard {
							t.Fatalf("%s split", g.Fixture)
						}
					}
				}
			}
		}
	}
	predictions := predictions(p, weights)
	total := 0.0
	for _, s := range predictions {
		total += s.Seconds
	}
	if total != 246 {
		t.Fatalf("shared fixture priced more than once: %v", total)
	}
	p.Units[0].Shard = (p.Units[0].Shard + 1) % p.Count
	if validateAffinity(p) == nil {
		t.Fatal("accepted split fixture")
	}
}

func TestVolumeParentStaysWhole(t *testing.T) {
	names, err := children("github.com/system-inc/adamic/stage1/cohere/typeaware", "TestVolumeAgreementAndMutants")
	if err != nil || names != nil {
		t.Fatalf("volume parent split: %v %v", names, err)
	}
}

func TestPartialAffinityRejected(t *testing.T) {
	g := knownAffinities()[0]
	p := plan{Count: 2, Units: []unit{{Package: g.Package, Test: g.Tests[0]}}}
	if assignUnits(&p, nil) == nil {
		t.Fatal("accepted incomplete fixture inventory")
	}
}

func TestOracleMeasuredTestsMaySpread(t *testing.T) {
	const pkg = "github.com/system-inc/adamic/internal/oracle"
	for _, g := range knownAffinities() {
		if g.Package == pkg {
			t.Fatal("sub-second gateOnce must not pin the oracle", g)
		}
	}
	p := plan{Count: 2, Units: []unit{{Package: pkg, Test: "TestCountsAreRecorded"}, {Package: pkg, Test: "TestLoopCountersAgreeWithNode"}}}
	weights := map[string]float64{p.Units[0].key(): 311.24, p.Units[1].key(): 79.31}
	if err := assignUnits(&p, weights); err != nil {
		t.Fatal(err)
	}
	if p.Units[0].Shard == p.Units[1].Shard {
		t.Fatal("oracle's independent tests remain pinned")
	}
}
