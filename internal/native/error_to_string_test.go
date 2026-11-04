package native

import (
	"os"
	"path/filepath"
	"strings"
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
// a message and without one (exceptions.c), a directory error and a directory's coded error
// (node_fs_directory.c), a file error, a file's coded error and realpath's system error (node_fs_file.c),
// process.chdir's error (node_process.c) and exitCode's coded RangeError (process.c). It is false for a
// plain object. Mutant, run by hand: adamic_error_is returning false turns every Error row red.
func TestErrorIsEveryRuntimeErrorShape(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <errno.h>
#include <stdio.h>
static adamic_string message = ADAMIC_STRING("boom");
static adamic_string absent = ADAMIC_STRING("/adamic-error-is-absent/missing");
static adamic_string nul = ADAMIC_STRING("a\0b");
static adamic_string read_flag = ADAMIC_STRING("r");
static const char *const plain_names[] = {"name"};
static const bool plain_references[] = {true};
static const adamic_field_kind plain_kinds[] = {adamic_field_reference};
static const adamic_shape plain_shape = {1, plain_names, plain_references, NULL, plain_kinds};
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
    adamic_node_fs_readdir(&nul, NULL);
    report("directory-coded", take());
    adamic_fs_file_unlink(&absent);
    report("file", take());
    adamic_fs_file_read_file(&nul, &read_flag);
    report("file-coded", take());
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
	const want = "error true\nerror-without-message true\ndirectory true\ndirectory-coded true\nfile true\nfile-coded true\nrealpath true\nchdir true\nexit-code true\nplain false\n"
	binary := filepath.Join(t.TempDir(), "error-is")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != want {
		t.Fatalf("native %q; want %q", got, want)
	}
}

// adamic_error_to_string of every Node-coded error a harness reaches, each made by the runtime call a
// program's host call lowers to, so by the raise site's own shape, held to Node's String(error) of the
// same call: name [code]: message where Node builds the error with makeNodeErrorWithCode or as a
// SystemError, and name: message where it doesn't (a descriptor's RangeError, and the errno errors).
// An emptied message or name keeps the brackets and ": ", as the class's toString is a template. Every
// row runs before any effect: rm of the working directory without recursive refuses it, and the
// utimes path is never reached. Mutant, run by hand: error_text's coded test false turns every
// bracketed row red, and the descriptor and errno rows stay green (#z28avqd).
func TestCodedErrorToStringIsNodes(t *testing.T) {
	t.Parallel()
	const source = `#define ADAMIC_NODE_HOST 1
#include "adamic.h"
#include <math.h>
#include <stdio.h>
static adamic_string nul = ADAMIC_STRING("a\0b");
static adamic_string file = ADAMIC_STRING("x");
static adamic_string here = ADAMIC_STRING(".");
static adamic_string missing = ADAMIC_STRING("/adamic-coded-host-error-missing");
static adamic_string missing_file = ADAMIC_STRING("/adamic-coded-host-error-missing/file");
static adamic_string read_flag = ADAMIC_STRING("r");
static adamic_string bad_flag = ADAMIC_STRING("not-a-flag");
static adamic_string empty = ADAMIC_STRING("");
static adamic_object *take(void) {
    adamic_object *thrown = adamic_thrown;
    adamic_thrown = NULL;
    return thrown;
}
static void show(adamic_object *thrown) {
    if (thrown == NULL) { printf("no throw\n"); return; }
    adamic_string *text = adamic_error_to_string(thrown);
    printf("[%.*s]\n", (int)text->length, text->bytes);
    adamic_release(text);
    adamic_release(thrown);
}
static adamic_object *emptied(size_t slot) {
    adamic_fs_file_read_file(&nul, &read_flag);
    adamic_object *thrown = take();
    adamic_release(thrown->slots[slot].reference);
    thrown->slots[slot].reference = adamic_retain(&empty);
    return thrown;
}
int main(void) {
    adamic_array *one = adamic_array_make(1, adamic_ir_Number);
    adamic_array_push(one, (adamic_value){.number = 0});
    adamic_array *none = adamic_array_make(0, adamic_ir_Number);
    adamic_fs_file_read_file(&nul, &read_flag); show(take());
    adamic_fs_file_open(&file, &bad_flag, 0666); show(take());
    adamic_fs_file_read_sync(0, one, 0, 1, 1.5); show(take());
    adamic_fs_file_read_sync(0, none, 0, 1, 0); show(take());
    adamic_fs_file_close(-1); show(take());
    adamic_process_set_exit_code((adamic_maybe_number){true, 1.5}); show(take());
    adamic_fs_file_rm(&here, false, false); show(take());
    adamic_fs_file_utimes(&missing, INFINITY, 0); show(take());
    adamic_node_fs_readdir(&nul, NULL); show(take());
    adamic_fs_file_mkdtemp(&nul); show(take());
    show(emptied(1));
    show(emptied(0));
    adamic_fs_file_read_file(&missing_file, &read_flag); show(take());
    adamic_node_fs_readdir(&missing, NULL); show(take());
    adamic_release(one);
    adamic_release(none);
    return 0;
}
`
	const oracle = `const fs = require('node:fs'); const missing = '/adamic-coded-host-error-missing'; const show = (run, edit) => { try { run(); console.log('no throw'); } catch (error) { if (edit !== undefined) { edit(error); } console.log('[' + String(error) + ']'); } }; show(() => fs.readFileSync('a\0b', 'utf8')); show(() => fs.openSync('x', 'not-a-flag')); show(() => fs.readSync(0, Buffer.alloc(1), 0, 1, 1.5)); show(() => fs.readSync(0, Buffer.alloc(0), 0, 1, 0)); show(() => fs.closeSync(-1)); show(() => { process.exitCode = 1.5; }); show(() => fs.rmSync('.')); show(() => fs.utimesSync(missing, Infinity, 0)); show(() => fs.readdirSync('a\0b')); show(() => fs.mkdtempSync('a\0b')); show(() => fs.readFileSync('a\0b', 'utf8'), (error) => { error.message = ''; }); show(() => fs.readFileSync('a\0b', 'utf8'), (error) => { error.name = ''; }); show(() => fs.readFileSync(missing + '/file', 'utf8')); show(() => fs.readdirSync(missing));`
	want := runWithInput(t, "", "node", "--eval", oracle)
	if want != codedErrorTexts {
		t.Fatalf("Node changed: %q", want)
	}
	binary := filepath.Join(t.TempDir(), "coded-error-to-string")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got, wanted := strings.Split(runWithInput(t, "", binary), "\n"), strings.Split(want, "\n")
	if len(got) != len(wanted) {
		t.Fatalf("native printed %d lines; Node %d: %q", len(got), len(wanted), got)
	}
	for row := range wanted {
		if got[row] != wanted[row] {
			t.Errorf("row %d: native %q; Node %q", row+1, got[row], wanted[row])
		}
	}
}

// codedErrorTexts is what TestCodedErrorToStringIsNodes's calls print on Node (24.14.1, measured).
const codedErrorTexts = `[TypeError [ERR_INVALID_ARG_VALUE]: The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'a\x00b']
[TypeError [ERR_INVALID_ARG_VALUE]: The argument 'flags' is invalid. Received 'not-a-flag']
[RangeError [ERR_OUT_OF_RANGE]: The value of "position" is out of range. It must be an integer. Received 1.5]
[TypeError [ERR_INVALID_ARG_VALUE]: The argument 'buffer' is empty and cannot be written. Received <Buffer >]
[RangeError: The value of "fd" is out of range. It must be >= 0 && <= 2147483647. Received -1]
[RangeError [ERR_OUT_OF_RANGE]: The value of "code" is out of range. It must be an integer. Received 1.5]
[SystemError [ERR_FS_EISDIR]: Path is a directory: rm returned EISDIR (is a directory) .]
[TypeError [ERR_INVALID_ARG_TYPE]: The "time" argument must be an instance of Date or an Time in seconds. Received type number (Infinity)]
[TypeError [ERR_INVALID_ARG_VALUE]: The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'a\x00b']
[TypeError [ERR_INVALID_ARG_VALUE]: The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'a\x00b']
[TypeError [ERR_INVALID_ARG_VALUE]: ]
[ [ERR_INVALID_ARG_VALUE]: The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'a\x00b']
[Error: ENOENT: no such file or directory, open '/adamic-coded-host-error-missing/file']
[Error: ENOENT: no such file or directory, scandir '/adamic-coded-host-error-missing']
`

// ERR_STRING_TOO_LONG is a coded error Node prints plain, `Error: Cannot create a string longer than
// 0x1fffffe8 characters`, so its raise keeps the errno shape. readFileSync of a sparse file exactly that
// long reaches it on Node and natively, with no program to lower, since a fixture would need the file.
// The coded mutant leaves this row green; one marking the site coded turns it red (#z28avqd).
func TestStringTooLongErrorIsPlain(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "long")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(path, 0x1fffffe8); err != nil {
		t.Fatal(err)
	}
	const source = `#include "adamic.h"
#include <stdio.h>
#include <string.h>
static adamic_string read_flag = ADAMIC_STRING("r");
int main(int argc, char **argv) {
    if (argc != 2) { return 2; }
    adamic_string *path = adamic_decode_utf8((const unsigned char *)argv[1], strlen(argv[1]));
    adamic_string *read = adamic_fs_file_read_file(path, &read_flag);
    adamic_release(path);
    if (read != NULL) { adamic_release(read); printf("no throw\n"); return 0; }
    adamic_object *thrown = adamic_thrown;
    adamic_thrown = NULL;
    adamic_string *text = adamic_error_to_string(thrown);
    printf("[%.*s]\n", (int)text->length, text->bytes);
    adamic_release(text);
    adamic_release(thrown);
    return 0;
}
`
	const oracle = `try { require('node:fs').readFileSync(process.argv[1], 'utf8'); console.log('no throw'); } catch (error) { console.log('[' + String(error) + ']'); }`
	want := runWithInput(t, "", "node", "--eval", oracle, path)
	if want != "[Error: Cannot create a string longer than 0x1fffffe8 characters]\n" {
		t.Fatalf("Node changed: %q", want)
	}
	binary := filepath.Join(t.TempDir(), "string-too-long")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary, path); got != want {
		t.Fatalf("native %q; Node %q", got, want)
	}
}
