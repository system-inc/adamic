package oracle

import (
	"bytes"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mutate the stack's existing runtime callback protocol, keeping its counted ABI.
func librarySmallRuntimeMutant(t *testing.T, fixture, rule string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/regexp_replace.c"))
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string][2]string{
		"collection-order": {"adamic_array *match = matches->elements[k].reference;", "adamic_array *match = matches->elements[k].reference; regex->slots[1].number = match->properties->slots[0].number + adamic_string_length(match->elements[0].reference);"},
		"match":            {"value = match->elements[j];", "value = match->elements[j]; if (j == 0) value.reference = &adamic_string_empty;"},
		"literal":          {"adamic_array_push(pieces, (adamic_value){.reference = replacement});", "if (adamic_string_length(replacement) == 2 && memcmp(replacement->bytes, \"$&\", 2) == 0) { adamic_release(replacement); replacement = adamic_retain(match->elements[0].reference); } adamic_array_push(pieces, (adamic_value){.reference = replacement});"},
		"unicode":          {"(regex->slots[7].boolean || regex->slots[10].boolean) && first >= 0xd800", "first >= 0xd800"},
		"global":           {"if (!global)", "if (true)"},
		"offset":           {"size_t offset = (size_t)match->properties->slots[0].number;", "size_t offset = (size_t)match->properties->slots[0].number + 1;"},
		"reset":            {"regex->slots[1].number = 0;", "regex->slots[1].number = 1;"},
		"input":            {"value.reference = input;", "value.reference = &adamic_string_empty;"},
	}
	mutation := mutations[rule]
	if bytes.Count(runtime, []byte(mutation[0])) != 1 {
		t.Fatal("mutation site moved")
	}
	changed := strings.Replace(string(runtime), mutation[0], mutation[1], 1)
	changed = strings.ReplaceAll(changed, "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	source := changed + "\n" + strings.ReplaceAll(native.C(program), "adamic_regex_replace_callback", "adamic_regex_replace_callback_mutant")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed before Node comparison: %+v", actual)
	}
	if diff := disagreement(onNode(t, path), actual); diff != "stdout differs" {
		t.Fatalf("want Node stdout mismatch, got %q", diff)
	}
	t.Log("caught by source Node stdout comparison with sanitizers and leaks")
}
