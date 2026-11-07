package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestWASILegacyFileCwdNodeMutant(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	directory := test.TempDir()
	if writeError := os.WriteFile(filepath.Join(directory, "marker"), []byte("present"), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	path := filepath.Join(directory, "main.a")
	if writeError := os.WriteFile(
		path,
		[]byte(`import {readTextFile} from 'adamic'; console.log(readTextFile('marker').kind);`),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	truth := onNodeWith(test, inputRun{directory: directory}, path)
	code := `#include "adamic.h"
#include <stdio.h>
int main(void) {
 adamic_start(0,NULL);
 adamic_string path=ADAMIC_STRING("marker");
 adamic_object *result=adamic_read_text_file(&path);
 adamic_string *kind=result->slots[0].reference;
 printf("%.*s\n",(int)kind->length,kind->bytes);
 adamic_release(result);
 return 0;
}
`
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	observe := func(source string) run {
		binary := filepath.Join(test.TempDir(), "legacy.wasm")
		if buildError := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); buildError != nil {
			test.Fatal(buildError)
		}
		return executeInput(
			test,
			inputRun{directory: directory},
			nil,
			"node",
			"--disable-warning=ExperimentalWarning",
			runner,
			binary,
		)
	}
	control := observe(code)
	if difference := disagreement(truth, control); difference != "" {
		test.Fatalf("%s: Node %+v, WASI %+v", difference, truth, control)
	}
	data, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime/input.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	implementation := string(data)
	const anchor = "adamic_node_wasi_cwd();"
	if strings.Count(implementation, anchor) != 1 {
		test.Fatal("legacy cwd mutant anchor changed")
	}
	implementation = strings.Replace(implementation, anchor, "(void)0;", 1)
	for _, name := range []string{
		"adamic_decode_utf8",
		"adamic_arguments_save",
		"adamic_program_arguments",
		"adamic_read_text_file",
		"adamic_write_text_file",
		"adamic_path_bytes",
	} {
		implementation = strings.ReplaceAll(implementation, name, "mutant_"+name)
		code = strings.ReplaceAll(code, name, "mutant_"+name)
	}
	bad := observe(implementation + "\n" + code)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		test.Fatalf("missing cwd must fail only Node stdout comparison: %+v", bad)
	}
	test.Log("omitted legacy cwd initializer compiled and ran cleanly; only Node stdout caught it")
}
