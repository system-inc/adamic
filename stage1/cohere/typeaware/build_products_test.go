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
func typeAwareProductHarness(test *testing.T) *harness {
	test.Helper()
	h := volumeProfileHarness(test)
	h.sharedSetup = true
	h.sixBuilds = true
	return h
}

// typeAwareOverlayOracle is a Go oracle built in cohere from a main its overlay synthesizes: virtualName, read from
// testdata/<source> with mutations planted. It is a buildcache.GoBuildIn product named binaryName, keyed by the
// oracle's own graph (cohere's lint rules and the TypeScript port, no compiler code) and the overlay's replacement,
// where as an ordinary product of this package it was keyed by the test binary, which links the compiler, and rebuilt
// on every compiler edit (#gr5y0yf). The virtual file's name is in the binary's line table, so two oracles over one
// source are two products.
func typeAwareOverlayOracle(
	h *harness,
	binaryName, virtualName, source string,
	mutations ...buildcache.Mutation,
) string {
	h.t.Helper()
	virtual := filepath.Join(h.repository, "cohere", virtualName)
	data, encodeError := json.Marshal(map[string]any{
		"Replace": map[string]string{virtual: filepath.Join(h.repository, "stage1/cohere/typeaware/testdata", source)},
	})
	if encodeError != nil {
		h.t.Fatal(encodeError)
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	overlay := filepath.Join(h.t.TempDir(), "overlay.json")
	if writeError := os.WriteFile(overlay, data, 0644); writeError != nil {
		h.t.Fatal(writeError)
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
	return h.build(
		"",
		name,
		filepath.Join(h.repository, "stage1/cohere/typeaware", entry),
		h.archive(sanitize),
		sanitize,
	)
}

func volumeProfileControlProduct(test *testing.T, sanitize bool) string {
	h := typeAwareProductHarness(test)
	return volumeProfileControlsBinary(h, "", h.archive(sanitize), sanitize)
}

func volumeProfileDeclaredMutant(test *testing.T, index int) string {
	return volumeProfileMutantProduct(typeAwareProductHarness(test), volumeProfileMutantChanges()[index])
}

func TestProduct_stage0(test *testing.T) { test.Parallel(); typeAwareProductHarness(test).stage0() }
func TestProduct_checker(test *testing.T) {
	test.Parallel()
	typeAwareProductHarness(test).archive(false)
}
func TestProduct_checker_asan(test *testing.T) {
	test.Parallel()
	typeAwareProductHarness(test).archive(true)
}
func TestProduct_six_lowered(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/sharded_suite.ts"), "suite")
}
func TestProduct_suite(test *testing.T) {
	test.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(test), "suite", "testdata/sharded_suite.ts", false)
}
func TestProduct_suite_asan(test *testing.T) {
	test.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(test), "suite-asan", "testdata/sharded_suite.ts", true)
}

// The two mutant natives whose ASan link would otherwise fill their shard's own time (mutantNative says why): six's
// facts-length (TestSixRuleAgreementAndMutants_028) and TypeAwareAgreement's wrong-node (_003).
func TestProduct_six_facts_length_native(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	h.mutantNative(
		"facts-length",
		filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/sharded_suite.ts"),
		checkerMutantArchive(test, "facts-length"),
		true,
	)
}

// Six's cross-file declaration-source mutant (TestSixRuleAgreementAndMutants_016), whose compile took 19 s of its
// shard's own time under load; the corpus mutants' natives are sixMutantNativeProduct's.
func TestProduct_six_declaration_source_native(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	h.mutantNative(
		"declaration-source-suite",
		filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/sharded_suite.ts"),
		checkerMutantArchive(test, "declaration-source"),
		false,
	)
}
func TestProduct_typeaware_wrong_node_native(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	typeAwareWrongNodeNative(h, h.archive(false), true)
}

// TypeAwareAgreement's other mutant natives, and the bad-kind refusal's, each built ahead so its shard only runs it.
func TestProduct_typeaware_stale_native(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	h.mutantNative(
		"stale-native",
		filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/released.ts"),
		checkerMutantArchive(test, "retained-handle"),
		false,
	)
}
func TestProduct_typeaware_empty_frame_native(test *testing.T) {
	test.Parallel()
	typeAwareFrameMutantProduct(test, "empty-frame")
}
func TestProduct_typeaware_missing_header_native(test *testing.T) {
	test.Parallel()
	typeAwareFrameMutantProduct(test, "missing-header")
}
func TestProduct_typeaware_bad_length_native(test *testing.T) {
	test.Parallel()
	typeAwareFrameMutantProduct(test, "bad-length")
}
func typeAwareFrameMutantProduct(test *testing.T, name string) {
	test.Helper()
	h := typeAwareProductHarness(test)
	h.mutantNative(
		name+"-native",
		filepath.Join(h.repository, "stage1/cohere/typeaware/main.ts"),
		checkerMutantArchive(test, name),
		false,
	)
}
func TestProduct_typeaware_bad_kind_native(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	typeAwareBadKindNative(h, "bad-kind", h.archive(false))
}
func TestProduct_typeaware_ignored_kind_native(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	typeAwareBadKindNative(h, "ignored-kind-native", checkerMutantArchive(test, "ignored-kind"))
}
func TestProduct_six_oracle(test *testing.T) {
	test.Parallel()
	typeAwareOracle(typeAwareProductHarness(test), "six")
}
func TestProduct_typeaware_lowered(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/main.ts"), "native")
}
func TestProduct_typeaware_native(test *testing.T) {
	test.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(test), "native", "main.ts", false)
}
func TestProduct_typeaware_native_asan(test *testing.T) {
	test.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(test), "native-asan", "main.ts", true)
}
func TestProduct_typeaware_oracle(test *testing.T) {
	test.Parallel()
	typeAwareOracle(typeAwareProductHarness(test), "typeaware")
}
func TestProduct_facts_cost_lowered(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/fact_cost.ts"), "facts-cost")
}
func TestProduct_facts_cost_native(test *testing.T) {
	test.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(test), "facts-cost", "testdata/fact_cost.ts", false)
}
func TestProduct_facts_cost_go(test *testing.T) {
	test.Parallel()
	sixFactsCostGo(typeAwareProductHarness(test))
}
func TestProduct_query_cost_lowered(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/query_cost.ts"), "query-cost")
}
func TestProduct_query_cost_native(test *testing.T) {
	test.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(test), "query-cost", "testdata/query_cost.ts", false)
}

// The released-handle refusal reads released.ts's C and its native over the checker archive; no product built either,
// so Workshop's combined proof (Oct 10) found the C missing on the reading checkout.
func TestProduct_released_lowered(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/released.ts"), "released")
}
func TestProduct_released_native(test *testing.T) {
	test.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(test), "released", "testdata/released.ts", false)
}
func TestProduct_volume_lowered(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	h.lowered(filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts"), "volume")
}
func TestProduct_volume_native(test *testing.T) {
	test.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(test), "volume", "volume_suite.ts", false)
}
func TestProduct_volume_oracle(test *testing.T) {
	test.Parallel()
	volumeOracle(typeAwareProductHarness(test), "volume-oracle", "oracle_volume.go")
}
func TestProduct_profile_controls_lowered(test *testing.T) {
	test.Parallel()
	volumeProfileControlsLowered(typeAwareProductHarness(test), "")
}
func TestProduct_profile_controls_native(test *testing.T) {
	test.Parallel()
	volumeProfileControlProduct(test, false)
}
func TestProduct_profile_missing_binding(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	volumeProfileMissingBinding(h, "", h.archive(false))
}
func TestProduct_profile_mutant_binding_slot(test *testing.T) {
	test.Parallel()
	volumeProfileDeclaredMutant(test, 0)
}
func TestProduct_profile_mutant_scope_containment(test *testing.T) {
	test.Parallel()
	volumeProfileDeclaredMutant(test, 1)
}
func TestProduct_profile_mutant_first_binding(test *testing.T) {
	test.Parallel()
	volumeProfileDeclaredMutant(test, 2)
}
func TestProduct_type_symbol_archive(test *testing.T) {
	test.Parallel()
	typeSymbolArchive(typeAwareProductHarness(test))
}
func TestProduct_type_symbol_native(test *testing.T) {
	test.Parallel()
	typeSymbolNativeProduct(typeAwareProductHarness(test))
}
func TestProduct_type_symbol_oracle(test *testing.T) {
	test.Parallel()
	typeSymbolOracle(typeAwareProductHarness(test))
}
func TestProduct_type_symbol_fixtures(test *testing.T) {
	test.Parallel()
	typeSymbolFixtures(typeAwareProductHarness(test))
}
func TestProduct_type_symbol_truth(test *testing.T) {
	test.Parallel()
	h := typeAwareProductHarness(test)
	oracle := typeSymbolOracle(h)
	paths, manifest := typeSymbolFixtures(h)
	typeSymbolTruth(h, oracle, paths, manifest)
}

// volume_config_guard_shards_test.go belongs to sf-volumeguard. Its declarations
// must be supplied there; this file deliberately does not duplicate them.
