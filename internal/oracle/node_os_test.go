package oracle

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"eol", "platform", "homedir", "tmpdir"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/node_os_" + name + ".a", true, false})
	}
}

func TestNodeOSAgreesWithNode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		environments [][]string
		args         []string
	}{
		{"eol", [][]string{nil}, nil},
		{"platform", [][]string{nil}, nil},
		{"tmpdir", [][]string{{"TMPDIR=", "TMP=", "TEMP="}, {"TMPDIR=/", "TMP=second", "TEMP=third"}, {"TMPDIR=//", "TMP=second", "TEMP=third"}, {"TMPDIR=/tmp///", "TMP=second", "TEMP=third"}, {"TMPDIR=relative/", "TMP=", "TEMP="}, {"TMPDIR=héllo 🌍/", "TMP=", "TEMP="}, {"TMPDIR=", "TMP=second/", "TEMP=third"}, {"TMPDIR=", "TMP=", "TEMP=third/"}, {"TMPDIR=bad\xff/", "TMP=", "TEMP="}}, nil},
		{"homedir", [][]string{{"HOME="}, {"HOME=/"}, {"HOME=relative/"}, {"HOME=héllo 🌍/"}, {"HOME=bad\xff/"}, {"HOME=" + strings.Repeat("x", 5000)}}, nil},
		{"homedir", [][]string{nil}, []string{"unset-home"}},
	}
	for _, one := range cases {
		t.Run(one.name+strings.Join(one.args, ""), func(t *testing.T) {
			path, binary, script := sanitized(t, "internal/oracle/testdata/node_os_"+one.name+".a")
			for _, env := range one.environments {
				truth := executeWith(t, env, "node", append([]string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path}, one.args...)...)
				if truth.exitCode != 0 || len(truth.stderr) != 0 {
					t.Fatalf("Node failed: %+v", truth)
				}
				for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script}} {
					got := executeWith(t, env, command[0], append(command[1:], one.args...)...)
					if diff := disagreement(truth, got); diff != "" {
						t.Fatalf("%s: Node %q; backend %q %q", diff, truth.stdout, got.stdout, got.stderr)
					}
				}
			}
		})
	}
}

func TestNodeOSMutants(t *testing.T) {
	t.Parallel()
	for _, one := range []struct {
		name, function, helper string
		env                    []string
	}{
		{"eol", "adamic_node_eol", `static adamic_string *os_mutant(void) {static adamic_string wrong=ADAMIC_STRING("wrong");return &wrong;}`, nil},
		{"platform", "adamic_node_os_platform", `static adamic_string *os_mutant(void) {static adamic_string wrong=ADAMIC_STRING("wrong");return &wrong;}`, nil},
		{"homedir", "adamic_node_homedir", `static adamic_string *os_mutant(void) {static adamic_string wrong=ADAMIC_STRING("wrong");return &wrong;}`, []string{"HOME=relative/"}},
		{"tmpdir", "adamic_node_tmpdir", `static adamic_string *os_mutant(void) {adamic_string *value=adamic_node_tmpdir();static adamic_string suffix=ADAMIC_STRING("/");adamic_string *result=adamic_string_concat(2,(adamic_string *const[]){value,&suffix});adamic_release(value);return result;}`, []string{"TMPDIR=relative/", "TMP=", "TEMP="}},
	} {
		t.Run(one.name, func(t *testing.T) {
			path := filepath.Join(repository, "internal/oracle/testdata/node_os_"+one.name+".a")
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			changed := strings.ReplaceAll(original, one.function+"(", "os_mutant(")
			if changed == original {
				t.Fatal("mutant changed nothing")
			}
			changed = strings.Replace(changed, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+one.helper, 1)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			env := append([]string{}, one.env...)
			env = append(env, "UBSAN_OPTIONS=halt_on_error=1")
			if runtime.GOOS == "linux" {
				env = append(env, "ASAN_OPTIONS=detect_leaks=1")
			}
			truth := executeWith(t, env, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			got := executeWith(t, env, binary)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
				t.Fatalf("mutant not caught only by Node stdout: Node %+v mutant %+v", truth, got)
			}
			t.Log("caught only by Node stdout; sanitizer and leak checks clean")
		})
	}
}

func TestNodeOSWASI(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, name := range []string{"eol", "platform", "homedir", "tmpdir"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(repository, "internal/oracle/testdata/node_os_"+name+".a")
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "program.wasm")
			err = native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"})
			if name == "platform" || name == "homedir" {
				if err == nil || !strings.Contains(err.Error(), "wasm32-wasi refuses") {
					t.Fatalf("expected compile-time target refusal, got %v", err)
				}
				t.Log(err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			env := []string{"TMPDIR=relative/", "TMP=", "TEMP="}
			truth := executeWith(t, env, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			actual := executeWith(t, env, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
			if diff := disagreement(truth, actual); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestNodeOSHomeUncaught(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/node_os_homedir.a")
	env := []string{"HOME=" + strings.Repeat("x", 5000)}
	truth := executeWith(t, env, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path, "uncaught-home")
	if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "SystemError [ERR_SYSTEM_ERROR]") {
		t.Fatalf("unexpected Node error: %+v", truth)
	}
	for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script}} {
		got := executeWith(t, env, command[0], append(command[1:], "uncaught-home")...)
		if diff := disagreement(truth, got); diff != "" {
			t.Fatalf("%s: Node %+v backend %+v", diff, truth, got)
		}
	}
}
