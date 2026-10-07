package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This proves JSX extraction after integration, not complete rule or React analysis parity.
func TestWave30JsxPrerequisites(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE30_JSX_PREREQUISITES")
	if directory == "" {
		directory = t.TempDir()
	}
	if err = os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	binary := filepath.Join(directory, "parser-probe")
	h.must("parser-probe-build", exec.Command(stage0, "build", filepath.Join(repository, "stage1/cohere/typeaware/wave_30_react_parse_probe.a"), "-o", binary))
	oracle := volumeOracle(h, "react-oracle", "oracle_wave_30_jsx.go")
	sources := []struct{ name, rule, source string }{
		{"fragments", "react/jsx-fragments", `const value = <React.Fragment />;`},
		{"context", "react/jsx-no-constructed-context-values", `function Component(){return <Ctx.Provider value={{a:1}}/>;}`},
		{"undef", "react/jsx-no-undef", `const value = <Missing />;`},
	}
	paths := []string{}
	for _, row := range sources {
		file := h.write(row.name+".a", row.source)
		alias := filepath.Join(directory, row.name+".tsx")
		if err = os.Symlink(filepath.Base(file), alias); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, alias)
	}
	config := h.write("tsconfig.json", `{"compilerOptions":{"target":"ESNext","module":"ESNext","jsx":"preserve","strict":true,"lib":["ESNext","DOM"],"types":[]},"files":["fragments.tsx","context.tsx","undef.tsx"]}`)
	manifest := h.write("react.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.must("react-controls-go", exec.Command(oracle, config, manifest))
	for at, row := range sources {
		if !bytes.Contains(truth.stdout, []byte("\t"+row.rule+"\t")) {
			t.Fatalf("%s lacks positive Go control: %s", row.rule, truth.stdout)
		}
		got := h.must(row.name+"-native-parser", exec.Command(binary, paths[at]))
		kind := "JsxSelfClosingElement"
		if len(got.stderr) != 0 || !wave30ParserHasJsxKind(got.stdout, kind) {
			t.Fatalf("native JSX prerequisite missing: %s %s", got.stdout, got.stderr)
		}
		changed := bytes.ReplaceAll(got.stdout, []byte(kind+"\n"), []byte("Identifier\n"))
		if wave30ParserHasJsxKind(changed, kind) {
			t.Fatal("missing-JSX-node mutant escaped extraction check")
		}
		t.Logf("%s: Go reports; native parser extracts %s; missing-node mutant caught; full rule/analysis parity is not established", row.rule, kind)

	}
}

func wave30ParserHasJsxKind(output []byte, kind string) bool {
	return bytes.Contains(output, []byte(kind+"\n")) && !bytes.Contains(output, []byte("TypeAssertionExpression\n"))
}
