package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	regex "github.com/system-inc/adamic/internal/regexp"
)

func checkedLibraryPushSource(t *testing.T) string {
	t.Helper()
	program, err := regex.Compile("a", "g")
	if err != nil {
		t.Fatal(err)
	}
	declarations, err := program.NativeDeclarations("pattern")
	if err != nil {
		t.Fatal(err)
	}
	return "#include \"adamic.h\"\n#include \"count.h\"\n#include <stdio.h>\n" + declarations + `
static const char *const fields[] = {"done"};
static const bool optional[] = {true};
static const adamic_field_contract children[] = {{.kind=9, .nullable=true, .declared="boolean | undefined"}};
static const adamic_field_contract element = {.kind=4, .declared="{ readonly done?: boolean | undefined }", .reference=true, .structural=true, .field_count=1, .field_names=fields, .field_optional=optional, .field_contracts=children};
int main(int argc, char **argv) {
    adamic_start(argc, argv);
    static adamic_string pattern_text=ADAMIC_STRING("a"), flags=ADAMIC_STRING("g"), input=ADAMIC_STRING("aba");
    adamic_object *regex=adamic_regex_new(&pattern, &pattern_text, &flags);
    adamic_object *iterator=adamic_regex_match_all(&input, regex);
    adamic_array *source=adamic_array_new(3, true);
    for(size_t i=0; i<3; i++) {
        adamic_object *result=adamic_regex_next(iterator);
        if(result->shape->contracts != NULL) { return 2; }
        adamic_array_push(source, (adamic_value){.reference=result});
    }
    adamic_array *destination=adamic_array_new(3, true);
    destination->element_contract=&element;
    adamic_array_append(destination, source);
    adamic_release(source);
    for(size_t i=0; i<destination->length; i++) {
        adamic_object *result=destination->elements[i].reference;
        printf("%s%s", i ? "," : "", result->slots[0].boolean ? "done" : "match");
    }
    printf("\n");
    adamic_release(destination); adamic_release(iterator); adamic_release(regex);
    if(adamic_counted.live != 0) { return 3; }
    return 0;
}
`
}

func TestRuntimeCheckedLibraryPush(t *testing.T) {
	t.Parallel()
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := filepath.Abs("testdata/checked_library_push.a")
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", runner, fixture)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "library-push")
		if err := Build(checkedLibraryPushSource(t), binary, Options{Sanitize: sanitize, Count: true}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %v: got %q, Node %q", sanitize, got, want)
		}
	}
}

// A broad runtime boolean slot cannot promise the literal false even when its current payload is false.
func TestRuntimeCheckedLibraryPushRejectsLiteral(t *testing.T) {
	t.Parallel()
	source := strings.ReplaceAll(checkedLibraryPushSource(t), "{.kind=9, .nullable=true, .declared=\"boolean | undefined\"}", "{.kind=2, .declared=\"false\", .count=1, .allowed=(const adamic_value[]){{.boolean=false}}}")
	binary := filepath.Join(t.TempDir(), "library-push-literal")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	output, err := command.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 70 || !strings.Contains(string(output), "expects { readonly done?: boolean | undefined }") {
		t.Fatalf("want exit 70 at contract, got %v: %s", err, output)
	}
}
