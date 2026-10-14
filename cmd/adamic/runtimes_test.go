package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// These are references to native-runtime products, with exactly the oracle's key and
// recipe. Shard numbers are deliberately absent: runtime's six-shard repartition
// changes which unit builds an archive, never the archive a command test reads.
type commandRuntimeSpec struct {
	name     string
	options  native.Options
	features []string
}

var commandRuntimes = []commandRuntimeSpec{
	{"sanitized", native.Options{Sanitize: true}, nil},
	{"wasm", native.Options{Target: "wasm32-wasi"}, nil},
	{"wasm-counted", native.Options{Target: "wasm32-wasi", Count: true}, nil},
}

func init() { native.SupplyRuntimes(commandRuntime) }

func commandRuntime(request native.RuntimeRequest) (string, error) {
	options := request.Options
	// Request selects the link ABI only. Flags and therefore the runtime key are
	// identical to an ordinary WASI build, whose existing archive we reuse.
	options.Request = false
	declared := slices.ContainsFunc(commandRuntimes, func(product commandRuntimeSpec) bool {
		return product.options == options && slices.Equal(product.features, request.Features)
	})
	if !declared {
		return "", fmt.Errorf("cmd/adamic: no runtime product declared for features %v with options %+v (commandRuntimes)", request.Features, request.Options)
	}
	// Even an ordinary developer test must build products ahead. A cache miss must
	// never compile C inside a behavior test, including its driver subprocesses.
	return buildcache.Get(commandRuntimeInputs(request), func(string) error {
		return fmt.Errorf("cmd/adamic: runtime product not built ahead for features %v with options %+v; run TestProduct_CommandRuntime tests first", request.Features, request.Options)
	})
}

// Keep the native-runtime identity shared with internal/oracle/runtimes_test.go.
// Portable holds library.go's key parts without machine-specific tool paths;
// tool reports key the compiler and archiver instead.
func commandRuntimeInputs(request native.RuntimeRequest) buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "native-runtime",
		Files:     []string{"internal/native/runtime", "internal/native/library.go", "internal/native/native.go", "internal/native/target.go"},
		Flags:     slices.Clone(request.Portable),
		Toolchain: []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version")},
	}
}

func commandRuntimeProduct(t *testing.T, name string) {
	t.Helper()
	for _, product := range commandRuntimes {
		if product.name != name {
			continue
		}
		if product.options.Target != "" && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			t.Skip("set ADAMIC_ORACLE_WASI=1")
		}
		request, err := native.RuntimeRequestFor(product.options, product.features...)
		if err != nil {
			t.Fatal(err)
		}
		buildcache.Product(t, commandRuntimeInputs(request), func(directory string) error { return request.Build(directory, 4) })
		return
	}
	t.Fatalf("no command runtime product %s", name)
}

func TestProduct_CommandRuntimeSanitized(t *testing.T) {
	t.Parallel()
	commandRuntimeProduct(t, "sanitized")
}
func TestProduct_CommandRuntimeWASI(t *testing.T) { t.Parallel(); commandRuntimeProduct(t, "wasm") }
func TestProduct_CommandRuntimeWASICounted(t *testing.T) {
	t.Parallel()
	commandRuntimeProduct(t, "wasm-counted")
}

// Dropping any native key part must change the product identity. Mutant: omit
// Portable's first part when declaring the product.
func TestCommandRuntimeProductKeyHoldsEveryPart(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	request, err := native.RuntimeRequestFor(native.Options{Sanitize: true})
	if err != nil {
		t.Fatal(err)
	}
	key := buildcache.NameKey(root, commandRuntimeInputs(request))
	for index, part := range request.Portable {
		changed := request
		changed.Portable = slices.Clone(request.Portable)
		changed.Portable[index] += "!"
		if buildcache.NameKey(root, commandRuntimeInputs(changed)) == key {
			t.Errorf("native key part %.60q is missing from the product key", part)
		}
	}
}
