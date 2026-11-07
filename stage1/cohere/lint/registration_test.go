package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Validate every descriptor before even a filtered test. No rule can hide behind a filter.
// Not parallel: TestMain owns the shared temporary directory and registry initialization.
func TestMain(suite *testing.M) {
	// The committed registry is checked, never rewritten: the tests don't write the tree (#k01pgng).
	if _, staleError := registry.Stale("."); staleError != nil {
		fmt.Fprintln(os.Stderr, staleError)
		os.Exit(1)
	}
	directory, temporaryDirectoryError := os.MkdirTemp("", "lint-shared-")
	if temporaryDirectoryError != nil {
		fmt.Fprintln(os.Stderr, temporaryDirectoryError)
		os.Exit(1)
	}
	sharedDirectory = directory
	if decodedOptionsBeforeTestsError := decodedOptionsBeforeTests(); decodedOptionsBeforeTestsError != nil {
		fmt.Fprintln(os.Stderr, decodedOptionsBeforeTestsError)
		os.RemoveAll(directory)
		os.Exit(1)
	}
	code := suite.Run()
	if removeError := os.RemoveAll(directory); removeError != nil {
		fmt.Fprintln(os.Stderr, removeError)
	}
	os.Exit(code)
}

func TestRegistrationMutant(test *testing.T) {
	test.Parallel()
	// This remains a valid descriptor and compiled rule: only its subscription is wrong.
	directory := mutant(test, `"DebuggerStatement"`, `"EmptyStatement"`, "rules/no-debugger/rule.json")
	source := ownedWitnesses(test, ".", "no-debugger")[0]
	path := manifest(test, []string{source + "\tno-debugger"})
	want := execute(test, "", lintGoOracleProduct(test), "--manifest", path).output
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(test, directory, path, false)},
		{"emitted JavaScript", emittedNode(test, directory, path, false)},
	} {
		if bytes.Equal(side.run.output, want) {
			test.Fatalf("listener omission survived on %s", side.name)
		}
		test.Logf("wrong descriptor kind caught on %s: %s", side.name, difference(side.run.output, want))
	}
}

// Not parallel: initializes the shared immutable hook fixtures before the parallel shards.
func TestFactoryHooks_Setup(test *testing.T) {
	factoryHooksSetup(test)
}

// Raw corpus text stays outside the project's TypeScript module graph.
func ownedWitnesses(test *testing.T, directory, slug string) []string {
	test.Helper()
	paths, witnessesError := registry.Witnesses(filepath.Join(directory, "rules", slug))
	if witnessesError != nil {
		test.Fatal(witnessesError)
	}
	var sources []string
	for index, path := range paths {
		data, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		target := filepath.Join(
			test.TempDir(),
			fmt.Sprintf("%s-%d%s", slug, index, filepath.Ext(strings.TrimSuffix(path, ".txt"))),
		)
		// A witness in a directory under testdata is linted at that relative path, for a rule that judges it.
		relative, relativePathError := filepath.Rel(filepath.Join(directory, "rules", slug, "testdata"), path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		if filepath.Dir(relative) != "." {
			target = filepath.Join(test.TempDir(), strings.TrimSuffix(relative, ".txt"))
			if createDirectoryError := os.MkdirAll(filepath.Dir(target), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
		}
		if writeError := os.WriteFile(target, data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
		sources = append(sources, target)
	}
	if len(sources) == 0 {
		test.Fatalf("%s has no witnesses", slug)
	}
	return sources
}

// ownedWitnessRows returns a manifest row selecting the rule for each of its witnesses. A witness may carry
// the rule's options beside it, as foo.options.json next to foo.ts.txt: a rule that reports nothing by
// default, such as one that bans only the types it is configured with, can only witness a finding with
// options. They go in the row's options field, as a captured upstream case's do, so the oracle hands them to
// the rule's adapter and the port reads the same settings. The "all" row stays unconfigured, because there
// the options would be every rule's.
func ownedWitnessRows(test *testing.T, directory string, descriptor registry.Descriptor) []string {
	test.Helper()
	paths, witnessesError := registry.Witnesses(filepath.Join(directory, "rules", descriptor.Slug))
	if witnessesError != nil {
		test.Fatal(witnessesError)
	}
	sources := ownedWitnesses(test, directory, descriptor.Slug)
	var rows []string
	for index, path := range paths {
		row := sources[index] + "\t" + descriptor.Name
		sidecar := strings.TrimSuffix(
			strings.TrimSuffix(path, ".txt"),
			filepath.Ext(strings.TrimSuffix(path, ".txt")),
		) +
			".options.json"
		data, readError := os.ReadFile(sidecar)
		if readError == nil {
			var options any
			if decodeError := json.Unmarshal(data, &options); decodeError != nil {
				test.Fatalf("%s: %v", sidecar, decodeError)
			}
			compact, encodeError := json.Marshal(options)
			if encodeError != nil {
				test.Fatal(encodeError)
			}
			row += "\t\t\tfalse\t" + string(compact)
		} else if !os.IsNotExist(readError) {
			test.Fatal(readError)
		}
		rows = append(rows, row)
	}
	return rows
}

func TestNestedOutsideModuleCopy(test *testing.T) {
	test.Parallel()
	directory := mutant(test, "", "")
	original := `import { written } from '../../../../typescript/parser/nodes.ts';
console.log(written('copied'));
`
	source := rewritePortImports(test, "rules/probe/rule.ts", original)
	if writeError := os.WriteFile(filepath.Join(directory, "main.ts"), []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	path := manifest(test, nil)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(test, directory, path, false)},
		{"native", execute(test, "", buildPort(test, directory, true))},
	} {
		if string(side.run.output) != "copied\n" {
			test.Fatalf("outside module copy on %s: %q", side.name, side.run.output)
		}
	}
	// Put the old root-only rewrite back: a nested rule keeps two unwanted parents.
	parserDirectory, absolutePathError := filepath.Abs("../../typescript")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	bad := strings.ReplaceAll(original, "../../typescript", parserDirectory)
	main := filepath.Join(directory, "main.ts")
	if writeError := os.WriteFile(main, []byte(bad), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "bad-copy-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	command := toolchain.Command("node", "--disable-warning=ExperimentalWarning", runner, main)
	command.Stdout = output
	command.Stderr = output
	runError := command.Run()
	if closeError := output.Close(); closeError != nil {
		test.Fatal(closeError)
	}
	data, readError := os.ReadFile(output.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	if runError == nil || !bytes.Contains(data, []byte("ERR_MODULE_NOT_FOUND")) {
		test.Fatalf("root-only import mutant survived on Node: %v %s", runError, data)
	}
	if _, loadError := load.Load([]string{main}); loadError == nil {
		test.Fatal("root-only import mutant survived native module loading")
	}
	test.Log("root-only import rewrite mutant caught by Node and native module loading")
}
