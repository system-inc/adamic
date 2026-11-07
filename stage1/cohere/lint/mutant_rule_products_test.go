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
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

type mutantRuleChange struct{ Name, File, From, To string }

func mutantRuleDescriptor(test *testing.T, slug string) registry.Descriptor {
	test.Helper()
	for _, descriptor := range prepareRegistry(test, packageDirectory) {
		if descriptor.Slug == slug {
			return descriptor
		}
	}
	test.Fatalf("mutant shard names unregistered rule %s", slug)
	return registry.Descriptor{}
}

func mutantRuleMutation(test *testing.T, descriptor registry.Descriptor) mutantRuleChange {
	test.Helper()
	data, readError := os.ReadFile(filepath.Join("rules", descriptor.Slug, "mutant.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var change mutantRuleChange
	if decodeError := json.Unmarshal(data, &change); decodeError != nil {
		test.Fatal(decodeError)
	}
	if change.File == "" {
		change.File = descriptor.Module
	}
	return change
}

var mutantRuleInputsOnce sync.Once
var mutantRuleFiles, mutantRuleTools []string

func mutantRuleInputs(test *testing.T, slug string) buildcache.Inputs {
	test.Helper()
	mutantRuleInputsOnce.Do(func() {
		mutantRuleFiles, mutantRuleTools = ownedWitnessInputs(test)
		mutantRuleFiles = append(mutantRuleFiles,
			"stage1/cohere/lint/mutant_rule_products_test.go", "stage1/cohere/lint/lint_test.go",
			"stage1/cohere/lint/registry", "stage1/cohere/lint/rules",
			"cohere/static_single_assignment", "cohere/mutation_aliasing")
	})
	if len(mutantRuleFiles) == 0 {
		test.Fatal("mutant product input preparation failed earlier in this process")
	}
	return buildcache.Inputs{
		Name:  "lint-mutant-javascript-" + slug,
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
func mutantRuleProduct(test *testing.T, slug string) string {
	test.Helper()
	stored, _ := mutantRuleProducts.LoadOrStore(slug, &mutantRuleProductValue{})
	value := stored.(*mutantRuleProductValue)
	value.once.Do(func() {
		descriptor := mutantRuleDescriptor(test, slug)
		change := mutantRuleMutation(test, descriptor)
		value.directory = buildcache.Product(test, mutantRuleInputs(test, slug), func(out string) error {
			started := time.Now()
			port := copyPort(
				test,
				filepath.Join(out, "port"),
				change.From,
				change.To,
				filepath.Join("rules", slug, change.File),
			)
			if _, generateError := registry.Generate(port); generateError != nil {
				return generateError
			}
			program, loadError := load.Load([]string{filepath.Join(port, "main.ts")})
			if loadError != nil {
				return loadError
			}
			lowered, lowerError := lower.Lower(context.Background(), program)
			if lowerError != nil {
				return lowerError
			}
			if writeError := os.WriteFile(
				filepath.Join(out, "program.mjs"),
				[]byte(javascript.JavaScript(lowered)),
				0644,
			); writeError != nil {
				return writeError
			}
			test.Logf("cold mutant %s JavaScript=%s", slug, time.Since(started))
			// The port's imports name this checkout for the build's own lowering; the product names it <repository>
			// (#tqrqx60).
			return buildcache.RelativeFiles(out)
		})
	})
	if value.directory == "" {
		test.Fatalf("mutant %s product failed earlier in this process", slug)
	}
	// Node and the loader read the port, so callers get the copy whose imports name this checkout.
	directory, resolvedError := buildcache.Resolved(value.directory)
	if resolvedError != nil {
		test.Fatal(resolvedError)
	}
	return directory
}

var mutantRuleNativeOnce sync.Once
var mutantRuleNativePath string

// The canary's native is the yepesta product's: the same mutation of the same rule, built with the same options from
// its relocatable port, so the canary runs that binary rather than building the mutant again from a port of its own.
func mutantRuleNative(test *testing.T) string {
	test.Helper()
	mutantRuleNativeOnce.Do(func() {
		descriptor := mutantRuleDescriptor(test, "react-jsx-no-comment-textnodes")
		if descriptor.Name != nativeCanaryRule {
			test.Fatal("native mutant canary descriptor changed")
		}
		mutantRuleNativePath = jsxTextnodesFoundNative(test)
	})
	if mutantRuleNativePath == "" {
		test.Fatal("native mutant product failed earlier in this process")
	}
	return mutantRuleNativePath
}

func TestProduct_LintMutantNativeCanary(test *testing.T) {
	test.Parallel()
	requireLintExecutable(test, mutantRuleNative(test))
}

// A shard's work uses the runner directly; products have already generated the registry.
func mutantRuleNode(test *testing.T, callerContext context.Context, module, manifest string) execution {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return ownedWitnessExecute(
		test,
		callerContext,
		"node",
		"--disable-warning=ExperimentalWarning",
		runner,
		module,
		"--manifest",
		manifest,
	)
}

func mutantRuleKilled(change, side string, got, want []byte) error {
	if string(got) == string(want) {
		return fmt.Errorf("%s mutant survived on %s", change, side)
	}
	return nil
}

func mutantRuleShard(test *testing.T, slug string) {
	test.Helper()
	descriptor := mutantRuleDescriptor(test, slug)
	change := mutantRuleMutation(test, descriptor)
	// Independently selected units fetch every shared product before their work clock.
	oracle := lintGoOracleProduct(test)
	product := mutantRuleProduct(test, slug)
	binary := ""
	if descriptor.Name == nativeCanaryRule {
		binary = mutantRuleNative(test)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { test.Logf("mutant work=%s; rule=%s", time.Since(started), descriptor.Name) }()

	// Preserve the original per-rule corpus, including every generated all-rule row
	// with its complete options bag and the rule's owned witnesses.
	var rows []string
	for _, row := range generated(test) {
		selected := "all"
		if fields := strings.Split(row, "\t"); len(fields) > 1 && fields[1] != "" {
			selected = fields[1]
		}
		if selected == "all" || selected == descriptor.Name {
			rows = append(rows, row)
		}
	}
	rows = append(
		rows,
		ownedWitnessRecoveryRows(test, timeoutContext, oracle, ownedWitnessRows(test, packageDirectory, descriptor))...,
	)
	path := manifest(test, rows)
	want := ownedWitnessExecute(test, timeoutContext, oracle, "--manifest", path).output
	mutated := mutantRuleNode(test, timeoutContext, filepath.Join(product, "port/main.ts"), path).output
	emitted := mutantRuleNode(test, timeoutContext, filepath.Join(product, "program.mjs"), path).output
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
		if mutantRuleKilledError := mutantRuleKilled(change.Name, side.name, got, want); mutantRuleKilledError != nil {
			test.Fatal(mutantRuleKilledError)
		}
		test.Logf("%s caught on %s: %s", change.Name, side.name, difference(got, want))
	}
	if binary != "" {
		got := ownedWitnessExecute(test, timeoutContext, binary, "--manifest", path).output
		if diff := difference(got, mutated); diff != "" {
			test.Fatalf("native canary differs from mutated Node: %s", diff)
		}
		test.Logf("sanitized native canary equals mutated Node: %d bytes", len(got))
	}
}
