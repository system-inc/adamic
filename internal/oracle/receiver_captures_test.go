package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/receiver_captures.a", true, false})
}

// A valid IR mutant loses the lexical receiver while keeping the closure's calling convention.
// Both backends must disagree with the source, independently of compiler warnings.
func TestReceiverCaptureNullMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/receiver_captures.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 {
		t.Fatalf("source failed: %#v", source)
	}
	changed := 0
	for index, function := range program.Functions {
		if !function.Closure {
			continue
		}
		lost := len(program.Locals)
		program.Locals = append(program.Locals, ir.Local{Name: "lost_receiver", Type: ir.Object, Function: index})
		before := changed
		program.Functions[index].Body = mutateReadiness(function.Body, func(value any) any {
			if read, ok := value.(ir.Read); ok && program.Locals[read.Local].Name == "this" {
				changed++
				return ir.Read{Local: lost, Of: ir.Object}
			}
			return value
		})
		if changed > before {
			program.Functions[index].Body = append([]ir.Statement{ir.Declare{Local: lost, Value: ir.Undefined{Of: ir.Object}}}, program.Functions[index].Body...)
		}
	}
	if changed == 0 {
		t.Fatal("mutant changed no receiver reads")
	}
	native, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if disagreement(source, got) == "" {
			t.Fatalf("%s lost its receiver without detection", name)
		}
		t.Logf("%s null receiver caught: exit=%d stderr=%q", name, got.exitCode, got.stderr)
	}
}

// Removing only the receiver retain leaves the escaping closure with a stale receiver.
// The dynamically allocated instance must be freed before its first callback invocation.
func TestReceiverCaptureRetainMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/receiver_captures.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	receiver := regexp.MustCompile(`adamic_cell_new\(\(adamic_value\)\{\.reference = adamic_retain\((adamic_local_[0-9]+_this)\)\}, true\)`)
	if len(receiver.FindAllString(code, -1)) != 2 {
		t.Fatal("expected the make and nested receiver cells")
	}
	code = receiver.ReplaceAllString(code, `adamic_cell_new((adamic_value){.reference = $1}, true)`)
	binary := filepath.Join(t.TempDir(), "lost-retain")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if got.exitCode == 0 || !strings.Contains(string(got.stderr), "heap-use-after-free") {
		t.Fatalf("want ASan to catch the stale receiver, got %#v", got)
	}
	if disagreement(onNode(t, path), got) == "" {
		t.Fatal("stale receiver mutant survived the oracle")
	}
	t.Log("receiver retain removal caught by ASan heap-use-after-free")
}
