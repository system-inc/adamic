package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

// A shared archive has one builder package. The owner names the product test
// cmd/adamic waits for in cloud/fast-gate/product-order.json before finding it.
// An empty owner means only cmd/adamic needs this key and builds it here.
type commandRuntimeSpec struct {
	name     string
	options  native.Options
	features []string
	owner    string
}

var commandRuntimes = []commandRuntimeSpec{
	{"shipped", native.Options{Release: true}, nil, "internal/oracle TestProduct_NativeRuntimesShipped0"},
	{"counted", native.Options{Count: true}, nil, "internal/oracle TestProduct_NativeRuntimesCounted0"},
	{"sanitized", native.Options{Sanitize: true}, nil, "internal/oracle TestProduct_NativeRuntimesSanitized0"},
	{"wasm", native.Options{Target: "wasm32-wasi"}, nil, "internal/oracle TestProduct_NativeRuntimesWASI0"},
	// commandRuntime drops Request, so this is the archive internal/oracle's wasm-request variant builds for the empty set.
	{"wasm-counted", native.Options{Target: "wasm32-wasi", Count: true}, nil, "internal/oracle TestProduct_NativeRuntimesWASIRequest0"},
}

func init() { native.SupplyRuntimes(commandRuntime) }

func commandRuntime(request native.RuntimeRequest) (string, error) {
	options := request.Options
	// Request selects the link ABI only. Flags and therefore the runtime key are
	// identical to an ordinary WASI build, whose existing archive we reuse.
	options.Request = false
	// Release does not change these archives, so find their existing products.
	if options.Count || options.Sanitize || options.ThreadSanitize || options.Target != "" {
		options.Release = false
	}
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

// The shared package owns the identity, so readers and builders cannot drift apart.
func commandRuntimeInputs(request native.RuntimeRequest) buildcache.Inputs {
	return runtimeproducts.Inputs(request)
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
		jobserver, err := ccjobserver.Process()
		if err != nil {
			t.Fatal(err)
		}
		build := func(directory string) error { return request.Build(directory, 4, jobserver) }
		if product.owner != "" {
			buildcache.BuiltElsewhere(t, "the "+product.name+" command runtime archive", commandRuntimeInputs(request), product.owner, build)
		} else {
			buildcache.Product(t, commandRuntimeInputs(request), build)
		}
		return
	}
	t.Fatalf("no command runtime product %s", name)
}

func TestProduct_CommandRuntimeShipped(t *testing.T) {
	t.Parallel()
	commandRuntimeProduct(t, "shipped")
}
func TestProduct_CommandRuntimeCounted(t *testing.T) {
	t.Parallel()
	commandRuntimeProduct(t, "counted")
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
