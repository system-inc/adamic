package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"sort"
	"sync"

	"path/filepath"
	"testing"
)

func TestProduct_MarkdownTextLowered(t *testing.T) {
	t.Parallel()
	textBuildProduct(t, markdownLayoutProductRoot(t), "lowered")
}

func TestProduct_MarkdownTextNativeSanitized(t *testing.T) {
	t.Parallel()
	textBuildProduct(t, markdownLayoutProductRoot(t), "sanitized")
}

func TestProduct_MarkdownTextNativeRelease(t *testing.T) {
	t.Parallel()
	textBuildProduct(t, markdownLayoutProductRoot(t), "release")
}

func TestProduct_MarkdownTextGo(t *testing.T) {
	t.Parallel()
	textBuildProduct(t, markdownLayoutProductRoot(t), "go")
}

func TestProduct_MarkdownTextFormatter(t *testing.T) {
	t.Parallel()
	textBuildProduct(t, markdownLayoutProductRoot(t), "formatter")
}

func TestProduct_MarkdownTextGenerator(t *testing.T) {
	t.Parallel()
	textBuildProduct(t, markdownLayoutProductRoot(t), "generator")
}

func TestProduct_MarkdownMalformedEventsGo(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	malformedEventsGoProduct(t, ctx, markdownLayoutProductRoot(t))
}
func TestProduct_MarkdownMalformedEventsLowered(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	malformedEventsBuildProduct(t, ctx, filepath.Join(markdownLayoutProductRoot(t), "stage1/cohere/markdownblocks/testdata/mdast_probe.ts"), true)
}
func TestProduct_MarkdownMalformedEventsNative(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	malformedEventsBuildProduct(t, ctx, filepath.Join(markdownLayoutProductRoot(t), "stage1/cohere/markdownblocks/testdata/mdast_probe.ts"), false)
}

// Both product declarations and independently selected shards fetch these recipes.
func pilotGoProduct(t *testing.T, family string) string {
	t.Helper()
	root := markdownLayoutProductRoot(t)
	cohere := filepath.Join(root, "cohere")
	command, output := "adamic_identifier", "go-identifier"
	targets := map[string]string{"cmd/adamic_identifier/main.go": "identifier_go.go"}
	if family == "events" {
		command, output = "adamic_chunks", "go-chunks"
		targets = map[string]string{"cmd/adamic_chunks/main.go": "events_go.go", "internal/format/markdown/micromark/adamic_chunks.go": "events_bridge.go", "internal/format/markdown/micromark/adamic_event_transport.go": "events_transport.go"}
	}
	files := []string{"cohere"}
	replace := make(map[string]string)
	for target, source := range targets {
		files = append(files, "stage1/cohere/markdownblocks/testdata/"+source)
		replace[filepath.Join(cohere, target)] = filepath.Join(root, "stage1/cohere/markdownblocks/testdata", source)
	}
	// Sort map-derived inputs to keep the recipe key deterministic.
	sort.Strings(files)
	flags := []string{"go build", "overlay:" + command}
	for _, name := range []string{"GOFLAGS", "GOTOOLCHAIN", "GOOS", "GOARCH", "CGO_ENABLED", "GOAMD64", "CC", "CXX"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	dir := buildcache.Product(t, buildcache.Inputs{Name: "markdownblocks-pilot-" + family + "-go", Files: files, Flags: flags, Toolchain: []string{buildcache.Tool("go", "version")}}, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			return err
		}
		overlayPath := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
			return err
		}
		ctx, cancel := markdownLayoutSetupContext(t.Context())
		defer cancel()
		build := tableLayoutCommand(ctx, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(dir, output), filepath.Join(cohere, "cmd", command, "main.go"))
		build.Dir = cohere
		if result, err := build.CombinedOutput(); err != nil {
			return fmt.Errorf("%s Go product: %w\n%s", family, err, result)
		}
		return nil
	})
	return filepath.Join(dir, output)
}

func pilotLoweredProduct(t *testing.T, family string) string {
	t.Helper()
	root := markdownLayoutProductRoot(t)
	flags := []string{}
	for _, name := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	return buildcache.Product(t, buildcache.Inputs{Name: "markdownblocks-pilot-" + family + "-lowered", Files: whitespaceLayoutBuildFiles(), Flags: flags, Toolchain: []string{buildcache.Tool("go", "version")}}, func(dir string) error {
		program, err := loweredResult(filepath.Join(root, "stage1/cohere/markdownblocks/testdata", family+"_probe.ts"))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
}

func pilotNativeProduct(t *testing.T, family string, sanitize bool) string {
	t.Helper()
	lowered := pilotLoweredProduct(t, family)
	source, err := os.ReadFile(filepath.Join(lowered, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	options := native.Options{Sanitize: sanitize}
	flags := append(native.Flags(options), fmt.Sprintf("source=%x", sha256.Sum256(source)))
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	dir := buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprintf("markdownblocks-pilot-%s-native-%t", family, sanitize), Files: []string{"internal/native"}, Flags: flags, Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}}, func(dir string) error {
		ctx, cancel := markdownLayoutSetupContext(t.Context())
		defer cancel()
		return whitespaceLayoutNativeBuild(ctx, string(source), filepath.Join(dir, "port"), options)
	})
	return filepath.Join(dir, "port")
}

type pilotProducts struct{ main, goBinary, backend, sanitized, release string }

func pilotProductsReady(t *testing.T, family string) pilotProducts {
	return pilotProducts{main: filepath.Join(markdownLayoutProductRoot(t), "stage1/cohere/markdownblocks/testdata", family+"_probe.ts"), goBinary: pilotGoProduct(t, family), backend: filepath.Join(pilotLoweredProduct(t, family), "program.mjs"), sanitized: pilotNativeProduct(t, family, true), release: pilotNativeProduct(t, family, false)}
}

var identifierProductsOnce sync.Once
var identifierProductsShared pilotProducts

func identifierProductsReady(t *testing.T) pilotProducts {
	identifierProductsOnce.Do(func() { identifierProductsShared = pilotProductsReady(t, "identifier") })
	if identifierProductsShared.goBinary == "" {
		t.Fatal("identifier preparation failed")
	}
	return identifierProductsShared
}

type tokenizerEventSetup struct {
	products  pilotProducts
	inputs    [][]uint16
	names     []string
	files     int
	shards    [][]int
	fullBatch []byte
	fullWant  run
	answers   [][]byte
	fork      string
	slots     chan struct{}
}

var tokenizerEventSetupOnce sync.Once
var tokenizerEventShared tokenizerEventSetup

func tokenizerEventReady(t *testing.T) tokenizerEventSetup {
	t.Helper()
	tokenizerEventSetupOnce.Do(func() {
		configureMarkdownMemory(t)
		markdownMemory.acquire(4)
		defer markdownMemory.release(4)
		root, inputs, names, files := tokenizerEventLiveEnumeration(t)
		shards := tokenizerEventShards(tokenizerEventKeys(names))
		validateTokenizerEventUnion(t, len(inputs), shards)
		products := pilotProductsReady(t, "events")
		// The full-corpus oracle uses temporary transport; mutant shards copy the batch.
		dir, err := os.MkdirTemp("", "markdown-events-cases-")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dir)
		cases := filepath.Join(dir, "cases.txt")
		batch := numericBatch(inputs)
		write(t, cases, batch)
		ctx, cancel := markdownLayoutSetupContext(t.Context())
		defer cancel()
		want := malformedEventsExecute(t, ctx, nil, products.goBinary, cases)
		clean(t, "full Go mutant oracle", want)
		answers := bytes.Split(bytes.TrimSuffix(want.stdout, []byte("\n")), []byte("\n"))
		if len(answers) != len(inputs) {
			t.Fatalf("Go oracle rows %d, want %d", len(answers), len(inputs))
		}
		cohere := filepath.Join(root, "cohere")
		fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
		if fork == "" {
			fork = filepath.Join(cohere, "internal/format/prettier/bundles")
		}
		installed, err := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
		if err != nil {
			t.Fatal(err)
		}
		pinned, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
		if err != nil {
			t.Fatal(err)
		}
		equal(t, "pinned bundle", installed, pinned)
		tokenizerEventShared = tokenizerEventSetup{products: products, inputs: inputs, names: names, files: files, shards: shards, fullBatch: batch, fullWant: want, answers: answers, fork: fork, slots: make(chan struct{}, 4)}
	})
	if tokenizerEventShared.products.goBinary == "" {
		t.Fatal("tokenizer preparation failed")
	}
	return tokenizerEventShared
}

// Child cancellation kills the process group; the caller owns the work deadline.
func pilotExecuteResult(ctx context.Context, environment []string, name string, args ...string) (run, error) {
	command := tableLayoutCommand(ctx, name, args...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if ctx.Err() != nil {
		return run{}, ctx.Err()
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return run{}, err
	}
	return run{stdout.Bytes(), stderr.Bytes(), command.ProcessState.ExitCode()}, nil
}

func TestProduct_MarkdownIdentifierGo(t *testing.T) {
	t.Parallel()
	pilotGoProduct(t, "identifier")
}

func TestProduct_MarkdownIdentifierLowered(t *testing.T) {
	t.Parallel()
	pilotLoweredProduct(t, "identifier")
}

func TestProduct_MarkdownIdentifierNativeSanitized(t *testing.T) {
	t.Parallel()
	pilotNativeProduct(t, "identifier", true)
}

func TestProduct_MarkdownIdentifierNativeRelease(t *testing.T) {
	t.Parallel()
	pilotNativeProduct(t, "identifier", false)
}

func TestProduct_TokenizerEventsGo(t *testing.T) {
	t.Parallel()
	pilotGoProduct(t, "events")
}

func TestProduct_TokenizerEventsLowered(t *testing.T) {
	t.Parallel()
	pilotLoweredProduct(t, "events")
}

func TestProduct_TokenizerEventsNativeSanitized(t *testing.T) {
	t.Parallel()
	pilotNativeProduct(t, "events", true)
}

func TestProduct_TokenizerEventsNativeRelease(t *testing.T) {
	t.Parallel()
	pilotNativeProduct(t, "events", false)
}
