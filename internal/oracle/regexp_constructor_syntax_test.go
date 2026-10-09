package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

const regexpConstructorSyntaxFixture = "internal/oracle/testdata/regexp_constructor_syntax.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{regexpConstructorSyntaxFixture, true, false})
}

func TestRegExpConstructorSyntaxWASI(t *testing.T) {
	if os.Getenv("WASI_SYSROOT") == "" {
		t.Skip("WASI_SYSROOT is required")
	}
	path, _ := filepath.Abs(filepath.Join(repository, regexpConstructorSyntaxFixture))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if diff := disagreement(expected, onWASI(t, native.C(program))); diff != "" {
		t.Fatal(diff)
	}
}

// Mutations are in each generated translation unit; the reference, JavaScript
// backend and files used by concurrent directory surveys stay unchanged.
func TestRegExpConstructorSyntaxNodeMutants(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, regexpConstructorSyntaxFixture))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	parser, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_compile_parser.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, old, new string }{
		{"accept-duplicate-modifier", "flag == 0 || (seen & flag) != 0", "flag == 0"},
		{"accept-invalid-group", `} else return regex_parse_fail(p, p->position, "invalid group", "Invalid group");`, `} else { p->position += 2; node->kind = 1; }`},
		{"accept-incomplete-unicode-escape", `return regex_parse_fail(p, start, "invalid hex escape", "Invalid Unicode escape");`, `return regex_parse_character(p, 'u', 1, start);`},
		{"accept-reversed-range", "if (character->value > right->value) return regex_parse_fail", "if (false && character->value > right->value) return regex_parse_fail"},
		{"accept-duplicate-flags", "byte >= 128 || seen[byte]", "byte >= 128"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if strings.Count(string(parser), mutant.old) != 1 {
				t.Fatal("mutation site moved")
			}
			changed := strings.Replace(string(parser), mutant.old, mutant.new, 1)
			source := strings.Replace(native.C(program), `#include "regexp_compile_parser.c"`, changed, 1)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 || !bytes.Contains(actual.stdout, []byte("accepted\n")) {
				t.Fatalf("mutant failed outside the Node comparison: %+v", actual)
			}
			if diff := disagreement(expected, actual); diff == "" {
				t.Fatal("Node did not catch mutant")
			} else {
				t.Logf("Node caught %s: %s", mutant.name, diff)
			}
		})
	}
}

func TestRegExpConstructorSyntaxCounts(t *testing.T) {
	row := counted(t, regexpConstructorSyntaxFixture, false, nil, false, false)
	t.Log(row)
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, row+"\n") {
		return
	}
	if !*updateCounts {
		t.Fatalf("constructor syntax counts missing or changed: %s", row)
	}
	if strings.Contains(text, "| "+regexpConstructorSyntaxFixture+" |") {
		t.Fatal("existing row changed; review it explicitly")
	}
	at := strings.Index(text, "| internal/oracle/testdata/regexp_cycle_fields.a |")
	if at < 0 {
		t.Fatal("regexp fixture boundary missing")
	}
	if err := os.WriteFile(countsPath, []byte(text[:at]+row+"\n"+text[at:]), 0644); err != nil {
		t.Fatal(err)
	}
}
