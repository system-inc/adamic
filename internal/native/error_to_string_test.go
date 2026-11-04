package native

import (
	"path/filepath"
	"testing"
)

// adamic_error_to_string is Error.prototype.toString over an Error's name and message, held to Node's
// String(error) for each of the four ways the two can be empty, and for an Error made without a message,
// whose fallback lives in its private slot (adamic_error_field_name). Mutant, run by hand: error_text's
// separator always 2 (name + ": " + message whatever is empty) turns every row but the first red.
func TestErrorToStringIsNodes(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <stdio.h>
static adamic_string error = ADAMIC_STRING("Error");
static adamic_string unnamed = ADAMIC_STRING("");
static adamic_string boom = ADAMIC_STRING("boom");
static adamic_string empty = ADAMIC_STRING("");
int main(void) {
    adamic_string *names[] = {&error, &error, &unnamed, &unnamed, &error, &unnamed};
    adamic_string *messages[] = {&boom, &empty, &boom, &empty, NULL, NULL};
    for (size_t i = 0; i < 6; i++) {
        adamic_object *made = adamic_error_new(messages[i]);
        adamic_release(made->slots[0].reference);
        made->slots[0].reference = adamic_retain(names[i]);
        adamic_string *text = adamic_error_to_string(made);
        printf("[%.*s]\n", (int)text->length, text->bytes);
        adamic_release(text);
        adamic_release(made);
    }
    return 0;
}
`
	const oracle = `for (const [name, message] of [['Error', 'boom'], ['Error', ''], ['', 'boom'], ['', ''], ['Error', undefined], ['', undefined]]) { const error = new Error(message); error.name = name; console.log('[' + String(error) + ']'); }`
	want := runWithInput(t, "", "node", "--eval", oracle)
	if want != "[Error: boom]\n[Error]\n[boom]\n[]\n[Error]\n[]\n" {
		t.Fatalf("Node changed: %q", want)
	}
	binary := filepath.Join(t.TempDir(), "error-to-string")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != want {
		t.Fatalf("native %q; Node %q", got, want)
	}
}

// adamic_error_is is true for every Error shape the runtime makes that a harness reaches: new Error with
// a message and without one (exceptions.c), a directory error (node_fs_directory.c), a file error and
// realpath's system error (node_fs_file.c), process.chdir's error (node_process.c) and exitCode's
// RangeError (process.c). It is false for a plain object. Mutant, run by hand: adamic_error_is returning
// false turns every Error row red.
func TestErrorIsEveryRuntimeErrorShape(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <errno.h>
#include <stdio.h>
static adamic_string message = ADAMIC_STRING("boom");
static adamic_string absent = ADAMIC_STRING("/adamic-error-is-absent/missing");
static const char *const plain_names[] = {"name"};
static const bool plain_references[] = {true};
static const adamic_field_kind plain_kinds[] = {adamic_field_reference};
static const adamic_shape plain_shape = {1, plain_names, plain_references, NULL, plain_kinds, false};
static adamic_object *take(void) {
    adamic_object *thrown = adamic_thrown;
    adamic_thrown = NULL;
    return thrown;
}
static void report(const char *label, adamic_object *object) {
    if (object == NULL) { printf("%s nothing thrown\n", label); return; }
    printf("%s %s\n", label, adamic_error_is(object) ? "true" : "false");
    adamic_release(object);
}
int main(void) {
    report("error", adamic_error_new(&message));
    report("error-without-message", adamic_error_new(NULL));
    adamic_node_fs_raise(&absent, ENOENT, "scandir");
    report("directory", take());
    adamic_fs_file_unlink(&absent);
    report("file", take());
    adamic_fs_file_realpath_walk_error(&absent, ENOENT, "lstat");
    report("realpath", take());
    adamic_node_chdir(&absent);
    report("chdir", take());
    adamic_process_set_exit_code((adamic_maybe_number){true, 1.5});
    report("exit-code", take());
    adamic_object *plain = adamic_object_new(&plain_shape);
    plain->slots[0].reference = adamic_retain(&message);
    report("plain", plain);
    return 0;
}
`
	const want = "error true\nerror-without-message true\ndirectory true\nfile true\nrealpath true\nchdir true\nexit-code true\nplain false\n"
	binary := filepath.Join(t.TempDir(), "error-is")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != want {
		t.Fatalf("native %q; want %q", got, want)
	}
}
