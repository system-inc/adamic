package typeaware

import "testing"

func TestProduct_volume_guard_stage0(test *testing.T) {
	test.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(test), "stage0")
}

func TestProduct_volume_guard_oracle(test *testing.T) {
	test.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(test), "oracle")
}

func TestProduct_volume_guard_archive_checker(test *testing.T) {
	test.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(test), "archive-checker")
}

func TestProduct_volume_guard_archive_checker_asan(test *testing.T) {
	test.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(test), "archive-checker-asan")
}

func TestProduct_volume_guard_archive_strict_this_checker(test *testing.T) {
	test.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(test), "archive-strict-this-checker")
}

func TestProduct_volume_guard_volume(test *testing.T) {
	test.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(test), "volume")
}

func TestProduct_volume_guard_volume_asan(test *testing.T) {
	test.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(test), "volume-asan")
}

func TestProduct_volume_guard_strict_this_native(test *testing.T) {
	test.Parallel()
	volumeGuardProduct(volumeGuardPreparationHarness(test), "strict-this-native")
}
