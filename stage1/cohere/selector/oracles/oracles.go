// Package oracles owns the Go reference products without linking the compiler.
// Both test binaries resolve relative paths from the parent package directory.
package oracles

import (
	"bytes"
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
	"strconv"
	"testing"
)

const repository = "../../.."

var oracleProducer bool

// EnableProduction enables product recipes in the owner test binary's TestMain.
func EnableProduction() { oracleProducer = true }
func Parser(t testing.TB) (string, string) {
	t.Helper()
	directory := t.TempDir()
	cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	packageDirectory := filepath.Join(cohere, "internal", "format", "css", "selector")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	const output = "selector.test-oracle-owner-v1"
	arguments := []string{"-overlay=" + overlayPath}
	var binary string
	if oracleProducer {
		binary = buildcache.GoTest(t, "cohere", output, "./internal/format/css/selector", arguments)
	} else {
		inputs, err := buildcache.GoTestInputs("cohere", output, "./internal/format/css/selector", arguments, nil)
		if err != nil {
			t.Fatal(err)
		}
		product := buildcache.BuiltElsewhere(t, output, inputs, "stage1/cohere/selector/oracles TestProduct_SelectorGoOracle", func(string) error { return fmt.Errorf("settle the Go oracle owner first") })
		binary = filepath.Join(product, output)
	}
	return binary, packageDirectory
}

// DefaultSeed fixes the shared generated-selector corpus.
const DefaultSeed = 20261005

func Cases(t *testing.T) (string, string) {
	t.Helper()
	seed := int64(DefaultSeed)
	var err error
	if value := os.Getenv("COHERE_SELECTOR_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			t.Fatal(err)
		}
	}
	generated := 4000
	if value := os.Getenv("COHERE_SELECTOR_GENERATED"); value != "" {
		if generated, err = strconv.Atoi(value); err != nil {
			t.Fatal(err)
		}
	}
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	// The root also contains the Go test constants the generator extracts.
	// Validate it before running that generator, even without the npm oracle.
	files := corpusfiles.Upstream(t, root, corpusfiles.CohereCommit,
		[]string{"internal/format/css", "internal/lint/rules/tailwind"}, []string{"*.css"})
	directory := selectorAnswers(t, seed, generated, root, files, os.Getenv("ADAMIC_SELECTOR_LIBRARY"))
	casesPath := filepath.Join(directory, "cases.txt")
	answers, err := os.ReadFile(filepath.Join(directory, "answers.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_SELECTOR_KEEP"); keep != "" {
		cases, err := os.ReadFile(casesPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keep, cases, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("seed %d, %d generated selectors", seed, generated)
	return casesPath, string(answers)
}

func AnswerInputs(seed int64, generated int, oracleDigest, libraryDigest, node string) buildcache.Inputs {
	flags := []string{fmt.Sprintf("seed=%d", seed), fmt.Sprintf("generated=%d", generated), "oracle-content-sha256=" + oracleDigest}
	toolchain := []string{runtime.Version()}
	if libraryDigest != "" {
		flags = append(flags, "library-content-sha256="+libraryDigest)
		toolchain = append(toolchain, node)
	}
	return buildcache.Inputs{
		Name:      "selector Go oracle answers oracle-owner-v1",
		Files:     []string{"cohere/internal/format/css", "cohere/internal/lint/rules/tailwind", "stage1/cohere/selector/testdata", "stage1/cohere/selector/selector_test.go", "stage1/cohere/selector/oracles/oracles.go"},
		Flags:     flags,
		Toolchain: toolchain,
	}
}

func selectorAnswers(t *testing.T, seed int64, generated int, root string, files []string, library string) string {
	t.Helper()
	oracle, _ := Parser(t)
	var libraryDigest, node string
	if library != "" {
		libraryDigest, node = contentDigest(t, library), buildcache.Tool("node", "--version")
	}
	inputs := AnswerInputs(seed, generated, contentDigest(t, oracle), libraryDigest, node)
	recipe := func(product string) error {
		directory := t.TempDir()
		corpus := filepath.Join(directory, "corpus.txt")
		testTexts := filepath.Join(directory, "test-texts.json")
		request := map[string]any{"seed": seed, "generated": generated, "cases": filepath.Join(product, "cases.txt"), "answers": filepath.Join(product, "answers.txt"), "corpus": corpus, "testTexts": testTexts}
		cohereSide(t, request)
		if library != "" {
			script, err := filepath.Abs(filepath.Join("testdata", "library.mjs"))
			if err != nil {
				return err
			}
			fileList := filepath.Join(directory, "css-files.json")
			encoded, err := json.Marshal(files)
			if err != nil {
				return err
			}
			if err := os.WriteFile(fileList, encoded, 0644); err != nil {
				return err
			}
			command := toolchain.CommandIn("", "node", script, library, "corpus", root, corpus, testTexts, fileList)
			command.Env = nil
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr
			if err := childguard.Run(command, childguard.Options{}); err != nil {
				return fmt.Errorf("corpus: %v: %s", err, stderr.Bytes())
			}
			t.Logf("%s", stdout.Bytes())
			cohereSide(t, request)
		} else {
			t.Log("CSS corpus extraction skipped: set ADAMIC_SELECTOR_LIBRARY")
		}
		return nil
	}
	if oracleProducer {
		return buildcache.Product(t, inputs, recipe)
	}
	return buildcache.BuiltElsewhere(t, inputs.Name, inputs, "stage1/cohere/selector/oracles TestProduct_SelectorGoAnswers", func(string) error { return fmt.Errorf("settle the answers owner first") })
}

func contentDigest(t *testing.T, path string) string {
	t.Helper()
	hash := sha256.New()
	err := filepath.WalkDir(path, func(name string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(path, name)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%s\x00%x\n", filepath.ToSlash(relative), sha256.Sum256(content))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func cohereSide(t *testing.T, request map[string]any) {
	t.Helper()
	directory := t.TempDir()
	requestPath := filepath.Join(directory, "request.json")
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(requestPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	// The package's test binary with the harness laid over it is a product built ahead (cohereSideOracle), run here in
	// the package's directory, where go test would run it; the child guard owns hang detection.
	binary, packageDirectory := Parser(t)
	command := toolchain.CommandIn("", binary, "-test.timeout=0", "-test.v", "-test.run=^TestAdamicPortCases$")
	command.Dir = packageDirectory
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, err := childguard.CombinedOutput(command, childguard.Options{})
	if err != nil {
		t.Fatalf("cohere's side: %v\n%s", err, output)
	}
	t.Logf("Go oracle: %s", output)
}
