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
	"github.com/system-inc/adamic/internal/childguard"
)

const helperMutantShards = 4

type helperMutation struct{ file, old, replacement string }

func helperMutations() []helperMutation {
	return []helperMutation{
		{"options_json.ts", "if(char.charCodeAt(0) < 32)", "if(false)"},
		{"option_schema.ts", "matched !== 1", "matched === 0"},
		{"policy_message.ts", "text = text.split(`{{${name}}}`).join(value);", "text = text;"},
		{"strict_options.ts", "if(field < 0) { return false; }", "if(field < 0) { continue; }"},
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

var helperProducts [helperMutantShards + 1]struct {
	once   sync.Once
	binary string
}
var helperNativeBuild sync.Mutex

// index -1 is the Go oracle; other indices name independently built native mutants.
// Both build-phase declarations and shards call this exact recipe and key.
func helperMutantProduct(t *testing.T, index int) string {
	t.Helper()
	product := &helperProducts[index+1]
	product.once.Do(func() {
		name := "helper-mutants-oracle"
		flags := []string{"go build", "overlay=oracle.go,catalog.go,descriptors.go"}
		tools := []string{buildcache.Tool("go", "version")}
		files := []string{"cohere", "stage1/cohere/lint/helpers/testdata/oracle.go", "stage1/cohere/lint/helpers/testdata/catalog.go", "stage1/cohere/lint/helpers/testdata/descriptors.go", "stage1/cohere/lint/helpers/helpers_test.go", "stage1/cohere/lint/helpers/helper_mutants_test.go"}
		if index >= 0 {
			m := helperMutations()[index]
			name = "helper-mutants-" + m.file
			flags = []string{"native.Build", "Sanitize=true", "mutation=" + m.old + "=>" + m.replacement}
			tools = append(tools, buildcache.Tool("clang", "--version"))
			files = append(files, "go.mod", "go.work", "internal", "stage1/cohere/lint/helpers/main.ts", "stage1/cohere/lint/helpers/options_json.ts", "stage1/cohere/lint/helpers/option_schema.ts", "stage1/cohere/lint/helpers/policy_message.ts", "stage1/cohere/lint/helpers/strict_options.ts")
		}
		directory := buildcache.Product(t, buildcache.Inputs{Name: name, Files: files, Flags: flags, Toolchain: tools}, func(directory string) error {
			var binary string
			if index < 0 {
				binary = oracle(t)
			} else {
				m := helperMutations()[index]
				source := filepath.Join(directory, "source")
				if err := os.Mkdir(source, 0755); err != nil {
					return err
				}
				for _, file := range []string{"main.ts", "options_json.ts", "option_schema.ts", "policy_message.ts", "strict_options.ts"} {
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
				binary = build(t, source)
			}
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

func TestProduct_HelperMutantsOracle(t *testing.T)  { t.Parallel(); helperMutantProduct(t, -1) }
func TestProduct_HelperMutantsJSON(t *testing.T)    { t.Parallel(); helperMutantProduct(t, 0) }
func TestProduct_HelperMutantsSchema(t *testing.T)  { t.Parallel(); helperMutantProduct(t, 1) }
func TestProduct_HelperMutantsMessage(t *testing.T) { t.Parallel(); helperMutantProduct(t, 2) }
func TestProduct_HelperMutantsStrict(t *testing.T)  { t.Parallel(); helperMutantProduct(t, 3) }

func TestHelperMutants_000(t *testing.T) { t.Parallel(); runHelperMutant(t, 0) }
func TestHelperMutants_001(t *testing.T) { t.Parallel(); runHelperMutant(t, 1) }
func TestHelperMutants_002(t *testing.T) { t.Parallel(); runHelperMutant(t, 2) }
func TestHelperMutants_003(t *testing.T) { t.Parallel(); runHelperMutant(t, 3) }

func requireHelperMutantKilled(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		t.Fatal("compiled mutant survived")
	}
}

func runHelperMutant(t *testing.T, index int) {
	t.Helper()
	if len(helperMutations()) != helperMutantShards {
		t.Fatal("mutant count differs from declared shards")
	}
	// Inject a survivor into each actual top-level child's production assertion.
	if os.Getenv("ADAMIC_HELPER_MUTANT_PROBE") == strconv.Itoa(index) {
		requireHelperMutantKilled(t, []byte("survivor\n"), []byte("survivor\n"))
		return
	}
	setup := time.Now()
	goOracle := helperMutantProduct(t, -1)
	binary := helperMutantProduct(t, index)
	cases := fixture(t)
	catalog, err := filepath.Abs("testdata/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("setup %.3fs", time.Since(setup).Seconds())
	// Own work is logged, not asserted. Each child runs under childguard's stall and ceiling, as
	// helpers_test.go's do, which bound a hang and name it; a 30 s context killed a loaded but healthy
	// run and reported only "signal: killed" (#he9xrrn).
	started := time.Now()
	execute := func(binary string, args ...string) []byte {
		cmd := exec.Command(binary, args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := childguard.Run(cmd, childguard.Options{}); err != nil || stderr.Len() != 0 {
			t.Fatalf("%s: %v stderr %s", binary, err, &stderr)
		}
		return stdout.Bytes()
	}
	want := execute(goOracle, cases)
	got := execute(binary, cases, catalog)
	requireHelperMutantKilled(t, got, want)
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Logf("%s mutant caught at output line %d", helperMutations()[index].file, i+1)
			break
		}
	}
	t.Logf("own %.3fs", time.Since(started).Seconds())
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

func TestHelperMutantsUnion(t *testing.T) {
	t.Parallel()
	if len(helperMutations()) != helperMutantShards {
		t.Fatal("mutant count differs from declared shards")
	}
	var corpus struct{ Cases []json.RawMessage }
	data, err := os.ReadFile(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) == 0 {
		t.Fatal("empty witness corpus")
	}
	groups := make([][]helperMutation, helperMutantShards+1)
	for i := range groups {
		groups[i] = helperMutantGroup(i)
	}
	if !helperUnionValid(groups, len(corpus.Cases)) {
		t.Fatal("union lost or repeated cases")
	}
	// Include an empty group: mutation probes are no-ops there, never false failures.
	for i, group := range groups {
		omitted := append([][]helperMutation(nil), groups...)
		omitted[i] = nil
		if helperUnionValid(omitted, len(corpus.Cases)) != (len(group) == 0) {
			t.Fatalf("omission self-check group %d", i)
		}
		doubled := append(append([][]helperMutation(nil), groups...), group)
		if helperUnionValid(doubled, len(corpus.Cases)) != (len(group) == 0) {
			t.Fatalf("duplicate self-check group %d", i)
		}
	}
	t.Logf("union: %d mutants x %d cases = %d unique pairs", len(helperMutations()), len(corpus.Cases), len(helperMutations())*len(corpus.Cases))
}

func TestHelperMutantsPlantedFailure(t *testing.T) {
	t.Parallel()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
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
		if err == nil || !bytes.Contains(output, []byte("compiled mutant survived")) || !bytes.Contains(output, []byte("--- FAIL: "+name)) {
			t.Fatalf("%s missed planted failure: %v\n%s", name, err, output)
		}
		t.Logf("%s rejected planted survivor", name)
	}
}
