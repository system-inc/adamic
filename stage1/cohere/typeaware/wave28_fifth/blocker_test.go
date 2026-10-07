package wave28fifth

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// A blocker reproduction, not a rule parity gate. Outputs always go to files.
func TestSharedParserBlocker(t *testing.T) {
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE28_FIFTH_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	stage0 := os.Getenv("ADAMIC_WAVE28_STAGE0")
	if stage0 == "" {
		t.Skip("set ADAMIC_WAVE28_STAGE0 to run native parser blocker reproduction")
	}
	write := func(name, text string) string {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	run := func(name string, command *exec.Cmd) ([]byte, []byte, error) {
		out, err := os.Create(filepath.Join(directory, name+".stdout"))
		if err != nil {
			t.Fatal(err)
		}
		defer out.Close()
		report, err := os.Create(filepath.Join(directory, name+".stderr"))
		if err != nil {
			t.Fatal(err)
		}
		defer report.Close()
		command.Stdout, command.Stderr = out, report
		if command.Dir == "" {
			command.Dir = repository
		}
		failure := command.Run()
		stdout, err := os.ReadFile(out.Name())
		if err != nil {
			t.Fatal(err)
		}
		stderr, err := os.ReadFile(report.Name())
		if err != nil {
			t.Fatal(err)
		}
		return stdout, stderr, failure
	}
	binary := filepath.Join(directory, "parser-probe")
	if _, stderr, err := run("native-build", exec.Command(stage0, "build", filepath.Join(repository, "stage1/cohere/typeaware/wave28_fifth/parser_probe.a"), "-o", binary)); err != nil {
		t.Fatalf("build: %v %s", err, stderr)
	}
	oracle := filepath.Join(directory, "oracle")
	overlayBytes, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repository, "cohere/adamic_wave28_fifth.go"): filepath.Join(repository, "stage1/cohere/typeaware/wave28_fifth/oracle.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := write("overlay.json", string(overlayBytes))
	command := exec.Command("go", "build", "-overlay", overlay, "-o", oracle, "./adamic_wave28_fifth.go")
	command.Dir = filepath.Join(repository, "cohere")
	if _, stderr, err := run("oracle-build", command); err != nil {
		t.Fatalf("oracle: %v %s", err, stderr)
	}
	write("seed.d.ts", `declare const React: any;
declare const Ctx: any;
`)
	config := write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","jsx":"preserve","skipLibCheck":true},"include":["seed.d.ts","*.tsx"]}`)
	probePath := filepath.Join(repository, "stage1/cohere/typeaware/wave28_fifth/parser_probe.a")
	probe, err := os.ReadFile(probePath)
	if err != nil {
		t.Fatal(err)
	}
	mutantSource := strings.Replace(string(probe), "../../../typescript/parser/parser.ts", filepath.Join(repository, "stage1/typescript/parser/parser.ts"), 1)
	mutantSource = strings.Replace(mutantSource, "parser.file();", "// Mutant skips parsing and exits successfully.", 1)
	mutantEntry := write("mutant.a", mutantSource)
	mutant := filepath.Join(directory, "parser-mutant")
	if _, stderr, err := run("mutant-build", exec.Command(stage0, "build", mutantEntry, "-o", mutant)); err != nil {
		t.Fatalf("mutant build: %v %s", err, stderr)
	}
	names := []string{"fragments", "context", "undef"}
	sources := []string{
		`declare const React: any; const element = <React.Fragment><div /></React.Fragment>;`,
		`declare const Ctx: any; function Component() { return <Ctx.Provider value={{a: 1}} />; }`,
		`const element = <MissingComponent />;`,
	}
	rules := []string{"react/jsx-fragments", "react/jsx-no-constructed-context-values", "react/jsx-no-undef"}
	for i, source := range sources {
		path := write(names[i]+".tsx", source+"\n")
		manifest := write(names[i]+".manifest", path+"\n")
		stdout, stderr, err := run(names[i]+"-go", exec.Command(oracle, config, manifest))
		if err != nil || !strings.Contains(string(stdout), "\t"+rules[i]+"\t") || !strings.HasSuffix(string(stdout), "findings 1\n") {
			t.Fatalf("Go positive %s: %v %s %s", rules[i], err, stdout, stderr)
		}
		if strings.Contains(string(stderr), "parse diagnostics") {
			t.Fatalf("invalid fixture: %s", stderr)
		}
		stdout, stderr, err = run(names[i]+"-native", exec.Command(binary, path))
		if err == nil || !strings.Contains(string(stderr), "adamic: panic: parser slice") {
			t.Fatalf("missing JSX blocker %s: %v %s %s", rules[i], err, stdout, stderr)
		}
		t.Logf("%s: Go 1 finding, native parser rejects JSX: %s", rules[i], strings.TrimSpace(string(stderr)))
		stdout, stderr, err = run(names[i]+"-mutant", exec.Command(mutant, path))
		if err != nil || string(stdout) != "parsed\n" || len(stderr) != 0 {
			t.Fatalf("mutant must exit normally: %v %s %s", err, stdout, stderr)
		}
		t.Logf("%s: parser bypass mutant exits 0 and is rejected by expected-rejection comparison", rules[i])
	}
	stdout, stderr, err := run("listener-kinds", exec.Command(oracle, config, filepath.Join(directory, "fragments.manifest"), "--listener-kinds"))
	if err != nil {
		t.Fatalf("listener kinds: %v %s", err, stderr)
	}
	var expected map[string][]int
	if err := json.Unmarshal(stdout, &expected); err != nil {
		t.Fatal(err)
	}
	for _, name := range rules {
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/wave28_fifth", strings.TrimPrefix(name, "react/"), "rule.json"))
		if err != nil {
			t.Fatal(err)
		}
		var declaration struct {
			Name  string `json:"name"`
			Kinds []int  `json:"kinds"`
		}
		if err := json.Unmarshal(data, &declaration); err != nil {
			t.Fatal(err)
		}
		matches := func(kinds []int) bool { return declaration.Name == name && slices.Equal(kinds, expected[name]) }
		if !matches(declaration.Kinds) {
			t.Fatalf("%s numeric listener mismatch: %v versus %v", name, declaration.Kinds, expected[name])
		}
		mutant := slices.Clone(declaration.Kinds)
		mutant[0] = 0
		if matches(mutant) {
			t.Fatalf("%s listener mutant survived", name)
		}
		t.Logf("%s listener kinds %v agree with production Go; wrong-kind mutant caught", name, declaration.Kinds)
	}
	ordinary := write("ordinary.a", "function useThing() { return 1; }\n")
	stdout, stderr, err = run("ordinary-native", exec.Command(binary, ordinary))
	if err != nil || string(stdout) != "parsed\n" || len(stderr) != 0 {
		t.Fatalf("ordinary source: %v %s %s", err, stdout, stderr)
	}
	t.Log("ordinary non-JSX parser control passes")
}
