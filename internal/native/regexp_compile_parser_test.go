package native

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	reference "github.com/system-inc/adamic/internal/regexp"
)

// The oracle process evaluates the entire corpus once; each C parse owns and
// releases its tree under ASan and UBSan, including rejected parses.
func TestRegExpRuntimeParserAcceptance(t *testing.T) { runtimeParserAcceptance(t, false) }
func TestRegExpRuntimeParserWASI(t *testing.T)       { runtimeParserAcceptance(t, true) }

func runtimeParserAcceptance(t *testing.T, wasi bool) {
	t.Helper()
	cases, corpus := runtimeRegexCompilerCases(t)
	cohereCount := 875
	input, _ := json.Marshal(cases)
	node := exec.Command("node", "-e", `const c=JSON.parse(require('fs').readFileSync(0,'utf8'));process.stdout.write(JSON.stringify(c.map(x=>{try{new RegExp(x[0],x[1]);return ''}catch(e){if(!(e instanceof SyntaxError))throw e;return e.message}})));`)
	node.Stdin = bytes.NewReader(input)
	output, err := node.CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v %s", err, output)
	}
	var messages []string
	if err := json.Unmarshal(output, &messages); err != nil {
		t.Fatal(err)
	}
	var source strings.Builder
	source.WriteString("#include <stdio.h>\n#include <string.h>\n#include \"regexp_compile_parser.h\"\nstruct test { const unsigned char *pattern, *flags; size_t length, flag_length; int expected; const unsigned char *reason; };\nstatic const struct test cases[] = {\n")
	emit := func(value string) string {
		var a strings.Builder
		a.WriteString("(const unsigned char[]){")
		for _, b := range []byte(value) {
			fmt.Fprintf(&a, "%d,", b)
		}
		a.WriteString("0}")
		return a.String()
	}
	disagreements := 0
	for i, c := range cases {
		_, err := reference.Parse(c[0], c[1])
		valid := err == nil
		var divergence *reference.V8DivergenceError
		refused := errors.As(err, &divergence)
		if !refused && valid != (messages[i] == "") {
			disagreements++
			t.Logf("reference versus Node %q %q: %v / %s", c[0], c[1], err, messages[i])
		}
		expected := 1
		if valid {
			expected = 0
		}
		reason := messages[i]
		if refused {
			expected = 3
			reason = err.Error()
		}

		fmt.Fprintf(&source, "{%s,%s,%d,%d,%d,%s},\n", emit(c[0]), emit(c[1]), len(c[0]), len(c[1]), expected, emit(reason))
	}
	source.WriteString("};\nint main(void) { int failures=0; for(size_t i=0;i<sizeof(cases)/sizeof(cases[0]);i++) { const struct test *c=&cases[i]; adamic_regex_parse_result r; adamic_regex_parse(c->pattern,c->length,c->flags,c->flag_length,&r); if(r.status!=c->expected || ((r.status==1 || r.status==3) && strcmp(r.message,(const char *)c->reason)!=0)) {fprintf(stderr,\"case %zu: status %d reason %s expected %s\\n\",i,r.status,r.message?r.message:\"none\",c->reason);failures++;} adamic_regex_parse_free(&r); } return failures!=0;}\n")
	directory := t.TempDir()
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	runtimeDirectory, _ := filepath.Abs("runtime")
	binary := filepath.Join(directory, "check")
	args := []string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O1", "-g", "-fsanitize=address,undefined", "-fno-omit-frame-pointer", "-DADAMIC_REGEXP_RUNTIME_COMPILER=1", "-I", runtimeDirectory, main, filepath.Join(runtimeDirectory, "regexp_compile_parser.c"), filepath.Join(runtimeDirectory, "regexp_compile_properties.c"), "-o", binary}
	compiler := "clang"
	if wasi {
		sysroot := os.Getenv("WASI_SYSROOT")
		if sysroot == "" {
			t.Fatal("parser WASI proof requires WASI_SYSROOT")
		}
		compiler = filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin", "clang")
		args = []string{"--target=wasm32-wasi", "--sysroot=" + sysroot, "-DADAMIC_TARGET_WASI=1", "-mno-atomics", "-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O1", "-DADAMIC_REGEXP_RUNTIME_COMPILER=1", "-I", runtimeDirectory, main, filepath.Join(runtimeDirectory, "regexp_compile_parser.c"), filepath.Join(runtimeDirectory, "regexp_compile_properties.c"), "-Wl,-z,stack-size=1048576", "-o", binary}
	}
	command := exec.Command(compiler, args...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, output)
	}
	command = exec.Command(binary)
	if wasi {
		command = exec.Command("node", "--no-warnings", "-e", `const fs=require('fs');const {WASI}=require('node:wasi');const wasi=new WASI({version:'preview1',args:[],env:{},returnOnExit:true});WebAssembly.instantiate(fs.readFileSync(process.argv[1]),{wasi_snapshot_preview1:wasi.wasiImport}).then(({instance})=>{process.exitCode=wasi.start(instance)});`, binary)
	}
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("C versus Go: %v\n%s", err, output)
	}
	if disagreements != 0 {
		t.Fatalf("%d reference versus Node disagreements", disagreements)
	}
	t.Logf("%d test262 + %d cohere + 4000 seeded + 4 ruling probes agree on acceptance, complete Node SyntaxError messages and Go divergence reasons (WASI=%v)", corpus-cohereCount, cohereCount, wasi)
}

func runtimeRegexCompilerCases(t *testing.T) ([][]string, int) {
	t.Helper()
	data, err := os.ReadFile("../regexp/testdata/test262.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases [][]string
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	corpus := len(cases)
	cohereData, err := os.ReadFile("../regexp/testdata/runtime-cohere.json")
	if err != nil {
		t.Fatal(err)
	}
	var cohere [][]string
	if err := json.Unmarshal(cohereData, &cohere); err != nil {
		t.Fatal(err)
	}
	if len(cohere) != 875 {
		t.Fatalf("cohere count %d", len(cohere))
	}
	cases = append(cases, cohere...)
	cases = append(cases, [][]string{{"(?<℘>a)", "u", "Other_ID_Start"}, {"(?<a·>a)", "u", "Other_ID_Continue"}, {`\u{10000000000000000000000}`, "u", "overflow"}, {"a{2147483648,2147483647}", "", "clamp"}}...)
	corpus += len(cohere)
	rng := rand.New(rand.NewSource(0xEC2025))
	atoms := []string{"a", ".", "[a-z]", "[^]", "(a)", "(?:a)", "(?=a)", "(?<=a)", `\d`, `\p{Letter}`, "[", "(", `\k<x>`, "(?<x>a)", "[a&&b]", `[\q{ab|}]`}
	quantifiers := []string{"", "*", "+", "?", "{0}", "{1,}", "{2,4}", "{4,2}", "*?"}
	flags := []string{"", "u", "v", "i", "gimsy", "uv", "uu"}
	for i := 0; i < 4000; i++ {
		p := atoms[rng.Intn(len(atoms))] + quantifiers[rng.Intn(len(quantifiers))]
		if rng.Intn(2) == 0 {
			p += "|" + atoms[rng.Intn(len(atoms))]
		}
		cases = append(cases, []string{p, flags[rng.Intn(len(flags))], "seeded"})
	}
	return cases, corpus
}
