package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"boolean_optional_selector.a", "boolean_optional_values.a", "boolean_optional_unknown.a", "boolean_optional_static.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + path, true, false})
	}
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{
		"internal/oracle/testdata/boolean_optional_fields.a", true, false,
	})
}

// The stray byte is identical in both builds; only the production domain guard changes.
func TestBooleanOptionalDomainMutant(t *testing.T) {
	t.Parallel()
	contents, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/maybe.c"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(contents)
	start := strings.Index(text, "adamic_maybe_boolean adamic_maybe_boolean_unpack(")
	if start < 0 {
		t.Fatal("domain helper missing")
	}
	function := strings.ReplaceAll(text[start:], "adamic_maybe_boolean_unpack", "test_unpack")
	for _, mutate := range []bool{false, true} {
		code := function
		if mutate {
			code = strings.Replace(code, "if (packed > 2)", "if (false)", 1)
			if code == function {
				t.Fatal("domain guard missing")
			}
		}
		code = "#include \"adamic.h\"\n" + code + `
int main(int argc, char **argv) {
    adamic_start(argc, argv);
    volatile uint8_t stray = 255;
    adamic_maybe_boolean value = test_unpack(stray);
    return value.present ? 0 : 1;
}
`
		binary := filepath.Join(t.TempDir(), "domain")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := execute(t, binary)
		if mutate {
			if result.exitCode != 0 {
				t.Fatalf("mutant must finish: %d %s", result.exitCode, result.stderr)
			}
		} else if result.exitCode != 70 || !strings.Contains(string(result.stderr), "invalid boolean field domain") {
			t.Fatalf("stray byte escaped: %d %s", result.exitCode, result.stderr)
		}
	}
	t.Log("domain mutant admits the same stray byte that production stops")
}

func TestBooleanOptionalPresenceMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/boolean_optional_fields.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_object_optional_field(", "adamic_object_field(")
	if mutant == source {
		t.Fatal("optional lookup missing")
	}
	binary := filepath.Join(t.TempDir(), "presence")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if result.exitCode != 0 {
		t.Fatalf("presence mutant stopped: %d %s", result.exitCode, result.stderr)
	}
	if disagreement(onNode(t, path), result) == "" {
		t.Fatal("Node did not catch ignored presence")
	}
	t.Logf("Node catches ignored boolean presence: %q", result.stdout)
}

func TestBooleanOptionalViewPending(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/boolean_optional_view_pending.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "true\n" {
		t.Fatalf("Node: %d %q %s", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "checked view field") {
		t.Fatalf("pending boundary changed: %v", err)
	}
	t.Skip("acceptance dependency: compiler/views-rehearsal 2b6c032a")
}

// An invalid physical tag exercises only the JSON representation guard.
func TestBooleanJSONMetadataMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/boolean_optional_fields.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	generated := native.C(program)
	tags := regexp.MustCompile(`(static const int adamic_shape_[0-9]+_types\[\] = \{)9(\};)`)
	corrupted := tags.ReplaceAllString(generated, `${1}6${2}`)
	if corrupted == generated {
		t.Fatal("boolean metadata missing")
	}
	runtimeCode, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/json_stringify.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []bool{false, true} {
		implementation := string(runtimeCode)
		if mutate {
			implementation = strings.Replace(implementation, "else if (type != 2)", "else if (false)", 1)
			if implementation == string(runtimeCode) {
				t.Fatal("metadata guard missing")
			}
		}
		code := "#define adamic_json_stringify probe_json_stringify\n" + implementation + "\n" + corrupted
		binary := filepath.Join(t.TempDir(), "metadata")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := execute(t, binary)
		if mutate {
			if result.exitCode != 0 {
				t.Fatalf("unguarded metadata must finish: %d %s", result.exitCode, result.stderr)
			}
		} else if result.exitCode != 70 || !strings.Contains(string(result.stderr), "JSON boolean object lacks proven scalar metadata") {
			t.Fatalf("invalid metadata escaped: %d %s", result.exitCode, result.stderr)
		}
	}
	t.Log("JSON representation guard stops invalid metadata; its mutant admits the same invalid tag")
}
