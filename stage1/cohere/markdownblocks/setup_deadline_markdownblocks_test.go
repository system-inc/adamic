package markdownblocks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
)

// Shared preparation has no deadline; the unit runner still bounds the process.
func markdownLayoutSetupContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}

func markdownLayoutProductRoot(test *testing.T) string {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return root
}

func TestMarkdownLayoutSetupHasNoDeadline(test *testing.T) {
	test.Parallel()
	ctx, cancel := markdownLayoutSetupContext(context.Background())
	defer cancel()
	if deadline, ok := ctx.Deadline(); ok {
		test.Fatalf("setup has deadline %v", deadline)
	}
	// The same process-group command path used by shards honors its work context.
	work, stop := context.WithTimeout(ctx, 25*time.Millisecond)
	defer stop()
	binary, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	command := tableLayoutCommand(work, binary, "-test.run=^TestMarkdownLayoutDeadlineWorker$", "-test.timeout=0")
	command.Env = append(os.Environ(), "ADAMIC_MARKDOWN_LAYOUT_DEADLINE_WORKER=1")
	if stdout, stderr, captureError := childguard.Capture(work, command); captureError == nil {
		test.Fatalf("work deadline did not stop child: %s%s", stdout, stderr)
	}
	if !errors.Is(work.Err(), context.DeadlineExceeded) {
		test.Fatalf("work ended with %v", work.Err())
	}
	if ctx.Err() != nil {
		test.Fatalf("work deadline canceled setup: %v", ctx.Err())
	}
}

func TestMarkdownLayoutDeadlineWorker(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_MARKDOWN_LAYOUT_DEADLINE_WORKER") == "1" {
		time.Sleep(time.Hour)
	}
}

func TestProduct_MarkdownQuote_lowered(test *testing.T) {
	test.Parallel()
	quoteLayoutBuildProduct(test, markdownLayoutProductRoot(test), "lowered")
}

func TestProduct_MarkdownQuote_native_true(test *testing.T) {
	test.Parallel()
	products := quoteLayoutBuildProduct(test, markdownLayoutProductRoot(test), "lowered")
	quoteLayoutNativeProduct(test, products, true)
}

func TestProduct_MarkdownQuote_native_false(test *testing.T) {
	test.Parallel()
	products := quoteLayoutBuildProduct(test, markdownLayoutProductRoot(test), "lowered")
	quoteLayoutNativeProduct(test, products, false)
}

func TestProduct_MarkdownStructure_lowered(test *testing.T) {
	test.Parallel()
	structureLayoutBuildProduct(test, markdownLayoutProductRoot(test), "lowered")
}

func TestProduct_MarkdownStructure_native(test *testing.T) {
	test.Parallel()
	structureLayoutBuildProduct(test, markdownLayoutProductRoot(test), "native")
}

func TestProduct_MarkdownTable_lowered(test *testing.T) {
	test.Parallel()
	tableLayoutBuildProduct(test, "lowered")
}

func TestProduct_MarkdownTable_native_true(test *testing.T) {
	test.Parallel()
	tableLayoutBuildProduct(test, "native-true")
}

func TestProduct_MarkdownTable_native_false(test *testing.T) {
	test.Parallel()
	tableLayoutBuildProduct(test, "native-false")
}

func TestProduct_MarkdownWhitespace_lowered(test *testing.T) {
	test.Parallel()
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	defer cancel()
	whitespaceLayoutBuildProduct(test, ctx, markdownLayoutProductRoot(test), "list_probe.ts", "lowered")
}

func TestProduct_MarkdownWhitespace_native_true(test *testing.T) {
	test.Parallel()
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	defer cancel()
	products := whitespaceLayoutBuildProduct(test, ctx, markdownLayoutProductRoot(test), "list_probe.ts", "lowered")
	whitespaceLayoutNativeProduct(test, ctx, products, true)
}

func TestProduct_MarkdownWhitespace_native_false(test *testing.T) {
	test.Parallel()
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	defer cancel()
	products := whitespaceLayoutBuildProduct(test, ctx, markdownLayoutProductRoot(test), "list_probe.ts", "lowered")
	whitespaceLayoutNativeProduct(test, ctx, products, false)
}

func TestProduct_MarkdownWhitespace_policy_lowered(test *testing.T) {
	test.Parallel()
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	defer cancel()
	whitespaceLayoutBuildProduct(test, ctx, markdownLayoutProductRoot(test), "whitespace_probe.ts", "lowered")
}
func TestProduct_MarkdownWhitespace_policy_native(test *testing.T) {
	test.Parallel()
	ctx, cancel := markdownLayoutSetupContext(test.Context())
	defer cancel()
	whitespaceLayoutPolicySetup(test, ctx)
}
func TestProduct_MarkdownWhitespace_manifest(test *testing.T) {
	test.Parallel()
	whitespaceLayoutReadyProducts(test)
}
