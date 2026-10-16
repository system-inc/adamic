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
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
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

func pilotLoweredProduct(t *testing.T, family string) string {
	t.Helper()
	root := markdownLayoutProductRoot(t)
	flags := []string{}
	for _, name := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	return buildcache.Product(t, buildcache.Inputs{Name: "markdownblocks-pilot-" + family + "-lowered", Files: oracles.LoweringFiles(), Flags: flags, Toolchain: []string{buildcache.Tool("go", "version")}}, func(dir string) error {
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
	return pilotNativeProductFromLowered(t, family, pilotLoweredProduct(t, family), sanitize)
}

func pilotNativeProductFromLowered(t *testing.T, family, lowered string, sanitize bool) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(lowered, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	options := native.Options{Sanitize: sanitize}
	flags := append(native.Flags(options), fmt.Sprintf("source=%x", sha256.Sum256(source)))
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT"} {
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
	lowered := pilotLoweredProduct(t, family)
	return pilotProducts{main: filepath.Join(markdownLayoutProductRoot(t), "stage1/cohere/markdownblocks/testdata", family+"_probe.ts"), goBinary: oracles.PilotGoProduct(t, family), backend: filepath.Join(lowered, "program.mjs"), sanitized: pilotNativeProductFromLowered(t, family, lowered, true), release: pilotNativeProductFromLowered(t, family, lowered, false)}
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
		root, inputs, names, files := oracles.TokenizerEventLiveEnumeration(t)
		shards := tokenizerEventShards(tokenizerEventKeys(names))
		validateTokenizerEventUnion(t, len(inputs), shards)
		products := pilotProductsReady(t, "events")
		batch := oracles.NumericBatch(inputs)
		want := run{stdout: oracles.TokenizerEventAnswers(t, products.goBinary, batch)}
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
