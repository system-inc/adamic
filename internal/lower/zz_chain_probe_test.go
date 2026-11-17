package lower

// A probe for #r11wppy, in the proof commit only, never for landing: what the JavaScript backend emits
// for the chain probe, whose run printed nothing where Node prints 2.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
)

func TestZZChainJavaScriptProbe(t *testing.T) {
	const source = "const lists: number[][] = [];\nlet list = lists[0];\nif (!list) lists[0] = list = [];\nlist.push(2);\nconsole.log(String(lists[0]?.[0]));\n"
	path := writeAgreementSource(t, "main.ts", source)
	program := lowerAgreementEntry(t, path)
	text := javascript.JavaScript(program)
	generated := filepath.Join(t.TempDir(), "generated.mjs")
	if err := os.WriteFile(generated, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	got := runAgreementNode(t, generated)
	fmt.Printf("PROBE stdout %q stderr %q code %d\n", got.stdout, got.stderr, got.code)
	lines := strings.Split(text, "\n")
	fmt.Printf("PROBE javascript, %d lines; the last 60:\n", len(lines))
	for _, line := range lines[max(0, len(lines)-60):] {
		fmt.Printf("PROBE| %s\n", line)
	}
	for index, statement := range program.Main {
		fmt.Printf("PROBE main[%d] %#v\n", index, statement)
	}
}
