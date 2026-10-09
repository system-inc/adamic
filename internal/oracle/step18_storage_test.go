package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

const storageDirectory = "docs/step-18/storage-fixtures/"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{storageDirectory + "optional-method.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{storageDirectory + "field.a", true, false})
}

// The same checker-accepted narrowing is a promise in .a and a runtime type lie
// in .ts. Source Node evaluates arguments before TypeError; both backends must
// preserve that order and make the ruled terminal stop instead.
func TestStep18StorageTypeLie(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile(filepath.Join(repository, storageDirectory, "noncallable.a"))
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.Replace(string(source), "// a-check: refused storage not proven callable\n", "", 1))
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "main"+extension)
			if err := os.WriteFile(path, source, 0644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if extension == ".a" {
				var refused *lower.Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.What, "storage not proven callable") {
					t.Fatalf("got %v, want non-callable storage refusal", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			wantNode := run{stdout: []byte("arguments\n"), stderr: []byte("adamic: panic: TypeError: host.callback is not a function\n"), exitCode: 70}
			if difference := disagreement(wantNode, node); difference != "" {
				t.Fatalf("Node order changed: %s; got %#v", difference, node)
			}

			want := run{stdout: []byte("arguments\n"), stderr: []byte("adamic: panic: TypeError: optional call value is not callable\n"), exitCode: 70}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s; got %#v", name, difference, got)
				}
			}
		})
	}
}
