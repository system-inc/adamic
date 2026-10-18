// Package oracles owns GraphQL printer Go reference products without linking the compiler.
// Owner TestMain keeps relative paths rooted in the parent printer package.
package oracles

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/toolchain"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

const repository = "../../../.."

var oracleProducer bool

// EnableProduction permits recipes only in the owner's test binary.
func EnableProduction() { oracleProducer = true }

var printerOracleOnce sync.Once
var printerOraclePath string

func Printer(t *testing.T) string {
	t.Helper()
	printerOracleOnce.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		side, _ := filepath.Abs("testdata/cohere_side_test.go")
		generator, _ := filepath.Abs("../testdata/cohere_side_test.go")
		cohere := filepath.Join(root, "cohere")
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
			cohere + "/internal/format/graphql/adamic_printer_test.go":   side,
			cohere + "/internal/format/graphql/adamic_generator_test.go": generator,
		}})
		if err != nil {
			t.Fatal(err)
		}
		// The overlay names this checkout's paths: it keys and builds the product and is never part of it (#tqrqx60).
		path := filepath.Join(t.TempDir(), "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			t.Fatal(err)
		}
		// Each corpus keys the owner-suffixed binary by its name and bytes.
		const output = "oracle-oracle-owner-v1"
		arguments := []string{"-overlay=" + path}
		if oracleProducer {
			printerOraclePath = buildcache.GoTest(t, "cohere", output, "./internal/format/graphql", arguments)
		} else {
			inputs, err := buildcache.GoTestInputs("cohere", output, "./internal/format/graphql", arguments, nil)
			if err != nil {
				t.Fatal(err)
			}
			product := buildcache.BuiltElsewhere(t, output, inputs, "stage1/cohere/graphql/printer/oracles TestProduct_GraphQLPrinterGoOracle", func(string) error { return fmt.Errorf("settle the Go printer owner first") })
			printerOraclePath = filepath.Join(product, output)
		}
	})
	// A build that failed inside the Once leaves the path empty for every later caller in this process.
	// Say so, rather than running an empty command.
	if printerOraclePath == "" {
		t.Fatal("Go GraphQL printer oracle unavailable: its build failed earlier in this process (see that test's log)")
	}
	return printerOraclePath
}

// printerCorpora holds each mode's corpus once per process: the as-Go, mutant, preflight and whitespace tests all read
// the same Go oracle's answers, so a process runs the git checks, hashes the oracle and looks the product up once per
// mode rather than once per caller (about 30 calls in a package run).
var printerCorpora sync.Map

type printerCorpus struct {
	once       sync.Once
	path, want string
}

var printerUpstream struct {
	once    sync.Once
	checked bool
}

// printerCases is one mode's corpus: the cases file and the Go oracle's answers to it.
func Cases(t *testing.T, mode string) (string, string) {
	t.Helper()
	value, _ := printerCorpora.LoadOrStore(mode, &printerCorpus{})
	corpus := value.(*printerCorpus)
	corpus.once.Do(func() { corpus.path, corpus.want = printerCorpusProduct(t, mode) })
	// A preparation that failed inside the Once leaves the path empty for every later caller in this process.
	if corpus.path == "" {
		t.Fatalf("GraphQL %s corpus unavailable: its preparation failed earlier in this process (see that test's log)", mode)
	}
	return corpus.path, corpus.want
}

func printerCorpusProduct(t *testing.T, mode string) (string, string) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	// Mode and mutant shards use this fixed corpus: Cohere commit
	// 7945d102a6c18dd36adf9114a758ce646e8b2359 plus generated controls
	// from cohere_side_test.go, whose random seed is 20261006. Compare
	// live case totals in the union; guard the pinned file count separately.
	// The pin is the same for every mode, so a process checks it once.
	printerUpstream.once.Do(func() {
		files := corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit,
			[]string{"internal/format/graphql"}, []string{"*_test.go"})
		if len(files) != 3 {
			t.Fatalf("pinned GraphQL corpus: %d files, want 3", len(files))
		}
		t.Log("GraphQL documents: no tracked .graphql corpus; pinned upstream test constants and generated controls")
		printerUpstream.checked = true
	})
	if !printerUpstream.checked {
		t.Fatal("pinned GraphQL corpus: its check failed earlier in this process (see that test's log)")
	}
	oracle := Printer(t)
	inputs := Key(t, Inputs{
		Name:  "GraphQL printer oracle outputs oracle-owner-v1",
		Files: []string{oracle, root + "/stage1/cohere/graphql/printer/printer_test.go", root + "/stage1/cohere/graphql/printer/shards_test.go", root + "/stage1/cohere/graphql/printer/oracles/oracles.go"},
		Flags: []string{"mode=" + mode, "test-run=TestAdamicPrinter"}, Toolchain: runtime.Version(),
	})
	recipe := func(directory string) error {
		request, err := json.Marshal(map[string]any{"Sources": []string(nil), "Cases": directory + "/cases.txt", "Answers": directory + "/answers.txt", "Mode": mode, "Coverage": directory + "/coverage.json"})
		if err != nil {
			return err
		}
		path := filepath.Join(t.TempDir(), "request.json")
		if err := os.WriteFile(path, request, 0644); err != nil {
			return err
		}
		command := toolchain.CommandContextIn(context.Background(), "", oracle, "-test.run=^TestAdamicPrinter$", "-test.count=1", "-test.timeout=0")
		command.Dir = root + "/cohere"
		command.Env = append(os.Environ(), "ADAMIC_PRINTER_REQUEST="+path)
		if output, err := childguard.CombinedOutput(command, childguard.Options{Unbounded: true}); err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		return nil
	}
	var directory string
	if oracleProducer {
		directory = buildcache.Product(t, inputs, recipe)
	} else {
		directory = buildcache.BuiltElsewhere(t, inputs.Name, inputs, "stage1/cohere/graphql/printer/oracles TestProduct_GraphQLPrinter"+strings.ToUpper(mode[:1])+mode[1:], func(string) error { return fmt.Errorf("settle the %s corpus owner first", mode) })
	}
	cases, answers := directory+"/cases.txt", directory+"/answers.txt"
	data, err := os.ReadFile(answers)
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_GRAPHQL_PRINTER_KEEP"); keep != "" {
		for _, name := range []string{"cases.txt", "answers.txt", "coverage.json"} {
			data, err := os.ReadFile(filepath.Join(directory, name))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(keep, name), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return cases, string(data)
}

type Inputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func Key(t *testing.T, inputs Inputs) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	key := buildcache.Inputs{Name: inputs.Name, Flags: append([]string{}, inputs.Flags...), Toolchain: []string{inputs.Toolchain, runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	for _, file := range inputs.Files {
		relative, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			// Generated inputs live in a read-only product outside the repository.
			// Its bytes, rather than the machine-specific cache path, identify it.
			key.Flags = append(key.Flags, fmt.Sprintf("generated:%s:%s", filepath.Base(file), printerGeneratedSum(t, file)))
		} else {
			key.Files = append(key.Files, filepath.ToSlash(relative))
		}
	}
	key.Flags = append(key.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	return key
}

func printerGeneratedSum(t *testing.T, file string) string {
	t.Helper()
	info, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	identity := fmt.Sprintf("%s\x00%d\x00%d", file, info.Size(), info.ModTime().UnixNano())
	value, _ := printerGeneratedSums.LoadOrStore(identity, &printerGeneratedSumOnce{})
	entry := value.(*printerGeneratedSumOnce)
	entry.once.Do(func() {
		var data []byte
		if data, entry.err = os.ReadFile(file); entry.err == nil {
			entry.sum = fmt.Sprintf("%x", sha256.Sum256(data))
		}
	})
	if entry.err != nil {
		t.Fatal(entry.err)
	}
	return entry.sum
}

type printerGeneratedSumOnce struct {
	once sync.Once
	sum  string
	err  error
}

var printerGeneratedSums sync.Map
