package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
)

const helperMutantShards = 4

type helperMutation struct{ file, old, replacement string }

func helperMutations() []helperMutation {
	return []helperMutation{
		{"options_json.ts", "if(char.charCodeAt(0) < 32)", "if(false)"},
		{"option_schema.ts", "matched !== 1", "matched === 0"},
		{"policy_message.ts", "text = text.split(`{{${name}}}`).join(value);", "text = text;"},
		{
			"strict_options.ts",
			"if(field < 0) {\n                    return false;\n                }",
			"if(field < 0) {\n                    continue;\n                }",
		},
	}
}

// One full corpus per mutant preserves every original mutant/case pair.
func helperMutantGroup(index int) []helperMutation {
	var group []helperMutation
	for i, m := range helperMutations() {
		if i%helperMutantShards == index {
			group = append(group, m)
		}
	}
	return group
}

// helperClean is the index of the unmutated native build, after the mutants.
const helperClean = helperMutantShards

var helperProducts [helperMutantShards + 1]struct {
	once   sync.Once
	binary string
}
var helperNativeBuild sync.Mutex

// helperClean is the unmutated native build; other indices name independently built native mutants. Both build-phase
// declarations and shards call this exact recipe and key.
func helperMutantProduct(test *testing.T, index int) string {
	test.Helper()
	product := &helperProducts[index]
	product.once.Do(func() {
		// The clean build is a mutant's recipe and key with no mutation planted.
		name := "helper-native"
		var m helperMutation
		var mutation []string
		if index != helperClean {
			m = helperMutations()[index]
			name = "helper-mutants-" + m.file
			mutation = []string{"mutation=" + m.old + "=>" + m.replacement}
		}
		flags := append(
			append([]string{"native.Build", "Sanitize=true"}, mutation...),
			"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
		)
		tools := []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}
		files := append(
			compilerinputs.Lowering(),
			"stage1/cohere/lint/helpers/helpers_test.go",
			"stage1/cohere/lint/helpers/helper_mutants_test.go",
			"stage1/cohere/lint/helpers/main.ts",
			"stage1/cohere/lint/helpers/options_json.ts",
			"stage1/cohere/lint/helpers/option_schema.ts",
			"stage1/cohere/lint/helpers/policy_message.ts",
			"stage1/cohere/lint/helpers/strict_options.ts",
		)
		directory := buildcache.Product(test, buildcache.Inputs{
			Name:      name,
			Files:     files,
			Flags:     flags,
			Toolchain: tools,
		}, func(directory string) error {
			source := filepath.Join(directory, "source")
			if err := os.Mkdir(source, 0755); err != nil {
				return err
			}
			for _, file := range []string{
				"main.ts",
				"options_json.ts",
				"option_schema.ts",
				"policy_message.ts",
				"strict_options.ts",
			} {
				data, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				if file == m.file {
					if strings.Count(string(data), m.old) != 1 {
						return fmt.Errorf("mutant anchor changed: %s", file)
					}
					data = []byte(strings.Replace(string(data), m.old, m.replacement, 1))
				}
				if err := os.WriteFile(filepath.Join(source, file), data, 0644); err != nil {
					return err
				}
			}
			// native.Build's process-local caches are shared by parallel declarations.
			helperNativeBuild.Lock()
			defer helperNativeBuild.Unlock()
			binary := build(test, source)
			data, err := os.ReadFile(binary)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "binary"), data, 0755)
		})
		product.binary = filepath.Join(directory, "binary")
	})
	return product.binary
}

func TestProduct_HelperMutantsJSON(test *testing.T)    { test.Parallel(); helperMutantProduct(test, 0) }
func TestProduct_HelperMutantsSchema(test *testing.T)  { test.Parallel(); helperMutantProduct(test, 1) }
func TestProduct_HelperMutantsMessage(test *testing.T) { test.Parallel(); helperMutantProduct(test, 2) }
func TestProduct_HelperMutantsStrict(test *testing.T)  { test.Parallel(); helperMutantProduct(test, 3) }
func TestProduct_HelperNative(test *testing.T) {
	test.Parallel()
	helperMutantProduct(test, helperClean)
}

func TestHelperMutants_000(test *testing.T) { test.Parallel(); runHelperMutant(test, 0) }
func TestHelperMutants_001(test *testing.T) { test.Parallel(); runHelperMutant(test, 1) }
func TestHelperMutants_002(test *testing.T) { test.Parallel(); runHelperMutant(test, 2) }
func TestHelperMutants_003(test *testing.T) { test.Parallel(); runHelperMutant(test, 3) }

func requireHelperMutantKilled(test *testing.T, got, want []byte) {
	test.Helper()
	if bytes.Equal(got, want) {
		test.Fatal("compiled mutant survived")
	}
}

func runHelperMutant(test *testing.T, index int) {
	test.Helper()
	if len(helperMutations()) != helperMutantShards {
		test.Fatal("mutant count differs from declared shards")
	}
	// Inject a survivor into each actual top-level child's production assertion.
	if os.Getenv("ADAMIC_HELPER_MUTANT_PROBE") == strconv.Itoa(index) {
		requireHelperMutantKilled(test, []byte("survivor\n"), []byte("survivor\n"))
		return
	}
	setup := time.Now()
	goOracle := oracle(test)
	binary := helperMutantProduct(test, index)
	cases := fixture(test)
	catalog, err := filepath.Abs("testdata/catalog.json")
	if err != nil {
		test.Fatal(err)
	}
	test.Logf("setup %.3fs", time.Since(setup).Seconds())
	// Setup has no test-side deadline. Only executing the prepared corpus does.
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	execute := func(binary string, args ...string) []byte {
		cmd := exec.CommandContext(ctx, binary, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		if err != nil || stderr.Len() != 0 {
			test.Fatalf("%s: %v stderr %s", binary, err, &stderr)
		}
		return output
	}
	want := execute(goOracle, cases)
	got := execute(binary, cases, catalog)
	requireHelperMutantKilled(test, got, want)
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			test.Logf("%s mutant caught at output line %d", helperMutations()[index].file, i+1)
			break
		}
	}
	test.Logf("own %.3fs", time.Since(started).Seconds())
}

func helperUnionValid(groups [][]helperMutation, rows int) bool {
	expected := map[string]bool{}
	for _, m := range helperMutations() {
		for row := 0; row < rows; row++ {
			expected[fmt.Sprintf("%s/%d", m.file, row)] = true
		}
	}
	seen := map[string]bool{}
	for _, group := range groups {
		for _, m := range group {
			for row := 0; row < rows; row++ {
				key := fmt.Sprintf("%s/%d", m.file, row)
				if seen[key] || !expected[key] {
					return false
				}
				seen[key] = true
			}
		}
	}
	return len(seen) == len(expected)
}

func TestHelperMutantsUnion(test *testing.T) {
	test.Parallel()
	if len(helperMutations()) != helperMutantShards {
		test.Fatal("mutant count differs from declared shards")
	}
	var corpus struct{ Cases []json.RawMessage }
	data, err := os.ReadFile(fixture(test))
	if err != nil {
		test.Fatal(err)
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		test.Fatal(err)
	}
	if len(corpus.Cases) == 0 {
		test.Fatal("empty witness corpus")
	}
	groups := make([][]helperMutation, helperMutantShards+1)
	for i := range groups {
		groups[i] = helperMutantGroup(i)
	}
	if !helperUnionValid(groups, len(corpus.Cases)) {
		test.Fatal("union lost or repeated cases")
	}
	// Include an empty group: mutation probes are no-ops there, never false failures.
	for i, group := range groups {
		omitted := append([][]helperMutation(nil), groups...)
		omitted[i] = nil
		if helperUnionValid(omitted, len(corpus.Cases)) != (len(group) == 0) {
			test.Fatalf("omission self-check group %d", i)
		}
		doubled := append(append([][]helperMutation(nil), groups...), group)
		if helperUnionValid(doubled, len(corpus.Cases)) != (len(group) == 0) {
			test.Fatalf("duplicate self-check group %d", i)
		}
	}
	test.Logf(
		"union: %d mutants x %d cases = %d unique pairs",
		len(helperMutations()),
		len(corpus.Cases),
		len(helperMutations())*len(corpus.Cases),
	)
}

func TestHelperMutantsPlantedFailure(test *testing.T) {
	test.Parallel()
	binary, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	for index := 0; index < helperMutantShards; index++ {
		name := fmt.Sprintf("TestHelperMutants_%03d", index)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := exec.CommandContext(ctx, binary, "-test.run=^"+name+"$", "-test.timeout=9s", "-test.v")
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "ADAMIC_HELPER_MUTANT_PROBE=") {
				cmd.Env = append(cmd.Env, value)
			}
		}
		cmd.Env = append(cmd.Env, "ADAMIC_HELPER_MUTANT_PROBE="+strconv.Itoa(index))
		output, err := cmd.CombinedOutput()
		cancel()
		if err == nil ||
			!bytes.Contains(output, []byte("compiled mutant survived")) ||
			!bytes.Contains(output, []byte("--- FAIL: "+name)) {
			test.Fatalf("%s missed planted failure: %v\n%s", name, err, output)
		}
		test.Logf("%s rejected planted survivor", name)
	}
}
