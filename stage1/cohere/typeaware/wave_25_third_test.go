package typeaware

import (
	"bytes"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Read production inputs only; expected findings still come from the independent Go rules.
func wave25ThirdControls(h *harness) []string {
	values := map[string]string{}
	trees := map[string]*goast.File{}
	var text func(goast.Expr) string
	var lines func(goast.Expr) []string
	lines = func(e goast.Expr) []string {
		if list, ok := e.(*goast.CompositeLit); ok {
			var result []string
			for _, item := range list.Elts {
				result = append(result, text(item))
			}
			return result
		}
		return nil
	}
	text = func(e goast.Expr) string {
		switch n := e.(type) {
		case *goast.BasicLit:
			if n.Kind == token.STRING {
				v, err := strconv.Unquote(n.Value)
				if err != nil {
					h.t.Fatal(err)
				}
				return v
			}
		case *goast.Ident:
			return values[n.Name]
		case *goast.BinaryExpr:
			if n.Op == token.ADD {
				return text(n.X) + text(n.Y)
			}
		case *goast.CallExpr:
			if f, ok := n.Fun.(*goast.Ident); ok && strings.HasSuffix(f.Name, "Lines") {
				var result []string
				for _, arg := range n.Args {
					result = append(result, text(arg))
				}
				return strings.Join(result, "\n") + "\n"
			}
			if f, ok := n.Fun.(*goast.SelectorExpr); ok && f.Sel.Name == "Join" && len(n.Args) == 2 {
				return strings.Join(lines(n.Args[0]), text(n.Args[1]))
			}
		}
		return ""
	}
	for _, name := range []string{"correctness_no_process_exit_after_output", "correctness_no_uncleared_race_timeout", "correctness_require_blocking_standard_streams"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/nexus", name+"_test.go"), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		trees[name] = tree
		goast.Inspect(tree, func(n goast.Node) bool {
			if assignment, ok := n.(*goast.AssignStmt); ok && len(assignment.Lhs) == len(assignment.Rhs) {
				for at, lhs := range assignment.Lhs {
					if name, ok := lhs.(*goast.Ident); ok {
						if value := text(assignment.Rhs[at]); value != "" {
							values[name.Name] = value
						}
					}
				}
			}
			if value, ok := n.(*goast.ValueSpec); ok && len(value.Names) == len(value.Values) {
				for at, name := range value.Names {
					values[name.Name] = text(value.Values[at])
				}
			}
			return true
		})
	}
	write := func(path, source string) string {
		physical := filepath.Join(h.directory, "controls", strings.TrimSuffix(path, ".ts")+".a")
		if err := os.MkdirAll(filepath.Dir(physical), 0755); err != nil {
			h.t.Fatal(err)
		}
		if err := os.WriteFile(physical, []byte(source), 0644); err != nil {
			h.t.Fatal(err)
		}
		logical := strings.TrimSuffix(physical, ".a") + ".ts"
		if _, err := os.Lstat(logical); os.IsNotExist(err) {
			if err := os.Symlink(filepath.Base(physical), logical); err != nil {
				h.t.Fatal(err)
			}
		} else if err != nil {
			h.t.Fatal(err)
		} else if target, err := os.Readlink(logical); err != nil || target != filepath.Base(physical) {
			h.t.Fatal("unexpected fixture link", logical, err)
		}
		return logical
	}
	write("types/node.d.ts", values["correctnessNoProcessExitAfterOutputNodeTypes"])
	write("types/web-console.d.ts", values["correctnessNoProcessExitAfterOutputWebConsole"])
	write("types/timers.d.ts", values["correctnessNoUnclearedRaceTimeoutNodeTimers"])
	write("types/prelude.d.ts", values["correctnessRequireBlockingStandardStreamsPrelude"])
	var roots []string
	for _, row := range []struct{ name, prefix string }{{"correctness_no_process_exit_after_output", "correctnessNoProcessExitAfterOutput"}, {"correctness_no_uncleared_race_timeout", "correctnessNoUnclearedRaceTimeout"}} {
		var sources []string
		parents := map[goast.Node]goast.Node{}
		var stack []goast.Node
		goast.Inspect(trees[row.name], func(n goast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return false
			}
			if len(stack) > 0 {
				parents[n] = stack[len(stack)-1]
			}
			stack = append(stack, n)
			return true
		})
		goast.Inspect(trees[row.name], func(n goast.Node) bool {
			if call, ok := n.(*goast.CallExpr); ok {
				if f, ok := call.Fun.(*goast.Ident); ok && f.Name == row.prefix+"Source" {
					var result []string
					good := true
					for _, arg := range call.Args {
						if _, ok := arg.(*goast.BasicLit); !ok {
							good = false
						}
						result = append(result, text(arg))
					}
					if good {
						sources = append(sources, strings.Join(result, "\n"))
					}
				}
			}
			if list, ok := n.(*goast.CompositeLit); ok {
				parent, table := parents[list].(*goast.CompositeLit)
				input := table && len(parent.Elts) >= 2 && parent.Elts[1] == list
				if pair, ok := parents[list].(*goast.KeyValueExpr); ok {
					if key, ok := pair.Key.(*goast.Ident); ok {
						input = key.Name == "lines" || key.Name == "before" || key.Name == "after"
					}
				}
				if !input {
					return true
				}
				if typ, ok := list.Type.(*goast.ArrayType); ok {
					if item, ok := typ.Elt.(*goast.Ident); ok && item.Name == "string" {
						good := true
						for _, e := range list.Elts {
							if _, ok := e.(*goast.BasicLit); !ok {
								good = false
							}
						}
						joined := strings.Join(lines(list), "\n")
						if good && !strings.HasPrefix(joined, "declare") && !strings.HasPrefix(joined, "export {};") && (strings.Contains(joined, "process.exit(") || strings.Contains(joined, "Promise.race(")) {
							sources = append(sources, joined)
						}
					}
				}
			}
			return true
		})
		seen := map[string]bool{}
		for _, source := range sources {
			if seen[source] {
				continue
			}
			seen[source] = true
			directory := fmt.Sprintf("%s-%03d/source", row.name, len(roots))
			write(directory+"/Help.ts", values["correctnessNoProcessExitAfterOutputHelpModule"])
			write(directory+"/ScriptHelp.ts", values["correctnessNoProcessExitAfterOutputScript"])
			roots = append(roots, write(directory+"/Subject.ts", values[row.prefix+"Prelude"]+source+"\nexport {};\n"))
		}
	}
	// Preserve each blocking rule fixture's import graph and logical module identity.
	goast.Inspect(trees["correctness_require_blocking_standard_streams"], func(n goast.Node) bool {
		list, ok := n.(*goast.CompositeLit)
		if !ok {
			return true
		}
		fields := map[string]goast.Expr{}
		for _, e := range list.Elts {
			if pair, ok := e.(*goast.KeyValueExpr); ok {
				if key, ok := pair.Key.(*goast.Ident); ok {
					fields[key.Name] = pair.Value
				}
			}
		}
		typed, explicit := list.Type.(*goast.Ident)
		ownCase := explicit && typed.Name == "correctnessRequireBlockingStandardStreamsCase"
		if fields["lines"] == nil || (fields["name"] == nil && !ownCase) {
			return true
		}
		base := fmt.Sprintf("blocking-%03d/", len(roots))
		write(base+"libraries/nexus/source/system/StandardStreams.ts", values["correctnessRequireBlockingStandardStreamsNexus"])
		write(base+"libraries/nexus/source/command-line/CommandLineInterface.ts", values["correctnessRequireBlockingStandardStreamsCommandLine"])
		source := ""
		if b, ok := fields["shebang"].(*goast.Ident); ok && b.Name == "true" {
			source = "#!/usr/bin/env tsx\n"
		}
		source += strings.Join(lines(fields["imports"]), "\n") + "\n" + strings.Join(lines(fields["lines"]), "\n") + "\n"
		roots = append(roots, write(base+"modules/subject/Subject.ts", source))
		if others, ok := fields["others"].(*goast.CompositeLit); ok {
			for _, e := range others.Elts {
				pair := e.(*goast.KeyValueExpr)
				contents := text(pair.Value)
				if contents == "" {
					h.t.Fatal("unresolved fixture input", text(pair.Key))
				}
				roots = append(roots, write(base+strings.TrimPrefix(text(pair.Key), "/repository/"), contents))
			}
		}
		return true
	})
	roots = append(roots, write("edge/Subject.ts", "/* 世界 🌍 */\r\nconsole.log('before'); if(flag) process.exit(1); else process.exit(2);\r\nPromise.race([work(),new Promise((resolve)=>{ setTimeout(resolve, milliseconds); })]);\r\nexport {};\r\n"))
	roots = append(roots, write("bound-before/Subject.ts", values["correctnessNoUnclearedRaceTimeoutPrelude"]+"export async function bounded() { await Promise.race([work(), new Promise((resolve) => { setTimeout(resolve, milliseconds); })]); }\n"))
	roots = append(roots, write("bound-after/Subject.ts", values["correctnessNoUnclearedRaceTimeoutPrelude"]+"export async function bounded() { let timer: NodeJS.Timeout | number | undefined; try { await Promise.race([work(), new Promise((resolve) => { timer = setTimeout(resolve, milliseconds); })]); } finally { clearTimeout(timer); } }\n"))
	if len(roots) < 75 {
		h.t.Fatalf("missing controls: %d", len(roots))
	}
	h.t.Logf("%d roots including imported graph fixtures", len(roots))
	return roots
}

func TestWave25ThirdAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE25_THIRD_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_25_third_suite.a")
	binary := h.build(stage0, "wave25-third", entry, archive, false)
	oracle := volumeOracle(h, "wave25-third-oracle", "oracle_wave_25_third.go")
	config := h.write("controls-config.json", fmt.Sprintf(`{"extends":%q,"compilerOptions":{"target":"ES2024","lib":["ES2024","DOM"],"moduleDetection":"auto"},"include":["controls/**/*.d.ts"]}`, filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")))
	paths := wave25ThirdControls(h)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-process-exit-after-output", "no-uncleared-race-timeout", "require-blocking-standard-streams"} {
		if !bytes.Contains(truth.stdout, []byte("\tnexus/correctness-"+name+"\t")) {
			t.Fatal("no positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave25-third-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, mutation := range []struct{ name, file, from, to string }{
		{"exit-state", "process_exit_after_output.a", "state.length > 0", "state.length > 999"},
		{"race-drop", "uncleared_race_timeout.a", "if(['ExpressionStatement', 'VoidExpression'].includes(node.kind) || parent === executor) { return true; }", "if(['ExpressionStatement', 'VoidExpression'].includes(node.kind) || parent === executor) { return false; }"},
		{"blocking-count", "blocking_standard_streams.a", "this.exits.length > 1 ?", "this.exits.length > 999 ?"},
	} {
		mutantDirectory := filepath.Join(directory, mutation.name)
		if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"wave_25_third_suite.a", "blocking_standard_streams.a", "process_exit_after_output.a", "uncleared_race_timeout.a", "process_output_facts.a", "syntax_projection.a", "syntax_control_flow.a", "process_symbol_details.a", "resolved_call_declaration.a", "program_imports.a"} {
			data, err := os.ReadFile(filepath.Join(filepath.Dir(entry), file))
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if file == mutation.file {
				if strings.Count(source, mutation.from) != 1 {
					t.Fatal("nonunique mutant", mutation.name)
				}
				source = strings.Replace(source, mutation.from, mutation.to, 1)
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
			for _, dependency := range []string{"unary_minus", "rules", "facts", "frames", "diagnostic", "type_fact"} {
				source = strings.ReplaceAll(source, "./"+dependency+".ts", filepath.Join(repository, "stage1/cohere/typeaware", dependency+".ts"))
			}
			if err := os.WriteFile(filepath.Join(mutantDirectory, file), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
		}
		mutant := h.build(stage0, mutation.name+"-native", filepath.Join(mutantDirectory, "wave_25_third_suite.a"), archive, false)
		got := h.must(mutation.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", mutation.name)
		}
		t.Logf("%s exits 0; only Go byte comparison catches byte %d", mutation.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, population := range []struct{ name, root, config string }{{"repository", repository, filepath.Join(repository, "tsconfig.json")}, {"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")}} {
		if population.root == "" {
			t.Log("compiler corpus not supplied")
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.name+".manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var roots []string
		for _, path := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			roots = append(roots, filepath.Join(population.root, path))
		}
		corpus := h.write(population.name+".manifest", strings.Join(roots, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, corpus)
		h.compare(population.name+"-asan", oracle, asan, population.config, corpus)
		want := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, corpus, "--count"))
		got := h.must(population.name+"-timed-native", exec.Command(binary, population.config, corpus, "--count"))
		if !bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("timing count mismatch")
		}
		t.Logf("%s native %s Go %s", population.name, got.elapsed, want.elapsed)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic'; const args=programArguments();const path=args[1]??'';const program=tsgoProgram(args[0]??'',[path]);tsgoRelease(program);console.log(tsgoInspect(program,path,0,3,'SourceFile','syntax-projection'));`)
	stale := h.build(stage0, "released", released, archive, false)
	probe := h.write("released-probe.a", "x;\n")
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released syntax question: panic 70, invalid or released checker handle")
}
