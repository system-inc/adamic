package tsgo_test

import (
	"path/filepath"
	"testing"
)

func bridgeRepository(test testing.TB) string {
	test.Helper()
	repository, err := filepath.Abs("../..")
	if err != nil {
		test.Fatal(err)
	}
	return repository
}

func bridgeCaseProducts(piece string) []string {
	switch piece {
	case "abi", "input-length":
		return []string{"api"}
	case "unlinked-build", "unlinked-c", "unlinked-js":
		return []string{"stage0"}
	case "output-length":
		return []string{"length-driver"}
	case "stale-handle":
		return []string{"stale-driver"}
	case "linkage":
		return []string{"linkage.test"}
	case "output-free":
		return []string{"leak-driver"}
	case "region":
		return []string{"oracle", "healthy-region"}
	case "region-ownership":
		return []string{"region-native"}
	case "oracle":
		return []string{"oracle", "native-asan"}
	case "timing-1", "timing-2", "timing-3":
		return []string{"oracle", "native"}
	case "wrong-position":
		return []string{"oracle", "wrong-native"}
	default:
		panic("unknown bridge piece: " + piece)
	}
}

func TestProduct_TsgoArchive(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "tsgo.a")
}

func TestProduct_TsgoSanitizedArchive(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "tsgo-asan.a")
}

func TestProduct_LengthArchive(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "length.a")
}

func TestProduct_StaleArchive(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "stale.a")
}

func TestProduct_WrongArchive(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "wrong.a")
}

func TestProduct_LeakArchive(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "leak.a")
}

func TestProduct_Stage0(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "stage0")
}

func TestProduct_Oracle(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "oracle")
}

func TestProduct_ApiDriver(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "api")
}

func TestProduct_LengthDriver(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "length-driver")
}

func TestProduct_StaleDriver(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "stale-driver")
}

func TestProduct_LeakDriver(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "leak-driver")
}

func TestProduct_SanitizedNative(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "native-asan")
}

func TestProduct_Native(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "native")
}

func TestProduct_WrongNative(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "wrong-native")
}

func TestProduct_HealthyRegion(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "healthy-region")
}

func TestProduct_RegionStage0(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "region-stage0")
}

func TestProduct_RegionNative(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "region-native")
}

func TestProduct_LinkageTest(test *testing.T) {
	test.Parallel()
	bridgeProduct(test, bridgeRepository(test), "linkage.test")
}
