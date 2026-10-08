package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"enums_tag_narrowing.a", "enums_tag_never.a", "enums_tag_singleton_checked.a", "enums_tag_singleton_cast.a", "enums_tag_singleton_class.a", "enums_tag_property_checked.a", "enums_tag_literal_checked.a", "enums_tag_boolean_checked.a", "enums_tag_remainder.a", "enums_tag_object_never.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/" + name, true, name != "enums_tag_narrowing.a" && name != "enums_tag_remainder.a",
		})
	}
}

func TestEnumTagNeverPinned(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums_tag_never.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("before\ndefault\n"), stderr: []byte("adamic: panic: unreachable value 42 for numeric enum SyntaxKind\n"), exitCode: 70}
	native, _ := natively(t, program)
	for backend, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Errorf("%s: %s; %+v", backend, difference, result)
		}
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "before\ndefault\nunreachable\nafter\n" {
		t.Fatalf("independent Node: %+v", node)
	}
}

// Treating the open remainder as a proof of never erases a reachable default.
// The mutant must build and finish cleanly; only the pinned stop catches it.
func TestEnumTagOpenRemainderMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums_tag_never.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for index := range program.Functions {
		if program.Functions[index].Name != "show" {
			continue
		}
		for position, statement := range program.Functions[index].Body {
			if switched, ok := statement.(ir.Switch); ok && len(switched.Default) > 0 {
				switched.Default = nil
				program.Functions[index].Body[position] = switched
				changed = true
			}
		}
	}
	if !changed {
		t.Fatal("mutant erased no default")
	}
	native, _ := natively(t, program)
	for backend, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != "before\nafter\n" {
			t.Fatalf("%s mutant must finish cleanly: %+v", backend, result)
		}
		want := run{stdout: []byte("before\ndefault\n"), stderr: []byte("adamic: panic: unreachable value 42 for numeric enum SyntaxKind\n"), exitCode: 70}
		if disagreement(want, result) == "" {
			t.Fatalf("%s open-remainder mutant survived", backend)
		}
	}
	t.Log("reachable default erased; pinned stop caught clean exit 0 in both backends")
}

func TestEnumTagViewsPinned(t *testing.T) {
	for _, probe := range []struct{ name, stdout, message, node string }{
		{"property_checked", "before\n", "numeric enum object view failed: field value", "before\ntext1\nafter\n"},
		{"singleton_class", "8\nbefore\n", "numeric enum object view failed: class", "8\nbefore\ntext1\nafter\n"},
		{"literal_checked", "before\n", "numeric enum object view failed: field value", "before\nbad\nafter\n"},
		{"singleton_cast", "before\n", "numeric enum member view failed: field kind", "before\n42\nafter\n"},
		{"singleton_checked", "before\n", "numeric enum object view failed: field value", "before\ntext1\nafter\n"},
		{"boolean_checked", "8\nbefore\n", "numeric enum object view failed: field value", "8\nbefore\n2\nafter\n"},
		{"object_never", "before\ndefault\n", "unreachable value 42 for numeric enum AKind", "before\ndefault\nunreachable\nafter\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums_tag_"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(probe.stdout), stderr: []byte("adamic: panic: " + probe.message + "\n"), exitCode: 70}
			native, _ := natively(t, program)
			for name, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, result); difference != "" {
					t.Errorf("%s: %s; %+v", name, difference, result)
				}
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != probe.node {
				t.Fatalf("independent Node: %+v", node)
			}
		})
	}
}

// Erasing the checked field reads must produce a clean, wrong result, rather
// than an invalid native pointer read. The string payload is viewed as a number.
func TestEnumTagPayloadMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums_tag_singleton_checked.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for i := range program.Functions {
		if program.Functions[i].Name != "enum_tag_view" {
			continue
		}
		body := []ir.Statement{}
		for _, statement := range program.Functions[i].Body {
			if _, ok := statement.(ir.Evaluate); ok {
				changed = true
				continue
			}
			body = append(body, statement)
		}
		program.Functions[i].Body = body
	}
	if !changed {
		t.Fatal("open_tag mutant erased no payload check")
	}
	native, _ := natively(t, program)
	for name, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("%s mutant must finish cleanly: %+v", name, result)
		}
		want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: numeric enum object view failed: field value\n"), exitCode: 70}
		if disagreement(want, result) == "" {
			t.Fatalf("%s open_tag mutant survived", name)
		}
	}
	t.Log("open_tag payload-check erasure caught by pinned stop in both backends")
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/enums_tag_split_union.a", true, false})
}
