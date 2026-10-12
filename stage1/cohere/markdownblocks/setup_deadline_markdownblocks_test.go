package markdownblocks

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Shared preparation has no deadline; the unit runner still bounds the process.
func markdownLayoutSetupContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}

func markdownLayoutProductRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// markdownLayoutGoOracle is cohere's command (adamic_markdown_lists or adamic_markdown_doclayout) with the testdata
// driver its overlay lays over the command's main, and for the lists testdata/list_bridge.go over adamic_lists.go: a
// product (buildcache.GoBuildIn) the whitespace, quote, structure, table, leaf and list layout families share. The
// overlay names this checkout's paths, so it sits in a temporary directory and is never part of the product (#tqrqx60).
func markdownLayoutGoOracle(t *testing.T, command string) string {
	t.Helper()
	cohere := filepath.Join(markdownLayoutProductRoot(t), "cohere")
	driverName := "document_go.go"
	if command == "adamic_markdown_lists" {
		driverName = "list_go.go"
	}
	driver, err := filepath.Abs(filepath.Join("testdata", driverName))
	if err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(cohere, "cmd", command, "main.go")
	replace := map[string]string{mainPath: driver}
	if command == "adamic_markdown_lists" {
		bridge, err := filepath.Abs("testdata/list_bridge.go")
		if err != nil {
			t.Fatal(err)
		}
		replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(t.TempDir(), "overlay.json")
	write(t, overlayPath, overlay)
	return buildcache.GoBuildIn(t, "cohere", command, mainPath, []string{"-overlay=" + overlayPath})
}

func TestMarkdownLayoutSetupHasNoDeadline(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(context.Background())
	defer cancel()
	if deadline, ok := ctx.Deadline(); ok {
		t.Fatalf("setup has deadline %v", deadline)
	}
	// The same process-group command path used by shards honors its work context.
	work, stop := context.WithTimeout(ctx, 25*time.Millisecond)
	defer stop()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := tableLayoutCommand(work, binary, "-test.run=^TestMarkdownLayoutDeadlineWorker$", "-test.timeout=0")
	command.Env = append(os.Environ(), "ADAMIC_MARKDOWN_LAYOUT_DEADLINE_WORKER=1")
	if output, err := command.CombinedOutput(); err == nil {
		t.Fatalf("work deadline did not stop child: %s", output)
	}
	if !errors.Is(work.Err(), context.DeadlineExceeded) {
		t.Fatalf("work ended with %v", work.Err())
	}
	if ctx.Err() != nil {
		t.Fatalf("work deadline canceled setup: %v", ctx.Err())
	}
}

func TestMarkdownLayoutDeadlineWorker(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_MARKDOWN_LAYOUT_DEADLINE_WORKER") == "1" {
		time.Sleep(time.Hour)
	}
}

func TestProduct_MarkdownQuote_lowered(t *testing.T) {
	t.Parallel()
	quoteLayoutBuildProduct(t, markdownLayoutProductRoot(t), "lowered")
}

func TestProduct_MarkdownQuote_lists(t *testing.T) {
	t.Parallel()
	quoteLayoutBuildProduct(t, markdownLayoutProductRoot(t), "lists")
}

func TestProduct_MarkdownQuote_layout(t *testing.T) {
	t.Parallel()
	quoteLayoutBuildProduct(t, markdownLayoutProductRoot(t), "layout")
}

func TestProduct_MarkdownQuote_native_true(t *testing.T) {
	t.Parallel()
	products := quoteLayoutBuildProduct(t, markdownLayoutProductRoot(t), "lowered")
	quoteLayoutNativeProduct(t, products, true)
}

func TestProduct_MarkdownQuote_native_false(t *testing.T) {
	t.Parallel()
	products := quoteLayoutBuildProduct(t, markdownLayoutProductRoot(t), "lowered")
	quoteLayoutNativeProduct(t, products, false)
}

func TestProduct_MarkdownStructure_lowered(t *testing.T) {
	t.Parallel()
	structureLayoutBuildProduct(t, markdownLayoutProductRoot(t), "lowered")
}

func TestProduct_MarkdownStructure_native(t *testing.T) {
	t.Parallel()
	structureLayoutBuildProduct(t, markdownLayoutProductRoot(t), "native")
}

func TestProduct_MarkdownStructure_adamic_markdown_lists(t *testing.T) {
	t.Parallel()
	structureLayoutBuildProduct(t, markdownLayoutProductRoot(t), "adamic_markdown_lists")
}

func TestProduct_MarkdownStructure_adamic_markdown_doclayout(t *testing.T) {
	t.Parallel()
	structureLayoutBuildProduct(t, markdownLayoutProductRoot(t), "adamic_markdown_doclayout")
}

func TestProduct_MarkdownTable_lowered(t *testing.T) {
	t.Parallel()
	tableLayoutBuildProduct(t, "lowered")
}

func TestProduct_MarkdownTable_native_true(t *testing.T) {
	t.Parallel()
	tableLayoutBuildProduct(t, "native-true")
}

func TestProduct_MarkdownTable_native_false(t *testing.T) {
	t.Parallel()
	tableLayoutBuildProduct(t, "native-false")
}

func TestProduct_MarkdownTable_list(t *testing.T) {
	t.Parallel()
	tableLayoutBuildProduct(t, "list")
}

func TestProduct_MarkdownTable_document(t *testing.T) {
	t.Parallel()
	tableLayoutBuildProduct(t, "document")
}

func TestProduct_MarkdownWhitespace_lowered(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	whitespaceLayoutBuildProduct(t, ctx, markdownLayoutProductRoot(t), "list_probe.ts", "lowered")
}

func TestProduct_MarkdownWhitespace_lists(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	whitespaceLayoutBuildProduct(t, ctx, markdownLayoutProductRoot(t), "list_probe.ts", "lists")
}

func TestProduct_MarkdownWhitespace_layout(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	whitespaceLayoutBuildProduct(t, ctx, markdownLayoutProductRoot(t), "list_probe.ts", "layout")
}

func TestProduct_MarkdownWhitespace_native_true(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	products := whitespaceLayoutBuildProduct(t, ctx, markdownLayoutProductRoot(t), "list_probe.ts", "lowered")
	whitespaceLayoutNativeProduct(t, ctx, products, true)
}

func TestProduct_MarkdownWhitespace_native_false(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	products := whitespaceLayoutBuildProduct(t, ctx, markdownLayoutProductRoot(t), "list_probe.ts", "lowered")
	whitespaceLayoutNativeProduct(t, ctx, products, false)
}

func TestProduct_MarkdownWhitespace_policy_lowered(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	whitespaceLayoutBuildProduct(t, ctx, markdownLayoutProductRoot(t), "whitespace_probe.ts", "lowered")
}
func TestProduct_MarkdownWhitespace_policy_native(t *testing.T) {
	t.Parallel()
	ctx, cancel := markdownLayoutSetupContext(t.Context())
	defer cancel()
	whitespaceLayoutPolicySetup(t, ctx)
}
func TestProduct_MarkdownWhitespace_manifest(t *testing.T) {
	t.Parallel()
	whitespaceLayoutReadyProducts(t)
}
