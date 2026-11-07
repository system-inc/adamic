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
func Parser(test testing.TB) (string, string) {
	test.Helper()
	directory := test.TempDir()
	cohere, absolutePathError := filepath.Abs(filepath.Join(repository, "cohere"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	side, absolutePathError := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	packageDirectory := filepath.Join(cohere, "internal", "format", "css", "selector")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	const output = "selector.test-oracle-owner-v1"
	arguments := []string{"-overlay=" + overlayPath}
	var binary string
	if oracleProducer {
		binary = buildcache.GoTest(test, "cohere", output, "./internal/format/css/selector", arguments)
	} else {
		inputs, goTestInputsError := buildcache.GoTestInputs(
			"cohere",
			output,
			"./internal/format/css/selector",
			arguments,
			nil,
		)
		if goTestInputsError != nil {
			test.Fatal(goTestInputsError)
		}
		product := buildcache.BuiltElsewhere(
			test,
			output,
			inputs,
			"stage1/cohere/selector/oracles TestProduct_SelectorGoOracle",
			func(string) error { return fmt.Errorf("settle the Go oracle owner first") },
		)
		binary = filepath.Join(product, output)
	}
	return binary, packageDirectory
}

// DefaultSeed fixes the shared generated-selector corpus.
const DefaultSeed = 20261005

func Cases(test *testing.T) (string, string) {
	test.Helper()
	seed := int64(DefaultSeed)
	var parseError error
	if value := os.Getenv("COHERE_SELECTOR_SEED"); value != "" {
		if seed, parseError = strconv.ParseInt(value, 10, 64); parseError != nil {
			test.Fatal(parseError)
		}
	}
	generated := 4000
	if value := os.Getenv("COHERE_SELECTOR_GENERATED"); value != "" {
		if generated, parseError = strconv.Atoi(value); parseError != nil {
			test.Fatal(parseError)
		}
	}
	root, absolutePathError := filepath.Abs(filepath.Join(repository, "cohere"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	// The root also contains the Go test constants the generator extracts.
	// Validate it before running that generator, even without the npm oracle.
	files := corpusfiles.Upstream(test, root, corpusfiles.CohereCommit,
		[]string{"internal/format/css", "internal/lint/rules/tailwind"}, []string{"*.css"})
	directory := selectorAnswers(test, seed, generated, root, files, os.Getenv("ADAMIC_SELECTOR_LIBRARY"))
	casesPath := filepath.Join(directory, "cases.txt")
	answers, readError := os.ReadFile(filepath.Join(directory, "answers.txt"))
	if readError != nil {
		test.Fatal(readError)
	}
	if keep := os.Getenv("ADAMIC_SELECTOR_KEEP"); keep != "" {
		cases, readError := os.ReadFile(casesPath)
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(keep, cases, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	test.Logf("seed %d, %d generated selectors", seed, generated)
	return casesPath, string(answers)
}

func AnswerInputs(seed int64, generated int, oracleDigest, libraryDigest, node string) buildcache.Inputs {
	flags := []string{
		fmt.Sprintf("seed=%d", seed),
		fmt.Sprintf("generated=%d", generated),
		"oracle-content-sha256=" + oracleDigest,
	}
	toolchain := []string{runtime.Version()}
	if libraryDigest != "" {
		flags = append(flags, "library-content-sha256="+libraryDigest)
		toolchain = append(toolchain, node)
	}
	return buildcache.Inputs{
		Name: "selector Go oracle answers oracle-owner-v1",
		Files: []string{
			"cohere/internal/format/css",
			"cohere/internal/lint/rules/tailwind",
			"stage1/cohere/selector/testdata",
			"stage1/cohere/selector/selector_test.go",
			"stage1/cohere/selector/oracles/oracles.go",
		},
		Flags:     flags,
		Toolchain: toolchain,
	}
}

func selectorAnswers(test *testing.T, seed int64, generated int, root string, files []string, library string) string {
	test.Helper()
	oracle, _ := Parser(test)
	var libraryDigest, node string
	if library != "" {
		libraryDigest, node = contentDigest(test, library), buildcache.Tool("node", "--version")
	}
	inputs := AnswerInputs(seed, generated, contentDigest(test, oracle), libraryDigest, node)
	recipe := func(product string) error {
		directory := test.TempDir()
		corpus := filepath.Join(directory, "corpus.txt")
		testTexts := filepath.Join(directory, "test-texts.json")
		request := map[string]any{
			"seed":      seed,
			"generated": generated,
			"cases":     filepath.Join(product, "cases.txt"),
			"answers":   filepath.Join(product, "answers.txt"),
			"corpus":    corpus,
			"testTexts": testTexts,
		}
		cohereSide(test, request)
		if library != "" {
			script, absolutePathError := filepath.Abs(filepath.Join("testdata", "library.mjs"))
			if absolutePathError != nil {
				return absolutePathError
			}
			fileList := filepath.Join(directory, "css-files.json")
			encoded, encodeError := json.Marshal(files)
			if encodeError != nil {
				return encodeError
			}
			if writeError := os.WriteFile(fileList, encoded, 0644); writeError != nil {
				return writeError
			}
			command := toolchain.CommandIn("", "node", script, library, "corpus", root, corpus, testTexts, fileList)
			command.Env = nil
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr
			if runError := childguard.Run(command, childguard.Options{}); runError != nil {
				return fmt.Errorf("corpus: %v: %s", runError, stderr.Bytes())
			}
			test.Logf("%s", stdout.Bytes())
			cohereSide(test, request)
		} else {
			test.Log("CSS corpus extraction skipped: set ADAMIC_SELECTOR_LIBRARY")
		}
		return nil
	}
	if oracleProducer {
		return buildcache.Product(test, inputs, recipe)
	}
	return buildcache.BuiltElsewhere(
		test,
		inputs.Name,
		inputs,
		"stage1/cohere/selector/oracles TestProduct_SelectorGoAnswers",
		func(string) error { return fmt.Errorf("settle the answers owner first") },
	)
}

func contentDigest(test *testing.T, path string) string {
	test.Helper()
	hash := sha256.New()
	walkError := filepath.WalkDir(path, func(name string, entry os.DirEntry, pathError error) error {
		if pathError != nil || entry.IsDir() {
			return pathError
		}
		content, readError := os.ReadFile(name)
		if readError != nil {
			return readError
		}
		relative, relativePathError := filepath.Rel(path, name)
		if relativePathError != nil {
			return relativePathError
		}
		fmt.Fprintf(hash, "%s\x00%x\n", filepath.ToSlash(relative), sha256.Sum256(content))
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func cohereSide(test *testing.T, request map[string]any) {
	test.Helper()
	directory := test.TempDir()
	requestPath := filepath.Join(directory, "request.json")
	encoded, encodeError := json.Marshal(request)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(requestPath, encoded, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	// The package's test binary with the harness laid over it is a product built ahead (cohereSideOracle), run here in
	// the package's directory, where go test would run it; the child guard owns hang detection.
	binary, packageDirectory := Parser(test)
	command := toolchain.CommandIn("", binary, "-test.timeout=0", "-test.v", "-test.run=^TestAdamicPortCases$")
	command.Dir = packageDirectory
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, combinedOutputError := childguard.CombinedOutput(command, childguard.Options{})
	if combinedOutputError != nil {
		test.Fatalf("cohere's side: %v\n%s", combinedOutputError, output)
	}
	test.Logf("Go oracle: %s", output)
}
