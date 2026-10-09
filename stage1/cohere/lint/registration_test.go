package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Validate every descriptor before even a filtered test. No rule can hide behind a filter.
func TestMain(m *testing.M) {
	descriptors, err := registry.Generate(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := validateRuleScope(descriptors); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	directory, err := os.MkdirTemp("", "lint-shared-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sharedDirectory = directory
	code := m.Run()
	cleanupCheckerArchives()
	if err := os.RemoveAll(directory); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}

const testOwnedWitnessesShards = 16

// Fixed shard count with headroom; live repository witnesses use stable-key hashing.
// ADAMIC_TEST_SHARD=i/n selects every nth shard starting at i; unset runs all.
// Build products are prepared once before the parallel, independently runnable units.
func TestOwnedWitnesses(t *testing.T) {
	t.Parallel()
	started := time.Now()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	cases := ownedWitnessCases(t, directory)
	plan, err := planOwnedWitnesses(cases, testOwnedWitnessesShards)
	if err != nil {
		t.Fatal(err)
	}
	selected := ownedShardSelection(t)
	oracle, binary, module, builds := ownedBuildProducts(t, directory)
	t.Logf("setup: with builds=%s without builds=%s; union=%d witnesses, %d manifest cases; shards=%d", time.Since(started), time.Since(started)-builds, len(cases), 2*len(cases), len(plan))
	enumerated := 0
	for index, group := range plan {
		enumerated++
		if !selected(index) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) {
			t.Parallel()
			var rows []string
			for _, witness := range group {
				row, d := witness.row, witness.descriptor
				path := strings.SplitN(row, "\t", 2)[0]
				if d.Typed {
					config := filepath.Join(t.TempDir(), "tsconfig.json")
					options := fmt.Sprintf(`{"compilerOptions":{"strict":true},"files":[%q]}`, path)
					if err := os.WriteFile(config, []byte(options), 0644); err != nil {
						t.Fatal(err)
					}
					projectManifest := manifest(t, []string{"program " + config, row, path + "\tall"})
					compareOwnedWitness(t, oracle, binary, directory, projectManifest, module, witness.id)
					answer := execute(t, "", oracle, "--manifest", manifest(t, []string{"program " + config, row}), "--count")
					if string(answer.output) == "0\n" {
						t.Fatalf("%s typed witness reports no findings", d.Name)
					}
					continue
				}
				pair := recoveryRows(t, oracle, []string{row, path + "\tall"})
				answer := execute(t, "", oracle, "--manifest", manifest(t, pair[:1]), "--count")
				if string(answer.output) == "0\n" {
					t.Fatalf("%s witness reports no findings", d.Name)
				}
				rows = append(rows, pair...)
			}
			if len(rows) != 0 {
				compareOwnedWitness(t, oracle, binary, directory, manifest(t, rows), module, group[0].id)
			}
		})
	}
	if enumerated != testOwnedWitnessesShards {
		t.Fatalf("enumerated %d shards, want %d", enumerated, testOwnedWitnessesShards)
	}
}

func TestRegistrationMutant(t *testing.T) {
	t.Parallel()
	// This remains a valid descriptor and compiled rule: only its subscription is wrong.
	directory := mutant(t, `"DebuggerStatement"`, `"EmptyStatement"`, "rules/no-debugger/rule.json")
	source := ownedWitnesses(t, ".", "no-debugger")[0]
	path := manifest(t, []string{source + "\tno-debugger"})
	want := execute(t, "", goOracle(t), "--manifest", path).output
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"emitted JavaScript", emittedNode(t, directory, path, false)},
	} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("listener omission survived on %s", side.name)
		}
		t.Logf("wrong descriptor kind caught on %s: %s", side.name, difference(side.run.output, want))
	}
}

const testFactoryHooksShards = 2

// ADAMIC_TEST_SHARD=i/n selects every nth shard starting at i; unset runs all.
// The intact hook sequence and omission mutant use immutable, prebuilt products.
func TestFactoryHooks(t *testing.T) {
	testFactoryHooksUnits(t)
}

// Raw corpus text stays outside the project's TypeScript module graph.
func ownedWitnesses(t *testing.T, directory, slug string) []string {
	t.Helper()
	paths, err := registry.Witnesses(filepath.Join(directory, "rules", slug))
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), fmt.Sprintf("%s-%d%s", slug, index, filepath.Ext(strings.TrimSuffix(path, ".txt"))))
		// A witness in a directory under testdata is linted at that relative path, for a rule that judges it.
		relative, err := filepath.Rel(filepath.Join(directory, "rules", slug, "testdata"), path)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Dir(relative) != "." {
			target = filepath.Join(t.TempDir(), strings.TrimSuffix(relative, ".txt"))
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			t.Fatal(err)
		}
		sources = append(sources, target)
	}
	if len(sources) == 0 {
		t.Fatalf("%s has no witnesses", slug)
	}
	return sources
}

// ownedWitnessRows returns a manifest row selecting the rule for each of its witnesses. A witness may carry
// the rule's options beside it, as foo.options.json next to foo.ts.txt: a rule that reports nothing by
// default, such as one that bans only the types it is configured with, can only witness a finding with
// options. They go in the row's options field, as a captured upstream case's do, so the oracle hands them to
// the rule's adapter and the port reads the same settings. The "all" row stays unconfigured, because there
// the options would be every rule's.
func ownedWitnessRows(t *testing.T, directory string, d registry.Descriptor) []string {
	t.Helper()
	paths, err := registry.Witnesses(filepath.Join(directory, "rules", d.Slug))
	if err != nil {
		t.Fatal(err)
	}
	sources := ownedWitnesses(t, directory, d.Slug)
	var rows []string
	for index, path := range paths {
		row := sources[index] + "\t" + d.Name
		sidecar := strings.TrimSuffix(strings.TrimSuffix(path, ".txt"), filepath.Ext(strings.TrimSuffix(path, ".txt"))) +
			".options.json"
		data, err := os.ReadFile(sidecar)
		if err == nil {
			var options any
			if err := json.Unmarshal(data, &options); err != nil {
				t.Fatalf("%s: %v", sidecar, err)
			}
			compact, err := json.Marshal(options)
			if err != nil {
				t.Fatal(err)
			}
			row += "\t\t\tfalse\t" + string(compact)
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		rows = append(rows, row)
	}
	return rows
}

func TestNestedOutsideModuleCopy(t *testing.T) {
	t.Parallel()
	directory := mutant(t, "", "")
	original := `import { written } from '../../../../typescript/parser/nodes.ts';
console.log(written('copied'));
`
	source := rewritePortImports(t, "rules/probe/rule.ts", original)
	if err := os.WriteFile(filepath.Join(directory, "main.ts"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, nil)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"native", execute(t, "", buildPort(t, directory, true))},
	} {
		if string(side.run.output) != "copied\n" {
			t.Fatalf("outside module copy on %s: %q", side.name, side.run.output)
		}
	}
	// Put the old root-only rewrite back: a nested rule keeps two unwanted parents.
	parserDirectory, err := filepath.Abs("../../typescript")
	if err != nil {
		t.Fatal(err)
	}
	bad := strings.ReplaceAll(original, "../../typescript", parserDirectory)
	main := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(main, []byte(bad), 0644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	output, err := os.CreateTemp(t.TempDir(), "bad-copy-")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, main)
	command.Stdout = output
	command.Stderr = output
	runError := command.Run()
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	if runError == nil || !bytes.Contains(data, []byte("ERR_MODULE_NOT_FOUND")) {
		t.Fatalf("root-only import mutant survived on Node: %v %s", runError, data)
	}
	if _, err := load.Load([]string{main}); err == nil {
		t.Fatal("root-only import mutant survived native module loading")
	}
	t.Log("root-only import rewrite mutant caught by Node and native module loading")
}

const testDecodedOptionsAndMutantShards = 2

// ADAMIC_TEST_SHARD=i/n selects every nth shard starting at i; unset runs all.
// Decoded-option agreement and its omission mutant consume immutable build products.
func TestDecodedOptionsAndMutant(t *testing.T) {
	testDecodedOptionsUnits(t)
}
