package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var jsonDecodeFixtures = []string{"json_decode_supplement.a", "json_decode_scalars.a", "json_decode_objects.a", "json_decode_grammar.a", "json_decode_depth.a"}

func init() {
	for _, name := range jsonDecodeFixtures {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
func TestJSONDecode(t *testing.T) {
	for _, name := range jsonDecodeFixtures {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			got, binary := natively(t, p)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("native: %s; Node %q; native %q; stderr %q", difference, truth.stdout, got.stdout, got.stderr)
			}
			if report := leaks(t, p, binary); report != "" {
				t.Fatal(report)
			}
			if js := onJavaScriptBackend(t, p); disagreement(truth, js) != "" {
				t.Fatalf("JavaScript differs: %q %q %q", truth.stdout, js.stdout, js.stderr)
			}
		})
	}
}
func TestJSONDecodeRefusals(t *testing.T) {
	cases := []struct{ name, source, want string }{
		{"missing", "decodeJson('1');", "name the type: decodeJson<YourType>(text)"},
		{"null_prerequisite", "decodeJson<null>('null');", "nullable value representation"},
		{"nullable_prerequisite", "decodeJson<number|null>('null');", "JSON data"},
		{"undefined", "decodeJson<undefined>('null');", "JSON data"},
		{"function", "decodeJson<() => number>('{}');", "JSON data"},
		{"class", "class Thing { readonly n = 1; } decodeJson<Thing>('{}');", "JSON data"},
		{"map", "decodeJson<Map<string,number>>('{}');", "JSON data"},
		{"set", "decodeJson<Set<string>>('[]');", "JSON data"},
		{"ambiguous", "decodeJson<{readonly x:number}|{readonly y:number}>('{}');", "JSON data"},
		{"overlap", "decodeJson<{readonly kind:'A';readonly x:number}|{readonly kind:'A';readonly y:number}>('{}');", "JSON data"},
		{"open", "function read<T>(text:string):void { decodeJson<T>(text); } read<number>('1');", "JSON data"},
		{"field_undefined", "decodeJson<{readonly n?:number|undefined}>('{}');", "JSON data"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "probe.a")
			if err := os.WriteFile(path, []byte("import { decodeJson } from 'adamic';\n"+c.source), 0600); err != nil {
				t.Fatal(err)
			}
			p, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), p)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want %q, got %v", c.want, err)
			}
		})
	}
}
