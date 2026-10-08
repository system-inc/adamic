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

func TestRegExpRuntimeBytecodeIdentity(t *testing.T) { runtimeBytecodeIdentity(t, false, "") }
func TestRegExpRuntimeBytecodeWASI(t *testing.T)     { runtimeBytecodeIdentity(t, true, "") }

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

func runtimeBytecodeIdentity(t *testing.T, wasi bool, mutant string) {
	t.Helper()
	cases, _ := runtimeRegexCompilerCases(t)
	checked := strings.HasPrefix(mutant, "v8")
	if checked {
		cases = append(cases, []string{`[\q{a}]{18446744073709551616}`, "iv"})
		for _, prefix := range []string{"", "(?i:^)", "(?-i:^)", "(?i:^)(?:)", "(?i:^)(?-i:)"} {
			for _, atom := range []string{`[b]`, `[a-z]`, `[0-9]`, `\w`, `[\w]`, `\W`, `[\W]`, `[\w\W]`, `[^\w]`, `[^\W]`, `[^b]`, `\p{Lowercase_Letter}`, `[\q{a}]`, `[a\q{a}]`, `[\q{AB}]`, `[\q{ab}]`, `[\q{Ss|x}]`, `[[\q{a}]&&[A]]`, `[[A]--[\q{a}]]`, `[\q{ab|a|}]`} {
				for _, flags := range []string{"u", "iu", "v", "iv"} {
					cases = append(cases, []string{prefix + atom, flags})
				}
			}
		}
	}
	var source strings.Builder
	source.WriteString("#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n#include \"regexp_compile_v8.h\"\n")
	source.WriteString(runtimeBytecodeWriter)
	source.WriteString("struct test {const unsigned char *pattern,*flags;size_t length,flag_length;};\nstatic const struct test cases[]={\n")
	emit := func(value string) string {
		var out strings.Builder
		out.WriteString("(const unsigned char[]){")
		for _, b := range []byte(value) {
			fmt.Fprintf(&out, "%d,", b)
		}
		out.WriteString("0}")
		return out.String()
	}
	expected := make([][]byte, len(cases))
	accepted, refused := 0, 0
	for i, c := range cases {
		p, err := reference.Compile(c[0], c[1])
		status := uint64(0)
		refusal := ""
		if err == nil && checked {
			err = p.NativeCompatibility()
		}
		if err != nil {
			var divergence *reference.V8DivergenceError
			if errors.As(err, &divergence) {
				status = 3
				refusal = err.Error()
			} else {
				var syntax *reference.SyntaxError
				if !errors.As(err, &syntax) {
					t.Fatalf("unexpected Go compilation error: %v", err)
				}
				status = 1
			}
		}
		var wire []byte
		if status == 0 {
			wire, err = p.NativeBytecodeSnapshot()
			if err != nil {
				status = 4
			} else {
				accepted++
			}
		}
		if status != 0 {
			refused++
		}
		expected[i] = binary.LittleEndian.AppendUint64(nil, status)
		expected[i] = append(expected[i], wire...)
		if checked && status == 3 {
			expected[i] = binary.LittleEndian.AppendUint64(expected[i], uint64(len(refusal)))
			expected[i] = append(expected[i], refusal...)
		}
		fmt.Fprintf(&source, "{%s,%s,%d,%d},\n", emit(c[0]), emit(c[1]), len(c[0]), len(c[1]))
	}
	if checked {
		source.WriteString("#define adamic_regex_compile_bytecode adamic_regex_compile_checked\n")
	}
	source.WriteString("};\nint main(void){for(size_t j=0;j<sizeof(cases)/sizeof(cases[0]);j++){const struct test *c=&cases[j];adamic_regex_parse_result r;adamic_regex_parse(c->pattern,c->length,c->flags,c->flag_length,&r);adamic_regex_program *p=NULL;if(r.status==0)p=adamic_regex_compile_bytecode(&r);number((uint64_t)r.status);if(r.status==0)program(p);adamic_regex_parse_free(&r);}return 0;}\n")
	if checked {
		text := source.String()
		text = strings.Replace(text, "if(r.status==0)program(p);", "if(r.status==0)program(p);if(r.status==3){number(r.message_length);fwrite(r.message,1,r.message_length,stdout);}", 1)
		source.Reset()
		source.WriteString(text)
	}
	dir := t.TempDir()
	main := filepath.Join(dir, "main.c")
	if err := os.WriteFile(main, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	runtimeDirectory, _ := filepath.Abs("runtime")
	bytecode := filepath.Join(runtimeDirectory, "regexp_compile_bytecode.c")
	parser := filepath.Join(runtimeDirectory, "regexp_compile_parser.c")
	if mutant == "set opcode" {
		data, err := os.ReadFile(bytecode)
		if err != nil {
			t.Fatal(err)
		}
		old := "base.op=1;base.ranges"
		if bytes.Count(data, []byte(old)) != 1 {
			t.Fatal("mutation site moved")
		}
		changed := bytes.Replace(data, []byte(old), []byte("base.op=5;base.ranges"), 1)
		bytecode = filepath.Join(dir, "mutant.c")
		if err := os.WriteFile(bytecode, changed, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if mutant == "accept syntax" {
		data, err := os.ReadFile(parser)
		if err != nil {
			t.Fatal(err)
		}
		old := `if (regex_parse_in(byte, "*+?")) return regex_parse_fail(p, p->position, "nothing to repeat", "Nothing to repeat");`
		if bytes.Count(data, []byte(old)) != 1 {
			t.Fatal("syntax mutant site moved")
		}
		changed := bytes.Replace(data, []byte(old), []byte(`if (regex_parse_in(byte, "*+?")) return regex_parse_literal(p);`), 1)
		parser = filepath.Join(dir, "parser-mutant.c")
		if err := os.WriteFile(parser, changed, 0600); err != nil {
			t.Fatal(err)
		}
	}
	v8 := filepath.Join(runtimeDirectory, "regexp_compile_v8.c")
	if mutant == "v8 bypass" {
		data, err := os.ReadFile(v8)
		if err != nil {
			t.Fatal(err)
		}
		old := []byte("!adamic_regex_compile_v8_check(result)")
		if bytes.Count(data, old) != 1 {
			t.Fatal("v8 mutant site moved")
		}
		data = bytes.Replace(data, old, []byte("false"), 1)
		v8 = filepath.Join(dir, "v8-mutant.c")
		if err := os.WriteFile(v8, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	binaryPath := filepath.Join(dir, "check")
	args := []string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O1", "-g", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-DADAMIC_REGEXP_RUNTIME_COMPILER=1", "-I", runtimeDirectory, main, parser, filepath.Join(runtimeDirectory, "regexp_compile_properties.c"), filepath.Join(runtimeDirectory, "regexp_compile_sets.c"), bytecode, "-o", binaryPath}
	if checked {
		args = append(args, v8)
	}
	compiler := "clang"
	if wasi {
		sysroot := os.Getenv("WASI_SYSROOT")
		if sysroot == "" {
			t.Fatal("WASI_SYSROOT required")
		}
		compiler = filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin", "clang")
		args = append([]string{"--target=wasm32-wasi", "--sysroot=" + sysroot, "-DADAMIC_TARGET_WASI=1", "-mno-atomics", "-Wl,-z,stack-size=1048576"}, args...)
		filtered := args[:0]
		for _, arg := range args {
			if arg != "-fsanitize=address,undefined" && arg != "-fno-sanitize-recover=all" && arg != "-g" {
				filtered = append(filtered, arg)
			}
		}
		args = filtered
	}
	command := exec.Command(compiler, args...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compile: %v %s", err, output)
	}
	outputFile := filepath.Join(dir, "actual.bin")
	out, err := os.Create(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	command = exec.Command(binaryPath)
	if wasi {
		command = exec.Command("node", "--no-warnings", "-e", `const fs=require('fs');const {WASI}=require('node:wasi');const wasi=new WASI({version:'preview1',args:[],env:{},returnOnExit:true});WebAssembly.instantiate(fs.readFileSync(process.argv[1]),{wasi_snapshot_preview1:wasi.wasiImport}).then(({instance})=>{process.exitCode=wasi.start(instance)});`, binaryPath)
	}
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	var diagnostics bytes.Buffer
	command.Stdout = out
	command.Stderr = &diagnostics
	err = command.Run()
	out.Close()
	if err != nil {
		t.Fatalf("execution: %v %s", err, diagnostics.Bytes())
	}
	actual, err := os.Open(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	defer actual.Close()
	mismatch := ""
	for i, want := range expected {
		got := make([]byte, len(want))
		if _, err := io.ReadFull(actual, got); err != nil {
			mismatch = fmt.Sprintf("case %d %q/%s truncated: %v", i, cases[i][0], cases[i][1], err)
			break
		}
		if !bytes.Equal(want, got) {
			at := 0
			for at < len(want) && want[at] == got[at] {
				at++
			}
			mismatch = fmt.Sprintf("case %d %q/%s byte %d Go=%d C=%d", i, cases[i][0], cases[i][1], at, want[at], got[at])
			break
		}
	}
	if mutant != "" && mutant != "v8" {
		if mismatch == "" {
			t.Fatal("bytecode mutant survived")
		}
		t.Logf("caught %s mutant: %s", mutant, mismatch)
		return
	}
	if mismatch != "" {
		t.Fatal(mismatch)
	}
	var extra [1]byte
	if n, err := actual.Read(extra[:]); n != 0 || err != io.EOF {
		t.Fatal("extra C bytecode output")
	}
	t.Logf("byte-identical: %d programs, %d syntax/refusal cases (WASI=%v)", accepted, refused, wasi)
}
func TestRegExpRuntimeBytecodeEmissionMutant(t *testing.T) {
	runtimeBytecodeIdentity(t, false, "set opcode")
}

func TestRegExpRuntimeRejectedPatternMutant(t *testing.T) {
	runtimeBytecodeIdentity(t, false, "accept syntax")
}

func TestRegExpRuntimeV8Refusals(t *testing.T)      { runtimeBytecodeIdentity(t, false, "v8") }
func TestRegExpRuntimeV8WASI(t *testing.T)          { runtimeBytecodeIdentity(t, true, "v8") }
func TestRegExpRuntimeV8RefusalMutant(t *testing.T) { runtimeBytecodeIdentity(t, false, "v8 bypass") }
