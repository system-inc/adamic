package main

import "testing"

func TestKnownAffinitiesStayWhole(t *testing.T) {
	p := plan{Count: 15}
	weights := map[string]float64{}
	groups := planningAffinities()
	if len(groups) != 4 || !groups[0].WholePackage || len(groups[1].Tests) != 4 || len(groups[2].Tests) != 5 || len(groups[3].Tests) != 3 {
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
	if len(p.Affinity) != 4 {
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
	if total != 492 {
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
	g := knownAffinities()[1]
	p := plan{Count: 2, Units: []unit{{Package: g.Package, Test: g.Tests[0]}}}
	if assignUnits(&p, nil) == nil {
		t.Fatal("accepted incomplete fixture inventory")
	}
}

func TestWholePackageAffinityOwnsComplements(t *testing.T) {
	g := knownAffinities()[0]
	p := plan{Count: 15, Units: []unit{{Package: g.Package, Test: "TestParent/one", RequiredEnvironment: []string{"REQUIRED"}}, {Package: g.Package, Test: "TestParent/two"}, {Package: g.Package, Test: "TestOther"}}}
	w := map[string]float64{g.key(): 100}
	if err := assignUnits(&p, w); err != nil {
		t.Fatal(err)
	}
	p.Complements = complements(p)
	if len(p.Complements) != 1 || p.Complements[0].Shard != 14 {
		t.Fatal("fixture complement moved to another shard", p.Complements)
	}
	if err := validateAffinity(p); err != nil {
		t.Fatal(err)
	}
	p.Complements[0].Shard = 13
	if validateAffinity(p) == nil {
		t.Fatal("accepted split fixture complement")
	}
}
