// Package oracles owns estree's compiler-free Go reference products.
package oracles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

type ComparisonInput struct {
	ID, Text, Filename string
	Audit              bool
}
type ComparisonAnswer struct {
	Status string
	Data   []byte
}

// EnableProductProduction is called only by this package's product-test binary.
func EnableProductProduction() { oracleProducer = true }

var oracleProducer bool

func oracleProduct(t *testing.T, inputs buildcache.Inputs, build func(string) error) string {
	t.Helper()
	inputs.Name += " oracle-owner-v1"
	return buildcache.Product(t, inputs, func(directory string) error {
		if !oracleProducer {
			return fmt.Errorf("oracle %s is missing: settle stage1/cohere/estree/oracles owners first", inputs.Name)
		}
		return build(directory)
	})
}

func oracleInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	return buildcache.Inputs{Name: name, Files: []string{"cohere", "go.mod", "go.work", "stage1/cohere/estree/testdata/oracle.go"},
		Flags:     []string{"repository=" + root(t), "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED")},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("go", "version")}}
}

// oracleExecute reads the oracle's output through childguard.Capture, to its end with no timer deciding when: exec's
// copying under a two-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF two seconds
// after the oracle exited, so a fixture product whose run had finished could fail under load (#n3gk22h's shape). The
// run had no stall guard to lose (Unbounded), and t.Context() and the unit's own ceiling bound it.
func oracleExecute(t *testing.T, oracle string, arguments ...string) []byte {
	t.Helper()
	stdout, stderr, err := childguard.Capture(t.Context(), toolchain.CommandContextIn(t.Context(), "", oracle, arguments...))
	if err != nil || len(stderr) != 0 {
		t.Fatalf("oracle %v: %v\n%s", arguments, err, stderr)
	}
	return stdout
}

var estreeGoOracleReady struct {
	once   sync.Once
	binary string
}

func RecoveredExpressions() []string {
	return []string{"const C = @d class {};", "const C = @d @e(1) class Named extends Base { static x=1; };", "(@d class { m(){return 1;} });", "function f(){ await 'x'; await 1; await true; await this; }", "function f(){ yield 'x'; yield 1; yield true; }", "await\nx; yield\nx;", "await(x); yield(x);"}
}

func RecoveredGrammar() []string {
	return []string{
		"class C { readonly!: number; static #x=1; 1n=2; }",
		"class C { private public x: number; protected public m(){} }",
		"class C { constructor(static x: number, readonly y: number){} }",
		"type X = { static x: number; new: string; #x: number; [a,b]: number; };",
		"({ x=1, y?, z?:2 });",
		"f<>(); new C<>(); class C<> {} function f<>(){}",
		"class C { constructor<>(){} }",
		"class C implements string, number {}",
		"class C extends A extends B implements X implements Y {}",
		"export {}; await(x); await 1; function f(){await(x);} namespace N {await(x);}",
		"export default async(x); export default async () => 1; export default async;",
		"declare\nmodule\n'x'\n{}",
		"let obj = /** @satisfies {{ f(s:string):void }} */ ({f(s){}});",
		"const expected = /** @type {{name:string}} */ (null);",
		"a ? (b) : c => (d) : e => f;",
		"class C { static { await(x); class D { x=await(x); } } }",
	}
}

func RecoveryLibraryGaps() []string { return []string{"f<>();", "class C<> {}"} }

func DecoratedExports() []string {
	return []string{"@d export class C {}", "@d export default class C {}", "export @d class C {}", "export default @d class C {}", "@a @b() export abstract class C {}", "// first\n@d\nexport default class {}"}
}

func JSXCases() []string {
	return []string{"const x=<A<>/>;", "const x = <A/>;", "const x = <A> text </A>;", "const x = <><A/>{x}{/* comment */}</>;", "const x = <A.B.C ns:x='&amp;&#x1f600;' data-test={x} bool {...rest}><ns:a/>{...children}</A.B.C>;", "const x = <A<T,> x=\"\\n\"/>;", "const x = <A> // text /* not comment */ &amp;&unknown;&#1114112; </A>;", "let x = <A value={<B/>}/>;", "<A>\n é😀\r\n {x+1}\n</A>;"}
}

func BoundedCases(t *testing.T) []ComparisonInput {
	t.Helper()
	cases := []ComparisonInput{
		{"stage1/cohere/estree/stalls_test.go:0:eof-type", "type X = {", "input.ts", false},
		{"stage1/cohere/estree/stalls_test.go:0:eof-interface", "interface I {", "input.ts", false},
		{"stage1/cohere/estree/stalls_test.go:0:eof-method", "type X = { m(a: string): void;", "input.ts", false},
	}
	fixtures, err := filepath.Glob("validation/followup/stalls/*.input")
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("stall fixtures: %d %v", len(fixtures), err)
	}
	for _, fixture := range fixtures {
		body, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, ComparisonInput{"stage1/cohere/estree/" + filepath.ToSlash(fixture) + ":0:audit", string(body), strings.TrimSuffix(filepath.Base(fixture), ".input"), true})
	}
	return cases
}

func root(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func manifest(t *testing.T, cases []string) string {
	t.Helper()
	dir := t.TempDir()
	var listing strings.Builder
	for index, text := range cases {
		path := filepath.Join(dir, fmt.Sprintf("%04d.ts", index))
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		listing.WriteString(path + "\n")
	}
	path := filepath.Join(dir, "manifest")
	if err := os.WriteFile(path, []byte(listing.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func miscDigest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// ParserOracle is the shared Go oracle recipe used by owners and consumers.
func ParserOracle(t *testing.T) string {
	t.Helper()
	estreeGoOracleReady.once.Do(func() {
		inputs := oracleInputs(t, "estree-go-oracle-v1")
		inputs.Name = "estree-go-oracle-v1"
		inputs.Flags = append(inputs.Flags, "go build "+strings.Join(buildcache.ReproducibleGoFlags(), " "))
		dir := oracleProduct(t, inputs, func(dir string) error {
			estreeGoOracleBuild(t, dir)
			return nil
		})
		estreeGoOracleReady.binary = filepath.Join(dir, "oracle")
	})
	if estreeGoOracleReady.binary == "" {
		t.Fatal("Go oracle preparation failed")
	}
	return estreeGoOracleReady.binary
}

func estreeGoOracleBuild(t *testing.T, dir string) string {
	t.Helper()
	repo := root(t)
	source, err := filepath.Abs("testdata/oracle.go")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(repo, "cohere/adamic_estree_oracle.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	defer os.Remove(path)
	binary := filepath.Join(dir, "oracle")
	command := toolchain.CommandContextIn(t.Context(), filepath.Join(repo, "cohere"), "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+path, "-o", binary, virtual)...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := childguard.Run(command, childguard.Options{Unbounded: true}); err != nil || stderr.Len() != 0 {
		t.Fatalf("go build %s: %v\n%s", virtual, err, &stderr)
	}
	return binary
}

// AcceptanceMutantsOracle is the shared Go oracle recipe used by owners and consumers.
func AcceptanceMutantsOracle(t *testing.T) string {
	t.Helper()
	acceptanceOracleOnce.Do(func() {
		inputs := buildcache.Inputs{Name: "estree-acceptance-mutants-go-oracle",
			Files:     []string{"stage1/cohere/estree", "cohere", "go.mod", "go.work"},
			Flags:     []string{"repository=" + root(t), "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ")},
			Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version")}}
		directory := oracleProduct(t, inputs, func(directory string) error {
			// Keep this copy product distinct from the underlying parser oracle.
			binary := ParserOracle(t)
			data, err := os.ReadFile(binary)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "oracle"), data, 0755)
		})
		acceptanceOraclePath = filepath.Join(directory, "oracle")
	})
	return acceptanceOraclePath
}

// ScalarEdgesOracle is the shared Go oracle recipe used by owners and consumers.
func ScalarEdgesOracle(t *testing.T) string {
	t.Helper()
	repo := root(t)
	source := filepath.Join(repo, "stage1/cohere/estree/testdata/oracle.go")
	virtual := filepath.Join(repo, "cohere/adamic_estree_oracle.go")
	product := oracleProduct(t, buildcache.Inputs{
		Name: "scalar-edges-go-oracle", Files: []string{"cohere", "stage1/cohere/estree/testdata/oracle.go"},
		Flags: []string{"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ") + " overlay oracle"}, Toolchain: []string{buildcache.Tool("go", "version")},
	}, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(path)
		command := toolchain.CommandContextIn(t.Context(), filepath.Join(repo, "cohere"), "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+path, "-o", filepath.Join(dir, "oracle"), virtual)...)
		output, err := childguard.CombinedOutput(command, childguard.Options{Unbounded: true})
		if err != nil {
			return fmt.Errorf("oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(product, "oracle")
}

// SyntaxMutantsOracle is the shared Go oracle recipe used by owners and consumers.
func SyntaxMutantsOracle(t *testing.T) string {
	t.Helper()
	inputs := oracleInputs(t, "syntax-mutants-go-oracle-v1")
	inputs.Name = "syntax-mutants-go-oracle-v1"
	inputs.Flags = append(inputs.Flags, "go build "+strings.Join(buildcache.ReproducibleGoFlags(), " "))
	return oracleProduct(t, inputs, func(dir string) error {
		data, err := os.ReadFile(ParserOracle(t))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "oracle"), data, 0755)
	})
}

// ThreePortMutantsOracle is the shared Go oracle recipe used by owners and consumers.
func ThreePortMutantsOracle(t *testing.T) string {
	t.Helper()
	inputs := oracleInputs(t, "three-port-mutants-go-oracle-v1")
	directory := oracleProduct(t, inputs, func(dir string) error {
		repo := root(t)
		source := filepath.Join(repo, "stage1/cohere/estree/testdata/oracle.go")
		virtual := filepath.Join(repo, "cohere/adamic_estree_oracle.go")
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(path)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		command := toolchain.CommandContextIn(ctx, "", "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+path, "-o", filepath.Join(dir, "oracle"), virtual)...)
		command.Dir = filepath.Join(repo, "cohere")
		command.Env = append(command.Env, "PWD="+command.Dir)
		// Read through childguard.Capture for oracleExecute's reason: a WaitDelay could fail a build that had
		// finished. The build had no stall guard to lose, and the product unit's ceiling bounds it.
		stdout, stderr, err := childguard.Capture(ctx, command)
		output := append(stdout, stderr...)
		if err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}

var acceptanceOracleOnce sync.Once
var acceptanceOraclePath string

// DeepMutantsFixture is the shared Go oracle recipe used by owners and consumers.
func DeepMutantsFixture(t *testing.T) string {
	t.Helper()
	deepMutantsFixturePrepared.once.Do(func() {
		oracle := ThreePortMutantsOracle(t)
		inputs := buildcache.Inputs{
			Name:      "deep-mutants-oracle-output-v2",
			Files:     []string{"stage1/cohere/estree/deep_mutants_product_shards_test.go"},
			Flags:     []string{DeepMutantsInput, "oracle=" + miscDigest(t, oracle)},
			Toolchain: []string{runtime.Version()},
		}
		deepMutantsFixturePrepared.directory = oracleProduct(t, inputs, func(dir string) error {
			list := manifest(t, []string{DeepMutantsInput})
			// Shared setup has no deadline of its own. Loom bounds the entire unit.
			output := oracleExecute(t, oracle, "--manifest", list)
			return os.WriteFile(filepath.Join(dir, "want"), output, 0644)
		})
	})
	if deepMutantsFixturePrepared.directory == "" {
		t.Fatal("deep fixture preparation failed")
	}
	return deepMutantsFixturePrepared.directory
}

const DeepMutantsInput = "type T=(A); a+b+c; const t=tag`a${b}c`; a&&(b&&c);"

var deepMutantsFixturePrepared struct {
	once      sync.Once
	directory string
}

func CachedComparisonAnswers(t *testing.T, oracle string, cases []ComparisonInput) []ComparisonAnswer {
	t.Helper()
	encoded, err := json.Marshal(casesToInputs(cases))
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "estree-misc-oracle-answers", Files: []string{"stage1/cohere/estree/misc_shards_test.go"}, Flags: []string{string(encoded)}, Toolchain: []string{runtime.Version(), miscDigest(t, oracle)}}
	directory := oracleProduct(t, inputs, func(directory string) error {
		answers := make([]ComparisonAnswer, len(cases))
		for i, item := range cases {
			subdir := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
			if err := os.Mkdir(subdir, 0755); err != nil {
				return err
			}
			path := filepath.Join(subdir, item.Filename)
			if err := os.WriteFile(path, []byte(item.Text), 0644); err != nil {
				return err
			}
			if item.Audit {
				list := filepath.Join(subdir, "manifest")
				if err := os.WriteFile(list, []byte(path+"\n"), 0644); err != nil {
					return err
				}
				output, err := miscReferenceCommand(t, oracle, "--audit", list, subdir)
				if err != nil {
					return err
				}
				if err := json.Unmarshal(output, &answers[i]); err != nil {
					return err
				}
			} else {
				answers[i].Status = "ok"
			}
			if answers[i].Status == "ok" {
				output, err := miscReferenceCommand(t, oracle, path)
				if err != nil {
					return err
				}
				answers[i].Data = output
			}
			if err := os.RemoveAll(subdir); err != nil {
				return err
			}
		}
		data, err := json.Marshal(answers)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "answers.json"), data, 0644)
	})
	data, err := os.ReadFile(filepath.Join(directory, "answers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var answers []ComparisonAnswer
	if err := json.Unmarshal(data, &answers); err != nil {
		t.Fatal(err)
	}
	if len(answers) != len(cases) {
		t.Fatal("oracle answer enumeration differs")
	}
	return answers
}

func casesToInputs(cases []ComparisonInput) [][]string {
	result := make([][]string, len(cases))
	for i, item := range cases {
		result[i] = []string{item.ID, item.Text, item.Filename, strconv.FormatBool(item.Audit)}
	}
	return result
}

func miscReferenceCommand(t *testing.T, oracle string, args ...string) ([]byte, error) {
	command := toolchain.CommandContextIn(t.Context(), "", oracle, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	var stdout bytes.Buffer
	command.Stdout = &stdout
	err := childguard.Run(command, childguard.Options{Unbounded: true})
	output := stdout.Bytes()
	if err != nil || stderr.Len() != 0 {
		return nil, fmt.Errorf("oracle %v: %v\n%s", args, err, &stderr)
	}
	return output, nil
}

// CachedRecoveryAnswer is the shared Go oracle recipe used by owners and consumers.
func CachedRecoveryAnswer(t *testing.T, oracle, list, mode string) []byte {
	t.Helper()
	if mode != "--manifest" && mode != "--audit" {
		t.Fatalf("unsupported recovery oracle mode %q", mode)
	}
	binary, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	manifestBytes, err := os.ReadFile(list)
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	flags := []string{mode, fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary))}
	for _, path := range strings.Split(string(manifestBytes), "\n") {
		if path == "" {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sources = append(sources, string(content))
		flags = append(flags, string(content))
	}
	inputs := buildcache.Inputs{Name: "estree-oracle-answer", Files: []string{"stage1/cohere/estree/recovery_cache_test.go"}, Flags: flags, Toolchain: []string{runtime.Version()}}
	directory := oracleProduct(t, inputs, func(directory string) error {
		work := filepath.Join(directory, "run")
		if err := os.MkdirAll(work, 0755); err != nil {
			return err
		}
		defer os.RemoveAll(work)
		var names []string
		for i, source := range sources {
			name := fmt.Sprintf("input-%06d.ts", i)
			if err := os.WriteFile(filepath.Join(work, name), []byte(source), 0644); err != nil {
				return err
			}
			names = append(names, name)
		}
		if err := os.WriteFile(filepath.Join(work, "manifest"), []byte(strings.Join(names, "\n")+"\n"), 0644); err != nil {
			return err
		}
		args := []string{mode, "manifest"}
		if mode == "--audit" {
			args = append(args, "answers")
		}
		command := toolchain.CommandContextIn(t.Context(), "", oracle, args...)
		command.Dir = work
		var stderr strings.Builder
		command.Stderr = &stderr
		cold := time.Now()
		var stdout bytes.Buffer
		command.Stdout = &stdout
		err := childguard.Run(command, childguard.Options{Unbounded: true})
		answer := stdout.Bytes()
		t.Logf("BUILD estree-oracle-answer cold wall=%.6fs", time.Since(cold).Seconds())
		if err != nil || stderr.Len() != 0 {
			return fmt.Errorf("oracle answer: %v: %s", err, &stderr)
		}
		return os.WriteFile(filepath.Join(directory, "answer"), answer, 0644)
	})
	answer, err := os.ReadFile(filepath.Join(directory, "answer"))
	if err != nil {
		t.Fatal(err)
	}
	return answer
}
