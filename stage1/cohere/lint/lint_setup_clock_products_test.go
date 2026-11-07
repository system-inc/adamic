package lint

import (
	"bytes"
	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProduct_ShardsAgreeLowered(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintArtifacts(test, testShardsAgreeLowered(test, packageDirectory), "main.c")
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreeNative(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintExecutable(test, testShardsAgreeBinary(test, packageDirectory))
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_LintFindsShardsAgreeCapture(test *testing.T) {
	test.Parallel()
	started := time.Now()
	testShardsAgreeUpstream(test)
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreePrepared(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintArtifacts(test, testShardsAgreePrepare(test), "prepared.json")
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreeReferences(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintReferences(test, testShardsAgreeReferences(test))
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationLowered(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintArtifacts(test, compilationLowered(test, compilationInputs(test)), "program.gob")
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationC(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintArtifacts(test, compilationEmission(test, compilationInputs(test), "c"), "main.c")
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationJavaScript(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintArtifacts(test, compilationEmission(test, compilationInputs(test), "javascript"), "lint.mjs")
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationScanner(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintExecutable(test, filepath.Join(compilationProductKinds(test, []string{"scanner"}), "scanner"))
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationCounted(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintExecutable(test, filepath.Join(compilationProductKinds(test, []string{"counted"}), "counted"))
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationProfiled(test *testing.T) {
	test.Parallel()
	started := time.Now()
	directory := compilationProductKinds(test, []string{"profiled"})
	requireLintExecutable(test, filepath.Join(directory, "profiled"))
	test.Logf("product wall=%s", time.Since(started))
	// The binary names no checkout: clang ran in this package's directory, and its debug info says only
	// profiledCompilationDirectory. Mutant: drop -ffile-compilation-dir (DW_AT_comp_dir is this directory).
	here, workingDirectoryError := os.Getwd()
	if workingDirectoryError != nil {
		test.Fatal(workingDirectoryError)
	}
	binary, readError := os.ReadFile(filepath.Join(directory, "profiled"))
	if readError != nil {
		test.Fatal(readError)
	}
	if bytes.Contains(binary, []byte(here)) {
		test.Fatalf("the profiled binary names the checkout it was built in, %s", here)
	}
}

func TestProduct_LintFindsProfileOracleOutputs(test *testing.T) {
	test.Parallel()
	oracles.FindProfileOracleOutputs(test)
}

func TestProduct_SuggestionAlongsideLowered(test *testing.T) {
	test.Parallel()
	started := time.Now()
	suggestionAlongsideSetup(test)
	requireLintArtifacts(test, filepath.Dir(suggestionAlongside.module), "lint.c", "lint.mjs")
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_SuggestionAlongsideNative(test *testing.T) {
	test.Parallel()
	started := time.Now()
	suggestionAlongsideSetup(test)
	requireLintExecutable(test, suggestionAlongside.binary)
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_LintFindsSuggestionAlongsideGoOracle(test *testing.T) {
	test.Parallel()
	oracles.FindSuggestionAlongsideOracle(test)
}

func TestProduct_DecodedOptionsLowered(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintArtifacts(test, decodedOptionsSetupLowered(test, false), "lint.c", "lint.mjs", "port/main.ts")
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_DecodedOptionsMutantLowered(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintArtifacts(test, decodedOptionsSetupLowered(test, true), "lint.c", "lint.mjs", "port/main.ts")
	test.Logf("product wall=%s", time.Since(started))
}

func TestProduct_DecodedOptionsNative(test *testing.T) {
	test.Parallel()
	started := time.Now()
	requireLintExecutable(test, filepath.Join(decodedOptionsNativeProduct(test), "scanner"))
	test.Logf("product wall=%s", time.Since(started))
}
