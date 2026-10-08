package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// These mutants remain valid C, finish normally, and change only answers. Sanitizers and the
// compiler cannot reject them: the source run on Node has to catch the wrong result.
func TestDateOracleCatchesMutants(t *testing.T) {
	t.Parallel()
	mutants := []struct{ name, fixture, before, after string }{
		{"nullable_typeof", "json", "&adamic_typeof_object", "&adamic_typeof_undefined"},
		{"nullable_number", "json", " == NULL ? 0.0 :", " == NULL ? NAN :"},
		{"nullable_stringify", "json", "adamic_json_nullable_string", "adamic_json_string"},
		{"date_stringify", "json", "adamic_json_date", "adamic_json_map"},
		{"toJSON", "json", "adamic_date_json(", "date_mutant_json("},
		{"dynamic_parse", "dynamic_parse", "adamic_date_parse_iso(", "1 + adamic_date_parse_iso("},
		{"constructor_clip", "construct", "adamic_date_new(", "adamic_date_new(1 + "},
		{"UTC", "utc", "adamic_date_utc(", "1 + adamic_date_utc("},
		{"invalid_NaN", "get", "adamic_date_get(", "date_mutant_get("},
		{"getters", "get", "adamic_date_get(", "1 + adamic_date_get("},
		{"setters", "set", "adamic_date_set(", "1 + adamic_date_set("},
		{"iso_format", "iso", "adamic_date_iso(", "date_mutant_iso("},
		{"format", "format", "adamic_date_new(", "adamic_date_new(86400000 + "},
		{"parse", "parse", "adamic_date_parse_iso(", "1 + adamic_date_parse_iso("},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_date_"+mutant.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			if !strings.Contains(source, mutant.before) {
				t.Fatalf("mutant target %q absent", mutant.before)
			}
			source = strings.ReplaceAll(source, mutant.before, mutant.after)
			if mutant.name == "toJSON" {
				source = "#include \"adamic.h\"\nstatic adamic_string *date_mutant_json(const adamic_object *date) { return isnan(adamic_date_value(date)) ? adamic_date_format(date, 0) : adamic_date_iso(date); }\n" + source
			}
			if mutant.name == "iso_format" {
				source = "#include \"adamic.h\"\nstatic adamic_string *date_mutant_iso(const adamic_object *date) { adamic_string *text = adamic_date_iso(date); ((char *)text->bytes)[text->length - 2] = '0'; return text; }\n" + source
			}
			if mutant.name == "invalid_NaN" {
				source = "#include \"adamic.h\"\nstatic double date_mutant_get(const adamic_object *date, int field) { double result = adamic_date_get(date, field); return isnan(result) ? 0 : result; }\n" + source
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := execute(t, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: exit %d, stderr %q", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("got %q, want stdout differs", difference)
			}
			t.Log("Node caught the mutant: stdout differs; sanitizer clean")
		})
	}
}

// Pin before parallel tests lower fixtures or launch children, including counted and leak runs.
func init() {
	if err := os.Setenv("TZ", "UTC"); err != nil {
		panic(err)
	}
}
