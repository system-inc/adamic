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
func TestWave30ReactPrerequisites(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_30_REACT_ARTIFACTS")
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
	oracle := volumeOracle(h, "react-oracle", "oracle_wave_30_react.go")
	sources := []struct{ name, rule, source string }{
		{"purity", "react-hooks/purity", `function Component(){return <div>{Math.random()}</div>;}`},
		{"refs", "react-hooks/refs", `function Component(props){const value=props.ref.current;return <div>{value}</div>;}`},
		{"preserve", "react-hooks/preserve-manual-memoization", `function Component(props){const data=useMemo(()=>props.items.edges.nodes??[],[props.items?.edges?.nodes]);return <Foo data={data}/>;}`},
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
	config := h.write("tsconfig.json", `{"compilerOptions":{"target":"ESNext","module":"ESNext","jsx":"preserve","strict":true,"lib":["ESNext","DOM"],"types":[]},"files":["purity.tsx","refs.tsx","preserve.tsx"]}`)
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
