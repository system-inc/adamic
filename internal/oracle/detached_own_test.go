package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

func init() {
	for _, name := range []string{"parser", "records", "objects", "unready"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/detached_own_" + name + ".a", true, false})
	}
}

func TestDetachedOwnInheritedMutant(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"records", "objects"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/detached_own_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := javascript.JavaScript(p)
			changed := strings.ReplaceAll(code, "Object.hasOwn(", "detachedOwnMutant(")
			if changed == code {
				t.Fatal("mutant changed no own-property call")
			}
			changed = "const detachedOwnMutant = (target,key) => key in target;\n" + changed
			file := filepath.Join(t.TempDir(), "mutant.mjs")
			if err := os.WriteFile(file, []byte(changed), 0644); err != nil {
				t.Fatal(err)
			}
			got, want := onNode(t, file), onNode(t, path)
			if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(want, got) != "stdout differs" {
				t.Fatalf("inherited-key mutant survived: Node %+v; mutant %+v", want, got)
			}
			t.Log("inherited name reported as own caught by Node stdout")
		})
	}
}

func TestDetachedOwnReadinessMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/detached_own_unready.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for i := range p.Functions {
		for j, statement := range p.Functions[i].Body {
			returned, ok := statement.(ir.Return)
			if !ok {
				continue
			}
			call, ok := returned.Value.(ir.Call)
			if !ok || len(call.Arguments) == 0 {
				continue
			}
			read, ok := call.Arguments[0].(ir.Read)
			if !ok || !read.Checked {
				continue
			}
			read.Checked = false
			call.Arguments[0] = read
			returned.Value = call
			p.Functions[i].Body[j] = returned
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutant changed no readiness check")
	}
	got, want := onJavaScriptBackend(t, p), onNode(t, path)
	if got.exitCode != 0 || want.exitCode != 70 || disagreement(want, got) == "" {
		t.Fatalf("initialization mutant survived: Node %+v; mutant %+v", want, got)
	}
	t.Log("early intrinsic call caught by Node's initialization stop")
}
