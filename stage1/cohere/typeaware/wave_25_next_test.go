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

// Extract source strings from the production tables, not their expected answers.
func wave25NextControls(h *harness) []string {
	var roots []string
	for _, name := range []string{"collection_misuse", "discarded_outcome", "discarded_pure_result"} {
		file := filepath.Join(h.repository, "cohere/internal/lint/rules/nexus/correctness_no_"+name+"_test.go")
		tree, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		values := map[string]string{}
		var text func(goast.Expr) string
		text = func(expression goast.Expr) string {
			switch node := expression.(type) {
			case *goast.BasicLit:
				if node.Kind == token.STRING {
					value, err := strconv.Unquote(node.Value)
					if err != nil {
						h.t.Fatal(err)
					}
					return value
				}
			case *goast.Ident:
				return values[node.Name]
			case *goast.BinaryExpr:
				if node.Op == token.ADD {
					return text(node.X) + text(node.Y)
				}
			case *goast.CallExpr:
				if function, ok := node.Fun.(*goast.SelectorExpr); ok && function.Sel.Name == "Join" && len(node.Args) == 2 {
					if list, ok := node.Args[0].(*goast.CompositeLit); ok {
						var lines []string
						for _, line := range list.Elts {
							lines = append(lines, text(line))
						}
						return strings.Join(lines, text(node.Args[1]))
					}
				}
			}
			return ""
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			if value, ok := node.(*goast.ValueSpec); ok && len(value.Names) == 1 && len(value.Values) == 1 {
				values[value.Names[0].Name] = text(value.Values[0])
			}
			return true
		})
		prefix := "correctnessNo" + map[string]string{"collection_misuse": "CollectionMisuse", "discarded_outcome": "DiscardedOutcome", "discarded_pure_result": "DiscardedPureResult"}[name]
		prelude := values[prefix+"Prelude"]
		if prelude == "" {
			h.t.Fatal("missing production prelude", name)
		}
		if name == "discarded_outcome" {
			goast.Inspect(tree, func(node goast.Node) bool {
				value, ok := node.(*goast.ValueSpec)
				if !ok || len(value.Names) != 1 || value.Names[0].Name != prefix+"NexusFiles" {
					return true
				}
				list, ok := value.Values[0].(*goast.CompositeLit)
				if !ok {
					h.t.Fatal("missing Nexus source map")
				}
				for _, entry := range list.Elts {
					pair := entry.(*goast.KeyValueExpr)
					path := strings.TrimPrefix(text(pair.Key), "/repository/")
					physical := filepath.Join(h.directory, "controls", strings.TrimSuffix(path, ".ts")+".a")
					if err := os.MkdirAll(filepath.Dir(physical), 0755); err != nil {
						h.t.Fatal(err)
					}
					if err := os.WriteFile(physical, []byte(text(pair.Value)), 0644); err != nil {
						h.t.Fatal(err)
					}
					// The copied external TypeScript fixture keeps its oracle identity via a link.
					// Every actual source file written here has the .a extension.
					logical := strings.TrimSuffix(physical, ".a") + ".ts"
					if _, err := os.Lstat(logical); os.IsNotExist(err) {
						if err := os.Symlink(filepath.Base(physical), logical); err != nil {
							h.t.Fatal(err)
						}
					}
				}
				return false
			})
		}
		var sources []string
		goast.Inspect(tree, func(node goast.Node) bool {
			if call, ok := node.(*goast.CallExpr); ok {
				if function, ok := call.Fun.(*goast.Ident); ok && function.Name == prefix+"Source" {
					var lines []string
					good := true
					for _, arg := range call.Args {
						if _, ok := arg.(*goast.BasicLit); !ok {
							good = false
						}
						lines = append(lines, text(arg))
					}
					if good {
						sources = append(sources, strings.Join(lines, "\n"))
					}
				}
			}
			if pair, ok := node.(*goast.KeyValueExpr); ok {
				if key, ok := pair.Key.(*goast.Ident); ok && (key.Name == "before" || key.Name == "after") {
					list := pair.Value.(*goast.CompositeLit)
					var lines []string
					for _, line := range list.Elts {
						lines = append(lines, text(line))
					}
					sources = append(sources, strings.Join(lines, "\n"))
				}
			}
			// The table's lines field is a positional []string composite.
			if list, ok := node.(*goast.CompositeLit); ok {
				if typ, ok := list.Type.(*goast.ArrayType); ok {
					if element, ok := typ.Elt.(*goast.Ident); ok && element.Name == "string" {
						parentText := []string{}
						all := true
						for _, line := range list.Elts {
							literal, ok := line.(*goast.BasicLit)
							if !ok || literal.Kind != token.STRING {
								all = false
								break
							}
							parentText = append(parentText, text(line))
						}
						joined := strings.Join(parentText, "\n")
						if all && len(parentText) > 0 && (strings.Contains(joined, ";") || strings.Contains(joined, "{")) && !strings.HasPrefix(joined, "declare let name") && !strings.HasPrefix(joined, "type Role") && !strings.HasPrefix(joined, "import { parseJson }") && !strings.HasPrefix(joined, "export type ") {
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
			path := filepath.Join(h.directory, "controls/modules/meta", fmt.Sprintf("%s-%03d.a", name, len(roots)))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				h.t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(prelude+source+"\nexport {};\n"), 0644); err != nil {
				h.t.Fatal(err)
			}
			roots = append(roots, path)
		}
	}
	if len(roots) < 45 {
		h.t.Fatalf("missing production controls: %d", len(roots))
	}
	return roots
}

// Not parallel: native builds, sanitizer runs and timings share this machine.
func TestWave25NextAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE25_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_25_next_suite.a")
	binary := h.build(stage0, "wave25-next", entry, archive, false)
	oracle := volumeOracle(h, "wave25-next-oracle", "oracle_wave_25_next.go")
	config := h.write("controls-config.json", fmt.Sprintf(`{"extends":%q,"compilerOptions":{"target":"ES2024","lib":["ES2024"]}}`, filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")))
	paths := wave25NextControls(h)
	paths = append(paths, h.write("edge.a", "/* 世界 🌍 */\r\ndeclare const roles:string[];declare const key:'4294967294'|'4294967295'; '4294967295' in roles; key in roles; roles.length >= -0; roles.length < -0.1; roles.length < -(0x1); roles.length < +0; roles['length'] < 0; (roles.slice(1));\r\nexport {};\r\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"collection-misuse", "discarded-outcome", "discarded-pure-result"} {
		if !bytes.Contains(truth.stdout, []byte("\tnexus/correctness-no-"+name+"\t")) {
			t.Fatal("no positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave25-next-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, mutation := range []struct{ name, file, from, to string }{
		{"collection-zero", "collection_misuse.a", "return value <= 0;", "return value < 0;"},
		{"outcome-incomplete", "discarded_outcome.a", "(arms[at] ?? panic('missing outcome arms')).length === graph.node(unions[at] ?? 0).arms", "(arms[at] ?? panic('missing outcome arms')).length >= 1"},
		{"pure-callback", "discarded_pure_result.a", "if(count > 0)", "if(count > 999)"},
	} {
		mutantDirectory := filepath.Join(directory, mutation.name)
		if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"wave_25_next_suite.a", "collection_misuse.a", "discarded_outcome.a", "discarded_pure_result.a", "nexus_symbols.a", "type_declaration_ancestry.a", "nexus_collection_misuse_messages.a", "nexus_discarded_outcome_messages.a", "nexus_discarded_pure_result_messages.a"} {
			data, err := os.ReadFile(filepath.Join(filepath.Dir(entry), file))
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if file == mutation.file {
				if strings.Count(source, mutation.from) != 1 {
					t.Fatal("nonunique mutant")
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
		mutant := h.build(stage0, mutation.name+"-native", filepath.Join(mutantDirectory, "wave_25_next_suite.a"), archive, false)
		got := h.must(mutation.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", mutation.name)
		}
		t.Logf("%s exits 0; only Go byte oracle catches byte %d", mutation.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, population := range []struct{ name, root, config string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json")},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")},
	} {
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
		command := exec.Command(binary, population.config, corpus, "--count")
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(population.name+"-timed-native", command)
		if !bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("timing count mismatch")
		}
		t.Logf("%s native %s Go %s; %s", population.name, got.elapsed, want.elapsed, got.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const path=args[1]??'';const program=tsgoProgram(args[0]??'',[path]);tsgoRelease(program);
console.log(tsgoInspect(program,path,0,1,'Identifier','type-declaration-ancestry\nraw'));`)
	stale := h.build(stage0, "released", released, archive, false)
	probe := h.write("released-probe.a", "x;\n")
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released ancestry question: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released program.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-registry-native", released, mutantArchive, false)
	h.must("released-registry-run", exec.Command(mutant, config, probe))
	t.Log("released registry mutant exits 0; required panic 70 catches it")
}
