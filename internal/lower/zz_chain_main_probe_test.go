package lower

// A probe for #r11wppy, in a proof commit on main only, never for landing: whether main's JavaScript backend
// disagrees with Node on a chained assignment whose outer target is an element, with a value main accepts.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
)

func TestZZChainMainProbe(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"held array", "const lists: number[][] = [];\nconst fresh: number[] = [];\nlet list = lists[0];\nif (!list) lists[0] = list = fresh;\nlist.push(2);\nconsole.log(String(lists[0]?.[0]));\n"},
		{"held array, no if", "const lists: number[][] = [];\nconst fresh: number[] = [];\nlet list: number[] | undefined;\nlists[0] = list = fresh;\nlist.push(2);\nconsole.log(String(lists[0]?.[0]));\n"},
		{"field target", "const holder: { list?: number[] } = {};\nconst fresh: number[] = [];\nlet list: number[] | undefined;\nholder.list = list = fresh;\nlist.push(2);\nconsole.log(String(holder.list?.[0]));\n"},
		{"plain value", "const lists: number[] = [];\nlet value = 0;\nlists[0] = value = 7;\nconsole.log(String(lists[0]) + ' ' + String(value));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			path := writeAgreementSource(t, "main"+extension, probe.source)
			program, err := lowerSourceAs(t, probe.source, extension)
			if err != nil {
				fmt.Printf("PROBE %s%s: stops: %v\n", probe.name, extension, err)
				continue
			}
			text := javascript.JavaScript(program)
			generated := filepath.Join(t.TempDir(), "generated.mjs")
			if err := os.WriteFile(generated, []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
			want := runAgreementNode(t, path)
			got := runAgreementNode(t, generated)
			verdict := "AGREES"
			if string(got.stdout) != string(want.stdout) || got.code != want.code {
				verdict = "DISAGREES"
			}
			fmt.Printf("PROBE %s%s: %s; Node %q exit %d; JavaScript backend %q exit %d stderr %q\n", probe.name, extension, verdict, want.stdout, want.code, got.stdout, got.code, got.stderr)
			if verdict == "DISAGREES" {
				lines := strings.Split(text, "\n")
				for _, line := range lines[max(0, len(lines)-40):] {
					fmt.Printf("PROBE| %s\n", line)
				}
			}
		}
	}
}
