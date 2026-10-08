package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSymbolDeclarationContracts(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "file.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := "interface M {}\ninterface M { x: number }\nlet m: M;\nclass C {}\nC;\n({C});\nfunction f(){let C=1; C;}\nmissing;\nnew String();\nclass Computed { [Symbol.iterator](arg: number){ arg; } }\n"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022"}}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	// The parser's exact span includes leading trivia. Find it using the actual
	// compiler tree rather than guessing that a token starts its node's span.
	ask := func(marker, question string, fileWide bool) []string {
		t.Helper()
		position := strings.Index(source, marker)
		if position < 0 {
			t.Fatal(marker)
		}
		var start, end uint64
		found := false
		// Inspect's exact index gives us the nodes, but expected facts are checked
		// below against language-level relationships and declaration source text.
		for first := position; first >= 0 && first >= position-2; first-- {
			for last := position + 1; last <= position+len(marker); last++ {
				_, node, e := p.exact(file, uint64(first), uint64(last), "Identifier")
				if e == nil && node.Text() == strings.FieldsFunc(marker, func(r rune) bool { return !(r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') })[0] {
					start, end, found = uint64(first), uint64(last), true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Fatalf("no identifier for %q", marker)
		}
		kind := "Identifier"
		if fileWide {
			question = fmt.Sprintf("%s\n%d\n%d\nIdentifier", question, start, end)
			start, end, kind = 0, uint64(len(source)), "SourceFile"
		}
		wire, e := p.Inspect(file, start, end, kind, question)
		if e != nil {
			t.Fatal(e)
		}
		return decodedFields(t, wire)
	}
	merged := ask("M;", "node-symbol-details", true)
	if merged[2] != "1" || merged[5] != "M" || merged[6] != "2" {
		t.Fatalf("merged declarations: %q", merged)
	}
	// Complete declarations must contain both distinct merged source records,
	// not a count accompanied by duplicated or truncated payloads.
	cursor := 7
	var texts []string
	for declaration := 0; declaration < 2; declaration++ {
		if merged[cursor] != filepath.ToSlash(file) || merged[cursor+1] != "InterfaceDeclaration" {
			t.Fatalf("wrong merged declaration origin: %q", merged[cursor:])
		}
		start, e1 := strconv.Atoi(merged[cursor+2])
		end, e2 := strconv.Atoi(merged[cursor+3])
		if e1 != nil || e2 != nil || start < 0 || end > len(source) || end < start {
			t.Fatal("invalid merged declaration span")
		}
		texts = append(texts, strings.TrimSpace(source[start:end]))
		cursor += 10
		tags, e := strconv.Atoi(merged[cursor])
		if e != nil {
			t.Fatal(e)
		}
		cursor += 1 + tags*5
		parameters, e := strconv.Atoi(merged[cursor])
		if e != nil {
			t.Fatal(e)
		}
		cursor += 1 + parameters
	}
	if cursor != len(merged) || texts[0] != "interface M {}" || texts[1] != "interface M { x: number }" {
		t.Fatalf("incomplete or reordered merged declarations: %q", texts)
	}
	direct := ask("C;", "node-symbol-details", false)
	replay := ask("C;", "node-symbol-details", true)
	if fmt.Sprint(direct) != fmt.Sprint(replay) {
		t.Fatalf("selector changed answer: %q %q", direct, replay)
	}
	provenance := ask("C;", "symbol-provenance", true)
	if provenance[2] != direct[3] || provenance[4] != "1" || provenance[5] != filepath.ToSlash(file) || provenance[9] != "2" || provenance[10] != "ClassDeclaration" || provenance[16] != "SourceFile" {
		t.Fatalf("provenance must share identity and include complete ancestry: %q", provenance)
	}
	if fmt.Sprint(provenance) != fmt.Sprint(ask("C;", "symbol-provenance", false)) {
		t.Fatal("provenance selector changed answer")
	}
	shadow := ask("C;}", "node-symbol-details", true)
	if direct[3] == shadow[3] {
		t.Fatal("shadow shares symbol identity")
	}
	absent := ask("missing;", "node-symbol-details", true)
	if len(absent) != 3 || absent[2] != "0" {
		t.Fatalf("missing symbol: %q", absent)
	}
	computed := ask("arg;", "node-symbol-details", true)
	if computed[13] != "MethodDeclaration" || computed[14] != "" {
		t.Fatalf("computed parent name must use its span, not Node.Text: %q", computed)
	}
	global := ask("String();", "node-symbol-details", true)
	if global[2] != "1" || global[11] != "1" || global[12] != "1" {
		t.Fatalf("library declaration flags: %q", global)
	}
	binding := ask("C});", "binding-declarations", true)
	if binding[2] != "1" || binding[4] != "1" || binding[5] != filepath.ToSlash(file) || binding[6] != "ClassDeclaration" {
		t.Fatalf("shorthand value binding: %q", binding)
	}
	begin, e := strconv.Atoi(binding[7])
	if e != nil {
		t.Fatal(e)
	}
	finish, e := strconv.Atoi(binding[8])
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(source[begin:finish], "class C") {
		t.Fatalf("wrong binding declaration: %q", source[begin:finish])
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "node-symbol-details\n01\n2\nIdentifier"); err == nil {
		t.Fatal("noncanonical selector accepted")
	}
}
