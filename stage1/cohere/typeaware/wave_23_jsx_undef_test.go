package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: native and sanitized builds share compiler resources.
func TestWave23JsxNoUndefAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE23_JSX_UNDEF_ARTIFACTS")
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
	base := filepath.Join(repository, "stage1/cohere/typeaware")
	entry := filepath.Join(base, "wave_23_jsx_undef_probe.a")
	binary := h.build(stage0, "native", entry, archive, false)
	oracle := volumeOracle(h, "jsx-undef-oracle", "oracle_wave_23_jsx_undef.go")
	h.write("prelude.d.ts", "")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","jsx":"preserve","lib":["ESNext"],"noEmit":true},"files":["prelude.d.ts"]}`)
	controls := []string{
		"function C(){return <Missing/>;} export{};",
		"const App=()=>null; function C(){return <App/>;} export{};",
		"function C(){return <div/>;} export{};",
		"function C(){return <Foo-bar/>;} export{};",
		"function C(){return <_foo/>;} export{};",
		"function C(){return <$foo/>;} export{};",
		"function C(){return <테스트/>;} export{};",
		"function C(){return <app.Foo/>;} export{};",
		"function C(){return <this.foo/>;} export{};",
		"function C(){return <Map/>;} export{};",
		"function C(){return <A:B/>;} export{};",
		"{const App=null;} function C(){return <App/>;} export{};",
	}
	var paths []string
	for i, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%02d.tsx", i), source+"\n"))
	}
	manifest := h.write("manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\treact/jsx-no-undef\t")) {
		t.Fatal("missing positive finding")
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "native-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	want := h.must("allow-globals-go", exec.Command(oracle, config, manifest, "--allow-globals"))
	if bytes.Equal(want.stdout, truth.stdout) {
		t.Fatal("globals option has no observable effect")
	}
	for name, exe := range map[string]string{"native": binary, "asan": asan} {
		got := h.must("allow-globals-"+name, exec.Command(exe, config, manifest, "--allow-globals"))
		if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("globals option differs", name)
		}
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE23_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE23_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		a := h.must(corpus.name+"-timing-go", exec.Command(oracle, corpus.config, corpus.manifest))
		b := h.must(corpus.name+"-timing-native", exec.Command(binary, corpus.config, corpus.manifest))
		if !bytes.Equal(a.stdout, b.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s timing: native %s, Go %s", corpus.name, b.elapsed, a.elapsed)
	}
	data, err := os.ReadFile(filepath.Join(base, "jsx_no_undef.a"))
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(string(data), "first >= 97 && first <= 122", "first >= 65 && first <= 90", 1)
	for _, rel := range []string{"../../typescript/parser/nodes.ts", "./rules.ts", "./frames.ts", "./diagnostic.ts"} {
		source = strings.ReplaceAll(source, "'"+rel+"'", "'"+filepath.Clean(filepath.Join(base, rel))+"'")
	}
	module := h.write("mutant-rule.a", source)
	data, err = os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	probe := string(data)
	for _, rel := range []string{"../../typescript/parser/nodes.ts", "../../typescript/parser/parser.ts", "../../typescript/scanner/scanner.ts", "./unary_minus.ts", "./rules.ts", "./jsx_no_undef.a"} {
		path := filepath.Clean(filepath.Join(base, rel))
		if rel == "./jsx_no_undef.a" {
			path = module
		}
		probe = strings.ReplaceAll(probe, "'"+rel+"'", "'"+path+"'")
	}
	mutant := h.build(stage0, "mutant", h.write("mutant-probe.a", probe), archive, false)
	got := h.must("mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("classification mutant survived")
	}
	t.Logf("jsx-no-undef mutant: exit 0, empty stderr, Go byte oracle catches byte %d", firstDifference(got.stdout, truth.stdout))
	probe = strings.ReplaceAll(probe, "'"+module+"'", "'"+filepath.Join(base, "jsx_no_undef.a")+"'")
	probe = strings.Replace(probe, "let count = 0;", "tsgoRelease(program);\nlet count = 0;", 1)
	released := h.build(stage0, "released", h.write("released.a", probe), archive, false)
	rejected := h.run("released-run", exec.Command(released, config, manifest))
	if code, ok := rejected.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(rejected.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatal("released handle escaped")
	}
	t.Log("released handle rejected: exit 70 and required error")
}
