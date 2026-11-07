package oracle

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

func TestOracleNodeTransformPreservesEnumTabThrowPosition(test *testing.T) {
	test.Parallel()
	fixture, absolutePathError := filepath.Abs("testdata/node_source_map/enum_tab_throw.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, readError := os.ReadFile(fixture)
	if readError != nil {
		test.Fatal(readError)
	}
	// V8 points at the Error constructor, with a tab counting as one source column.
	offset := strings.Index(string(source), "new Error(")
	if offset < 0 {
		test.Fatal("fixture has no Error constructor")
	}
	prefix := string(source[:offset])
	line := strings.Count(prefix, "\n") + 1
	column := len(prefix) - strings.LastIndex(prefix, "\n")
	node, pathError := toolchain.Path("node")
	if pathError != nil {
		test.Fatal(pathError)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, node, "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", fixture)
	output, combinedOutputError := command.CombinedOutput()
	if ctx.Err() != nil {
		test.Fatal(ctx.Err())
	}
	if combinedOutputError != nil {
		test.Fatalf("Node: %v: %s", combinedOutputError, output)
	}
	want := fmt.Sprintf("at sourceMapThrow (%s:%d:%d)", fixture, line, column)
	if got := strings.TrimSpace(string(output)); got != want {
		test.Fatalf("top frame = %q, want %q", got, want)
	}
}
