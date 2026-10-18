package lint

import (
	"bytes"
	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProduct_ShardsAgreeLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintArtifacts(t, testShardsAgreeLowered(t, packageDirectory), "main.c")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreeNative(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintExecutable(t, testShardsAgreeBinary(t, packageDirectory))
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_LintFindsShardsAgreeCapture(t *testing.T) {
	t.Parallel()
	started := time.Now()
	testShardsAgreeUpstream(t)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreePrepared(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintArtifacts(t, testShardsAgreePrepare(t), "prepared.json")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreeReferences(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintReferences(t, testShardsAgreeReferences(t))
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintArtifacts(t, compilationLowered(t, compilationInputs(t)), "program.gob")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationC(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintArtifacts(t, compilationEmission(t, compilationInputs(t), "c"), "main.c")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationJavaScript(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintArtifacts(t, compilationEmission(t, compilationInputs(t), "javascript"), "lint.mjs")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationScanner(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintExecutable(t, filepath.Join(compilationProductKinds(t, []string{"scanner"}), "scanner"))
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationCounted(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintExecutable(t, filepath.Join(compilationProductKinds(t, []string{"counted"}), "counted"))
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationProfiled(t *testing.T) {
	t.Parallel()
	started := time.Now()
	directory := compilationProductKinds(t, []string{"profiled"})
	requireLintExecutable(t, filepath.Join(directory, "profiled"))
	t.Logf("product wall=%s", time.Since(started))
	// The binary names no checkout: clang ran in this package's directory, and its debug info says only
	// profiledCompilationDirectory. Mutant: drop -ffile-compilation-dir (DW_AT_comp_dir is this directory).
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(filepath.Join(directory, "profiled"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(binary, []byte(here)) {
		t.Fatalf("the profiled binary names the checkout it was built in, %s", here)
	}
}

func TestProduct_LintFindsProfileOracleOutputs(t *testing.T) {
	t.Parallel()
	oracles.FindProfileOracleOutputs(t)
}

func TestProduct_SuggestionAlongsideLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	suggestionAlongsideSetup(t)
	requireLintArtifacts(t, filepath.Dir(suggestionAlongside.module), "lint.c", "lint.mjs")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_SuggestionAlongsideNative(t *testing.T) {
	t.Parallel()
	started := time.Now()
	suggestionAlongsideSetup(t)
	requireLintExecutable(t, suggestionAlongside.binary)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_LintFindsSuggestionAlongsideGoOracle(t *testing.T) {
	t.Parallel()
	oracles.FindSuggestionAlongsideOracle(t)
}

func TestProduct_DecodedOptionsLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintArtifacts(t, decodedOptionsSetupLowered(t, false), "lint.c", "lint.mjs", "port/main.ts")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_DecodedOptionsMutantLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintArtifacts(t, decodedOptionsSetupLowered(t, true), "lint.c", "lint.mjs", "port/main.ts")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_DecodedOptionsNative(t *testing.T) {
	t.Parallel()
	started := time.Now()
	requireLintExecutable(t, filepath.Join(decodedOptionsNativeProduct(t), "scanner"))
	t.Logf("product wall=%s", time.Since(started))
}
