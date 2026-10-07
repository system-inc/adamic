package oracle

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var jsonDecodeFixtures = []string{"json_decode_coverage_arrays.a", "json_decode_coverage_shapes.a", "json_decode_coverage_depth_extra.a", "json_decode_coverage_unicode_tag.a", "json_decode_coverage_nested.a", "json_decode_coverage_grammar.a", "json_decode_coverage_unions.a", "json_decode_coverage_tags.a", "json_decode_coverage_layouts.a", "json_decode_coverage_empty.a", "json_decode_coverage_calls.a", "json_decode_coverage_object_scalar.a", "json_decode_supplement.a", "json_decode_scalars.a", "json_decode_objects.a", "json_decode_grammar.a", "json_decode_depth.a", "json_decode_empty_tuple.a", "json_decode_empty_object.a"}

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
		{"null_prerequisite", "decodeJson<null>('null');", "containing null"},
		{"nullable_prerequisite", "decodeJson<number|null>('null');", "containing null"},
		{"null_field", "decodeJson<{readonly value:null}>('{}');", "containing null"},
		{"nullable_field", "decodeJson<{readonly value?:string|null}>('{}');", "containing null"},
		{"nested_nullable", "interface Branch { readonly children:readonly Branch[]; readonly value:number|null; } decodeJson<readonly Branch[]>('[]');", "containing null"},
		{"nullable_tuple", "decodeJson<readonly [string,null]>('[]');", "containing null"},
		{"nullable_array", "decodeJson<readonly (number|null)[]>('[]');", "containing null"},
		{"nullable_member", "decodeJson<{readonly kind:'A';readonly value:number|null}|{readonly kind:'B';readonly value:string}>('{}');", "containing null"},
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
			if c.want == "containing null" {
				var gap *lower.NotYet
				const fix = "nullable JSON fields come with the representation of T | null; decode the field as a discriminated union or leave it out"
				if !errors.As(err, &gap) || !strings.Contains(gap.What, "decodeJson<") || !strings.Contains(gap.Where, path+":2:") || !strings.HasSuffix(err.Error(), "\nfix: "+fix) {
					t.Fatalf("wanted NotYet with what, where and exact fix line, got %v", err)
				}
			}
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want %q, got %v", c.want, err)
			}
		})
	}
}
