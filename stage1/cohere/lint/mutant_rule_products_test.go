package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

type mutantRuleChange struct{ Name, File, From, To string }

func mutantRuleDescriptor(t *testing.T, slug string) registry.Descriptor {
	t.Helper()
	for _, descriptor := range prepareRegistry(t, packageDirectory) {
		if descriptor.Slug == slug {
			return descriptor
		}
	}
	t.Fatalf("mutant shard names unregistered rule %s", slug)
	return registry.Descriptor{}
}

func mutantRuleMutation(t *testing.T, descriptor registry.Descriptor) mutantRuleChange {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("rules", descriptor.Slug, "mutant.json"))
	if err != nil {
		t.Fatal(err)
	}
	var change mutantRuleChange
	if err := json.Unmarshal(data, &change); err != nil {
		t.Fatal(err)
	}
	if change.File == "" {
		change.File = descriptor.Module
	}
	return change
}

var mutantRuleInputsOnce sync.Once
var mutantRuleFiles, mutantRuleTools []string

func mutantRuleInputs(t *testing.T, slug, kind string) buildcache.Inputs {
	t.Helper()
	mutantRuleInputsOnce.Do(func() {
		mutantRuleFiles, mutantRuleTools = ownedWitnessInputs(t)
		mutantRuleFiles = append(mutantRuleFiles,
			"stage1/cohere/lint/mutant_rule_products_test.go", "stage1/cohere/lint/lint_test.go",
			"stage1/cohere/lint/registry", "stage1/cohere/lint/rules",
			"cohere/static_single_assignment", "cohere/mutation_aliasing")
	})
	if len(mutantRuleFiles) == 0 {
		t.Fatal("mutant product input preparation failed earlier in this process")
	}
	return buildcache.Inputs{
		Name:  "lint-mutant-" + kind + "-" + slug,
		Files: append([]string(nil), mutantRuleFiles...),
		// Copied port imports outside this package resolve against this checkout.
		Flags:     []string{"rule=" + slug, "source-root=" + packageDirectory},
		Toolchain: append([]string(nil), mutantRuleTools...),
	}
}

type mutantRuleProductValue struct {
	once      sync.Once
	directory string
}

var mutantRuleProducts sync.Map

// The declaration and its independently selected shard call this one recipe.
// The immutable product contains the mutated Node port and its emitted JavaScript.
func mutantRuleProduct(t *testing.T, slug string) string {
	t.Helper()
	stored, _ := mutantRuleProducts.LoadOrStore(slug, &mutantRuleProductValue{})
	value := stored.(*mutantRuleProductValue)
	value.once.Do(func() {
		descriptor := mutantRuleDescriptor(t, slug)
		change := mutantRuleMutation(t, descriptor)
		value.directory = buildcache.Product(t, mutantRuleInputs(t, slug, "javascript"), func(out string) error {
			started := time.Now()
			port := copyPort(t, filepath.Join(out, "port"), change.From, change.To, filepath.Join("rules", slug, change.File))
			if _, err := registry.Generate(port); err != nil {
				return err
			}
			program, err := load.Load([]string{filepath.Join(port, "main.ts")})
			if err != nil {
				return err
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(out, "program.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
				return err
			}
			t.Logf("cold mutant %s JavaScript=%s", slug, time.Since(started))
			// The port's imports name this checkout for the build's own lowering; the product names it <repository> (#tqrqx60).
			return buildcache.RelativeFiles(out)
		})
	})
	if value.directory == "" {
		t.Fatalf("mutant %s product failed earlier in this process", slug)
	}
	// Node and the loader read the port, so callers get the copy whose imports name this checkout.
	directory, err := buildcache.Resolved(value.directory)
	if err != nil {
		t.Fatal(err)
	}
	return directory
}

var mutantRuleNativeOnce sync.Once
var mutantRuleNativePath string

func mutantRuleNative(t *testing.T) string {
	t.Helper()
	mutantRuleNativeOnce.Do(func() {
		descriptor := mutantRuleDescriptor(t, "react-jsx-no-comment-textnodes")
		if descriptor.Name != nativeCanaryRule {
			t.Fatal("native mutant canary descriptor changed")
		}
		port := filepath.Join(mutantRuleProduct(t, descriptor.Slug), "port")
		inputs := mutantRuleInputs(t, descriptor.Slug, "native")
		options := native.Options{Sanitize: true, Split: true, Jobs: 4}
		inputs.Flags = append(inputs.Flags, native.Flags(options)...)
		inputs.Flags = append(inputs.Flags, "Split=true", "Jobs=4", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		out := buildcache.Product(t, inputs, func(out string) error {
			started := time.Now()
			program, err := load.Load([]string{filepath.Join(port, "main.ts")})
			if err != nil {
				return err
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				return err
			}
			if err := native.Build(native.C(lowered), filepath.Join(out, "native"), options); err != nil {
				return err
			}
			t.Logf("cold native mutant=%s", time.Since(started))
			return nil
		})
		mutantRuleNativePath = filepath.Join(out, "native")
	})
	if mutantRuleNativePath == "" {
		t.Fatal("native mutant product failed earlier in this process")
	}
	return mutantRuleNativePath
}

func TestProduct_LintMutantNativeCanary(t *testing.T) {
	t.Parallel()
	mutantRuleNative(t)
}

// A shard's work uses the runner directly; products have already generated the registry.
func mutantRuleNode(t *testing.T, ctx context.Context, module, manifest string) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return ownedWitnessExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", runner, module, "--manifest", manifest)
}

func mutantRuleKilled(change, side string, got, want []byte) error {
	if string(got) == string(want) {
		return fmt.Errorf("%s mutant survived on %s", change, side)
	}
	return nil
}

func mutantRuleShard(t *testing.T, slug string) {
	t.Helper()
	descriptor := mutantRuleDescriptor(t, slug)
	change := mutantRuleMutation(t, descriptor)
	// Independently selected units fetch every shared product before their work clock.
	oracle := lintGoOracleProduct(t)
	product := mutantRuleProduct(t, slug)
	binary := ""
	if descriptor.Name == nativeCanaryRule {
		binary = mutantRuleNative(t)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { t.Logf("mutant work=%s; rule=%s", time.Since(started), descriptor.Name) }()

	// Preserve the original per-rule corpus, including every generated all-rule row
	// with its complete options bag and the rule's owned witnesses.
	var rows []string
	for _, row := range generated(t) {
		selected := "all"
		if fields := strings.Split(row, "\t"); len(fields) > 1 && fields[1] != "" {
			selected = fields[1]
		}
		if selected == "all" || selected == descriptor.Name {
			rows = append(rows, row)
		}
	}
	rows = append(rows, ownedWitnessRecoveryRows(t, ctx, oracle, ownedWitnessRows(t, packageDirectory, descriptor))...)
	path := manifest(t, rows)
	want := ownedWitnessExecute(t, ctx, oracle, "--manifest", path).output
	mutated := mutantRuleNode(t, ctx, filepath.Join(product, "port/main.ts"), path).output
	emitted := mutantRuleNode(t, ctx, filepath.Join(product, "program.mjs"), path).output
	for _, side := range []struct {
		name   string
		output []byte
	}{{"Node", mutated}, {"emitted JavaScript", emitted}} {
		got := side.output
		// The planted probe replaces a real mutant answer with the original oracle's
		// bytes immediately before the same survivor check used by every normal shard.
		if os.Getenv("ADAMIC_MUTANT_SHARD_PROBE") == slug+":"+side.name {
			got = want
		}
		if err := mutantRuleKilled(change.Name, side.name, got, want); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s caught on %s: %s", change.Name, side.name, difference(got, want))
	}
	if binary != "" {
		got := ownedWitnessExecute(t, ctx, binary, "--manifest", path).output
		if diff := difference(got, mutated); diff != "" {
			t.Fatalf("native canary differs from mutated Node: %s", diff)
		}
		t.Logf("sanitized native canary equals mutated Node: %d bytes", len(got))
	}
}
