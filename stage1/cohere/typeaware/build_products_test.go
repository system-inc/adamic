package typeaware

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Product declarations and check units use these same builders. Dependencies
// remain cache-backed, so any product or shard can be selected by itself.
func typeAwareProductHarness(t *testing.T) *harness {
	t.Helper()
	h := volumeProfileHarness(t)
	h.sharedSetup = true
	h.sixBuilds = true
	return h
}

// typeAwareOverlayOracle is a Go oracle built in cohere from a main its overlay synthesizes: virtualName, read from
// testdata/<source> with mutations planted. It is a buildcache.GoBuildIn product named binaryName, keyed by the oracle's
// own graph (cohere's lint rules and the TypeScript port, no compiler code) and the overlay's replacement, where as an
// ordinary product of this package it was keyed by the test binary, which links the compiler, and rebuilt on every
// compiler edit (#gr5y0yf). The virtual file's name is in the binary's line table, so two oracles over one source are
// two products.
func typeAwareOverlayOracle(h *harness, binaryName, virtualName, source string, mutations ...buildcache.Mutation) string {
	h.t.Helper()
	virtual := filepath.Join(h.repository, "cohere", virtualName)
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(h.repository, "stage1/cohere/typeaware/testdata", source)}})
	if err != nil {
		h.t.Fatal(err)
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	overlay := filepath.Join(h.t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, data, 0644); err != nil {
		h.t.Fatal(err)
	}
	return buildcache.GoBuildInMutated(h.t, "cohere", binaryName, virtual, []string{"-overlay=" + overlay}, mutations)
}

func typeAwareOracle(h *harness, kind string) string {
	source, virtual := "oracle.go", "adamic_typeaware_oracle.go"
	if kind == "six" {
		source, virtual = "oracle_six.go", "adamic_six_oracle.go"
	}
	return typeAwareOverlayOracle(h, "oracle", virtual, source)
}

// sixFactsCostGo is ./bridge/tsgo/cost, the facts' cost measured straight on the checker, a buildcache.GoBuild product.
func sixFactsCostGo(h *harness) string {
	return buildcache.GoBuild(h.t, "direct-cost", "./bridge/tsgo/cost", nil)
}

func typeAwareNativeProduct(h *harness, name, entry string, sanitize bool) string {
	return h.build("", name, filepath.Join(h.repository, "stage1/cohere/typeaware", entry), h.archive(sanitize), sanitize)
}

func volumeProfileControlProduct(t *testing.T, sanitize bool) string {
	h := typeAwareProductHarness(t)
	return volumeProfileControlsBinary(h, "", h.archive(sanitize), sanitize)
}

func volumeProfileDeclaredMutant(t *testing.T, index int) string {
	return volumeProfileMutantProduct(typeAwareProductHarness(t), volumeProfileMutantChanges()[index])
}

func TestProduct_stage0(t *testing.T) { t.Parallel(); typeAwareProductHarness(t).stage0() }
func TestProduct_checker(t *testing.T) {
	t.Parallel()
	typeAwareProductHarness(t).archive(false)
}
func TestProduct_checker_asan(t *testing.T) {
	t.Parallel()
	typeAwareProductHarness(t).archive(true)
}
func TestProduct_six_lowered(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/sharded_suite.ts"), "suite")
}
func TestProduct_suite(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "suite", "testdata/sharded_suite.ts", false)
}
func TestProduct_suite_asan(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "suite-asan", "testdata/sharded_suite.ts", true)
}

// The two mutant natives whose ASan link would otherwise fill their shard's own time (mutantNative says why): six's
// facts-length (TestSixRuleAgreementAndMutants_028) and TypeAwareAgreement's wrong-node (_003).
func TestProduct_six_facts_length_native(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	h.mutantNative("facts-length", filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/sharded_suite.ts"), checkerMutantArchive(t, "facts-length"), true)
}

// Six's cross-file declaration-source mutant (TestSixRuleAgreementAndMutants_016), whose compile took 19 s of its
// shard's own time under load; the corpus mutants' natives are sixMutantNativeProduct's.
func TestProduct_six_declaration_source_native(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	h.mutantNative("declaration-source-suite", filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/sharded_suite.ts"), checkerMutantArchive(t, "declaration-source"), false)
}
func TestProduct_typeaware_wrong_node_native(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	h.mutantNative("wrong-node", typeAwareWrongNodeEntry(h), h.archive(false), true)
}
func TestProduct_six_oracle(t *testing.T) {
	t.Parallel()
	typeAwareOracle(typeAwareProductHarness(t), "six")
}
func TestProduct_typeaware_lowered(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/main.ts"), "native")
}
func TestProduct_typeaware_native(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "native", "main.ts", false)
}
func TestProduct_typeaware_native_asan(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "native-asan", "main.ts", true)
}
func TestProduct_typeaware_oracle(t *testing.T) {
	t.Parallel()
	typeAwareOracle(typeAwareProductHarness(t), "typeaware")
}
func TestProduct_facts_cost_lowered(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/fact_cost.ts"), "facts-cost")
}
func TestProduct_facts_cost_native(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "facts-cost", "testdata/fact_cost.ts", false)
}
func TestProduct_facts_cost_go(t *testing.T) {
	t.Parallel()
	sixFactsCostGo(typeAwareProductHarness(t))
}
func TestProduct_query_cost_lowered(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/query_cost.ts"), "query-cost")
}
func TestProduct_query_cost_native(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "query-cost", "testdata/query_cost.ts", false)
}

// The released-handle refusal reads released.ts's C and its native over the checker archive; no product built either,
// so Workshop's combined proof (Oct 10) found the C missing on the reading checkout.
func TestProduct_released_lowered(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/released.ts"), "released")
}
func TestProduct_released_native(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "released", "testdata/released.ts", false)
}
func TestProduct_volume_lowered(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts"), "volume")
}
func TestProduct_volume_native(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "volume", "volume_suite.ts", false)
}
func TestProduct_volume_oracle(t *testing.T) {
	t.Parallel()
	volumeOracle(typeAwareProductHarness(t), "volume-oracle", "oracle_volume.go")
}
func TestProduct_profile_controls_lowered(t *testing.T) {
	t.Parallel()
	volumeProfileControlsLowered(typeAwareProductHarness(t), "")
}
func TestProduct_profile_controls_native(t *testing.T) {
	t.Parallel()
	volumeProfileControlProduct(t, false)
}
func TestProduct_profile_missing_binding(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	volumeProfileMissingBinding(h, "", h.archive(false))
}
func TestProduct_profile_mutant_binding_slot(t *testing.T) {
	t.Parallel()
	volumeProfileDeclaredMutant(t, 0)
}
func TestProduct_profile_mutant_scope_containment(t *testing.T) {
	t.Parallel()
	volumeProfileDeclaredMutant(t, 1)
}
func TestProduct_profile_mutant_first_binding(t *testing.T) {
	t.Parallel()
	volumeProfileDeclaredMutant(t, 2)
}
func TestProduct_type_symbol_archive(t *testing.T) {
	t.Parallel()
	typeSymbolArchive(typeAwareProductHarness(t))
}
func TestProduct_type_symbol_native(t *testing.T) {
	t.Parallel()
	typeSymbolNativeProduct(typeAwareProductHarness(t))
}
func TestProduct_type_symbol_oracle(t *testing.T) {
	t.Parallel()
	typeSymbolOracle(typeAwareProductHarness(t))
}
func TestProduct_type_symbol_fixtures(t *testing.T) {
	t.Parallel()
	typeSymbolFixtures(typeAwareProductHarness(t))
}
func TestProduct_type_symbol_truth(t *testing.T) {
	t.Parallel()
	h := typeAwareProductHarness(t)
	oracle := typeSymbolOracle(h)
	paths, manifest := typeSymbolFixtures(h)
	typeSymbolTruth(h, oracle, paths, manifest)
}

// volume_config_guard_shards_test.go belongs to sf-volumeguard. Its declarations
// must be supplied there; this file deliberately does not duplicate them.
