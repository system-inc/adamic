package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// A field read through a structural type can reach a class's getter at run time (#1w5bve2's
// getter_order.ts), and flow treats an ir.Property as running no code (internal/flow/infer.go). That
// holds because finishAccessors (class_accessors.go) rewrites every read of a name any class declares
// as an accessor into a call of that name's dispatcher, before flow, reuse, regions and emit see the
// IR (#gm8tx0n, Memory's read, Oct 30). So a getter's effects are a call's effects there. Each fixture
// reads `extra` through `Reader` from a class whose getter does what a borrow, a reuse or an escape
// decision must not miss: it replaces an element the caller holds from a loop over the array, drops the
// last reference to an object the caller read from a field, or stores its receiver in a module array.
// Each agrees with Node on native under ASan, its leak check and UBSan, and on the JavaScript backend,
// in .a and .ts, and no ir.Property reading `extra` survives lowering outside its dispatcher. Mutant:
// finishAccessors leaves reads alone, which reddens the IR check in all three.
func TestGetterReachedByAStructuralReadIsACall(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct{ name, stdout string }{
		{"array_borrow", "first7-1\nsecond7-2\n"},
		{"field_borrow", "start3-1\nnext1-2\n"},
		{"escape", "-made1\n-made2\nmade1\nmade2\n"},
	} {
		source, err := os.ReadFile("testdata/getter_structural_read/" + fixture.name + ".ts")
		if err != nil {
			t.Fatal(err)
		}
		for _, extension := range []string{".a", ".ts"} {
			t.Run(fixture.name+extension, func(t *testing.T) {
				t.Parallel()
				var program *ir.Program
				var err error
				if extension == ".a" {
					program, err = lowerSource(t, string(source))
				} else {
					program, err = lowerTypeScriptAssertionSource(t, string(source))
				}
				if err != nil {
					t.Fatal(err)
				}
				survivors := 0
				inspect := func(node any) bool {
					if read, ok := node.(ir.Property); ok && read.Name == "extra" {
						survivors++
					}
					return true
				}
				// The dispatcher, property_get:extra, keeps one: the slot read it falls back to when the
				// object holds no accessor (class_accessors.go), which runs no getter.
				walk(program.Main, inspect)
				for _, function := range program.Functions {
					if function.Name != "property_get:extra" {
						walk(function.Body, inspect)
					}
				}
				if survivors != 0 {
					t.Fatalf("%d ir.Property reads of the accessor extra survive lowering outside its dispatcher; each should be the dispatcher's call", survivors)
				}
				directory := t.TempDir()
				binary := filepath.Join(directory, "native")
				if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				runner, err := filepath.Abs("../../oracle/node.mjs")
				if err != nil {
					t.Fatal(err)
				}
				generated := filepath.Join(directory, "generated.mjs")
				if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
					t.Fatal(err)
				}
				for _, command := range []*exec.Cmd{exec.Command(binary), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated)} {
					command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
					var out, errs bytes.Buffer
					command.Stdout, command.Stderr = &out, &errs
					if err := command.Run(); err != nil {
						var exit *exec.ExitError
						if !errors.As(err, &exit) {
							t.Fatal(err)
						}
					}
					if out.String() != fixture.stdout || errs.Len() != 0 {
						t.Fatalf("%s: stdout %q stderr %q; want %q, as Node prints", command.Path, out.String(), errs.String(), fixture.stdout)
					}
				}
			})
		}
	}
}
