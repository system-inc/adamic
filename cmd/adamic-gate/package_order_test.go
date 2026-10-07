package main

import (
	"reflect"
	"testing"
)

func TestPackageFeedDescending(t *testing.T) {
	p := plan{Count: 2, Units: []unit{{Package: "a", Test: "TestShared", Shard: 0, Seconds: 200}, {Package: "z", Test: "TestParent/a", Shard: 0, Seconds: 10}, {Package: "b", Test: "TestOther", Shard: 0, Seconds: 10}}, Affinity: []affinity{{Package: "a", Fixture: "once", Tests: []string{"TestShared"}, Shard: 0, Seconds: 10}}}
	p.PackageSeconds = packagePredictions(p, map[string]float64{"z::TestParent/a": 10, "z::TestParent": 100})
	packages := []string{"a", "b", "z"}
	sortPackages(p, 0, packages)
	if !reflect.DeepEqual(packages, []string{"z", "a", "b"}) {
		t.Fatal("feed is not descending with deterministic name ties", packages, p.PackageSeconds)
	}
	if p.PackageSeconds[packagePredictionKey(0, "a")] != 10 || p.PackageSeconds[packagePredictionKey(0, "z")] != 100 {
		t.Fatal("lost affinity or parent setup", p.PackageSeconds)
	}
}
