package native

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

// Compile from counted source in C, then compare the matcher's capture spans
// and lastIndex directly with live Node. The arena outlives every borrowed
// program pointer and is freed after the regex and its results.
func runtimeCompilerMatching(test *testing.T, wasm bool) {
	test.Helper()
	probes := []struct{ pattern, flags, input string }{
		{"(a|aa)(a?)", "", "aa"}, {"(a|(b))+", "", "aba"},
		{"(a?){2,4}?", "", "a"}, {"(?<=([ab]+)([bc]+))$", "", "abc"},
		{"(?<x>a)\\k<x>", "g", "xaa"}, {"(?<x>a)|(?<x>b)", "", "b"},
		{"\\1(a)", "", "a"}, {"(?!b)a", "", "ba"},
		{"[[a-z]&&[^aeiou]]+", "v", "abc"}, {"[[a-z]--[b-d]]+", "v", "abcxyz"},
		{"[\\q{ab|cd}]", "v", "xcd"}, {"[\\q{AB}]", "iv", "ab"},
		{"\\p{Script=Greek}+", "u", "xαβ"}, {"\\p{RGI_Emoji}", "v", "x👨‍👩‍👧"},
		{"\\P{Lowercase_Letter}", "iu", "a"}, {"[^a]", "iu", "A"},
		{"\\p{ASCII}", "u", "\x7f"}, {"k", "iu", "K"}, {"k", "i", "K"}, {"(?i:a)(?-i:B)", "", "aB"},
		{"a\x00b", "u", "xa\x00b"}, {"[\\uDF0D]", "u", "🌍"},
		{"[\\q{a}]", "iv", "A"}, // A classified V8 divergence; bypassing its guard must disagree with Node.
	}
	cases := make([]regexCase, len(probes))
	for index, probe := range probes {
		cases[index] = regexCase{Pattern: probe.pattern, Flags: probe.flags, Input: utf16.Encode([]rune(probe.input))}
	}
	wire, encodeError := json.Marshal(cases)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(
		timeoutContext,
		"node",
		"-e",
		`const c=JSON.parse(require('fs').readFileSync(0,'utf8'));for(const x of c){const r=new RegExp(x.pattern,x.flags+'d');r.lastIndex=x.lastIndex;const m=r.exec(String.fromCharCode(...x.input));x.expected={captures:m?Array.from(m.indices,p=>p??null):null,lastIndex:r.lastIndex};}process.stdout.write(JSON.stringify(c));`,
	)
	command.Stdin = bytes.NewReader(wire)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v %s", combinedOutputError, output)
	}
	if decodeError := json.Unmarshal(output, &cases); decodeError != nil {
		test.Fatal(decodeError)
	}
	var source strings.Builder
	source.WriteString(
		"#include \"adamic.h\"\n#include \"regexp_compile_v8.h\"\n#include <stdio.h>\n#include <string.h>\nint main(int argc,char **argv){adamic_start(argc,argv);unsigned failures=0;\n",
	)
	for i, regexCase := range cases {
		fmt.Fprintf(
			&source,
			"{const unsigned char pattern[]=%s,flags[]=%s;adamic_regex_parse_result parsed;adamic_regex_parse_native(pattern,%d,flags,%d,&parsed);adamic_regex_program *p=adamic_regex_compile_checked(&parsed);\n",
			cString(regexCase.Pattern),
			cString(regexCase.Flags+"d"),
			len(regexCase.Pattern),
			len(regexCase.Flags)+1,
		)
		if i == len(cases)-1 {
			source.WriteString("if(p==NULL&&parsed.status==3){adamic_regex_parse_free(&parsed);}else{\n")
		} else {
			source.WriteString("if(p==NULL){fprintf(stderr,\"unexpected compilation refusal\\n\");return 1;}\n")
		}
		source.WriteString("double units[]={")
		for _, u := range regexCase.Input {
			fmt.Fprintf(&source, "%d,", u)
		}
		source.WriteString("0};\n")
		fmt.Fprintf(
			&source,
			"adamic_string *input=adamic_string_from_char_codes(%d,units);adamic_object *r=adamic_regex_new(p,&adamic_string_empty,&adamic_string_empty);r->slots[1].number=%d;adamic_array *m=adamic_regex_exec(r,input);bool same=(m!=NULL)==%t&&r->slots[1].number==%d;\n",
			len(regexCase.Input),
			regexCase.LastIndex,
			len(regexCase.Expected.Captures) > 0,
			regexCase.Expected.LastIndex,
		)
		if len(regexCase.Expected.Captures) > 0 {
			fmt.Fprintf(
				&source,
				"if(m!=NULL){same= same&&m->length==%d;adamic_array *indices=m->properties->slots[3].reference;\n",
				len(regexCase.Expected.Captures),
			)
			for captureIndex, pair := range regexCase.Expected.Captures {
				if pair == nil {
					fmt.Fprintf(&source, "same=same&&indices->elements[%d].reference==NULL;\n", captureIndex)
				} else {
					fmt.Fprintf(
						&source,
						"{adamic_object *pair=indices->elements[%d].reference;same=same&&pair!=NULL;if(pair!=NULL)same=same&&pair->slots[0].number==%d&&pair->slots[1].number==%d;}\n",
						captureIndex,
						pair[0],
						pair[1],
					)
				}
			}
			source.WriteString("}\n")
		}
		fmt.Fprintf(
			&source,
			"if(!same){fprintf(stderr,\"DISAGREEMENT Node case %d\\n\");failures++;}adamic_release(m);adamic_release(r);adamic_release(input);adamic_regex_parse_free(&parsed);\n",
			i,
		)
		if i == len(cases)-1 {
			source.WriteString("}\n")
		}
		source.WriteString("}\n")
	}
	source.WriteString("return failures!=0;}\n")
	options := Options{Sanitize: true}
	if wasm {
		options = Options{Target: "wasm32-wasi"}
	}
	runtimeDirectory := "runtime"
	if path := os.Getenv("ADAMIC_REGEX_COMPILER_TEST_RUNTIME"); path != "" {
		runtimeDirectory = path
	}
	compilerHeaders, absolutePathError := filepath.Abs(runtimeDirectory)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	flags := append(Flags(options), "-DADAMIC_REGEXP_RUNTIME_COMPILER=1", "-I", compilerHeaders)
	if os.Getenv("ADAMIC_REGEX_COMPILER_TEST_RUNTIME") != "" {
		for _, name := range []string{"parser", "properties", "sets", "bytecode", "v8"} {
			path, absolutePathError := filepath.Abs(filepath.Join(runtimeDirectory, "regexp_compile_"+name+".c"))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			flags = append(flags, path)
		}
	} else {
		// The ordinary comparison uses the separately cached compiler component.
		// Explicit runtime overrides still compile their mutated sources directly.
		sourceText := source.String()
		source.Reset()
		source.WriteString("#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n" + sourceText)
	}
	binary := filepath.Join(test.TempDir(), "compiler-matches")
	if buildFlagsError := BuildFlags(source.String(), binary, options, flags); buildFlagsError != nil {
		test.Fatal(buildFlagsError)
	}
	ctxRun, cancelRun := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelRun()
	command = exec.CommandContext(ctxRun, binary)
	if wasm {
		command = exec.CommandContext(
			ctxRun,
			"node",
			"--disable-warning=ExperimentalWarning",
			"../../oracle/wasi.mjs",
			binary,
		)
	}
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	if output, combinedOutputError = command.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("runtime compiler versus Node: %v %s", combinedOutputError, output)
	}
	test.Logf(
		"%d runtime-compiled matches agree with live Node (wasm=%v), plus one classified V8 refusal",
		len(cases)-1,
		wasm,
	)
}
func TestRegExpRuntimeCompilerMatchesNodeCached(test *testing.T) {
	test.Parallel()
	runtimeCompilerMatching(test, false)
}
func TestRegExpRuntimeCompilerMatchesNodeWasmCached(test *testing.T) {
	test.Parallel()
	runtimeCompilerMatching(test, true)
}
