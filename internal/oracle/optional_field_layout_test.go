package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A checked copy rejects a present unreadied field. Its NULL-expression path is
// also the runtime's raw copy: it must carry the bit to the eventual checked read.
// An absent unreadied slot exercises preservation on a successful checked copy.
func TestOptionalFieldCopyState(t *testing.T) {
	t.Parallel()
	implementation, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/object.c"))
	if err != nil {
		t.Fatal(err)
	}
	const probe = `
#include <stdio.h>
int main(void) {
    static const char *const names[] = {"pending", "slot", "ready"};
    static const bool references[] = {false, false, false};
    static const adamic_shape shape = {3, names, references, NULL};
    static const int types[] = {1, 7, 1};
    static adamic_shape_types metadata = {&shape, types, NULL};
    static const char *const extra_names[] = {"extra"};
    static const bool extra_references[] = {false};
    static const adamic_shape extra = {1, extra_names, extra_references, NULL};
    static const int extra_types[] = {7};
    static adamic_shape_types extra_metadata = {&extra, extra_types, NULL};
    adamic_register_shape_types(&metadata);
    adamic_register_shape_types(&extra_metadata);
    adamic_object *source = adamic_object_new(&shape);
    source->slots[1].number = adamic_maybe_number_pack((adamic_maybe_number){false, 0});
    adamic_object_absent(source, 1);
    for (size_t index = 0; index < 3; index++) adamic_object_field_types(source)[index] = (unsigned char)types[index];
    adamic_object_initialized(source)[0] = 0;
    for (int dynamic = 0; dynamic < 2; dynamic++) {
        adamic_object *copy = dynamic ? adamic_object_copy_reserving_checked(source, &extra, NULL) : adamic_object_copy_checked(source, NULL);
        printf("types %u %u %u\n", adamic_object_field_types(copy)[0], adamic_object_field_types(copy)[1], adamic_object_field_types(copy)[2]);
        adamic_heap *pending = adamic_dynamic_property(&copy->heap, "pending");
        printf("%d %d %d %d\n", adamic_has_property(&copy->heap, "slot"), adamic_object_initialized(copy)[0], adamic_has_property(&copy->heap, "pending"), pending == NULL);
        adamic_release(pending);
        adamic_release(copy);
    }
    // Absent storage is not read by a checked spread, but its state is still copied.
    adamic_object_absent(source, 0);
    for (int dynamic = 0; dynamic < 2; dynamic++) {
        adamic_object *copy = dynamic ? adamic_object_copy_reserving_checked(source, &extra, "copy") : adamic_object_copy_checked(source, "copy");
        printf("types %u %u %u\n", adamic_object_field_types(copy)[0], adamic_object_field_types(copy)[1], adamic_object_field_types(copy)[2]);
        printf("%d %d %d\n", adamic_has_property(&copy->heap, "slot"), adamic_object_initialized(copy)[0], adamic_has_property(&copy->heap, "pending"));
        adamic_slot_cache cache = {NULL, 0};
        adamic_value *slot = adamic_object_write_field(copy, "pending", &cache);
        slot->number = 0;
        adamic_heap *value = adamic_dynamic_property(&copy->heap, "pending");
        printf("%d %d\n", adamic_object_initialized(copy)[cache.index], value != NULL && value->kind == adamic_kind_number && ((adamic_number_box *)value)->number == 0);
        adamic_release(value);
        adamic_release(copy);
    }
    adamic_release(source);
    return 0;
}
`
	want := "types 1 7 1\n0 0 1 1\ntypes 1 7 1\n0 0 1 1\ntypes 1 7 1\n0 0 0\n1 1\ntypes 1 7 1\n0 0 0\n1 1\n"
	original := string(implementation)
	// Compile the production implementation under private symbol names. Runtime
	// link flags load the complete archive, so replacing public symbols would clash.
	definitions := regexp.MustCompile(`(?m)^(?:adamic_object \*|adamic_value \*|adamic_closure \*|adamic_maybe_number |adamic_maybe_boolean |adamic_value |void |bool )((?:adamic_object_|adamic_view_)[a-z_]+)\(`)
	prefix := ""
	for _, match := range definitions.FindAllStringSubmatch(original, -1) {
		prefix += "#define " + match[1] + " probe_" + match[1] + "\n"
	}
	for _, mutation := range []struct {
		name, from, to string
	}{
		{"baseline", "", ""},
		{"drop presence on checked copy", "if (source->class == NULL) adamic_object_orders(object)[index] = adamic_object_orders(source)[cache.index];", "(void)source;"},
		{"drop readiness on checked copy", "adamic_object_initialized(object)[index] = adamic_object_initialized(source)[cache.index];", "(void)source;"},
		{"drop representation on checked copy", "adamic_object_field_types(object)[index] = adamic_object_field_types(source)[cache.index];", "(void)source;"},
		{"overlap presence and readiness", "#include \"adamic.h\"", "#include \"adamic.h\"\n#define adamic_object_initialized(object) ((unsigned char *)(void *)adamic_object_orders(object))"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			code := original
			if mutation.from != "" {
				if strings.Count(code, mutation.from) != 1 {
					t.Fatal("mutant must change exactly one production site")
				}
				code = strings.Replace(code, mutation.from, mutation.to, 1)
			}
			// Each mutation changes the production copy implementation in isolation.
			for _, sanitize := range []bool{false, true} {
				binary := filepath.Join(t.TempDir(), "probe")
				if err := native.Build(prefix+code+probe, binary, native.Options{Sanitize: sanitize}); err != nil {
					t.Fatal(err)
				}
				got := execute(t, binary)
				if mutation.name == "overlap presence and readiness" && got.exitCode == 70 && strings.Contains(string(got.stderr), "read before assignment: field 'ready' in copy") {
					t.Logf("caught by checked read, sanitize=%t: %s", sanitize, got.stderr)
					continue
				}
				if got.exitCode != 0 || len(got.stderr) != 0 {
					t.Fatalf("probe must finish without diagnostics: %d %s", got.exitCode, got.stderr)
				}
				if mutation.name == "baseline" {
					if string(got.stdout) != want {
						t.Fatalf("copy state: want %q, got %q", want, got.stdout)
					}
				} else if string(got.stdout) == want {
					t.Fatal("mutant escaped the state assertions")
				} else {
					t.Logf("caught by state assertions, sanitize=%t: %q", sanitize, got.stdout)
				}
			}
		})
	}
}
