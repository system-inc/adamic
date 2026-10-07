package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This records a blocking shared-parser prerequisite, not completed rule ports.
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
		got := h.run(row.name+"-native-parser", exec.Command(binary, paths[at]))
		if got.err != nil {
			exit, ok := got.err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte("adamic: panic:")) {
				t.Fatalf("unexpected parser failure: %v %s", got.err, got.stderr)
			}
			t.Logf("%s: Go reports; native parser refuses JSX, exit 70: %s", row.rule, strings.TrimSpace(string(got.stderr)))
		} else {
			if bytes.Contains(got.stdout, []byte("Jsx")) || !bytes.Contains(got.stdout, []byte("TypeAssertionExpression")) {
				t.Fatalf("shared JSX prerequisite changed: %s", got.stdout)
			}
			t.Logf("%s: Go accepts TSX and reports; native parser treats JSX as TypeAssertionExpression, emits no JSX node", row.rule)
		}

	}
}
