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
func TestRegExpRuntimeParserAcceptance(test *testing.T) {
	test.Parallel()
	runtimeParserAcceptance(test, false)
}
func TestRegExpRuntimeParserWasi(test *testing.T) {
	test.Parallel()
	runtimeParserAcceptance(test, true)
}

func runtimeParserAcceptance(test *testing.T, wasi bool, mutateMessage ...bool) {
	test.Helper()
	cases, corpus := runtimeRegexCompilerCases(test)
	cohereCount := 875
	input, _ := json.Marshal(cases)
	node := exec.Command(
		"node",
		"-e",
		`const c=JSON.parse(require('fs').readFileSync(0,'utf8'));process.stdout.write(JSON.stringify(c.map(x=>{try{new RegExp(x[0],x[1]);return ''}catch(e){if(!(e instanceof SyntaxError))throw e;return e.message}})));`,
	)
	node.Stdin = bytes.NewReader(input)
	output, combinedOutputError := node.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v %s", combinedOutputError, output)
	}
	var messages []string
	if decodeError := json.Unmarshal(output, &messages); decodeError != nil {
		test.Fatal(decodeError)
	}
	var source strings.Builder
	source.WriteString(
		"#include <stdio.h>\n#include <string.h>\n#include \"regexp_compile_parser.h\"\nstruct test { const unsigned char *pattern, *flags; size_t length, flag_length; int expected; const unsigned char *reason; size_t reason_length; };\nstatic const struct test cases[] = {\n",
	)
	emit := func(value string) string {
		var builder strings.Builder
		builder.WriteString("(const unsigned char[]){")
		for _, character := range []byte(value) {
			fmt.Fprintf(&builder, "%d,", character)
		}
		builder.WriteString("0}")
		return builder.String()
	}
	disagreements := 0
	for index, c := range cases {
		_, parseError := reference.Parse(c[0], c[1])
		valid := parseError == nil
		var divergence *reference.V8DivergenceError
		refused := errors.As(parseError, &divergence)
		if !refused && valid != (messages[index] == "") {
			disagreements++
			test.Logf("reference versus Node %q %q: %v / %s", c[0], c[1], parseError, messages[index])
		}
		expected := 1
		if valid {
			expected = 0
		}
		reason := messages[index]
		if refused {
			expected = 3
			reason = parseError.Error()
		}

		fmt.Fprintf(
			&source,
			"{%s,%s,%d,%d,%d,%s,%d},\n",
			emit(c[0]),
			emit(c[1]),
			len(c[0]),
			len(c[1]),
			expected,
			emit(reason),
			len(reason),
		)
	}
	source.WriteString(
		"};\nint main(void) { int failures=0; for(size_t i=0;i<sizeof(cases)/sizeof(cases[0]);i++) { const struct test *c=&cases[i]; adamic_regex_parse_result r; adamic_regex_parse(c->pattern,c->length,c->flags,c->flag_length,&r); if(r.status!=c->expected || ((r.status==1 || r.status==3) && (r.message_length!=c->reason_length || memcmp(r.message,c->reason,c->reason_length)!=0))) {fprintf(stderr,\"case %zu: status %d reason %s expected %s\\n\",i,r.status,r.message?r.message:\"none\",c->reason);failures++;} adamic_regex_parse_free(&r); } return failures!=0;}\n",
	)
	directory := test.TempDir()
	main := filepath.Join(directory, "main.c")
	if writeError := os.WriteFile(main, []byte(source.String()), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	runtimeDirectory, _ := filepath.Abs("runtime")
	binary := filepath.Join(directory, "check")
	args := []string{
		"-std=c11",
		"-Wall",
		"-Wextra",
		"-Werror",
		"-pedantic",
		"-O1",
		"-g",
		"-fsanitize=address,undefined",
		"-fno-omit-frame-pointer",
		"-DADAMIC_REGEXP_RUNTIME_COMPILER=1",
		"-I",
		runtimeDirectory,
		main,
		filepath.Join(runtimeDirectory, "regexp_compile_parser.c"),
		filepath.Join(runtimeDirectory, "regexp_compile_properties.c"),
		"-o",
		binary,
	}
	compiler := "clang"
	if wasi {
		sysroot := os.Getenv("WASI_SYSROOT")
		if sysroot == "" {
			test.Fatal("parser WASI proof requires WASI_SYSROOT")
		}
		compiler = filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin", "clang")
		args = []string{
			"--target=wasm32-wasi",
			"--sysroot=" + sysroot,
			"-DADAMIC_TARGET_WASI=1",
			"-mno-atomics",
			"-std=c11",
			"-Wall",
			"-Wextra",
			"-Werror",
			"-pedantic",
			"-O1",
			"-DADAMIC_REGEXP_RUNTIME_COMPILER=1",
			"-I",
			runtimeDirectory,
			main,
			filepath.Join(runtimeDirectory, "regexp_compile_parser.c"),
			filepath.Join(runtimeDirectory, "regexp_compile_properties.c"),
			"-Wl,-z,stack-size=1048576",
			"-o",
			binary,
		}
	}
	if len(mutateMessage) > 0 && mutateMessage[0] {
		parser := filepath.Join(runtimeDirectory, "regexp_compile_parser.c")
		data, readError := os.ReadFile(parser)
		if readError != nil {
			test.Fatal(readError)
		}
		old := "result->message = message; result->message_length = position;"
		if bytes.Count(data, []byte(old)) != 1 {
			test.Fatal("counted message mutation site moved")
		}
		data = bytes.Replace(
			data,
			[]byte(old),
			[]byte("if(memchr(pattern,0,length)!=NULL)message[position-1]='!';\n    "+old),
			1,
		)
		replacement := filepath.Join(directory, "parser-mutant.c")
		if writeError := os.WriteFile(replacement, data, 0600); writeError != nil {
			test.Fatal(writeError)
		}
		for index, arg := range args {
			if arg == parser {
				args[index] = replacement
			}
		}
	}
	command := exec.Command(compiler, args...)
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("compile: %v\n%s", combinedOutputError, output)
	}
	command = exec.Command(binary)
	if wasi {
		command = exec.Command(
			"node",
			"--no-warnings",
			"-e",
			`const fs=require('fs');const {WASI}=require('node:wasi');const wasi=new WASI({version:'preview1',args:[],env:{},returnOnExit:true});WebAssembly.instantiate(fs.readFileSync(process.argv[1]),{wasi_snapshot_preview1:wasi.wasiImport}).then(({instance})=>{process.exitCode=wasi.start(instance)});`,
			binary,
		)
	}
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	output, combinedOutputError = command.CombinedOutput()
	if len(mutateMessage) > 0 && mutateMessage[0] {
		if combinedOutputError == nil ||
			!bytes.Contains(output, []byte("case ")) ||
			bytes.Contains(output, []byte("ERROR: AddressSanitizer")) {
			test.Fatalf("counted message mutant did not reach comparison: %v %s", combinedOutputError, output)
		}
		test.Log("caught error-message corruption after embedded NUL at counted-byte comparison")
		return
	}
	if combinedOutputError != nil {
		test.Fatalf("C versus Go: %v\n%s", combinedOutputError, output)
	}
	if disagreements != 0 {
		test.Fatalf("%d reference versus Node disagreements", disagreements)
	}
	test.Logf(
		"%d test262 + %d cohere + 4000 seeded + 9 ruling probes agree on acceptance, complete Node SyntaxError messages and Go divergence reasons (WASI=%v)",
		corpus-cohereCount,
		cohereCount,
		wasi,
	)
}

func runtimeRegexCompilerCases(test *testing.T) ([][]string, int) {
	test.Helper()
	data, readError := os.ReadFile("../regexp/testdata/test262.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var cases [][]string
	if decodeError := json.Unmarshal(data, &cases); decodeError != nil {
		test.Fatal(decodeError)
	}
	corpus := len(cases)
	cohereData, readError := os.ReadFile("../regexp/testdata/runtime-cohere.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var cohere [][]string
	if decodeError := json.Unmarshal(cohereData, &cohere); decodeError != nil {
		test.Fatal(decodeError)
	}
	if len(cohere) != 875 {
		test.Fatalf("cohere count %d", len(cohere))
	}
	cases = append(cases, cohere...)
	cases = append(cases, [][]string{
		{"\x00[", "u", "counted SyntaxError"},
		{"(?<ⸯ>a)", "u", "Pattern_Syntax"},
		{"(?<Ᲊ>a)", "u", "Unicode identifier"},
		{"a\x00b", "u", "embedded NUL"},
		{"\x00", "", "embedded NUL"},
		{"(?<℘>a)", "u", "Other_ID_Start"},
		{"(?<a·>a)", "u", "Other_ID_Continue"},
		{`\u{10000000000000000000000}`, "u", "overflow"},
		{"a{2147483648,2147483647}", "", "clamp"},
	}...)
	corpus += len(cohere)
	rng := rand.New(rand.NewSource(0xEC2025))
	atoms := []string{
		"a",
		".",
		"[a-z]",
		"[^]",
		"(a)",
		"(?:a)",
		"(?=a)",
		"(?<=a)",
		`\d`,
		`\p{Letter}`,
		"[",
		"(",
		`\k<x>`,
		"(?<x>a)",
		"[a&&b]",
		`[\q{ab|}]`,
	}
	quantifiers := []string{"", "*", "+", "?", "{0}", "{1,}", "{2,4}", "{4,2}", "*?"}
	flags := []string{"", "u", "v", "i", "gimsy", "uv", "uu"}
	for index := 0; index < 4000; index++ {
		p := atoms[rng.Intn(len(atoms))] + quantifiers[rng.Intn(len(quantifiers))]
		if rng.Intn(2) == 0 {
			p += "|" + atoms[rng.Intn(len(atoms))]
		}
		cases = append(cases, []string{p, flags[rng.Intn(len(flags))], "seeded"})
	}
	return cases, corpus
}

func TestRegExpRuntimeCountedErrorMutant(test *testing.T) {
	test.Parallel()
	runtimeParserAcceptance(test, false, true)
}
