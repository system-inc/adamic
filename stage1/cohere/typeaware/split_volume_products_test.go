package typeaware

import "testing"

func TestProduct_volume_native_asan(t *testing.T) {
	t.Parallel()
	typeAwareNativeProduct(typeAwareProductHarness(t), "volume-asan", "volume_suite.ts", true)
}

func TestProduct_profile_controls_native_asan(t *testing.T) {
	t.Parallel()
	volumeProfileControlProduct(t, true)
}
