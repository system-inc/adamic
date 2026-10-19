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

var nodeWASICwdBaseline = []string{"args_get", "args_sizes_get", "clock_time_get", "fd_close", "fd_fdstat_get", "fd_prestat_dir_name", "fd_prestat_get", "fd_seek", "fd_write", "proc_exit"}

func nodeWASICwdEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" || os.Getenv("WASI_SYSROOT") == "" {
		t.Skip("set ADAMIC_TEST_WASI=1 and WASI_SYSROOT")
	}
}

func nodeWASICwdRuntime(t *testing.T, host bool) {
	t.Helper()
	nodeWASICwdEnabled(t)
	source := ""
	if host {
		source = "#define ADAMIC_NODE_HOST 1\n"
	}
	if _, err := RuntimeLibraryForSource("", source, Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
}

func TestProduct_NodeWASICwdRuntime_000(t *testing.T) { t.Parallel(); nodeWASICwdRuntime(t, false) }
func TestProduct_NodeWASICwdRuntime_001(t *testing.T) { t.Parallel(); nodeWASICwdRuntime(t, true) }

func nodeWASICwdImports(t *testing.T, variant int) {
	t.Helper()
	nodeWASICwdEnabled(t)
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
		source = strings.Replace(source, "adamic_write_line(adamic_stdout, &message);", `static adamic_string path = ADAMIC_STRING("/one/two");
 adamic_string *name = adamic_node_path_basename(&path, NULL);
 adamic_write_line(adamic_stdout, name);
 adamic_release(name);`, 1)
	}
	if variant == 2 {
		source = strings.Replace(source, "adamic_start(argc, argv);", "adamic_start(argc, argv);\n adamic_node_wasi_cwd();", 1)
	}
	module := filepath.Join(t.TempDir(), "imports.wasm")
	if err := Build(source, module, Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	const inspect = `const fs=require('node:fs');console.log(JSON.stringify(WebAssembly.Module.imports(new WebAssembly.Module(fs.readFileSync(process.argv[1]))).map(i=>i.module==='wasi_snapshot_preview1'?i.name:i.module+'.'+i.name).sort()));`
	got := observeWASI(t, "", "node", "-e", inspect, module)
	if got.exit != 0 || len(got.stderr) != 0 {
		t.Fatalf("import inspection: %+v", got)
	}
	var imports []string
	if err := json.Unmarshal(got.stdout, &imports); err != nil {
		t.Fatal(err)
	}
	same := slices.Equal(imports, nodeWASICwdBaseline)
	if variant == 2 {
		if same {
			t.Fatal("eager cwd initialization survived the import comparison")
		}
		for _, name := range []string{"environ_get", "environ_sizes_get", "path_filestat_get"} {
			if !slices.Contains(imports, name) {
				t.Fatalf("startup mutant did not add %s: %v", name, imports)
			}
		}
		t.Log("eager initializer compiled; Node's module import inspection caught it")
	} else if !same {
		t.Fatalf("file-free imports: got %v, want %v", imports, nodeWASICwdBaseline)
	}
}

func TestNodeWASICwdFileFreeImports(t *testing.T) { t.Parallel(); nodeWASICwdImports(t, 0) }
func TestNodeWASICwdLexicalImports(t *testing.T)  { t.Parallel(); nodeWASICwdImports(t, 1) }
func TestNodeWASICwdStartupMutant(t *testing.T)   { t.Parallel(); nodeWASICwdImports(t, 2) }

func nodeWASICwdAgreement(t *testing.T, variant int) {
	t.Helper()
	nodeWASICwdEnabled(t)
	directory := t.TempDir()
	first := filepath.Join(directory, "first")
	moved := filepath.Join(first, "nested")
	if err := os.MkdirAll(moved, 0700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(first, "marker"), filepath.Join(moved, "marker")} {
		if err := os.WriteFile(path, []byte("marker"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	truthSource := `const fs=require('node:fs'),path=require('node:path');`
	if variant == 1 {
		truthSource += `console.log(fs.readdirSync('.').length);`
	}
	truthSource += `console.log(path.resolve('.'));process.chdir(process.argv[1]);console.log(path.resolve('.'));console.log(fs.readdirSync('.').length);`
	truth := observeWASI(t, first, "node", "-e", truthSource, moved)
	if truth.exit != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node cwd contract: %+v", truth)
	}
	helper := ""
	initializer := "adamic_node_wasi_cwd"
	if variant >= 3 {
		data, err := runtime.ReadFile("runtime/node_cwd_wasi.c")
		if err != nil {
			t.Fatal(err)
		}
		helper = string(data)
		before, after := `const char *directory = getenv("ADAMIC_WASI_CWD");`, `const char *directory = NULL;`
		if variant == 4 {
			before, after = "if (initialized) { return; }", "if (false) { return; }"
		}
		if strings.Count(helper, before) != 1 {
			t.Fatal("cwd mutation needs one target")
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
		entries = initializer + `(); DIR *dir = opendir("."); if (dir == NULL) { exit(3); } size_t count = 0; struct dirent *entry;
 while ((entry = readdir(dir)) != NULL) { if (strcmp(entry->d_name,".") != 0 && strcmp(entry->d_name,"..") != 0) { count++; } }
 closedir(dir); printf("%zu\n",count);`
		change = `if (chdir(moved.bytes) != 0) { return 4; }`
	}
	firstCall := ""
	if variant == 1 {
		firstCall = "show_entries();"
	}
	source = strings.NewReplacer("%HELPER%", helper, "%MOVED%", cString(moved), "%SHOW%", show, "%ENTRIES%", entries, "%CHDIR%", change, "%FIRST%", firstCall).Replace(source)
	module := filepath.Join(directory, "cwd.wasm")
	if err := Build(source, module, Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/wasi.mjs")
	if err != nil {
		t.Fatal(err)
	}
	got := observeWASI(t, first, "node", "--disable-warning=ExperimentalWarning", runner, module)
	if got.exit != 0 || len(got.stderr) != 0 {
		t.Fatalf("cwd check failed outside Node stdout: %+v", got)
	}
	if variant >= 3 {
		if string(got.stdout) == string(truth.stdout) {
			t.Fatal("wrong cwd initializer survived Node stdout")
		}
		t.Logf("cwd mutant %d compiled and ran cleanly; only Node stdout caught it", variant)
	} else if string(got.stdout) != string(truth.stdout) {
		t.Fatal(fmt.Sprintf("cwd: got %q want Node %q", got.stdout, truth.stdout))
	}
}

func TestNodeWASICwdPathFirst(t *testing.T)          { t.Parallel(); nodeWASICwdAgreement(t, 0) }
func TestNodeWASICwdDirectoryFirst(t *testing.T)     { t.Parallel(); nodeWASICwdAgreement(t, 1) }
func TestNodeWASICwdSharedOnce(t *testing.T)         { t.Parallel(); nodeWASICwdAgreement(t, 2) }
func TestNodeWASICwdMissingInitMutant(t *testing.T)  { t.Parallel(); nodeWASICwdAgreement(t, 3) }
func TestNodeWASICwdRepeatedInitMutant(t *testing.T) { t.Parallel(); nodeWASICwdAgreement(t, 4) }
