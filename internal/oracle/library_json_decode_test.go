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

var jsonDecodeFixtures = []string{
	"json_decode_coverage_arrays.a",
	"json_decode_coverage_shapes.a",
	"json_decode_coverage_depth_extra.a",
	"json_decode_coverage_unicode_tag.a",
	"json_decode_coverage_nested.a",
	"json_decode_coverage_grammar.a",
	"json_decode_coverage_unions.a",
	"json_decode_coverage_tags.a",
	"json_decode_coverage_layouts.a",
	"json_decode_coverage_empty.a",
	"json_decode_coverage_calls.a",
	"json_decode_coverage_object_scalar.a",
	"json_decode_supplement.a",
	"json_decode_scalars.a",
	"json_decode_objects.a",
	"json_decode_grammar.a",
	"json_decode_depth.a",
	"json_decode_empty_tuple.a",
	"json_decode_empty_object.a",
}

// Main's enumeration proof rejects the empty {} type before decoded ObjectKeys lowering.
// Keep these source fixtures intact and verify the refusal rather than bypassing that proof.
func jsonDecodeEnumerationRefused(name string) bool {
	return name == "json_decode_empty_object.a" || name == "json_decode_coverage_empty.a"
}
func assertJSONDecodeEnumerationRefusal(test *testing.T, sourceError error) {
	test.Helper()
	var gap *lower.NotYet
	if !errors.As(sourceError, &gap) || !strings.Contains(gap.What, "Object.keys on a non-object") {
		test.Fatalf("want main's empty-object enumeration refusal, got %v", sourceError)
	}
}

func init() {
	for _, name := range jsonDecodeFixtures {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, !jsonDecodeEnumerationRefused(name), false})
	}
}
func jsonDecodeKeys() []string { return jsonDecodeFixtures }

func runJSONDecodeShard(test *testing.T, shard int) {
	for _, index := range familyCases(test, "TestJSONDecode", shard, jsonDecodeKeys()) {
		name := jsonDecodeFixtures[index]
		test.Run(name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if jsonDecodeEnumerationRefused(name) {
				assertJSONDecodeEnumerationRefusal(test, loweredError)
				return
			}
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			truth := onNode(test, path)
			if _, probing := familyProbe("TestJSONDecode"); probing {
				truth.stdout = append(truth.stdout, '!')
			}
			got, binary := natively(test, program)
			if difference := disagreement(truth, got); difference != "" {
				test.Fatalf(
					"native: %s; Node %q; native %q; stderr %q",
					difference,
					truth.stdout,
					got.stdout,
					got.stderr,
				)
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatal(report)
			}
			if js := onJavaScriptBackend(test, program); disagreement(truth, js) != "" {
				test.Fatalf("JavaScript differs: %q %q %q", truth.stdout, js.stdout, js.stderr)
			}
		})
	}
}
func TestJSONDecodeRefusals(test *testing.T) {
	test.Parallel()
	cases := []struct{ name, source, want string }{
		{"missing", "decodeJson('1');", "name the type: decodeJson<YourType>(text)"},
		{"null_prerequisite", "decodeJson<null>('null');", "containing null"},
		{"nullable_prerequisite", "decodeJson<number|null>('null');", "containing null"},
		{"null_field", "decodeJson<{readonly value:null}>('{}');", "containing null"},
		{"nullable_field", "decodeJson<{readonly value?:string|null}>('{}');", "containing null"},
		{
			"nested_nullable",
			"interface Branch { readonly children:readonly Branch[]; readonly value:number|null; } decodeJson<readonly Branch[]>('[]');",
			"containing null",
		},
		{"nullable_tuple", "decodeJson<readonly [string,null]>('[]');", "containing null"},
		{"nullable_array", "decodeJson<readonly (number|null)[]>('[]');", "containing null"},
		{
			"nullable_member",
			"decodeJson<{readonly kind:'A';readonly value:number|null}|{readonly kind:'B';readonly value:string}>('{}');",
			"containing null",
		},
		{"undefined", "decodeJson<undefined>('null');", "JSON data"},
		{"function", "decodeJson<() => number>('{}');", "JSON data"},
		{"class", "class Thing { readonly n = 1; } decodeJson<Thing>('{}');", "JSON data"},
		{"map", "decodeJson<Map<string,number>>('{}');", "JSON data"},
		{"set", "decodeJson<Set<string>>('[]');", "JSON data"},
		{"ambiguous", "decodeJson<{readonly x:number}|{readonly y:number}>('{}');", "JSON data"},
		{
			"overlap",
			"decodeJson<{readonly kind:'A';readonly x:number}|{readonly kind:'A';readonly y:number}>('{}');",
			"JSON data",
		},
		{"open", "function read<T>(text:string):void { decodeJson<T>(text); } read<number>('1');", "JSON data"},
		{"field_undefined", "decodeJson<{readonly n?:number|undefined}>('{}');", "JSON data"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			path := filepath.Join(test.TempDir(), "probe.a")
			if writeError := os.WriteFile(
				path,
				[]byte("import { decodeJson } from 'adamic';\n"+testCase.source),
				0600,
			); writeError != nil {
				test.Fatal(writeError)
			}
			program, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			_, lowerError := lower.Lower(context.Background(), program)
			if testCase.want == "containing null" {
				var gap *lower.NotYet
				const fix = "nullable JSON fields come with the representation of T | null; decode the field as a discriminated union or leave it out"
				if !errors.As(lowerError, &gap) ||
					!strings.Contains(gap.What, "decodeJson<") ||
					!strings.Contains(gap.Where, path+":2:") ||
					!strings.HasSuffix(lowerError.Error(), "\nfix: "+fix) {
					test.Fatalf("wanted NotYet with what, where and exact fix line, got %v", lowerError)
				}
			}
			if lowerError == nil || !strings.Contains(lowerError.Error(), testCase.want) {
				test.Fatalf("want %q, got %v", testCase.want, lowerError)
			}
		})
	}
}
