package lint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Build-phase units and ordinary shards use these exact recipes and keys.
func TestProduct_WitnessScriptKindLowered(test *testing.T) {
	test.Parallel()
	requireLintArtifacts(test, witnessScriptKindLoweredProduct(test), "main.c", "lint.mjs")
}

func TestProduct_WitnessScriptKindNative(test *testing.T) {
	test.Parallel()
	requireLintExecutable(test, filepath.Join(witnessScriptKindNativeProduct(test), "scanner"))
}

func witnessScriptKindLoweredProduct(test *testing.T) string {
	test.Helper()
	inputs := witnessScriptKindInputs()
	return buildcache.Product(test, inputs, func(out string) error {
		// The manifest selects only no-debugger. Compile that unchanged rule
		// with the unchanged scanner and serializers, avoiding every unrelated
		// rule's lowering. Node and Go still use the complete registry.
		source := copyPort(test, filepath.Join(out, "source"), "", "")
		var selected []registry.Descriptor
		for _, descriptor := range prepareRegistry(test, ".") {
			if descriptor.Slug == "no-debugger" {
				selected = append(selected, descriptor)
			}
		}
		if len(selected) != 1 {
			return fmt.Errorf("no-debugger descriptor count: %d", len(selected))
		}
		ts, _ := registry.Render(selected)
		if err := os.WriteFile(filepath.Join(source, ".generated/registry.ts"), ts, 0644); err != nil {
			return err
		}
		program, err := load.Load([]string{filepath.Join(source, "main.ts")})
		if err != nil {
			return err
		}
		result, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, "main.c"), []byte(native.C(result)), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(
			filepath.Join(out, "lint.mjs"),
			[]byte(javascript.JavaScript(result)),
			0644,
		); err != nil {
			return err
		}
		// The source's imports name this checkout for the lowering above; the product names it <repository> (#tqrqx60).
		return buildcache.RelativeFiles(out)
	})
}

func witnessScriptKindNativeProduct(test *testing.T) string {
	test.Helper()
	lowered := witnessScriptKindLoweredProduct(test)
	inputs := witnessScriptKindInputs()
	options := native.Options{Sanitize: true, Split: true}
	inputs.Name = "witness-script-kind-native-no-debugger-v1"
	inputs.Flags = append(native.Flags(options), "Split=true")
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	return buildcache.Product(test, inputs, func(out string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(out, "scanner"), options)
	})
}
