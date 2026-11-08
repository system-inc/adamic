package native

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var nodeWasiCwdBaseline = []string{
	"args_get",
	"args_sizes_get",
	"clock_time_get",
	"fd_close",
	"fd_fdstat_get",
	"fd_prestat_dir_name",
	"fd_prestat_get",
	"fd_seek",
	"fd_write",
	"proc_exit",
}

func nodeWasiCwdEnabled(test *testing.T) {
	test.Helper()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" || os.Getenv("WASI_SYSROOT") == "" {
		test.Skip("set ADAMIC_TEST_WASI=1 and WASI_SYSROOT")
	}
}

func nodeWasiCwdRuntime(test *testing.T, host bool) {
	test.Helper()
	nodeWasiCwdEnabled(test)
	source := ""
	if host {
		source = "#define ADAMIC_NODE_HOST 1\n"
	}
	if _, runtimeLibraryForSourceError := RuntimeLibraryForSource(
		"",
		source,
		Options{Target: "wasm32-wasi"},
	); runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
}

func TestProduct_NodeWasiCwdRuntime_000(test *testing.T) {
	test.Parallel()
	nodeWasiCwdRuntime(test, false)
}
func TestProduct_NodeWasiCwdRuntime_001(test *testing.T) {
	test.Parallel()
	nodeWasiCwdRuntime(test, true)
}

func nodeWasiCwdImports(test *testing.T, variant int) {
	test.Helper()
	nodeWasiCwdEnabled(test)
	source := `#include "adamic.h"
#include "node_cwd_wasi.h"
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 static adamic_string message = ADAMIC_STRING("hello");
 adamic_write_line(adamic_stdout, &message);
 return 0;
}
`
	if variant == 1 {
		source = strings.Replace(
			source,
			"adamic_write_line(adamic_stdout, &message);",
			`static adamic_string path = ADAMIC_STRING("/one/two");
 adamic_string *name = adamic_node_path_basename(&path, NULL);
 adamic_write_line(adamic_stdout, name);
 adamic_release(name);`,
			1,
		)
	}
	if variant == 2 {
		source = strings.Replace(
			source,
			"adamic_start(argc, argv);",
			"adamic_start(argc, argv);\n adamic_node_wasi_cwd();",
			1,
		)
	}
	module := filepath.Join(test.TempDir(), "imports.wasm")
	if buildError := Build(source, module, Options{Target: "wasm32-wasi"}); buildError != nil {
		test.Fatal(buildError)
	}
	const inspect = `const fs=require('node:fs');console.log(JSON.stringify(WebAssembly.Module.imports(new WebAssembly.Module(fs.readFileSync(process.argv[1]))).map(i=>i.module==='wasi_snapshot_preview1'?i.name:i.module+'.'+i.name).sort()));`
	got := observeWasi(test, "", "node", "-e", inspect, module)
	if got.exit != 0 || len(got.stderr) != 0 {
		test.Fatalf("import inspection: %+v", got)
	}
	var imports []string
	if decodeError := json.Unmarshal(got.stdout, &imports); decodeError != nil {
		test.Fatal(decodeError)
	}
	same := slices.Equal(imports, nodeWasiCwdBaseline)
	if variant == 2 {
		if same {
			test.Fatal("eager cwd initialization survived the import comparison")
		}
		for _, name := range []string{"environ_get", "environ_sizes_get", "path_filestat_get"} {
			if !slices.Contains(imports, name) {
				test.Fatalf("startup mutant did not add %s: %v", name, imports)
			}
		}
		test.Log("eager initializer compiled; Node's module import inspection caught it")
	} else if !same {
		test.Fatalf("file-free imports: got %v, want %v", imports, nodeWasiCwdBaseline)
	}
}

func TestNodeWasiCwdFileFreeImports(test *testing.T) { test.Parallel(); nodeWasiCwdImports(test, 0) }
func TestNodeWasiCwdLexicalImports(test *testing.T)  { test.Parallel(); nodeWasiCwdImports(test, 1) }
func TestNodeWasiCwdStartupMutant(test *testing.T)   { test.Parallel(); nodeWasiCwdImports(test, 2) }

func nodeWasiCwdAgreement(test *testing.T, variant int) {
	test.Helper()
	nodeWasiCwdEnabled(test)
	directory := test.TempDir()
	first := filepath.Join(directory, "first")
	moved := filepath.Join(first, "nested")
	if createDirectoryError := os.MkdirAll(moved, 0700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	for _, path := range []string{filepath.Join(first, "marker"), filepath.Join(moved, "marker")} {
		if writeError := os.WriteFile(path, []byte("marker"), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	truthSource := `const fs=require('node:fs'),path=require('node:path');`
	if variant == 1 {
		truthSource += `console.log(fs.readdirSync('.').length);`
	}
	truthSource += `console.log(path.resolve('.'));process.chdir(process.argv[1]);console.log(path.resolve('.'));console.log(fs.readdirSync('.').length);`
	truth := observeWasi(test, first, "node", "-e", truthSource, moved)
	if truth.exit != 0 || len(truth.stderr) != 0 {
		test.Fatalf("Node cwd contract: %+v", truth)
	}
	helper := ""
	initializer := "adamic_node_wasi_cwd"
	if variant >= 3 {
		data, readFileError := runtime.ReadFile("runtime/node_cwd_wasi.c")
		if readFileError != nil {
			test.Fatal(readFileError)
		}
		helper = string(data)
		before, after := `const char *directory = getenv("ADAMIC_WASI_CWD");`, `const char *directory = NULL;`
		if variant == 4 {
			before, after = "if (initialized) { return; }", "if (false) { return; }"
		}
		if strings.Count(helper, before) != 1 {
			test.Fatal("cwd mutation needs one target")
		}
		helper = strings.Replace(helper, before, after, 1)
		initializer += "_mutant"
		helper = strings.ReplaceAll(helper, "adamic_node_wasi_cwd", initializer)
	}
	source := `#define _XOPEN_SOURCE 700
#define _POSIX_C_SOURCE 200809L
#define ADAMIC_NODE_HOST 1
%HELPER%
#include "adamic.h"
#include "node_cwd_wasi.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dirent.h>
#include <unistd.h>
static adamic_string dot = ADAMIC_STRING(".");
static adamic_string moved = ADAMIC_STRING(%MOVED%);
static void show_path(void) {
 %SHOW%
}
static void show_entries(void) {
 %ENTRIES%
}
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 %FIRST%
 show_path();
 %CHDIR%
 show_path();
 show_entries();
 return 0;
}
`
	show := `adamic_string *path = adamic_node_path_resolve(1, (adamic_string *const[]){&dot});
 if (path == NULL) { exit(2); }
 printf("%.*s\n", (int)path->length, path->bytes);
 adamic_release(path);`
	entries := `adamic_array *listed = adamic_node_fs_readdir(&dot, NULL);
 if (listed == NULL) { exit(3); }
 printf("%zu\n", listed->length);
 adamic_release(listed);`
	change := `adamic_fs_file_host_chdir(&moved);`
	if variant >= 2 {
		show = initializer + `(); char *path = getcwd(NULL,0); if (path == NULL) { exit(2); } puts(path); free(path);`
		entries = initializer +
			`(); DIR *dir = opendir("."); if (dir == NULL) { exit(3); } size_t count = 0; struct dirent *entry;
 while ((entry = readdir(dir)) != NULL) { if (strcmp(entry->d_name,".") != 0 && strcmp(entry->d_name,"..") != 0) { count++; } }
 closedir(dir); printf("%zu\n",count);`
		change = `if (chdir(moved.bytes) != 0) { return 4; }`
	}
	firstCall := ""
	if variant == 1 {
		firstCall = "show_entries();"
	}
	source = strings.NewReplacer(
		"%HELPER%",
		helper,
		"%MOVED%",
		cString(moved),
		"%SHOW%",
		show,
		"%ENTRIES%",
		entries,
		"%CHDIR%",
		change,
		"%FIRST%",
		firstCall,
	).Replace(source)
	module := filepath.Join(directory, "cwd.wasm")
	if buildError := Build(source, module, Options{Target: "wasm32-wasi"}); buildError != nil {
		test.Fatal(buildError)
	}
	runner, absolutePathError := filepath.Abs("../../oracle/wasi.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	got := observeWasi(test, first, "node", "--disable-warning=ExperimentalWarning", runner, module)
	if got.exit != 0 || len(got.stderr) != 0 {
		test.Fatalf("cwd check failed outside Node stdout: %+v", got)
	}
	if variant >= 3 {
		if string(got.stdout) == string(truth.stdout) {
			test.Fatal("wrong cwd initializer survived Node stdout")
		}
		test.Logf("cwd mutant %d compiled and ran cleanly; only Node stdout caught it", variant)
	} else if string(got.stdout) != string(truth.stdout) {
		test.Fatal(fmt.Sprintf("cwd: got %q want Node %q", got.stdout, truth.stdout))
	}
}

func TestNodeWasiCwdPathFirst(test *testing.T)      { test.Parallel(); nodeWasiCwdAgreement(test, 0) }
func TestNodeWasiCwdDirectoryFirst(test *testing.T) { test.Parallel(); nodeWasiCwdAgreement(test, 1) }
func TestNodeWasiCwdSharedOnce(test *testing.T)     { test.Parallel(); nodeWasiCwdAgreement(test, 2) }
func TestNodeWasiCwdMissingInitMutant(test *testing.T) {
	test.Parallel()
	nodeWasiCwdAgreement(test, 3)
}
func TestNodeWasiCwdRepeatedInitMutant(test *testing.T) {
	test.Parallel()
	nodeWasiCwdAgreement(test, 4)
}
