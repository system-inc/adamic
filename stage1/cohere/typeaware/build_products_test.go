package typeaware

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
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

func typeAwareOverlayOracle(h *harness, productName, binaryName, virtualName, source, overlayName string, buildVCSFalse bool) string {
	h.t.Helper()
	virtual := filepath.Join(h.repository, "cohere", virtualName)
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(h.repository, "stage1/cohere/typeaware/testdata", source)}})
	if err != nil {
		h.t.Fatal(err)
	}
	args := []string{"build"}
	if buildVCSFalse {
		args = append(args, "-buildvcs=false")
	}
	args = append(args, "-overlay", h.write(overlayName, string(data)), "-o", filepath.Join(h.directory, binaryName), virtual)
	command := exec.Command("go", args...)
	command.Dir = filepath.Join(h.repository, "cohere")
	return h.sixBuildProduct(productName, command)
}

func typeAwareOracle(h *harness, kind string) string {
	source, virtual := "oracle.go", "adamic_typeaware_oracle.go"
	if kind == "six" {
		source, virtual = "oracle_six.go", "adamic_six_oracle.go"
	}
	return typeAwareOverlayOracle(h, "oracle-build", "oracle", virtual, source, "oracle-overlay.json", false)
}

func sixFactsCostGo(h *harness) string {
	return h.sixBuildProduct("facts-cost-go-build", exec.Command("go", "build", "-o", filepath.Join(h.directory, "direct-cost"), "./bridge/tsgo/cost"))
}

func typeAwareNativeProduct(h *harness, name, entry string, sanitize bool) string {
	archiveName := "checker"
	if sanitize {
		archiveName += "-asan"
	}
	return h.build("", name, filepath.Join(h.repository, "stage1/cohere/typeaware", entry), h.archive(archiveName, "", sanitize), sanitize)
}

func volumeProfileControlProduct(t *testing.T, sanitize bool) string {
	h := typeAwareProductHarness(t)
	return volumeProfileControlsBinary(h, "", h.archive("checker", "", sanitize), sanitize)
}

func volumeProfileCorpusNativeProduct(t *testing.T, sanitize bool) string {
	h := typeAwareProductHarness(t)
	ctx := context.Background()
	stage0 := volumeProfileCorporaGoProduct(t, ctx, h.repository, "adamic")
	archiveName := "checker.a"
	if sanitize {
		archiveName = "checker-asan.a"
	}
	archive := volumeProfileCorporaGoProduct(t, ctx, h.repository, archiveName)
	inputs := volumeProfileCorporaNativeInputs(ctx, h, sanitize)
	return volumeProfileCorporaNative(ctx, h, stage0, archive, inputs, sanitize)
}

func volumeProfileCorpusGoProduct(t *testing.T, output string) string {
	h := typeAwareProductHarness(t)
	return volumeProfileCorporaGoProduct(t, context.Background(), h.repository, output)
}

func volumeProfileDeclaredMutant(t *testing.T, index int) string {
	return volumeProfileMutantProduct(typeAwareProductHarness(t), volumeProfileMutantChanges()[index])
}

func TestProduct_stage0(t *testing.T) { t.Parallel(); typeAwareProductHarness(t).stage0() }
func TestProduct_checker(t *testing.T) {
	t.Parallel()
	typeAwareProductHarness(t).archive("checker", "", false)
}
func TestProduct_checker_asan(t *testing.T) {
	t.Parallel()
	typeAwareProductHarness(t).archive("checker-asan", "", true)
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
	volumeProfileMissingBinding(h, "", h.archive("checker", "", false))
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
func TestProduct_corpus_stage0(t *testing.T) { t.Parallel(); volumeProfileCorpusGoProduct(t, "adamic") }
func TestProduct_corpus_checker(t *testing.T) {
	t.Parallel()
	volumeProfileCorpusGoProduct(t, "checker.a")
}
func TestProduct_corpus_checker_asan(t *testing.T) {
	t.Parallel()
	volumeProfileCorpusGoProduct(t, "checker-asan.a")
}
func TestProduct_corpus_oracle(t *testing.T) {
	t.Parallel()
	volumeProfileCorpusGoProduct(t, "volume-oracle")
}
func TestProduct_corpus_native(t *testing.T) {
	t.Parallel()
	volumeProfileCorpusNativeProduct(t, false)
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

// Narrow canary of staged tools ec73409f (#trp9kgz, #v5fgqc4): never merge.
