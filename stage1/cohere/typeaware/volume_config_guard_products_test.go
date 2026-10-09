package typeaware

import "testing"

func TestProduct_volume_guard_stage0(t *testing.T) {
	t.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(t), "stage0")
}

func TestProduct_volume_guard_oracle(t *testing.T) {
	t.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(t), "oracle")
}

func TestProduct_volume_guard_archive_checker(t *testing.T) {
	t.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(t), "archive-checker")
}

func TestProduct_volume_guard_archive_checker_asan(t *testing.T) {
	t.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(t), "archive-checker-asan")
}

func TestProduct_volume_guard_archive_strict_this_checker(t *testing.T) {
	t.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(t), "archive-strict-this-checker")
}

func TestProduct_volume_guard_volume(t *testing.T) {
	t.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(t), "volume")
}

func TestProduct_volume_guard_volume_asan(t *testing.T) {
	t.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(t), "volume-asan")
}

func TestProduct_volume_guard_strict_this_native(t *testing.T) {
	t.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(t), "strict-this-native")
}
