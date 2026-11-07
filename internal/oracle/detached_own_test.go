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

func TestDetachedOwnInheritedMutant(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"records", "objects"} {
		test.Run(name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/detached_own_"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			p, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			code := javascript.JavaScript(p)
			changed := strings.ReplaceAll(code, "Object.hasOwn(", "detachedOwnMutant(")
			if changed == code {
				test.Fatal("mutant changed no own-property call")
			}
			changed = "const detachedOwnMutant = (target,key) => key in target;\n" + changed
			file := filepath.Join(test.TempDir(), "mutant.mjs")
			if writeError := os.WriteFile(file, []byte(changed), 0644); writeError != nil {
				test.Fatal(writeError)
			}
			got, want := onNode(test, file), onNode(test, path)
			if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(want, got) != "stdout differs" {
				test.Fatalf("inherited-key mutant survived: Node %+v; mutant %+v", want, got)
			}
			test.Log("inherited name reported as own caught by Node stdout")
		})
	}
}

func TestDetachedOwnReadinessMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/detached_own_unready.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	p, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
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
		test.Fatal("mutant changed no readiness check")
	}
	got, want := onJavaScriptBackend(test, p), onNode(test, path)
	if got.exitCode != 0 ||
		want.exitCode != 70 ||
		!strings.Contains(string(want.stderr), "ReferenceError") ||
		disagreement(want, got) == "" {
		test.Fatalf("initialization mutant survived: Node %+v; mutant %+v", want, got)
	}
	test.Log("early intrinsic call caught by Node's initialization stop")
}
