package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// These mutants remain valid C, finish normally, and change only answers. Sanitizers and the
// compiler cannot reject them: the source run on Node has to catch the wrong result.
func TestDateOracleCatchesMutants(t *testing.T) {
	t.Parallel()
	mutants := []struct{ name, fixture, before, after string }{
		{"own_property", "own", "adamic_date_has_own(", "!adamic_date_has_own("},
		{"nullable_typeof", "json", "adamic_union_typeof(", "date_mutant_typeof("},
		{"nullable_number", "json", "adamic_temporary_10 = (0x0p+00);", "adamic_temporary_10 = NAN;"},
		{"nullable_stringify", "json", "adamic_json_stringify(", "date_mutant_stringify("},
		{"date_stringify", "json", "adamic_json_date", "adamic_json_map"},
		{"toJSON", "json", "adamic_date_json(", "date_mutant_json("},

		{"constructor_clip", "construct", "adamic_date_new(", "adamic_date_new(1 + "},
		{"UTC", "utc", "adamic_date_utc(", "1 + adamic_date_utc("},
		{"invalid_NaN", "get", "adamic_date_get(", "date_mutant_get("},
		{"getters", "get", "adamic_date_get(", "1 + adamic_date_get("},
		{"setters", "set", "adamic_date_set(", "1 + adamic_date_set("},
		{"iso_format", "iso", "adamic_date_iso(", "date_mutant_iso("},
		{"format", "area_format", "adamic_date_new(", "adamic_date_new(86400000 + "},
		{"parse", "area_parse", "adamic_date_parse_iso(", "1 + adamic_date_parse_iso("},
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
			if mutant.name == "nullable_typeof" {
				source = "#include \"adamic.h\"\nstatic adamic_string *date_mutant_typeof(const adamic_heap *value, bool nullable) { return value == NULL && nullable ? &adamic_typeof_undefined : adamic_union_typeof(value, nullable); }\n" + source
			}
			if mutant.name == "nullable_stringify" {
				source = "#include \"adamic.h\"\n#include \"json_stringify.h\"\n#include <string.h>\nstatic adamic_string *date_mutant_stringify(adamic_value value, const adamic_json_schema *schema, adamic_value replacer, const adamic_json_schema *replacer_schema, adamic_value space, const adamic_json_schema *space_schema) { adamic_string *text = adamic_json_stringify(value, schema, replacer, replacer_schema, space, space_schema); if (text != NULL && text->length == 4 && memcmp(text->bytes, \"null\", 4) == 0) { adamic_release(text); static adamic_string wrong = ADAMIC_STRING(\"undefined\"); return adamic_retain(&wrong); } return text; }\n" + source
			}
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
			if mutant.name == "own_property" {
				if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
					result := onWASI(t, source)
					if result.exitCode != 0 || len(result.stderr) != 0 {
						t.Fatalf("WASI mutant must finish cleanly: %s", result.stderr)
					}
					if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
						t.Fatalf("WASI Node comparison: %s", difference)
					}
				}
				script := filepath.Join(t.TempDir(), "mutant.mjs")
				text := strings.ReplaceAll(javascript.JavaScript(program), "Date.prototype.hasOwnProperty(", "!Date.prototype.hasOwnProperty(")
				if err := os.WriteFile(script, []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
				result := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script)
				if result.exitCode != 0 || len(result.stderr) != 0 {
					t.Fatalf("JavaScript mutant must finish cleanly: %s", result.stderr)
				}
				if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
					t.Fatalf("JavaScript Node comparison: %s", difference)
				}
			}
			t.Log("Node caught the mutant: stdout differs; sanitizer clean")
		})
	}
}
