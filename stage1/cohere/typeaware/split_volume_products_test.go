package typeaware

import (
	"os"
	"testing"
)

// The corpus compiler and profile controls share this exact emission recipe.
func TestProduct_corpus_lowered(t *testing.T) {
	t.Parallel()
	volumeProfileControlsLowered(typeAwareProductHarness(t), "")
}

func TestProduct_volume_native_asan(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "volume-asan", "volume_suite.ts", true)
}

func TestProduct_profile_controls_native_asan(t *testing.T) {
	t.Parallel()
	volumeProfileControlProduct(t, true)
}

func TestProduct_corpus_native_asan(t *testing.T) {
	t.Parallel()
	volumeProfileCorpusNativeProduct(t, true)
}

func TestProduct_corpus_record(t *testing.T) {
	t.Parallel()
	volumeProfileCorporaPrepare(t)
}

// Optional entry point; every selected shard prepares these products itself.
func TestVolumeProfileCorpora_Setup(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST") == "" && os.Getenv("ADAMIC_VOLUME_COMPILER_MANIFEST") == "" {
		t.Skip("set a corpus manifest")
	}
	volumeProfileCorporaPrepare(t)
}
