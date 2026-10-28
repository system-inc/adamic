package json

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// inventory.json, the corpus's largest case (17.5 MB), shared _0760 with a 3 KB case and was json#20/28's long pole:
// 17.7 s alone on Golem (Go 1.6 s, Node 1.3 s, release 1.9 s, the sanitized run 10.8 s, the JavaScript backend 1.6 s),
// 29.4 s in the unit on Cloud, and 77.6 s in the unit with 8 busy loops halving its CPUs, under the 90 s deadline with
// nothing to spare (#21kptp1). One case doesn't split, so it runs as class order's case does: alone in the unit, before
// go test resumes the parallel shards, with all five sides (jsonPortAgree, the same agreement a top shard checks).

// Not parallel: runs alone in the unit, before the parallel shards resume, so the indivisible case has its CPUs.
func TestPortMatchesGoCohereInventory(t *testing.T) {
	jsonPortDedicatedAgree(t, jsonPortInventoryKey, jsonPortWhole)
}

// More heavy cases each held a hash bucket nearly alone, and their test ran its five sides one after another, one child
// at a time as a parallel test does (testguard.Share). json#14/23 at 8 CPUs on Eve (Oct 19): _1643 27.9 s for
// default-contracts.json, 8.9 MB (Go 4.7 s, Node 3.8 s, release 4.1 s, the sanitized run 12.7 s, the JavaScript backend
// 2.3 s), _1649 14.7 s for selection.json, 4.9 MB (sanitized 7.0 s), and _1549 13.1 s for inspection.json, 2.7 MB
// (sanitized 4.8 s), the unit's three slowest tests in each of six runs over two days: the work, not the load. Then
// json#19/32 on Eve at load 63 (Oct 27): _1345 61.75 s over the 60 s law for sites.json, 8.6 MB (Go 8.0 s, Node 8.3 s,
// release 7.0 s, sanitized 20.1 s, the JavaScript backend 16.3 s), which turned another owner's change red; and in
// four runs at load 45 to 51 (Oct 23) _0713 38.4 s for discriminant-writes.json, 3.3 MB, _1396 36.2 s for the
// predicates REPORT.json, 14.2 MB, _1226 35.3 s for cases-reference.json, 10.2 MB, and _1246 27.8 s for
// cache-local-roots.json, 27.1 MB; and cohere's three tailwind collapse fixtures, about 3.7 MB each, _0731 33.1 s for
// descriptor_fixtures_public.json, _0195 25.5 s for descriptor_fixtures.json and _0565 24.2 s for utility_fixtures.json.
// One case doesn't split, but its sides do: each case is owned by three parallel tests,
// _000 its unsanitized part (the Go oracle, Node and release), _001 its sanitized part (the Go oracle and the sanitized
// run) and _002 its JavaScript backend part (the Go oracle and the JavaScript backend), so its sides run on three CPUs
// at once. sites.json's parts come to about 23, 28 and 24 s at load 63, where one test was 60.
const (
	jsonPortDefaultContractsKey     = "stage3/interface-downcasts/default-contracts.json"
	jsonPortVerdictSelectionKey     = "stage3/verdict/selection.json"
	jsonPortAnyReturnsInspectionKey = "stage3/census/any-returns/evidence/inspection.json"
	jsonPortCensusSitesKey          = "stage3/census/data/sites.json"
	jsonPortDiscriminantWritesKey   = "stage3/fixtures/assertions/discriminant-writes.json"
	jsonPortPredicatesReportKey     = "stage3/census/predicates/REPORT.json"
	jsonPortParserCasesKey          = "stage3/drivers/parser/cases-reference.json"
	jsonPortCacheLocalRootsKey      = "stage3/notyet-new-expression/evidence/cache-local-roots.json"
	jsonPortTailwindDescriptorsKey  = "cohere/internal/lint/rules/tailwind/collapse/testdata/descriptor_fixtures.json"
	jsonPortTailwindPublicKey       = "cohere/internal/lint/rules/tailwind/collapse/testdata/descriptor_fixtures_public.json"
	jsonPortTailwindUtilitiesKey    = "cohere/internal/lint/rules/tailwind/collapse/testdata/utility_fixtures.json"
)

func TestPortMatchesGoCohereAnyReturnsInspection_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortAnyReturnsInspectionKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereAnyReturnsInspection_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortAnyReturnsInspectionKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereAnyReturnsInspection_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortAnyReturnsInspectionKey, jsonPortJavaScript)
}

func TestPortMatchesGoCohereCacheLocalRoots_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortCacheLocalRootsKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereCacheLocalRoots_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortCacheLocalRootsKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereCacheLocalRoots_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortCacheLocalRootsKey, jsonPortJavaScript)
}

func TestPortMatchesGoCohereCensusSites_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortCensusSitesKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereCensusSites_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortCensusSitesKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereCensusSites_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortCensusSitesKey, jsonPortJavaScript)
}

func TestPortMatchesGoCohereDefaultContracts_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortDefaultContractsKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereDefaultContracts_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortDefaultContractsKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereDefaultContracts_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortDefaultContractsKey, jsonPortJavaScript)
}

func TestPortMatchesGoCohereDiscriminantWrites_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortDiscriminantWritesKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereDiscriminantWrites_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortDiscriminantWritesKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereDiscriminantWrites_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortDiscriminantWritesKey, jsonPortJavaScript)
}

func TestPortMatchesGoCohereParserCases_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortParserCasesKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereParserCases_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortParserCasesKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereParserCases_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortParserCasesKey, jsonPortJavaScript)
}

func TestPortMatchesGoCoherePredicatesReport_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortPredicatesReportKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCoherePredicatesReport_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortPredicatesReportKey, jsonPortSanitized)
}

func TestPortMatchesGoCoherePredicatesReport_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortPredicatesReportKey, jsonPortJavaScript)
}

func TestPortMatchesGoCohereVerdictSelection_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortVerdictSelectionKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereVerdictSelection_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortVerdictSelectionKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereVerdictSelection_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortVerdictSelectionKey, jsonPortJavaScript)
}

func TestPortMatchesGoCohereTailwindDescriptors_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortTailwindDescriptorsKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereTailwindDescriptors_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortTailwindDescriptorsKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereTailwindDescriptors_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortTailwindDescriptorsKey, jsonPortJavaScript)
}

func TestPortMatchesGoCohereTailwindPublic_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortTailwindPublicKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereTailwindPublic_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortTailwindPublicKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereTailwindPublic_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortTailwindPublicKey, jsonPortJavaScript)
}

func TestPortMatchesGoCohereTailwindUtilities_000(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortTailwindUtilitiesKey, jsonPortUnsanitized)
}

func TestPortMatchesGoCohereTailwindUtilities_001(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortTailwindUtilitiesKey, jsonPortSanitized)
}

func TestPortMatchesGoCohereTailwindUtilities_002(t *testing.T) {
	t.Parallel()
	jsonPortDedicatedAgree(t, jsonPortTailwindUtilitiesKey, jsonPortJavaScript)
}

// jsonPortDedicatedAgree is a dedicated owner's agreement: the corpus's one case named key, out of the hash buckets,
// held to the Go oracle by the sides part names.
func jsonPortDedicatedAgree(t *testing.T, key string, part jsonPortPart) {
	t.Helper()
	setup := time.Now()
	portMatchesPrepare(t)
	all, _ := jsonTopCorpus(t)
	ordinary, shards, dedicated := jsonPortDedicatedPartition(all)
	if err := jsonPortDedicatedUnion(all, ordinary, shards, dedicated); err != nil {
		t.Fatal(err)
	}
	items := jsonPortDedicatedCase(dedicated, key)
	if len(items) != 1 {
		t.Fatalf("%s is in the corpus %d times, want 1", key, len(items))
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("shared setup: %.3fs; exact union: %d ordinary + %d dedicated", time.Since(setup).Seconds(), len(ordinary), len(dedicated))
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()
	start := time.Now()
	binaries := jsonPortBinaries{
		oracle:    portMatchesShared.oracle,
		entry:     portMatchesShared.entry,
		script:    portMatchesShared.script,
		release:   portMatchesShared.release,
		sanitized: portMatchesShared.sanitized,
		runner:    runner,
	}
	jsonPortAgree(t, items, binaries, strings.ToLower(strings.TrimPrefix(t.Name(), "TestPortMatchesGoCohere")), part)
	t.Logf("%s: own work %.3fs", key, time.Since(start).Seconds())
}

// jsonPortPartsUnion is nil when the unsanitized, sanitized and JavaScript backend parts are the whole's sides between
// them, each side in exactly one part but the Go oracle, which every part runs.
func jsonPortPartsUnion(parts map[jsonPortPart][]int) error {
	owners := map[int]int{}
	for _, part := range jsonPortParts {
		if !slices.Contains(parts[part], 0) {
			return fmt.Errorf("part %d runs no Go oracle", part)
		}
		for _, side := range parts[part] {
			owners[side]++
		}
	}
	if len(owners) != len(parts[jsonPortWhole]) {
		return fmt.Errorf("the parts run %d sides, the whole %d", len(owners), len(parts[jsonPortWhole]))
	}
	for _, side := range parts[jsonPortWhole] {
		want := 1
		if side == 0 {
			want = len(jsonPortParts)
		}
		if owners[side] != want {
			return fmt.Errorf("side %d runs in %d parts, want %d", side, owners[side], want)
		}
	}
	return nil
}

// A split owner checks what one test running the whole checked: the parts are the whole between them, and a side
// dropped from a part, or run in both, fails here.
func TestJSONPortPartsAreTheWhole(t *testing.T) {
	t.Parallel()
	if !slices.Equal(jsonPortPartSides[jsonPortWhole], []int{0, 1, 2, 3, 4}) {
		t.Fatalf("the whole runs sides %v, want all five", jsonPortPartSides[jsonPortWhole])
	}
	if err := jsonPortPartsUnion(jsonPortPartSides); err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct {
		name  string
		parts map[jsonPortPart][]int
	}{
		{"JavaScript backend dropped", map[jsonPortPart][]int{jsonPortWhole: {0, 1, 2, 3, 4}, jsonPortUnsanitized: {0, 1, 2}, jsonPortSanitized: {0, 3}, jsonPortJavaScript: {0}}},
		{"sanitized run dropped", map[jsonPortPart][]int{jsonPortWhole: {0, 1, 2, 3, 4}, jsonPortUnsanitized: {0, 1, 2}, jsonPortSanitized: {0}, jsonPortJavaScript: {0, 4}}},
		{"release in two parts", map[jsonPortPart][]int{jsonPortWhole: {0, 1, 2, 3, 4}, jsonPortUnsanitized: {0, 1, 2}, jsonPortSanitized: {0, 2, 3}, jsonPortJavaScript: {0, 4}}},
		{"no Go oracle in the JavaScript backend part", map[jsonPortPart][]int{jsonPortWhole: {0, 1, 2, 3, 4}, jsonPortUnsanitized: {0, 1, 2}, jsonPortSanitized: {0, 3}, jsonPortJavaScript: {4}}},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if jsonPortPartsUnion(mutant.parts) == nil {
				t.Fatal("a broken split survived")
			}
		})
	}
}
