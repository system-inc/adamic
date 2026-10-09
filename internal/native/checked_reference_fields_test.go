package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func checkedReferenceArraySource() string {
	return `#include "adamic.h"
static adamic_string alpha = ADAMIC_STRING("alpha");
static const adamic_field_contract text = {.kind=3, .declared="string"};
static const adamic_field_contract array = {.kind=5, .declared="readonly string[]", .reference=true, .structural=true, .field_count=1, .field_contracts=&text};
static const char *const names[] = {"names"};
static const bool references[] = {true}, optional[] = {false};
static const adamic_shape shape = {1, names, references, NULL, NULL};
static const adamic_shape empty_shape = {0, NULL, NULL, NULL, NULL};
static const adamic_field_contract root = {.kind=4, .declared="Result", .reference=true, .structural=true, .field_count=1, .field_names=names, .field_optional=optional, .field_contracts=&array};
int main(int argc, char **argv) {
    adamic_start(argc, argv);
    adamic_array *values = adamic_array_new(1, true);
    adamic_array_push(values, (adamic_value){.reference=&alpha});
    adamic_object *result = adamic_object_new(&shape);
    result->slots[0].reference = values;
    adamic_check_contract(&root, 4, (adamic_value){.reference=result}, 0, "results[]");
    adamic_array_push(values, (adamic_value){.reference=adamic_object_new(&empty_shape)});
    adamic_release(result);
    return 0;
}
`
}

func checkedReferenceReject(t *testing.T, source, message string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "reference-fields")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	output, err := command.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 70 || string(output) != message {
		t.Fatalf("want exit 70 %q, got %v %q", message, err, output)
	}
}

func TestRuntimeCheckedReferenceArrayKeepsContract(t *testing.T) {
	t.Parallel()
	checkedReferenceReject(t, checkedReferenceArraySource(), "adamic: panic: write failed: array[] expects string, got object\n")
}
func TestRuntimeCheckedReferenceArrayRejectsWrongElement(t *testing.T) {
	t.Parallel()
	source := strings.Replace(checkedReferenceArraySource(), ".reference=&alpha", ".reference=adamic_object_new(&empty_shape)", 1)
	// The initial string remains declared for the literal-domain variants.
	source = strings.Replace(source, "adamic_start(argc, argv);", "adamic_start(argc, argv); (void)alpha;", 1)
	checkedReferenceReject(t, source, "adamic: panic: write failed: results[] expects Result, got object\n")
}
func TestRuntimeCheckedReferenceArrayRejectsLiteral(t *testing.T) {
	t.Parallel()
	source := strings.Replace(checkedReferenceArraySource(), "{.kind=3, .declared=\"string\"}", "{.kind=3, .declared=\"\\\"alpha\\\"\", .count=1, .allowed=(const adamic_value[]){{.reference=&alpha}}}", 1)
	checkedReferenceReject(t, source, "adamic: panic: write failed: results[] expects Result, got object\n")
}
func TestRuntimeCheckedReferenceArrayRejectsWeakerContract(t *testing.T) {
	t.Parallel()
	source := strings.Replace(checkedReferenceArraySource(), "result->slots[0].reference = values;", "result->slots[0].reference = values; static const adamic_field_contract weaker = {.kind=3, .nullable=true, .declared=\"string | undefined\"}; values->element_contract = &weaker;", 1)
	checkedReferenceReject(t, source, "adamic: panic: write failed: results[] expects Result, got object\n")
}
