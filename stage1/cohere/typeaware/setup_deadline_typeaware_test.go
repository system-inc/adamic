package typeaware

import (
	"os"
	"testing"
)

// Optional entry point; standalone shards prepare the same plan themselves.
func TestTypeAwareAgreementAndMutants_Setup(t *testing.T) {
	t.Parallel()
	typeAwareTopPlan(t, "typeaware")
}

// Optional entry point for older callers; each shard prepares independently.
func TestVolumeProfileCorpora_Setup(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST") == "" && os.Getenv("ADAMIC_VOLUME_COMPILER_MANIFEST") == "" {
		t.Skip("set a corpus manifest")
	}
	volumeProfileCorporaPrepare(t)
}
