package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
	"testing"
)

var host29PathMembers = []string{"resolve", "dirname", "join", "normalize", "relative", "basename", "extname", "isabsolute", "sep", "aliases", "tsc"}

func init() {
	for _, member := range host29PathMembers {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/node_path_host29_" + member + ".a", true, false})
	}
}

// Every mutant still compiles and finishes with clean sanitizers and leaks.
// Only comparing its observable result with unmodified Node detects the error.
func TestNodePathHost29Mutants(t *testing.T) {
	for _, one := range []struct{ member, signature, body string }{
		{"resolve", "size_t count, adamic_string *const paths[]", "return adamic_node_path_join(count,paths);"},
		{"join", "size_t count, adamic_string *const paths[]", "return adamic_node_path_resolve(count,paths);"},
		{"dirname", "const adamic_string *path", "return adamic_node_path_basename(path,NULL);"},
		{"normalize", "const adamic_string *path", "return adamic_node_path_basename(path,NULL);"},
		{"relative", "const adamic_string *from,const adamic_string *to", "return adamic_node_path_join(2,(adamic_string *const[]){(adamic_string*)from,(adamic_string*)to});"},
		{"basename", "const adamic_string *path,const adamic_string *suffix", "return adamic_node_path_dirname(path);"},
		{"extname", "const adamic_string *path", "return adamic_node_path_basename(path,NULL);"},
		{"extname", "const adamic_string *path", `adamic_string *value=adamic_node_path_extname(path); static adamic_string high=ADAMIC_STRING("\xed\xa0\x80"), low=ADAMIC_STRING("\xed\xbf\xbf"); adamic_string *result=adamic_string_replace(value,&high,&low,false); adamic_release(value); return result;`},
		{"isAbsolute", "const adamic_string *path", "return !adamic_node_path_isAbsolute(path);"},
		{"sep", "", ""},
	} {
		t.Run(one.member, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_path_host29_"+strings.ToLower(one.member)+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node failed: %+v", truth)
			}
			original := native.C(program)
			var source string
			if one.member == "sep" {
				changed := false
				for i, value := range program.Strings {
					if value == "/" {
						program.Strings[i] = "\\"
						changed = true
					}
				}
				if !changed {
					t.Fatal("separator mutant changed nothing")
				}
				source = native.C(program)
			} else {
				source = strings.ReplaceAll(original, "adamic_node_path_"+one.member+"(", "mutant_path(")
				if source == original {
					t.Fatal("mutant changed nothing")
				}
				returns := "adamic_string *"
				if one.member == "isAbsolute" {
					returns = "bool "
				}
				helper := "\nstatic " + returns + "mutant_path(" + one.signature + "){" + one.body + "}\n"
				source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+helper, 1)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, nodeFSDirectorySanitizerEnvironment(), binary)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside comparison: exit %d stderr %s", got.exitCode, got.stderr)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				t.Fatalf("mutant caught by %q", difference)
			}
			t.Log("caught only by Node stdout comparison; sanitizers and leaks clean")
		})
	}
}
