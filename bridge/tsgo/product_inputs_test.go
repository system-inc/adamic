package tsgo_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// A small source tree stands in for the recipe's real dependencies. Its exact
// Files, Flags and Toolchain feed Key, then Get. Changing the recipe source must
// fetch a new product; omitting it returns version one and fails this comparison.
func checkBridgeProductInput(test *testing.T, name string) {
	test.Helper()
	repository := bridgeRepository(test)
	inputs := bridgeProductInputs(test, repository, name)
	root := test.TempDir()
	// The stand-in tree is the repository here: its flags name it where they named the checkout, so Key spells both
	// <repository> and the proof's products key the same from any checkout, as Workshop builds them.
	for index, flag := range inputs.Flags {
		inputs.Flags[index] = strings.ReplaceAll(flag, repository, root)
	}
	for _, file := range inputs.Files {
		path := filepath.Join(root, file)
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0700); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(path, []byte("dependency"), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	recipe := filepath.Join(root, "bridge/tsgo/products_test.go")
	if createDirectoryError := os.MkdirAll(filepath.Dir(recipe), 0700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	var keys []string
	fetch := func(version string) string {
		if writeError := os.WriteFile(recipe, []byte(version), 0600); writeError != nil {
			test.Fatal(writeError)
		}
		key, keyError := buildcache.Key(root, inputs)
		if keyError != nil {
			test.Fatal(keyError)
		}
		keys = append(keys, key)
		directory, getError := buildcache.Get(buildcache.Inputs{
			Name:  "bridge-input-proof-" + name,
			Flags: []string{key},
		}, func(directory string) error {
			source, readError := os.ReadFile(recipe)
			if readError != nil {
				return readError
			}
			return os.WriteFile(filepath.Join(directory, "product"), source, 0600)
		})
		if getError != nil {
			test.Fatal(getError)
		}
		data, readError := os.ReadFile(filepath.Join(directory, "product"))
		if readError != nil {
			test.Fatal(readError)
		}
		return string(data)
	}
	if got := fetch("version one"); got != "version one" {
		test.Fatalf("first product: %q", got)
	}
	if got := fetch("version two"); got != "version two" {
		test.Fatalf("%s served stale product: %q, want version two", name, got)
	}
	if keys[0] == keys[1] {
		test.Fatalf("%s kept a stale key after its recipe changed", name)
	}
}

// TestProduct_BridgeInputProofs builds, on Workshop, the two products each TestBridgeProductInput_ reads: for every
// bridge product, the stand-in recipe's version one and version two, under the keys that recipe's inputs give them.
func TestProduct_BridgeInputProofs(test *testing.T) {
	test.Parallel()
	for _, name := range bridgeProductNames {
		checkBridgeProductInput(test, name)
	}
}

func TestBridgeProductUnitsCoverEveryProduct(test *testing.T) {
	test.Parallel()
	if len(bridgeProductNames) != 19 || len(bridgeProductBindings) != 19 {
		test.Fatalf(
			"product coverage: %d products, %d units, want 19",
			len(bridgeProductNames),
			len(bridgeProductBindings),
		)
	}
	for _, name := range bridgeProductNames {
		if bridgeProductBindings[name] == nil {
			test.Fatalf("no product unit for %s", name)
		}
	}
	for _, unit := range bridgeCases {
		for _, name := range bridgeCaseProducts(unit.piece) {
			if bridgeProductBindings[name] == nil {
				test.Fatalf("%s fetches unknown product %s", unit.name, name)
			}
		}
	}
}

var bridgeProductBindings = map[string]func(*testing.T){
	"tsgo.a":         TestProduct_TsgoArchive,
	"tsgo-asan.a":    TestProduct_TsgoSanitizedArchive,
	"length.a":       TestProduct_LengthArchive,
	"stale.a":        TestProduct_StaleArchive,
	"wrong.a":        TestProduct_WrongArchive,
	"leak.a":         TestProduct_LeakArchive,
	"stage0":         TestProduct_Stage0,
	"oracle":         TestProduct_Oracle,
	"api":            TestProduct_ApiDriver,
	"length-driver":  TestProduct_LengthDriver,
	"stale-driver":   TestProduct_StaleDriver,
	"leak-driver":    TestProduct_LeakDriver,
	"native-asan":    TestProduct_SanitizedNative,
	"native":         TestProduct_Native,
	"wrong-native":   TestProduct_WrongNative,
	"healthy-region": TestProduct_HealthyRegion,
	"region-stage0":  TestProduct_RegionStage0,
	"region-native":  TestProduct_RegionNative,
	"linkage.test":   TestProduct_LinkageTest,
}

func TestBridgeProductInput_TsgoArchive(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "tsgo.a")
}

func TestBridgeProductInput_TsgoSanitizedArchive(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "tsgo-asan.a")
}

func TestBridgeProductInput_LengthArchive(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "length.a")
}

func TestBridgeProductInput_StaleArchive(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "stale.a")
}

func TestBridgeProductInput_WrongArchive(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "wrong.a")
}

func TestBridgeProductInput_LeakArchive(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "leak.a")
}

func TestBridgeProductInput_Stage0(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "stage0")
}

func TestBridgeProductInput_Oracle(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "oracle")
}

func TestBridgeProductInput_ApiDriver(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "api")
}

func TestBridgeProductInput_LengthDriver(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "length-driver")
}

func TestBridgeProductInput_StaleDriver(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "stale-driver")
}

func TestBridgeProductInput_LeakDriver(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "leak-driver")
}

func TestBridgeProductInput_SanitizedNative(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "native-asan")
}

func TestBridgeProductInput_Native(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "native")
}

func TestBridgeProductInput_WrongNative(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "wrong-native")
}

func TestBridgeProductInput_HealthyRegion(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "healthy-region")
}

func TestBridgeProductInput_RegionStage0(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "region-stage0")
}

func TestBridgeProductInput_RegionNative(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "region-native")
}

func TestBridgeProductInput_LinkageTest(test *testing.T) {
	test.Parallel()
	checkBridgeProductInput(test, "linkage.test")
}
