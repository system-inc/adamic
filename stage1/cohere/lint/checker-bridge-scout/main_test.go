package main

import (
	"bytes"
	"crypto/sha256"
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

	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
)

func command(t *testing.T, directory, executable string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(executable, args...)
	cmd.Dir = directory
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s: %v\n%s", executable, err, stderr.Bytes())
	}
	return stdout.Bytes()
}
func oracle(t *testing.T, root string) string {
	t.Helper()
	dir := t.TempDir()
	side := filepath.Join(root, "stage1/cohere/lint/checker-bridge-scout/testdata/oracle.go")
	virtual := filepath.Join(root, "cohere/adamic_scout_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	path := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(path, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "oracle")
	command(t, filepath.Join(root, "cohere"), "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func upstreamSources(t *testing.T, root string) []string {
	t.Helper()
	path := filepath.Join(root, "cohere/internal/lint/rules/typescript/no_unnecessary_boolean_literal_compare_test.go")
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	sources := []string{}
	goast.Inspect(parsed, func(node goast.Node) bool {
		literal, ok := node.(*goast.CompositeLit)
		if !ok {
			return true
		}
		source, configuration := "", "nonliteral"
		for _, entry := range literal.Elts {
			pair, ok := entry.(*goast.KeyValueExpr)
			if !ok {
				continue
			}
			name, ok := pair.Key.(*goast.Ident)
			if !ok {
				continue
			}
			value, ok := pair.Value.(*goast.BasicLit)
			if !ok || value.Kind != token.STRING {
				continue
			}
			text, e := strconv.Unquote(value.Value)
			if e != nil {
				t.Fatal(e)
			}
			switch name.Name {
			case "sourceText":
				source = text
			case "configuration":
				configuration = text
			}
		}
		if source != "" && configuration == "" {
			source = strings.TrimSpace(source) + "\n"
			if !seen[source] {
				sources = append(sources, source)
				seen[source] = true
			}
		}
		return true
	})
	if len(sources) < 20 {
		t.Fatal("upstream default literal corpus disappeared")
	}
	return sources
}
func TestAgreementAndMutants(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	native := os.Getenv("ADAMIC_SCOUT_NATIVE")
	if native == "" {
		t.Fatal("set ADAMIC_SCOUT_NATIVE to driver.ts's built native binary")
	}
	corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if corpus == "" {
		t.Fatal("set ADAMIC_TYPESCRIPT_SOURCE to pinned TypeScript 6.0.3 checkout")
	}
	pin := strings.TrimSpace(string(command(t, corpus, "git", "rev-parse", "FETCH_HEAD")))
	if pin != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		t.Fatalf("wrong TypeScript pin: %s", pin)
	}
	goOracle := oracle(t, root)
	sources := upstreamSources(t, root)
	// Existing witness, shortest hard cases, and a real TypeScript compiler file.
	witness, err := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/rules/no-unnecessary-boolean-literal-compare/testdata/witness.ts.txt"))
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := os.ReadFile(filepath.Join(corpus, "src/compiler/core.ts"))
	if err != nil {
		t.Fatal(err)
	}
	sources = append(sources, string(witness), "function f<T extends boolean>(b:T){return b===true;}\n", "declare const b:boolean|undefined; b===true;\n", "// 🌍\ndeclare const b:boolean; ((b))===false;\n", string(compiler))

	for _, name := range []string{"conformance/types/literal/booleanLiteralTypes2.ts", "compiler/booleanAssignment.ts", "conformance/types/primitives/boolean/booleanPropertyAccess.ts"} {
		data, e := os.ReadFile(filepath.Join(root, "cohere/TypeScript/tsc/testdata/tests/cases", name))
		if e != nil {
			t.Fatal(e)
		}
		sources = append(sources, string(data))
	}
	public := os.Getenv("ADAMIC_SCOUT_PUBLIC_MANIFEST")
	if public == "" {
		t.Fatal("set ADAMIC_SCOUT_PUBLIC_MANIFEST to the 23 pinned public sample files")
	}
	manifest, e := os.ReadFile(public)
	if e != nil {
		t.Fatal(e)
	}
	paths := strings.Split(strings.TrimSpace(string(manifest)), "\n")
	if len(paths) != 23 {
		t.Fatalf("want 23 public samples, got %d", len(paths))
	}
	mutantCase := len(sources) - 5
	for _, path := range paths {
		data, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		sources = append(sources, string(data))
	}
	findings, questions := 0, 0
	for index, source := range sources {
		dir := t.TempDir()
		path := filepath.Join(dir, "source.ts")
		config := filepath.Join(dir, "tsconfig.json")
		for name, text := range map[string]string{path: source, config: `{"compilerOptions":{"strict":true,"target":"ESNext","noResolve":true},"files":["source.ts"]}`} {
			if err := os.WriteFile(name, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
		}
		p, err := bridge.Open(config, nil)
		if err != nil {
			t.Fatal(err)
		}
		asks, err := plan(p, path)
		if err != nil {
			t.Fatalf("case %d: %v", index, err)
		}
		if _, _, err = answers(p, asks, 2); err != nil {
			t.Fatal(err)
		}
		requests := filepath.Join(dir, "requests")
		replay := filepath.Join(dir, "replay")
		if err = os.WriteFile(requests, []byte(encodeRequests(asks)), 0600); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(replay, []byte(transcript(asks)), 0600); err != nil {
			t.Fatal(err)
		}
		got := command(t, root, native, config, requests, "2")
		if !bytes.Equal(got, expectedOutput(asks, 2)) {
			t.Fatalf("case %d native facts differ", index)
		}
		want := command(t, root, goOracle, config, path)
		pilot := os.Getenv("ADAMIC_SCOUT_PILOT")
		if pilot == "" {
			t.Fatal("set ADAMIC_SCOUT_PILOT to pilot.ts's native binary")
		}
		nativeReplay := filepath.Join(dir, "native-replay")
		nativeFindings := command(t, root, pilot, config, path, "1", nativeReplay)
		if !bytes.Equal(nativeFindings, want) {
			t.Fatalf("case %d native production findings/repairs differ\nGo %s\nNative %s", index, want, nativeFindings)
		}
		recorded, e := os.ReadFile(nativeReplay)
		if e != nil {
			t.Fatal(e)
		}
		if string(recorded) != transcript(asks) {
			t.Fatalf("case %d actual production ask plan differs", index)
		}

		nodeArgs := []string{"--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(root, "stage1/cohere/lint/checker-bridge-scout/testdata/pilot.mjs"), path, replay}
		got = command(t, root, "node", nodeArgs...)
		if !bytes.Equal(got, want) {
			t.Fatalf("case %d Go/port finding or repair differs\nGo %s\nNode %s", index, want, got)
		}
		var findingCount int
		if _, err = fmt.Sscanf(string(want), "findings %d", &findingCount); err != nil {
			t.Fatal(err)
		}
		findings += findingCount
		questions += len(asks)
		if index == mutantCase {
			changed := append([]request(nil), asks...)
			changed[0].Value += "mutant"
			if err = os.WriteFile(requests, []byte(encodeRequests(changed)), 0600); err != nil {
				t.Fatal(err)
			}
			bad := exec.Command(native, config, requests, "1")
			bad.Dir = root
			out, e := bad.CombinedOutput()
			if e == nil || !bytes.Contains(out, []byte("scout fact mismatch")) {
				t.Fatalf("answer mutant survived: %s", out)
			}
			t.Log("native answer mutant caught by byte comparison")
			for _, kind := range []string{"ReadsOtherFiles", "ReadsDefaultLibrary"} {
				for _, cmd := range []*exec.Cmd{exec.Command(pilot, config, path, "1", "", kind), exec.Command("node", append(nodeArgs, kind)...)} {
					cmd.Dir = root
					output, e := cmd.CombinedOutput()
					if e == nil || !bytes.Contains(output, []byte("undeclared program read "+kind)) {
						t.Fatalf("%s read guard failed: %s", kind, output)
					}
					if exit, ok := e.(*exec.ExitError); !ok || exit.ExitCode() != 70 {
						t.Fatalf("%s did not hit Adamic refusal: %v", kind, e)
					}
				}
				t.Logf("native and Node %s contract guard caught forbidden ask", kind)
			}

			mutantNative := command(t, root, pilot, config, path, "1", "", "", "x")
			mutantNode := command(t, root, "node", append(nodeArgs, "", "x")...)
			for name, output := range map[string][]byte{"native": mutantNative, "Node": mutantNode} {
				if bytes.Equal(output, want) {
					t.Fatalf("%s repair-output mutant survived", name)
				}
				t.Logf("%s repair-output mutant compiles/runs, caught by full fix-byte comparison", name)
			}
			for name, text := range map[string]string{"missing": "", "extra": transcript(asks) + transcript(asks[:1]), "wrong-question": strings.Replace(transcript(asks), "options", "optioNs", 1)} {
				if err = os.WriteFile(replay, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
				bad = exec.Command("node", nodeArgs...)
				bad.Dir = root
				out, e = bad.CombinedOutput()
				if e == nil {
					t.Fatalf("%s transcript mutant survived: %s", name, out)
				}
				if !bytes.Contains(out, []byte("Error:")) {
					t.Fatalf("%s failed outside replay guard: %s", name, out)
				}
				t.Logf("Node %s mutant caught by replay contract", name)
			}
		}
	}
	t.Logf("%d cases, %d findings including exact repair bytes, %d actual pilot questions; native facts and Node ask plans agree", len(sources), findings, questions)
	publicFindings, publicQuestions := 0, 0
	publicConfig := filepath.Join(filepath.Dir(public), "tsconfig.json")
	for _, path := range paths {
		dir := t.TempDir()
		m, e := run(publicConfig, path, native, os.Getenv("ADAMIC_SCOUT_PILOT"), goOracle, dir, 2, 1, false)
		if e != nil {
			t.Fatal(e)
		}
		stem := fmt.Sprintf("file-%x-r1", sha256.Sum256([]byte(path)))
		replay := filepath.Join(dir, stem+".transcript")
		node := command(t, root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(root, "stage1/cohere/lint/checker-bridge-scout/testdata/pilot.mjs"), path, replay)
		want, e := os.ReadFile(filepath.Join(dir, stem+"-oracle.stdout"))
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(node, want) {
			t.Fatalf("shared public project Node finding/fix differs for %s", path)
		}
		publicFindings += m.Findings
		publicQuestions += m.Questions
	}
	t.Logf("shared 23-file public project: %d findings, %d real pilot questions; Go/native/Node and timing counts agree", publicFindings, publicQuestions)
	nativeWhole := command(t, root, os.Getenv("ADAMIC_SCOUT_PILOT"), publicConfig, "--manifest", public, "1")
	oracleWhole := command(t, root, goOracle, publicConfig, "--manifest", public)
	if !bytes.Equal(nativeWhole, oracleWhole) {
		t.Fatal("single-program public run finding/fix bytes differ")
	}
	t.Log("single-program 23-file native production walk agrees with Go")

}
