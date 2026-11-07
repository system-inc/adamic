package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A product its builder audits is found by a later BuiltAhead with no audit and no error: a find's only build is its
// refusal, so auditing it read as a failed rebuild of a product that was there (#v08876, internal/worker's wasm mutants
// under 0.8.876's rekey). The same listed product looked up through Get with a real build is still rebuilt and
// compared.
// Mutant: BuiltAhead audits as Get does (its error is the refusal's "its rebuild failed"). Every lookup whose only build
// refuses is a BuiltAhead (cmd/adamic-gate/parallelcensus's TestNoLookupThatOnlyRefusesGoesThroughGet).
// Not parallel: newRig.
func TestABuiltAheadLookupNeverAuditsAndABuildingOneStillDoes(test *testing.T) {
	rig := newRig(test)
	Product(test, port, building("built"))
	build := rig.built(test)["port"]
	rig.window(build.Build, func() { rig.open(rig.pid, "source/a.c") })
	if settlement := rig.settle(test); len(settlement.Settled) != 1 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeldForLife()
	keys := filepath.Join(test.TempDir(), "keys")
	if writeError := os.WriteFile(keys, []byte(build.NameKey+"\n"), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	test.Setenv("ADAMIC_BUILD_AUDIT_KEYS", keys)
	log := filepath.Join(test.TempDir(), "builds.log")
	test.Setenv("ADAMIC_BUILD_LOG", log)
	refusal := errors.New("port not built ahead")
	if _, builtAheadError := BuiltAhead(port, refusal); builtAheadError != nil {
		test.Fatalf("a BuiltAhead lookup of a listed product that's there: %v, want it found", builtAheadError)
	}
	if lines, _ := os.ReadFile(log); strings.Contains(string(lines), "audit ") {
		test.Errorf("a BuiltAhead lookup audited its product: census %q", lines)
	}
	if _, getError := Get(port, building("built")); getError != nil {
		test.Fatal(getError)
	}
	if lines, _ := os.ReadFile(
		log,
	); !strings.Contains(string(lines), "audit port ") || !strings.Contains(string(lines), " clean ") {
		test.Errorf("a listed product looked up with a real build wasn't audited clean: census %q", lines)
	}
}

// A product BuiltAhead fetches under a trace, with every fetch sampled for an audit, is fetched and never audited: a
// find's only build is its refusal, so the audit returned that refusal, bare, for a product the store held (#v08876).
// A lookup that builds still audits a fetch (TestAnAuditedFetchThatDiffersFromARebuildIsPoisoning).
// Mutant: the fetch path audits a BuiltAhead lookup (its error is the refusal).
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through test.Setenv.
func TestABuiltAheadFetchUnderATraceIsNeverAudited(test *testing.T) {
	store, log := shared(test)
	test.Setenv("ADAMIC_BUILD_AUDIT", "1")
	test.Setenv("ADAMIC_BUILD_TRACE", test.TempDir())
	store.put(test, recorded(test), map[string]string{"product": "built"})
	directory, builtAheadError := BuiltAhead(thisPackage, errors.New("this package's product not built ahead"))
	if builtAheadError != nil {
		test.Fatalf("a BuiltAhead fetch of a product the store holds: %v, want it fetched", builtAheadError)
	}
	if content, _ := os.ReadFile(filepath.Join(directory, "product")); string(content) != "built" {
		test.Errorf("fetched %q, want the stored product", content)
	}
	if lines, _ := os.ReadFile(
		log,
	); strings.Contains(string(lines), " audited ") || !strings.Contains(string(lines), " fetched ") {
		test.Errorf("census %q: want the product fetched and never audited", lines)
	}
}
