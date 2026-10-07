package oracle

import (
	"bytes"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

// Reuse the existing input harness so source, JavaScript, and native see the same directories.
func init() {
	inputFixtures = append(inputFixtures, struct {
		path       string
		arguments  []string
		unreadable bool
		writes     bool
	}{"internal/oracle/testdata/node_fs_directory_symlink.a", nil, false, true})
	for _, path := range []string{
		"internal/oracle/testdata/node_path_posix.a",
		"internal/oracle/testdata/node_path_basename.a",
		"internal/oracle/testdata/node_path_relative.a",
		"internal/oracle/testdata/node_fs_directory_entries.a",
		"internal/oracle/testdata/node_fs_directory_realpath.a",
		"internal/oracle/testdata/node_fs_directory_system.a",
		"internal/oracle/testdata/node_fs_directory_stat_options.a",
		"internal/oracle/testdata/node_fs_directory_permissions.a",
	} {
		inputFixtures = append(inputFixtures, struct {
			path       string
			arguments  []string
			unreadable bool
			writes     bool
		}{path, nil, false, false})
	}
}

// Mutants run cleanly under sanitizers and leak checks. Only the source on Node catches them.
func TestNodeFSDirectoryMutants(t *testing.T) {
	cases := []struct{ name, fixture, from, to, helper, function string }{
		{name: "resolve", fixture: "node_path_posix.a", from: "adamic_node_path_resolve(", to: "adamic_node_path_join("},
		{name: "join", fixture: "node_path_posix.a", from: "adamic_node_path_join(", to: "adamic_node_path_resolve("},
		{name: "dirname", fixture: "node_path_posix.a", from: "adamic_node_path_dirname(", to: "mutant_dirname(", helper: `static adamic_string *mutant_dirname(const adamic_string *path) { return adamic_node_path_join(1, (adamic_string *const[]){(adamic_string *)path}); }`},
		{name: "readdir order", fixture: "node_fs_directory_entries.a", from: "adamic_node_fs_readdir(", to: "mutant_readdir(", helper: `static adamic_array *mutant_readdir(const adamic_string *path, const adamic_object *options) { adamic_array *result=adamic_node_fs_readdir(path, options); if(result!=NULL) adamic_array_reverse(result); return result; }`},
		{name: "Dirent name", fixture: "node_fs_directory_entries.a", from: "adamic_node_fs_readdir(", to: "mutant_readdir(", helper: `static adamic_array *mutant_readdir(const adamic_string *path, const adamic_object *options) { adamic_array *result=adamic_node_fs_readdir(path, options); if(result!=NULL && options!=NULL) { for(size_t i=0;i<result->length;i++) { adamic_object *entry=result->elements[i].reference; static adamic_string name=ADAMIC_STRING("wrong"); adamic_release(entry->slots[0].reference); entry->slots[0].reference=&name; } } return result; }`},
		{name: "isFile", fixture: "node_fs_directory_entries.a", from: `"isFile"`, to: `"isDirectory"`},
		{name: "isDirectory", fixture: "node_fs_directory_entries.a", from: `"isDirectory"`, to: `"isFile"`},
		{name: "isSymbolicLink", fixture: "node_fs_directory_entries.a", from: `"isSymbolicLink"`, to: `"isFile"`},
		{name: "realpath walk", fixture: "node_fs_directory_realpath.a", from: ", false)", to: ", true)"},
		{name: "realpath native", fixture: "node_fs_directory_realpath.a", from: ", true)", to: ", false)"},
		{name: "error code", fixture: "node_fs_directory_entries.a", from: "adamic_node_fs_readdir(", to: "mutant_readdir(", helper: `static adamic_array *mutant_readdir(const adamic_string *path, const adamic_object *options) { adamic_array *result=adamic_node_fs_readdir(path, options); if(adamic_thrown!=NULL) { static adamic_string code=ADAMIC_STRING("WRONG"); adamic_release(adamic_thrown->slots[2].reference); adamic_thrown->slots[2].reference=&code; } return result; }`},
		{name: "error message", fixture: "node_fs_directory_entries.a", from: "adamic_node_fs_readdir(", to: "mutant_readdir(", helper: `static adamic_array *mutant_readdir(const adamic_string *path, const adamic_object *options) { adamic_array *result=adamic_node_fs_readdir(path, options); if(adamic_thrown!=NULL) { static adamic_string message=ADAMIC_STRING("wrong"); adamic_release(adamic_thrown->slots[1].reference); adamic_thrown->slots[1].reference=&message; } return result; }`},
	}
	for _, function := range []string{"getAccessibleFileSystemEntries", "getDirectories", "directoryExists", "realpath", "resolvePath", "isFileSystemCaseSensitive", "readDirectory"} {
		cases = append(cases, struct{ name, fixture, from, to, helper, function string }{name: function, fixture: "node_fs_directory_system.a", function: function})
	}
	for _, one := range cases {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", one.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			if one.function == "readDirectory" {
				for index := range program.Functions {
					function := &program.Functions[index]
					for at, statement := range function.Body {
						returned, ok := statement.(ir.Return)
						if !ok {
							continue
						}
						call, ok := returned.Value.(ir.CallClosure)
						if !ok || len(call.Arguments) != 9 {
							continue
						}
						call.Arguments[1], call.Arguments[2] = call.Arguments[2], call.Arguments[1]
						returned.Value = call
						function.Body[at] = returned
						changed = true
					}
				}
			} else if one.function != "" {
				for index := range program.Functions {
					function := &program.Functions[index]
					if function.Name != one.function {
						continue
					}
					var value ir.Expression
					switch one.function {
					case "getAccessibleFileSystemEntries":
						value = ir.ObjectLiteral{Fields: []ir.Field{{Name: "files", Value: ir.ArrayLiteral{Element: ir.String}}, {Name: "directories", Value: ir.ArrayLiteral{Element: ir.String}}}}
					case "getDirectories":
						value = ir.ArrayLiteral{Element: ir.String}
					case "directoryExists", "isFileSystemCaseSensitive":
						value = ir.BooleanConstant{Value: false}
					default:
						value = ir.Read{Local: function.Parameters[0], Of: ir.String}
					}
					function.Body = []ir.Statement{ir.Return{Value: value}}
					changed = true
				}
			}
			source := native.C(program)
			if one.from != "" {
				original := source
				if strings.HasPrefix(one.name, "realpath ") {
					pattern := regexp.MustCompile(`adamic_node_fs_realpath\(([^,\n]+)` + regexp.QuoteMeta(one.from))
					source = pattern.ReplaceAllString(source, "adamic_node_fs_realpath(${1}"+one.to)
				} else {
					source = strings.ReplaceAll(source, one.from, one.to)
				}
				changed = source != original
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			if one.helper != "" {
				source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+one.helper, 1)
			}
			shared := sharedDirectory(t)
			binary := filepath.Join(shared, "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			how := inputRun{directory: filepath.Dir(path)}
			// LeakSanitizer is Linux's: macOS's AddressSanitizer aborts when asked for it.
			var environment []string
			if runtime.GOOS == "linux" {
				environment = []string{"ASAN_OPTIONS=detect_leaks=1"}
			}
			got := executeInput(t, how, environment, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside comparison: exit %d stderr %s", got.exitCode, got.stderr)
			}
			if difference := disagreement(onNodeWith(t, how, path), got); difference != "stdout differs" {
				t.Fatalf("mutant caught by %q, want stdout differs", difference)
			}
			t.Log("caught only by stdout comparison with Node")
		})
	}
}

func TestNodeFSDirectoryPermissions(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_directory_permissions.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	blocked := filepath.Join(shared, "blocked")
	if err := os.Mkdir(blocked, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blocked, "child"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(blocked, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(blocked, 0700) })
	how := inputRun{directory: filepath.Dir(path), arguments: []string{blocked}}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	truth := onNodeWith(t, how, path)
	if !bytes.Contains(truth.stdout, []byte("EACCES: permission denied")) {
		t.Fatalf("permission probe did not fail on Node: %s", truth.stdout)
	}
	backend := inputBackend(t, how, program, shared)
	got, binary := inputNatively(t, how, program, shared)
	for name, result := range map[string]run{"native": got, "JavaScript": backend} {
		if difference := disagreement(truth, result); difference != "" {
			t.Errorf("%s: %s\nNode %s\ngot %s", name, difference, truth.stdout, result.stdout)
		}
	}
	if leaked := inputLeaks(t, func() inputRun { return how }, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}
