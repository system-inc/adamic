package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/scout_union_keys.a", "stage3/map-keys/mixed_map.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestScoutUnionKeyMutants(t *testing.T) {
	for _, rule := range []struct{ name, body string }{
		{"number_string", `if (boxed != NULL && boxed->kind == adamic_kind_number && ((adamic_number_box *)boxed)->number == 1) {adamic_release(key.reference); key.reference = adamic_string_from_number(1);}`},
		{"NaN", `if (boxed != NULL && boxed->kind == adamic_kind_number && isnan(((adamic_number_box *)boxed)->number)) {adamic_release(key.reference); if (map->reference_values) adamic_release(value.reference); return;}`},
		{"negative_zero", ``},
	} {
		t.Run(rule.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_union_keys.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			code := strings.ReplaceAll(original, "adamic_map_set(", "scout_union_store(")
			if code == original {
				t.Fatal("no insertion changed")
			}
			after := ""
			if rule.name == "negative_zero" {
				after = `for(size_t i=0;i<map->used;i++){if(map->entries[i].deleted)continue;adamic_heap *stored=map->entries[i].key.reference;if(stored!=NULL && stored->kind==adamic_kind_number && ((adamic_number_box *)stored)->number==0)((adamic_number_box *)stored)->number=-0.0;}`
			}
			code = insertCollectionMutant(code, `static void scout_union_store(adamic_map *map, adamic_value key, adamic_value value) {adamic_heap *boxed=key.reference;(void)boxed;`+rule.body+`adamic_map_set(map,key,value);`+after+`}`)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("unclean mutant: %d %s", actual.exitCode, actual.stderr)
			}
			if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
				t.Fatalf("want stdout disagreement: %q", difference)
			}
		})
	}
}
