package lint

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func volumeGenerated(t *testing.T) []string {
	t.Helper()
	source := "interface Bare { n: number }; interface WrongType { value: number }; type Alias = string; const Choice={Yes:'Yes'} as const; function guard(x: unknown): x is string { return true; } console.log('one'); console['warn']('two'); let count=0; count++; for(let i=0;i<3;i++){count++;} interface CallableType { method(value: string): number; readonly property: (value: string) => number; } enum Direction { Left=1, Right=Left|2, Other=compute() } let boxed: Number; class Thing implements Boolean {} if(!flag){doOne();}else{doTwo();} const chosen=!flag?left:right; const assigning=()=>left=right; const wrapped=()=>(left=right); function returning(){return left=right;}\n"
	path := filepath.Join(t.TempDir(), "volume.ts")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return []string{path,
		path + "\tno-plusplus\t\t\t\t{\"AllowForLoopAfterthoughts\":true}",
		path + "\t@typescript-eslint/method-signature-style\t\t\t\t{\"Style\":\"method\"}",
		path + "\t@typescript-eslint/prefer-literal-enum-member\t\t\t\t{\"AllowBitwiseExpressions\":true}",
		path + "\tno-return-assign\t\t\t\t\"always\"",
	}
}

// Not parallel: sanitized mutant builds are bounded and completed before timing.
func TestVolumeMutants(t *testing.T) {
	path := manifest(t, volumeGenerated(t))
	want := execute(t, "", goOracle(t), "--manifest", path).output
	for _, change := range []struct{ name, from, to string }{
		{"predicate kind omitted", "node.kind === 'TypePredicate' &&", "node.kind === 'NeverKeyword' &&"},
		{"console receiver widened", "this.node(receiver).text === 'console'", "this.node(receiver).text !== 'console'"},
		{"for update option ignored", "this.settings.read('allowforloopafterthoughts', 'false') === 'true'", "this.settings.read('allowforloopafterthoughts', 'false') === 'false'"},
		{"method listener lost", "style === 'property' && node.kind === 'MethodSignature'", "style === 'property' && node.kind === 'NeverKeyword'"},
		{"wrapper Number ignored", "['BigInt', 'Boolean', 'Number', 'Object', 'String', 'Symbol']", "['BigInt', 'Boolean', 'Object', 'String', 'Symbol']"},
		{"enum bitwise option ignored", "this.settings.read('allowbitwiseexpressions', 'false') === 'true'", "this.settings.read('allowbitwiseexpressions', 'false') === 'false'"},
		{"enum declaration listener lost", "node.kind === 'EnumDeclaration' && this.enabled('nexus/consistency-no-enum')", "node.kind === 'EnumMember' && this.enabled('nexus/consistency-no-enum')"},
		{"condition polarity reversed", "this.negated(test)", "!this.negated(test)"},
		{"return parentheses option reversed", "this.settings.read('option', 'except-parens') !== 'always'", "this.settings.read('option', 'except-parens') === 'always'"},
		{"interface accepts type suffix", "? ['Interface', 'Properties', 'Options']", "? ['Type', 'Interface', 'Properties', 'Options']"},
	} {
		t.Run(change.name, func(t *testing.T) {
			directory := mutant(t, change.from, change.to, "volume.ts")
			binary := buildPort(t, directory, true)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("%s survived on %s", change.name, side.name)
				}
				t.Logf("%s caught on %s: %s", change.name, side.name, difference(side.run.output, want))
			}
		})
	}
}

// Recovery is a parser dependency, not successful lint parity. Keep the exact
// upstream malformed cases and prove that both ports refuse instead of silently
// returning the oracle's recovered findings.
func checkRecoveryRefusal(t *testing.T, oracle, binary, directory, row string) {
	t.Helper()
	recovered := manifest(t, []string{strings.TrimSuffix(row, "unsupported-recovery") + "recovery"})
	answer := execute(t, "", oracle, "--manifest", recovered)
	t.Logf("Go recovered output: %s", answer.output)
	path := manifest(t, []string{row})
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name string
		args []string
	}{
		{binary, []string{"--manifest", path}},
		{"node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path}},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		command := exec.CommandContext(ctx, side.name, side.args...)
		output, err := os.CreateTemp(t.TempDir(), "recovery-refusal-")
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout = output
		var stderr bytes.Buffer
		command.Stderr = &stderr
		err = command.Run()
		timedOut := ctx.Err() == context.DeadlineExceeded
		cancel()
		output.Close()
		if err == nil || (!timedOut && !strings.Contains(stderr.String(), "adamic: panic:")) {
			t.Fatalf("expected parser refusal from %s, got %v: %s", side.name, err, stderr.String())
		}
		t.Logf("explicit unsupported recovery: %s: timeout=%t: %v: %s", side.name, timedOut, err, stderr.String())
	}
}
