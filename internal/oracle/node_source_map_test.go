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
	fixture, err := filepath.Abs("testdata/node_source_map/enum_tab_throw.ts")
	if err != nil {
		test.Fatal(err)
	}
	source, err := os.ReadFile(fixture)
	if err != nil {
		test.Fatal(err)
	}
	// V8 points at the Error constructor, with a tab counting as one source column.
	offset := strings.Index(string(source), "new Error(")
	if offset < 0 {
		test.Fatal("fixture has no Error constructor")
	}
	prefix := string(source[:offset])
	line := strings.Count(prefix, "\n") + 1
	column := len(prefix) - strings.LastIndex(prefix, "\n")
	node, err := toolchain.Path("node")
	if err != nil {
		test.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, node, "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", fixture)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		test.Fatal(ctx.Err())
	}
	if err != nil {
		test.Fatalf("Node: %v: %s", err, output)
	}
	want := fmt.Sprintf("at sourceMapThrow (%s:%d:%d)", fixture, line, column)
	if got := strings.TrimSpace(string(output)); got != want {
		test.Fatalf("top frame = %q, want %q", got, want)
	}
}
