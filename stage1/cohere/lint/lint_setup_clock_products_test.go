package lint

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProduct_ShardsAgreeLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	testShardsAgreeLowered(t, packageDirectory)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreeNative(t *testing.T) {
	t.Parallel()
	started := time.Now()
	testShardsAgreeBinary(t, packageDirectory)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreeCapture(t *testing.T) {
	t.Parallel()
	started := time.Now()
	testShardsAgreeUpstream(t)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreePrepared(t *testing.T) {
	t.Parallel()
	started := time.Now()
	testShardsAgreePrepare(t)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ShardsAgreeReferences(t *testing.T) {
	t.Parallel()
	started := time.Now()
	testShardsAgreeReferences(t)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	compilationLowered(t, compilationInputs(t))
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationC(t *testing.T) {
	t.Parallel()
	started := time.Now()
	compilationEmission(t, compilationInputs(t), "c")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationJavaScript(t *testing.T) {
	t.Parallel()
	started := time.Now()
	compilationEmission(t, compilationInputs(t), "javascript")
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationScanner(t *testing.T) {
	t.Parallel()
	started := time.Now()
	compilationProductKinds(t, []string{"scanner"})
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationCounted(t *testing.T) {
	t.Parallel()
	started := time.Now()
	compilationProductKinds(t, []string{"counted"})
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_ProfileCompilationProfiled(t *testing.T) {
	t.Parallel()
	started := time.Now()
	directory := compilationProductKinds(t, []string{"profiled"})
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

func TestProduct_ProfileCompilationOracleOutputs(t *testing.T) {
	t.Parallel()
	started := time.Now()
	compilationPrepare(t)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_SuggestionAlongsideLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	suggestionAlongsideSetup(t)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_SuggestionAlongsideNative(t *testing.T) {
	t.Parallel()
	started := time.Now()
	suggestionAlongsideSetup(t)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_SuggestionAlongsideGoOracle(t *testing.T) {
	t.Parallel()
	started := time.Now()
	suggestionAlongsideSetup(t)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_DecodedOptionsLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	decodedOptionsSetupLowered(t, false)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_DecodedOptionsMutantLowered(t *testing.T) {
	t.Parallel()
	started := time.Now()
	decodedOptionsSetupLowered(t, true)
	t.Logf("product wall=%s", time.Since(started))
}

func TestProduct_DecodedOptionsNative(t *testing.T) {
	t.Parallel()
	started := time.Now()
	decodedOptionsNativeProduct(t)
	t.Logf("product wall=%s", time.Since(started))
}
