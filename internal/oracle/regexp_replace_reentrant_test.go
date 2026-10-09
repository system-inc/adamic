package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const reentrantRegexFixture = "internal/oracle/testdata/regexp_replace/reentrant.a"

func TestRegExpReplacementReentrant(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, reentrantRegexFixture))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if expected.exitCode != 0 || len(expected.stderr) != 0 {
		t.Fatalf("Node witness failed: %+v", expected)
	}
	for backend, actual := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": func() run {
		r, b := natively(t, program)
		if failure := leaks(t, program, b); failure != "" {
			t.Fatal(failure)
		}
		return r
	}()} {
		if d := disagreement(expected, actual); d != "" {
			t.Fatalf("%s: %s %+v", backend, d, actual)
		}
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if d := disagreement(expected, onWASI(t, native.C(program))); d != "" {
			t.Fatal("WASI: " + d)
		}
	}
}

func TestRegExpReplacementReentrantNodeMutant(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, reentrantRegexFixture))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	if !strings.Contains(source, "adamic_regex_test(") {
		t.Fatal("nested test call missing")
	}
	changed := `#include "adamic.h"
static bool reentrant_test_mutant(adamic_object *regex,adamic_string *input) { (void)regex;(void)input;return false; }
` + strings.ReplaceAll(source, "adamic_regex_test(", "reentrant_test_mutant(")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	expected := onNode(t, path)
	if actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(expected, actual) != "stdout differs" {
		t.Fatalf("reentrant mutant not caught only by Node: %+v", actual)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		wasi := onWASI(t, changed)
		if wasi.exitCode != 0 || len(wasi.stderr) != 0 || disagreement(expected, wasi) != "stdout differs" {
			t.Fatalf("WASI reentrant mutant survived Node: %+v", wasi)
		}
	}
	js := javascript.JavaScript(program)
	if !strings.Contains(js, ".test(") {
		t.Fatal("nested JavaScript test call missing")
	}
	js = strings.ReplaceAll(js, ".test(", ".reentrantTestMutant(")
	js = "RegExp.prototype.reentrantTestMutant = function(input) { return false; };\n" + js
	result := replacementJavaScriptMutant(t, js)
	if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(expected, result) != "stdout differs" {
		t.Fatalf("JavaScript reentrant mutant escaped Node: %+v", result)
	}
	t.Log("nested regex test returning false: clean sanitizer exit, caught only by Node stdout")
}

// Not parallel: can update the shared counts.md file when ADAMIC_UPDATE_COUNTS is set.
func TestRegExpReplacementReentrantCounts(t *testing.T) {
	row := counted(t, reentrantRegexFixture, false, nil, false, false)
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, row+"\n") {
		return
	}
	if !*updateCounts {
		t.Fatalf("new reentrant fixture counts missing: %s", row)
	}
	at := strings.Index(text, "| internal/oracle/testdata/regexp_replace/move_effect.a |")
	if at < 0 {
		t.Fatal("replacement fixture boundary missing")
	}
	at += strings.Index(text[at:], "\n") + 1
	if err := os.WriteFile(countsPath, []byte(text[:at]+row+"\n"+text[at:]), 0644); err != nil {
		t.Fatal(err)
	}
}
