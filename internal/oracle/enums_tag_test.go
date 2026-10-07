package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"enums_tag_narrowing.a", "enums_tag_never.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/" + name, true, name == "enums_tag_never.a",
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
