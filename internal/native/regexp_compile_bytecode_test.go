package native

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	reference "github.com/system-inc/adamic/internal/regexp"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegExpRuntimeBytecodeIdentity(test *testing.T) {
	test.Parallel()
	runtimeBytecodeIdentity(test, false, "")
}
func TestRegExpRuntimeBytecodeWASI(test *testing.T) {
	test.Parallel()
	runtimeBytecodeIdentity(test, true, "")
}

const runtimeBytecodeWriter = `
static void number(uint64_t value) { unsigned char bytes[8]; for(unsigned i=0;i<8;i++)bytes[i]=(unsigned char)(value>>(8*i)); if(fwrite(bytes,1,8,stdout)!=8)exit(2); }
static void text(const char *value) {size_t length=strlen(value);number(length);if(fwrite(value,1,length,stdout)!=length)exit(2);}
static void program(const adamic_regex_program *p) {
 size_t count=0;do{count++;}while(p->code[count-1].op!=0);
 number(p->flags);number(p->captures);number(p->repeats);number(count);number(p->group_count);
 for(size_t i=0;i<p->group_count;i++){text(p->groups[i].name);number(p->groups[i].count);for(size_t j=0;j<p->groups[i].count;j++)number(p->groups[i].captures[j]);}
 for(size_t index=0;index<count;index++) {
 const adamic_regex_instruction *i=&p->code[index];number((uint64_t)(int64_t)i->op);number((uint64_t)(int64_t)i->x);number((uint64_t)(int64_t)i->y);number((uint64_t)(int64_t)i->direction);number(i->flags);number((uint64_t)(int64_t)i->assertion);number(i->negative);number(i->greedy);number(i->unbounded);number(i->minimum);number(i->maximum);
 number(i->range_count);for(size_t j=0;j<i->range_count;j++){number(i->ranges[j].first);number(i->ranges[j].last);}
 number(i->string_count);for(size_t j=0;j<i->string_count;j++){number(i->strings[j].count);for(size_t k=0;k<i->strings[j].count;k++)number(i->strings[j].points[k]);}
 number(i->id_count);for(size_t j=0;j<i->id_count;j++)number(i->ids[j]);number(i->look!=NULL);if(i->look!=NULL)program(i->look);
 }
}
`

func runtimeBytecodeIdentity(test *testing.T, wasi bool, mutant string) {
	test.Helper()
	cases, _ := runtimeRegexCompilerCases(test)
	checked := strings.HasPrefix(mutant, "v8")
	if checked {
		cases = append(cases, []string{`[\q{a}]{18446744073709551616}`, "iv"})
		for _, prefix := range []string{"", "(?i:^)", "(?-i:^)", "(?i:^)(?:)", "(?i:^)(?-i:)"} {
			for _, atom := range []string{
				`[b]`,
				`[a-z]`,
				`[0-9]`,
				`\w`,
				`[\w]`,
				`\W`,
				`[\W]`,
				`[\w\W]`,
				`[^\w]`,
				`[^\W]`,
				`[^b]`,
				`\p{Lowercase_Letter}`,
				`[\q{a}]`,
				`[a\q{a}]`,
				`[\q{AB}]`,
				`[\q{ab}]`,
				`[\q{Ss|x}]`,
				`[[\q{a}]&&[A]]`,
				`[[A]--[\q{a}]]`,
				`[\q{ab|a|}]`,
			} {
				for _, flags := range []string{"u", "iu", "v", "iv"} {
					cases = append(cases, []string{prefix + atom, flags})
				}
			}
		}
	}
	var source strings.Builder
	source.WriteString(
		"#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n#include \"regexp_compile_v8.h\"\n",
	)
	source.WriteString(runtimeBytecodeWriter)
	source.WriteString(
		"struct test {const unsigned char *pattern,*flags;size_t length,flag_length;};\nstatic const struct test cases[]={\n",
	)
	emit := func(value string) string {
		var out strings.Builder
		out.WriteString("(const unsigned char[]){")
		for _, character := range []byte(value) {
			fmt.Fprintf(&out, "%d,", character)
		}
		out.WriteString("0}")
		return out.String()
	}
	expected := make([][]byte, len(cases))
	accepted, refused := 0, 0
	for index, c := range cases {
		program, err := reference.Compile(c[0], c[1])
		status := uint64(0)
		refusal := ""
		if err == nil && checked {
			err = program.NativeCompatibility()
		}
		if err != nil {
			var divergence *reference.V8DivergenceError
			if errors.As(err, &divergence) {
				status = 3
				refusal = err.Error()
			} else {
				var syntax *reference.SyntaxError
				if !errors.As(err, &syntax) {
					test.Fatalf("unexpected Go compilation error: %v", err)
				}
				status = 1
			}
		}
		var wire []byte
		if status == 0 {
			var nativeBytecodeSnapshotError error
			wire, nativeBytecodeSnapshotError = program.NativeBytecodeSnapshot()
			if nativeBytecodeSnapshotError != nil {
				status = 4
			} else {
				accepted++
			}
		}
		if status != 0 {
			refused++
		}
		expected[index] = binary.LittleEndian.AppendUint64(nil, status)
		expected[index] = append(expected[index], wire...)
		if checked && status == 3 {
			expected[index] = binary.LittleEndian.AppendUint64(expected[index], uint64(len(refusal)))
			expected[index] = append(expected[index], refusal...)
		}
		fmt.Fprintf(&source, "{%s,%s,%d,%d},\n", emit(c[0]), emit(c[1]), len(c[0]), len(c[1]))
	}
	if checked {
		source.WriteString("#define adamic_regex_compile_bytecode adamic_regex_compile_checked\n")
	}
	source.WriteString(
		"};\nint main(void){for(size_t j=0;j<sizeof(cases)/sizeof(cases[0]);j++){const struct test *c=&cases[j];adamic_regex_parse_result r;adamic_regex_parse_native(c->pattern,c->length,c->flags,c->flag_length,&r);adamic_regex_program *p=NULL;if(r.status==0)p=adamic_regex_compile_bytecode(&r);number((uint64_t)r.status);if(r.status==0)program(p);adamic_regex_parse_free(&r);}return 0;}\n",
	)
	if checked {
		text := source.String()
		text = strings.Replace(
			text,
			"if(r.status==0)program(p);",
			"if(r.status==0)program(p);if(r.status==3){number(r.message_length);fwrite(r.message,1,r.message_length,stdout);}",
			1,
		)
		source.Reset()
		source.WriteString(text)
	}
	dir := test.TempDir()
	main := filepath.Join(dir, "main.c")
	if writeError := os.WriteFile(main, []byte(source.String()), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	runtimeDirectory, _ := filepath.Abs("runtime")
	bytecode := filepath.Join(runtimeDirectory, "regexp_compile_bytecode.c")
	parser := filepath.Join(runtimeDirectory, "regexp_compile_parser.c")
	if mutant == "set opcode" {
		data, readError := os.ReadFile(bytecode)
		if readError != nil {
			test.Fatal(readError)
		}
		old := "base.op=1;base.ranges"
		if bytes.Count(data, []byte(old)) != 1 {
			test.Fatal("mutation site moved")
		}
		changed := bytes.Replace(data, []byte(old), []byte("base.op=5;base.ranges"), 1)
		bytecode = filepath.Join(dir, "mutant.c")
		if writeError := os.WriteFile(bytecode, changed, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if mutant == "accept syntax" {
		data, readError := os.ReadFile(parser)
		if readError != nil {
			test.Fatal(readError)
		}
		old := `if (regex_parse_in(byte, "*+?")) return regex_parse_fail(p, p->position, "nothing to repeat", "Nothing to repeat");`
		if bytes.Count(data, []byte(old)) != 1 {
			test.Fatal("syntax mutant site moved")
		}
		changed := bytes.Replace(
			data,
			[]byte(old),
			[]byte(`if (regex_parse_in(byte, "*+?")) return regex_parse_literal(p);`),
			1,
		)
		parser = filepath.Join(dir, "parser-mutant.c")
		if writeError := os.WriteFile(parser, changed, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	v8 := filepath.Join(runtimeDirectory, "regexp_compile_v8.c")
	if mutant == "v8 bypass" {
		data, readError := os.ReadFile(v8)
		if readError != nil {
			test.Fatal(readError)
		}
		old := []byte("!adamic_regex_compile_v8_check(result)")
		if bytes.Count(data, old) != 1 {
			test.Fatal("v8 mutant site moved")
		}
		data = bytes.Replace(data, old, []byte("false"), 1)
		v8 = filepath.Join(dir, "v8-mutant.c")
		if writeError := os.WriteFile(v8, data, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	binaryPath := filepath.Join(dir, "check")
	args := []string{
		"-std=c11",
		"-Wall",
		"-Wextra",
		"-Werror",
		"-pedantic",
		"-O1",
		"-g",
		"-fsanitize=address,undefined",
		"-fno-sanitize-recover=all",
		"-DADAMIC_REGEXP_RUNTIME_COMPILER=1",
		"-I",
		runtimeDirectory,
		main,
		parser,
		filepath.Join(runtimeDirectory, "regexp_compile_properties.c"),
		filepath.Join(runtimeDirectory, "regexp_compile_sets.c"),
		bytecode,
		"-o",
		binaryPath,
	}
	if checked {
		args = append(args, v8)
	}
	compiler := "clang"
	if wasi {
		sysroot := os.Getenv("WASI_SYSROOT")
		if sysroot == "" {
			test.Fatal("WASI_SYSROOT required")
		}
		compiler = filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin", "clang")
		args = append([]string{
			"--target=wasm32-wasi",
			"--sysroot=" + sysroot,
			"-DADAMIC_TARGET_WASI=1",
			"-mno-atomics",
			"-Wl,-z,stack-size=1048576",
		}, args...)
		filtered := args[:0]
		for _, arg := range args {
			if arg != "-fsanitize=address,undefined" && arg != "-fno-sanitize-recover=all" && arg != "-g" {
				filtered = append(filtered, arg)
			}
		}
		args = filtered
	}
	command := exec.Command(compiler, args...)
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("compile: %v %s", combinedOutputError, output)
	}
	outputFile := filepath.Join(dir, "actual.bin")
	out, createError := os.Create(outputFile)
	if createError != nil {
		test.Fatal(createError)
	}
	command = exec.Command(binaryPath)
	if wasi {
		command = exec.Command(
			"node",
			"--no-warnings",
			"-e",
			`const fs=require('fs');const {WASI}=require('node:wasi');const wasi=new WASI({version:'preview1',args:[],env:{},returnOnExit:true});WebAssembly.instantiate(fs.readFileSync(process.argv[1]),{wasi_snapshot_preview1:wasi.wasiImport}).then(({instance})=>{process.exitCode=wasi.start(instance)});`,
			binaryPath,
		)
	}
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	var diagnostics bytes.Buffer
	command.Stdout = out
	command.Stderr = &diagnostics
	runError := command.Run()
	out.Close()
	if runError != nil {
		test.Fatalf("execution: %v %s", runError, diagnostics.Bytes())
	}
	actual, openError := os.Open(outputFile)
	if openError != nil {
		test.Fatal(openError)
	}
	defer actual.Close()
	mismatch := ""
	for index, want := range expected {
		got := make([]byte, len(want))
		if _, readError := io.ReadFull(actual, got); readError != nil {
			mismatch = fmt.Sprintf("case %d %q/%s truncated: %v", index, cases[index][0], cases[index][1], readError)
			break
		}
		if !bytes.Equal(want, got) {
			at := 0
			for at < len(want) && want[at] == got[at] {
				at++
			}
			mismatch = fmt.Sprintf(
				"case %d %q/%s byte %d Go=%d C=%d",
				index,
				cases[index][0],
				cases[index][1],
				at,
				want[at],
				got[at],
			)
			break
		}
	}
	if mutant != "" && mutant != "v8" {
		if mismatch == "" {
			test.Fatal("bytecode mutant survived")
		}
		test.Logf("caught %s mutant: %s", mutant, mismatch)
		return
	}
	if mismatch != "" {
		test.Fatal(mismatch)
	}
	var extra [1]byte
	if n, readError := actual.Read(extra[:]); n != 0 || readError != io.EOF {
		test.Fatal("extra C bytecode output")
	}
	test.Logf("byte-identical: %d programs, %d syntax/refusal cases (WASI=%v)", accepted, refused, wasi)
}
func TestRegExpRuntimeBytecodeEmissionMutant(test *testing.T) {
	test.Parallel()
	runtimeBytecodeIdentity(test, false, "set opcode")
}

func TestRegExpRuntimeRejectedPatternMutant(test *testing.T) {
	test.Parallel()
	runtimeBytecodeIdentity(test, false, "accept syntax")
}

func TestRegExpRuntimeV8Refusals(test *testing.T) {
	test.Parallel()
	runtimeBytecodeIdentity(test, false, "v8")
}
func TestRegExpRuntimeV8WASI(test *testing.T) {
	test.Parallel()
	runtimeBytecodeIdentity(test, true, "v8")
}
func TestRegExpRuntimeV8RefusalMutant(test *testing.T) {
	test.Parallel()
	runtimeBytecodeIdentity(test, false, "v8 bypass")
}
