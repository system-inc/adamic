package tsgo_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// A small source tree stands in for the recipe's real dependencies. Its exact
// Files, Flags and Toolchain feed Key, then Get. Changing the recipe source must
// fetch a new product; omitting it returns version one and fails this comparison.
func checkBridgeProductInput(t *testing.T, name string) {
	t.Helper()
	inputs := bridgeProductInputs(t, bridgeRepository(t), name)
	root := t.TempDir()
	for _, file := range inputs.Files {
		path := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("dependency"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	recipe := filepath.Join(root, "bridge/tsgo/products_test.go")
	if err := os.MkdirAll(filepath.Dir(recipe), 0700); err != nil {
		t.Fatal(err)
	}
	var keys []string
	fetch := func(version string) string {
		if err := os.WriteFile(recipe, []byte(version), 0600); err != nil {
			t.Fatal(err)
		}
		key, err := buildcache.Key(root, inputs)
		if err != nil {
			t.Fatal(err)
		}
		keys = append(keys, key)
		directory, err := buildcache.Get(buildcache.Inputs{Name: "bridge-input-proof-" + name, Flags: []string{root, key}}, func(directory string) error {
			source, err := os.ReadFile(recipe)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "product"), source, 0600)
		})
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(directory, "product"))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	if got := fetch("version one"); got != "version one" {
		t.Fatalf("first product: %q", got)
	}
	if got := fetch("version two"); got != "version two" {
		t.Fatalf("%s served stale product: %q, want version two", name, got)
	}
	if keys[0] == keys[1] {
		t.Fatalf("%s kept a stale key after its recipe changed", name)
	}
}

func TestBridgeProductUnitsCoverEveryProduct(t *testing.T) {
	t.Parallel()
	if len(bridgeProductNames) != 19 || len(bridgeProductBindings) != 19 {
		t.Fatalf("product coverage: %d products, %d units, want 19", len(bridgeProductNames), len(bridgeProductBindings))
	}
	for _, name := range bridgeProductNames {
		if bridgeProductBindings[name] == nil {
			t.Fatalf("no product unit for %s", name)
		}
	}
	for _, unit := range bridgeCases {
		for _, name := range bridgeCaseProducts(unit.piece) {
			if bridgeProductBindings[name] == nil {
				t.Fatalf("%s fetches unknown product %s", unit.name, name)
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

func TestBridgeProductInput_TsgoArchive(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "tsgo.a")
}

func TestBridgeProductInput_TsgoSanitizedArchive(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "tsgo-asan.a")
}

func TestBridgeProductInput_LengthArchive(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "length.a")
}

func TestBridgeProductInput_StaleArchive(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "stale.a")
}

func TestBridgeProductInput_WrongArchive(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "wrong.a")
}

func TestBridgeProductInput_LeakArchive(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "leak.a")
}

func TestBridgeProductInput_Stage0(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "stage0")
}

func TestBridgeProductInput_Oracle(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "oracle")
}

func TestBridgeProductInput_ApiDriver(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "api")
}

func TestBridgeProductInput_LengthDriver(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "length-driver")
}

func TestBridgeProductInput_StaleDriver(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "stale-driver")
}

func TestBridgeProductInput_LeakDriver(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "leak-driver")
}

func TestBridgeProductInput_SanitizedNative(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "native-asan")
}

func TestBridgeProductInput_Native(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "native")
}

func TestBridgeProductInput_WrongNative(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "wrong-native")
}

func TestBridgeProductInput_HealthyRegion(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "healthy-region")
}

func TestBridgeProductInput_RegionStage0(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "region-stage0")
}

func TestBridgeProductInput_RegionNative(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "region-native")
}

func TestBridgeProductInput_LinkageTest(t *testing.T) {
	t.Parallel()
	checkBridgeProductInput(t, "linkage.test")
}
