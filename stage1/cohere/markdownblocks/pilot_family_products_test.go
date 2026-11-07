package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

func TestProduct_MarkdownTextLowered(test *testing.T) {
	test.Parallel()
	textBuildProduct(test, markdownLayoutProductRoot(test), "lowered")
}

func TestProduct_MarkdownTextNativeSanitized(test *testing.T) {
	test.Parallel()
	textBuildProduct(test, markdownLayoutProductRoot(test), "sanitized")
}

func TestProduct_MarkdownTextNativeRelease(test *testing.T) {
	test.Parallel()
	textBuildProduct(test, markdownLayoutProductRoot(test), "release")
}

func TestProduct_MarkdownMalformedEventsLowered(test *testing.T) {
	test.Parallel()
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	defer cancel()
	malformedEventsBuildProduct(
		test,
		ctx,
		filepath.Join(markdownLayoutProductRoot(test), "stage1/cohere/markdownblocks/testdata/mdast_probe.ts"),
		true,
	)
}
func TestProduct_MarkdownMalformedEventsNative(test *testing.T) {
	test.Parallel()
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	defer cancel()
	malformedEventsBuildProduct(
		test,
		ctx,
		filepath.Join(markdownLayoutProductRoot(test), "stage1/cohere/markdownblocks/testdata/mdast_probe.ts"),
		false,
	)
}

// Both product declarations and independently selected shards fetch these recipes.
func pilotGoProduct(test *testing.T, family string) string {
	test.Helper()
	return oracles.PilotGoProduct(test, family)
}

func pilotLoweredProduct(test *testing.T, family string) string {
	test.Helper()
	root := markdownLayoutProductRoot(test)
	flags := []string{}
	for _, name := range []string{
		"ADAMIC_CLOSURE_CONVENTION",
		"ADAMIC_CANONICAL_CLOSURES",
		"ADAMIC_CLOSURE_RECEIVERS",
		"ADAMIC_REGEXP_REPLACE_CALLBACK",
		"ADAMIC_NODE_HOST",
	} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	return buildcache.Product(test, buildcache.Inputs{
		Name:      "markdownblocks-pilot-" + family + "-lowered",
		Files:     whitespaceLayoutBuildFiles(),
		Flags:     flags,
		Toolchain: []string{buildcache.Tool("go", "version")},
	}, func(dir string) error {
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

func pilotNativeProduct(test *testing.T, family string, sanitize bool) string {
	test.Helper()
	return pilotNativeProductFromLowered(test, family, pilotLoweredProduct(test, family), sanitize)
}

func pilotNativeProductFromLowered(test *testing.T, family, lowered string, sanitize bool) string {
	test.Helper()
	source, err := os.ReadFile(filepath.Join(lowered, "program.c"))
	if err != nil {
		test.Fatal(err)
	}
	options := native.Options{Sanitize: sanitize}
	flags := append(native.Flags(options), fmt.Sprintf("source=%x", sha256.Sum256(source)))
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	dir := buildcache.Product(test, buildcache.Inputs{
		Name:      fmt.Sprintf("markdownblocks-pilot-%s-native-%t", family, sanitize),
		Files:     []string{"internal/native"},
		Flags:     flags,
		Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")},
	}, func(dir string) error {
		ctx, cancel := markdownLayoutSetupContext(test.Context())
		defer cancel()
		return whitespaceLayoutNativeBuild(ctx, string(source), filepath.Join(dir, "port"), options)
	})
	return filepath.Join(dir, "port")
}

type pilotProducts struct{ main, goBinary, backend, sanitized, release string }

func pilotProductsReady(test *testing.T, family string) pilotProducts {
	lowered := pilotLoweredProduct(test, family)
	return pilotProducts{
		main: filepath.Join(
			markdownLayoutProductRoot(test),
			"stage1/cohere/markdownblocks/testdata",
			family+"_probe.ts",
		),
		goBinary:  pilotGoProduct(test, family),
		backend:   filepath.Join(lowered, "program.mjs"),
		sanitized: pilotNativeProductFromLowered(test, family, lowered, true),
		release:   pilotNativeProductFromLowered(test, family, lowered, false),
	}
}

var identifierProductsOnce sync.Once
var identifierProductsShared pilotProducts

func identifierProductsReady(test *testing.T) pilotProducts {
	identifierProductsOnce.Do(func() { identifierProductsShared = pilotProductsReady(test, "identifier") })
	if identifierProductsShared.goBinary == "" {
		test.Fatal("identifier preparation failed")
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

func tokenizerEventReady(test *testing.T) tokenizerEventSetup {
	test.Helper()
	tokenizerEventSetupOnce.Do(func() {
		configureMarkdownMemory(test)
		markdownMemory.acquire(4)
		defer markdownMemory.release(4)
		root, inputs, names, files := tokenizerEventLiveEnumeration(test)
		shards := tokenizerEventShards(tokenizerEventKeys(names))
		validateTokenizerEventUnion(test, len(inputs), shards)
		products := pilotProductsReady(test, "events")
		batch := numericBatch(inputs)
		want := run{stdout: tokenizerEventAnswers(test, products.goBinary, batch)}
		answers := bytes.Split(bytes.TrimSuffix(want.stdout, []byte("\n")), []byte("\n"))
		if len(answers) != len(inputs) {
			test.Fatalf("Go oracle rows %d, want %d", len(answers), len(inputs))
		}
		cohere := filepath.Join(root, "cohere")
		fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
		if fork == "" {
			fork = filepath.Join(cohere, "internal/format/prettier/bundles")
		}
		installed, err := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
		if err != nil {
			test.Fatal(err)
		}
		pinned, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
		if err != nil {
			test.Fatal(err)
		}
		equal(test, "pinned bundle", installed, pinned)
		tokenizerEventShared = tokenizerEventSetup{
			products:  products,
			inputs:    inputs,
			names:     names,
			files:     files,
			shards:    shards,
			fullBatch: batch,
			fullWant:  want,
			answers:   answers,
			fork:      fork,
			slots:     make(chan struct{}, 4),
		}
	})
	if tokenizerEventShared.products.goBinary == "" {
		test.Fatal("tokenizer preparation failed")
	}
	return tokenizerEventShared
}

// The full-corpus Go answers are a product keyed by the oracle binary's bytes and the whole numeric batch's, so
// the first shard compares against Go's answers without rerunning Go on all 79000 cases, which cost it about 19 s
// of one core. Every shard still runs actual Go on its own cases. The full-corpus oracle uses temporary transport;
// mutant shards copy the batch.
func tokenizerEventAnswers(test *testing.T, binary string, batch []byte) []byte {
	test.Helper()
	return oracles.TokenizerEventAnswers(test, binary, batch)
}

// Child cancellation kills the process group; the caller owns the work deadline. Output is read through
// childguard.Capture for tableLayoutExecute's reason: a WaitDelay could fail a run that had finished.
func pilotExecuteResult(ctx context.Context, environment []string, name string, args ...string) (run, error) {
	command := tableLayoutCommand(ctx, name, args...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, err := childguard.Capture(ctx, command)
	if ctx.Err() != nil {
		return run{}, ctx.Err()
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return run{}, err
	}
	return run{stdout, stderr, command.ProcessState.ExitCode()}, nil
}

func TestProduct_MarkdownIdentifierLowered(test *testing.T) {
	test.Parallel()
	pilotLoweredProduct(test, "identifier")
}

func TestProduct_MarkdownIdentifierNativeSanitized(test *testing.T) {
	test.Parallel()
	pilotNativeProduct(test, "identifier", true)
}

func TestProduct_MarkdownIdentifierNativeRelease(test *testing.T) {
	test.Parallel()
	pilotNativeProduct(test, "identifier", false)
}

func TestProduct_TokenizerEventsLowered(test *testing.T) {
	test.Parallel()
	pilotLoweredProduct(test, "events")
}

func TestProduct_TokenizerEventsNativeSanitized(test *testing.T) {
	test.Parallel()
	pilotNativeProduct(test, "events", true)
}

func TestProduct_TokenizerEventsNativeRelease(test *testing.T) {
	test.Parallel()
	pilotNativeProduct(test, "events", false)
}
