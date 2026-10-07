package wave28fifth

import (
	"bytes"
	"encoding/json"
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

func TestUndefAgreement(t *testing.T) {
	stage0 := os.Getenv("ADAMIC_WAVE28_STAGE0")
	if stage0 == "" {
		t.Skip("set ADAMIC_WAVE28_STAGE0 for native rule comparison")
	}
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE28_UNDEF_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	archive := os.Getenv("ADAMIC_WAVE28_ARCHIVE")
	if archive == "" {
		archive = h.archive("checker", "", false)
	}
	binary := h.build(stage0, "undef", filepath.Join(repository, "stage1/cohere/typeaware/wave28_fifth/undef_suite.a"), archive, false)
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repository, "cohere/adamic_wave28_fifth.go"): filepath.Join(repository, "stage1/cohere/typeaware/wave28_fifth/oracle.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := h.write("oracle-overlay.json", string(overlay))
	oracle := filepath.Join(directory, "oracle")
	cmd := exec.Command("go", "build", "-overlay", overlayPath, "-o", oracle, "./adamic_wave28_fifth.go")
	cmd.Dir = filepath.Join(repository, "cohere")
	h.must("oracle-build", cmd)
	h.write("globals.d.ts", "declare var Text: any;\n")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","jsx":"preserve","skipLibCheck":true},"include":["globals.d.ts","*.tsx"]}`)
	tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/react/jsx_no_undef_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	sources := []string{}
	seen := map[string]bool{}
	goast.Inspect(tree, func(node goast.Node) bool {
		lit, ok := node.(*goast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(s, "<") && !seen[s] {
			seen[s] = true
			sources = append(sources, s)
		}
		return true
	})
	sources = append(sources, "// 💩\nexport {}; <Missing />;", "export {}; <테스트 />;", "export {}; <div />; <Foo-bar />; <Missing />;", "export {}; <Map />;")
	paths := []string{}
	for i, source := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.tsx", i), "export {};\n"+source+"\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	compare := func(name, native, conf, files string, allow bool) result {
		args := []string{conf, files}
		goargs := append([]string{}, args...)
		goargs = append(goargs, "--only-undef")
		if allow {
			args = append(args, "--allow-globals")
			goargs = append(goargs, "--allow-globals")
		}
		want := h.must(name+"-go", exec.Command(oracle, goargs...))
		got := h.must(name+"-native", exec.Command(native, args...))
		if len(got.stderr) != 0 {
			t.Fatalf("%s native stderr %s", name, got.stderr)
		}
		if !bytes.Equal(want.stdout, got.stdout) {
			at := firstDifference(want.stdout, got.stdout)
			t.Fatalf("%s mismatch byte %d: Go %q native %q", name, at, want.stdout[max(0, at-60):min(len(want.stdout), at+250)], got.stdout[max(0, at-60):min(len(got.stdout), at+250)])
		}
		t.Logf("%s: %d bytes, %s, native %s Go %s", name, len(want.stdout), summary(want.stdout), got.elapsed, want.elapsed)
		return want
	}
	want := compare("controls", binary, config, manifest, false)
	if !strings.Contains(string(want.stdout), "jsxIdentifierNotDefined") {
		t.Fatal("no positive finding")
	}
	t.Logf("%d control inputs", len(sources))
	compare("globals", binary, config, manifest, true)
	rulePath := filepath.Join(repository, "stage1/cohere/typeaware/wave28_fifth/jsx-no-undef/rule.a")
	data, err := os.ReadFile(rulePath)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	for _, name := range []string{"../../../../typescript/parser/nodes.ts", "../../rules.ts", "../../frames.ts", "../../diagnostic.ts"} {
		source = strings.ReplaceAll(source, name, filepath.Clean(filepath.Join(filepath.Dir(rulePath), name)))
	}
	source = strings.Replace(source, "first >= 97 && first <= 122", "first >= 0 && first <= 65535", 1)
	mutantRule := h.write("mutant-rule.a", source)
	suitePath := filepath.Join(repository, "stage1/cohere/typeaware/wave28_fifth/undef_suite.a")
	data, err = os.ReadFile(suitePath)
	if err != nil {
		t.Fatal(err)
	}
	source = string(data)
	for _, name := range []string{"../../../typescript/parser/nodes.ts", "../../../typescript/parser/parser.ts", "../../../typescript/scanner/scanner.ts", "../unary_minus.ts", "../rules.ts", "./jsx-fragments/rule.a"} {
		source = strings.ReplaceAll(source, name, filepath.Clean(filepath.Join(filepath.Dir(suitePath), name)))
	}
	source = strings.Replace(source, "./jsx-no-undef/rule.a", mutantRule, 1)
	mutantEntry := h.write("mutant-suite.a", source)
	mutant := h.build(stage0, "mutant", mutantEntry, archive, false)
	got := h.must("mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(want.stdout, got.stdout) {
		t.Fatalf("mutant invalid or survived: %s", got.stderr)
	}
	t.Logf("intrinsic-classification mutant exits zero, empty stderr; Go bytes catch byte %d", firstDifference(want.stdout, got.stdout))
	sanitizedArchive := os.Getenv("ADAMIC_WAVE28_SANITIZED_ARCHIVE")
	if sanitizedArchive == "" {
		sanitizedArchive = h.archive("checker-asan", "", true)
	}
	sanitized := h.build(stage0, "undef-asan", suitePath, sanitizedArchive, true)
	compare("controls-asan", sanitized, config, manifest, false)
	compare("globals-asan", sanitized, config, manifest, true)
	releasedEntry := h.write("released.a", `import { programArguments, tsgoProgram, tsgoRelease, tsgoInspect } from 'adamic';
const args = programArguments(); const path = args[1] ?? ''; const program = tsgoProgram(args[0] ?? '', [path]); tsgoRelease(program); console.log(tsgoInspect(program, path, 0, 1, 'Identifier', 'reference-symbol'));
`)
	for _, sanitized := range []bool{false, true} {
		lib := archive
		name := "released"
		if sanitized {
			lib = sanitizedArchive
			name += "-asan"
		}
		probe := h.build(stage0, name, releasedEntry, lib, sanitized)
		r := h.run(name+"-run", exec.Command(probe, config, paths[0]))
		exit, ok := r.err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || !strings.Contains(string(r.stderr), "invalid or released checker handle") {
			t.Fatalf("released checker guard %v %s", r.err, r.stderr)
		}
		if strings.Contains(string(r.stderr), "AddressSanitizer") || strings.Contains(string(r.stderr), "runtime error:") {
			t.Fatalf("released sanitizer error %s", r.stderr)
		}
		t.Logf("%s: required panic 70, no sanitizer findings", name)
	}
	for _, population := range []string{"repository", "compiler"} {
		files := filepath.Join(repository, "../wave-28-artifacts", population+".manifest")
		conf := filepath.Join(repository, "tsconfig.json")
		if population == "compiler" {
			conf = filepath.Join(repository, "../wave-28-corpus/src/compiler/tsconfig.json")
		}
		compare(population, binary, conf, files, false)
		compare(population+"-asan", sanitized, conf, files, false)
	}
}
