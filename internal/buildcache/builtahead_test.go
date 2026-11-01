package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A product its builder audits is found by a later BuiltAhead with no audit and no error: a find's only build is its refusal,
// so auditing it read as a failed rebuild of a product that was there (#v08876, internal/worker's wasm mutants under
// 0.8.876's rekey). The same listed product looked up through Get with a real build is still rebuilt and compared.
// Mutant: BuiltAhead audits as Get does (its error is the refusal's "its rebuild failed"). Every lookup whose only build
// refuses is a BuiltAhead (cmd/adamic-gate/parallelcensus's TestNoLookupThatOnlyRefusesGoesThroughGet).
// Not parallel: newRig.
func TestABuiltAheadLookupNeverAuditsAndABuildingOneStillDoes(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeldForLife()
	keys := filepath.Join(t.TempDir(), "keys")
	if err := os.WriteFile(keys, []byte(build.NameKey+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_BUILD_AUDIT_KEYS", keys)
	log := filepath.Join(t.TempDir(), "builds.log")
	t.Setenv("ADAMIC_BUILD_LOG", log)
	refusal := errors.New("port not built ahead")
	if _, err := BuiltAhead(port, refusal); err != nil {
		t.Fatalf("a BuiltAhead lookup of a listed product that's there: %v, want it found", err)
	}
	if lines, _ := os.ReadFile(log); strings.Contains(string(lines), "audit ") {
		t.Errorf("a BuiltAhead lookup audited its product: census %q", lines)
	}
	if _, err := Get(port, building("built")); err != nil {
		t.Fatal(err)
	}
	if lines, _ := os.ReadFile(log); !strings.Contains(string(lines), "audit port ") || !strings.Contains(string(lines), " clean ") {
		t.Errorf("a listed product looked up with a real build wasn't audited clean: census %q", lines)
	}
}

// A product BuiltAhead fetches under a trace, with every fetch sampled for an audit, is fetched and never audited: a
// find's only build is its refusal, so the audit returned that refusal, bare, for a product the store held (#v08876).
// A lookup that builds still audits a fetch (TestAnAuditedFetchThatDiffersFromARebuildIsPoisoning).
// Mutant: the fetch path audits a BuiltAhead lookup (its error is the refusal).
// Not parallel: points the build cache (ADAMIC_BUILD_CACHE_DIR, ADAMIC_BUILD_LOG) and the store at this test through t.Setenv.
func TestABuiltAheadFetchUnderATraceIsNeverAudited(t *testing.T) {
	store, log := shared(t)
	t.Setenv("ADAMIC_BUILD_AUDIT", "1")
	t.Setenv("ADAMIC_BUILD_TRACE", t.TempDir())
	store.put(t, recorded(t), map[string]string{"product": "built"})
	directory, err := BuiltAhead(thisPackage, errors.New("this package's product not built ahead"))
	if err != nil {
		t.Fatalf("a BuiltAhead fetch of a product the store holds: %v, want it fetched", err)
	}
	if content, _ := os.ReadFile(filepath.Join(directory, "product")); string(content) != "built" {
		t.Errorf("fetched %q, want the stored product", content)
	}
	if lines, _ := os.ReadFile(log); strings.Contains(string(lines), " audited ") || !strings.Contains(string(lines), " fetched ") {
		t.Errorf("census %q: want the product fetched and never audited", lines)
	}
}
