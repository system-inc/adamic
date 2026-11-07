package typeaware

import "testing"

func TestProduct_volume_native_asan(test *testing.T) {
	test.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(test), "volume-asan", "volume_suite.ts", true)
}

func TestProduct_profile_controls_native_asan(test *testing.T) {
	test.Parallel()
	volumeProfileControlProduct(test, true)
}
