// Command regex compares warmed native RegExp.test calls with Node 24.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp/syntax"
	"strconv"
	"strings"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
	regex "github.com/system-inc/adamic/internal/regexp"
)

type probe struct {
	Name, Pattern, Flags string
	Inputs               []string
	Calls                int
}

func main() {
	directory := flag.String("out", "/tmp/regex-speed", "artifact directory")
	buildOnly := flag.Bool("build-only", false, "emit benchmark without running")
	flag.Parse()
	if runError := run(*directory, *buildOnly); runError != nil {
		fmt.Fprintln(os.Stderr, runError)
		os.Exit(1)
	}
}
func run(directory string, buildOnly bool) error {
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		return createDirectoryError
	}
	var probes []probe
	walkError := filepath.WalkDir("cohere/internal/lint", func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files := token.NewFileSet()
		tree, parseFileError := parser.ParseFile(files, path, nil, 0)
		if parseFileError != nil {
			return parseFileError
		}
		constants := map[string]ast.Expr{}
		ast.Inspect(tree, func(node ast.Node) bool {
			if value, ok := node.(*ast.ValueSpec); ok && len(value.Names) == 1 && len(value.Values) == 1 {
				constants[value.Names[0].Name] = value.Values[0]
			}
			return true
		})
		var evaluate func(ast.Expr, int) (string, bool)
		evaluate = func(expr ast.Expr, depth int) (string, bool) {
			if depth > 10 {
				return "", false
			}
			switch typed := expr.(type) {
			case *ast.BasicLit:
				if typed.Kind == token.STRING {
					text, unquoteError := strconv.Unquote(typed.Value)
					return text, unquoteError == nil
				}
			case *ast.BinaryExpr:
				if typed.Op == token.ADD {
					a, ok := evaluate(typed.X, depth+1)
					b, yes := evaluate(typed.Y, depth+1)
					return a + b, ok && yes
				}
			case *ast.Ident:
				if value, ok := constants[typed.Name]; ok {
					return evaluate(value, depth+1)
				}
			}
			return "", false
		}
		ast.Inspect(tree, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "MustCompile" {
				return true
			}
			owner, ok := selector.X.(*ast.Ident)
			if !ok || owner.Name != "regexp" {
				return true
			}
			pattern, ok := evaluate(call.Args[0], 0)
			if !ok {
				line := files.Position(call.Pos()).Line
				switch {
				case strings.HasSuffix(path, "require_description.go"):
					pattern = `^(custom-rule|additional)(?:\s|$)`
				case strings.HasSuffix(path, "consistency_no_return_void.go"):
					pattern = `^return\s+void\s+answer\s*;$`
				case strings.HasSuffix(path, "abbreviation_vocabulary.go"):
					switch line {
					case 162:
						pattern = `^cfg[A-Z]`
					case 176:
						pattern = `(^|[^a-zA-Z])cfg($|[A-Z0-9])`
					case 177:
						pattern = `[a-z0-9]cfg($|[A-Z0-9])`
					case 178:
						pattern = `cfg($|[A-Z0-9])`
					default:
						panic("new dynamic pattern")
					}
				default:
					panic("unaccounted dynamic regexp: " + path)
				}
				fmt.Fprintf(os.Stderr, "dynamic pattern instantiated: %s:%d = %s\n", path, line, pattern)
			}
			sample := ""
			if tree, parseError := syntax.Parse(pattern, syntax.Perl); parseError == nil {
				sample = witness(tree)
			}
			flags := ""
			for _, f := range []string{"i", "m", "s"} {
				if strings.HasPrefix(pattern, "(?"+f+")") {
					pattern = strings.TrimPrefix(pattern, "(?"+f+")")
					flags += f
				}
			}
			// Translate Go's explicit code-point escapes to ECMAScript Unicode escapes.
			if strings.Contains(pattern, "\\x{") {
				pattern = strings.ReplaceAll(pattern, "\\x{", "\\u{")
				flags += "u"
			}
			probes = append(
				probes,
				probe{fmt.Sprintf("%s:%d", path, files.Position(call.Pos()).Line), pattern, flags, []string{
					sample,
					"shadow",
					"useEffect",
					"eslint-disable-next-line no-warning-comments -- explanation",
					"return undefined;",
					"bg-opacity-50",
					strings.Repeat("ordinary prose without a directive ", 32),
					"import React from 'react';\nconst x = useState(0);",
					"\t* @example 'CAPITAL TOKEN' ```code```",
				}, 9000},
			)
			return true
		})
		return nil
	})
	if walkError != nil {
		return walkError
	}
	keywords := strings.Fields(
		"abstract arguments await boolean break byte case catch char class const continue debugger default delete do double else enum eval export extends false final finally float for function goto if implements import in instanceof int interface let long native new null package private protected public return short static super switch synchronized this throw throws transient true try typeof var void volatile while with yield",
	)
	large := strings.Repeat("ordinary prose; padding 12345\n", 512)
	probes = append(probes,
		probe{
			"hard/keywords",
			`\b(?:` + strings.Join(keywords, "|") + `)\b`,
			"",
			[]string{large + "volatile", large},
			200,
		},
		probe{
			"hard/email",
			`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`,
			"",
			[]string{large + "hello@example.com", large},
			200,
		},
		probe{"hard/ignorecase", `canonicalize`, "iu", []string{large + "CANONICALIZE", large + "Kſ"}, 200},
		probe{"hard/lookbehind", `(?<=prefix )[a-z]+`, "", []string{large + "prefix suffix", large}, 200})
	originalData, readError := os.ReadFile("bench/regex/originals.json")
	if readError != nil {
		return readError
	}
	var originals []probe
	if decodeError := json.Unmarshal(originalData, &originals); decodeError != nil {
		return decodeError
	}
	probes = append(probes, originals...)
	data, encodeError := json.MarshalIndent(probes, "", "  ")
	if encodeError != nil {
		return encodeError
	}
	if writeError := os.WriteFile(filepath.Join(directory, "cases.json"), data, 0644); writeError != nil {
		return writeError
	}
	var source strings.Builder
	source.WriteString(
		"#define _POSIX_C_SOURCE 200809L\n#include \"adamic.h\"\n#include <stdio.h>\n#include <stdlib.h>\n#include <time.h>\n",
	)
	header := os.Getenv("REGEXP_CALLGRIND_HEADER")
	if header != "" {
		fmt.Fprintf(&source, "#include %q\n", header)
	} else {
		source.WriteString(
			"#define CALLGRIND_START_INSTRUMENTATION ((void)0)\n#define CALLGRIND_STOP_INSTRUMENTATION ((void)0)\n#define CALLGRIND_ZERO_STATS ((void)0)\n#define CALLGRIND_DUMP_STATS ((void)0)\n",
		)
	}
	for id, probe := range probes {
		compiled, compileError := regex.Compile(probe.Pattern, probe.Flags)
		if compileError != nil {
			return fmt.Errorf("%s: %w", probe.Name, compileError)
		}
		declarations, nativeDeclarationsError := compiled.NativeDeclarations(fmt.Sprintf("program_%d", id))
		if nativeDeclarationsError != nil {
			return nativeDeclarationsError
		}
		source.WriteString(declarations)
		for index, input := range probe.Inputs {
			fmt.Fprintf(&source, "static const char bytes_%d_%d[] = {", id, index)
			for _, character := range []byte(input) {
				fmt.Fprintf(&source, "%d,", character)
			}
			fmt.Fprintf(
				&source,
				"0};\nstatic adamic_string input_%d_%d = {{0,adamic_kind_string,0},%d,bytes_%d_%d,0,ADAMIC_LITERAL_INDEX,NULL,0};\n",
				id,
				index,
				len(input),
				id,
				index,
			)
		}
	}
	source.WriteString(
		"int main(int argc,char **argv) {adamic_start(argc,argv);int id=atoi(argv[1]);size_t calls=(size_t)strtoull(argv[2],NULL,10);switch(id){\n",
	)
	for id, probe := range probes {
		fmt.Fprintf(
			&source,
			"case %d: {adamic_object *regex=adamic_regex_new(&program_%d,&adamic_string_empty,&adamic_string_empty);adamic_string *inputs[]={",
			id,
			id,
		)
		for index := range probe.Inputs {
			fmt.Fprintf(&source, "&input_%d_%d,", id, index)
		}
		fmt.Fprintf(
			&source,
			"};size_t count=%d;for(size_t k=0;k<100;k++)adamic_regex_test(regex,inputs[k%%count]);struct timespec a,b;size_t sum=0;clock_gettime(CLOCK_MONOTONIC,&a);CALLGRIND_START_INSTRUMENTATION;CALLGRIND_ZERO_STATS;for(size_t k=0;k<calls;k++)sum+=adamic_regex_test(regex,inputs[k%%count]);CALLGRIND_STOP_INSTRUMENTATION;CALLGRIND_DUMP_STATS;clock_gettime(CLOCK_MONOTONIC,&b);struct timespec r;clock_getres(CLOCK_MONOTONIC,&r);printf(\"%%.9f %%zu %%.9f\\n\",(double)(b.tv_sec-a.tv_sec)+(double)(b.tv_nsec-a.tv_nsec)*1e-9,sum,(double)r.tv_sec+(double)r.tv_nsec*1e-9);adamic_release(regex);break;}\n",
			len(probe.Inputs),
		)
	}
	source.WriteString("default:return 1;}return 0;}\n")
	if writeError := os.WriteFile(
		filepath.Join(directory, "main.c"),
		[]byte(source.String()),
		0644,
	); writeError != nil {
		return writeError
	}
	if os.Getenv("REGEXP_BENCH_NO_BUILD") != "1" {
		if buildError := native.Build(
			source.String(),
			filepath.Join(directory, "native"),
			native.Options{},
		); buildError != nil {
			return buildError
		}
	}
	js := `const cases=require('./cases.json');if(!process.version.startsWith('v24.'))throw Error('Node 24 required');const p=cases[+process.argv[2]],n=+process.argv[3],r=new RegExp(p.Pattern,p.Flags);for(let k=0;k<10000;k++)r.test(p.Inputs[k%p.Inputs.length]);let sum=0;const start=process.hrtime.bigint();for(let k=0;k<n;k++)sum+=r.test(p.Inputs[k%p.Inputs.length]);const elapsed=Number(process.hrtime.bigint()-start)/1e9;let step=Infinity;for(let k=0;k<1000;k++){const t=process.hrtime.bigint();let u;do{u=process.hrtime.bigint();}while(u===t);step=Math.min(step,Number(u-t)/1e9);}console.log(elapsed,sum,step);`
	if writeError := os.WriteFile(filepath.Join(directory, "node.cjs"), []byte(js), 0644); writeError != nil {
		return writeError
	}
	if buildOnly {
		return nil
	}
	fmt.Println("pattern,native_seconds_best5,node_seconds_best5,calls,checksum")
	for id, probe := range probes {
		best := [2]float64{1e99, 1e99}
		checks := [2]string{}
		for round := 0; round < 5; round++ {
			for engine := 0; engine < 2; engine++ {
				var command *exec.Cmd
				if engine == 0 {
					command = exec.Command(
						filepath.Join(directory, "native"),
						strconv.Itoa(id),
						strconv.Itoa(probe.Calls),
					)
				} else {
					command = exec.Command(
						"node",
						filepath.Join(directory, "node.cjs"),
						strconv.Itoa(id),
						strconv.Itoa(probe.Calls),
					)
				}
				output, combinedOutputError := command.CombinedOutput()
				if combinedOutputError != nil {
					return fmt.Errorf("%s: %v %s", probe.Name, combinedOutputError, output)
				}
				seconds, checksum, readingError := reading(string(output))
				if readingError != nil {
					return fmt.Errorf("%s engine %d: %w", probe.Name, engine, readingError)
				}
				if seconds < best[engine] {
					best[engine] = seconds
				}
				if checks[engine] != "" && checks[engine] != checksum {
					return fmt.Errorf("unstable checksum %s", probe.Name)
				}
				checks[engine] = checksum
			}
		}
		if checks[0] != checks[1] {
			return fmt.Errorf("DISAGREEMENT %s native=%s node=%s", probe.Name, checks[0], checks[1])
		}
		fmt.Printf("%s,%.9f,%.9f,%d,%s\n", probe.Name, best[0], best[1], probe.Calls, checks[0])
	}
	return nil
}

// reading is one engine's timed run: its seconds, the checksum the loop consumed, and its timer's
// resolution, all three printed by the engine. A reading at or below that resolution measured
// nothing (Perry's report gave 0.0 ms in every cell) and is refused, never kept as the fastest.
func reading(output string) (float64, string, error) {
	fields := strings.Fields(output)
	if len(fields) != 3 {
		return 0, "", fmt.Errorf("want seconds, checksum and timer resolution, got %q", output)
	}
	seconds, parseError := strconv.ParseFloat(fields[0], 64)
	if parseError != nil {
		return 0, "", parseError
	}
	resolution, parseError := strconv.ParseFloat(fields[2], 64)
	if parseError != nil {
		return 0, "", parseError
	}
	if !(resolution > 0) || !(seconds > resolution) {
		return 0, "", fmt.Errorf("measured nothing: %.9f s at a timer resolution of %.9f s", seconds, resolution)
	}
	return seconds, fields[1], nil
}

// Construct a positive example from each Go source pattern, before translation.
func witness(r *syntax.Regexp) string {
	switch r.Op {
	case syntax.OpLiteral:
		return string(r.Rune)
	case syntax.OpCharClass:
		for index := 0; index < len(r.Rune); index += 2 {
			for _, character := range []rune{'a', 'A', '0', ' '} {
				if character >= r.Rune[index] && character <= r.Rune[index+1] {
					return string(character)
				}
			}
		}
		if len(r.Rune) > 0 {
			return string(r.Rune[0])
		}
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return "a"
	case syntax.OpCapture:
		return witness(r.Sub[0])
	case syntax.OpConcat:
		var builder strings.Builder
		for _, sub := range r.Sub {
			builder.WriteString(witness(sub))
		}
		return builder.String()
	case syntax.OpAlternate:
		if len(r.Sub) > 0 {
			return witness(r.Sub[0])
		}
	case syntax.OpPlus:
		return witness(r.Sub[0])
	case syntax.OpRepeat:
		return strings.Repeat(witness(r.Sub[0]), r.Min)
	}
	return ""
}
