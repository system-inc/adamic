package estree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func importSourceWitness(t *testing.T) string {
	t.Helper()
	return filepath.Join(root(t), "cohere/TypeScript/tsc/testdata/tests/cases/conformance/importSource/importSource5.ts")
}

func TestImportSourceConstructorRefusal(t *testing.T) {
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	oracle := goOracle(t)
	list := manifest(t, []string{
		`new import.source("./a.js");`,
		`new import.source("./a.js").a;`,
		`new import.meta;`,
		`new import("./a.js");`,
		`new new import.source("./a.js");`,
		"// é😀\r\nnew /* target */ import.source('x');",
		"\ufeff// é😀\r\nnew\r\nimport.source('x');",
		"const x = ; new import.source('x');", // Preserve the earlier diagnostic.
	})
	listing, err := os.ReadFile(list)
	if err != nil {
		t.Fatal(err)
	}
	paths := append([]string{importSourceWitness(t)}, strings.Fields(string(listing))...)
	for index, path := range paths {
		one := filepath.Join(t.TempDir(), "manifest")
		if err := os.WriteFile(one, []byte(path+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		var record struct{ Status, Error string }
		if err := json.Unmarshal(execute(t, "", oracle, "--audit", one, t.TempDir()), &record); err != nil {
			t.Fatal(err)
		}
		if record.Status != "error" || !strings.HasPrefix(record.Error, "parse error at ") {
			t.Fatalf("%s: Go did not give a parse refusal: %+v", path, record)
		}
		if index == 0 && record.Error != "parse error at 351: Expression expected." {
			t.Fatalf("witness Go diagnostic changed: %q", record.Error)
		}
		for _, runner := range []struct {
			name string
			argv []string
		}{
			{"Node", []string{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}},
			{"sanitized native", []string{binary, path}},
			{"emitted JavaScript", []string{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}},
		} {
			stdout, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			var stderr bytes.Buffer
			exit, timeout := runWithCPUBudget(t, runner.argv, stdout, &stderr, 2*time.Second)
			stdout.Close()
			info, err := os.Stat(stdout.Name())
			if err != nil {
				t.Fatal(err)
			}
			// Runtime panic envelopes differ; the complete diagnostic payload must be identical.
			got := regexp.MustCompile(`(?m)parse error at [0-9]+: [^\r\n]+`).FindString(stderr.String())
			if timeout || exit == nil || info.Size() != 0 || got != record.Error || strings.Contains(stderr.String(), "Sanitizer") {
				t.Fatalf("%s %s: timeout=%v exit=%v stdout=%d diagnostic=%q want=%q stderr=%s", runner.name, path, timeout, exit, info.Size(), got, record.Error, &stderr)
			}
		}
		t.Logf("%s: %s on Node, sanitized native and emitted JavaScript", filepath.Base(path), record.Error)
	}
	controls := manifest(t, []string{
		`import.source; import.source("./a.js"); import.meta;`,
		`new (import.source)("./a.js");`,
		`new C(import.source("./a.js"));`,
		`new C; new C(); new C<T>(); new.target;`,
	})
	want := execute(t, "", oracle, "--manifest", controls)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", controls), "sanitized native": execute(t, "", binary, "--manifest", controls), "emitted JavaScript": onNode(t, script, "--manifest", controls)} {
		if diff := firstDifference(want, got); diff != "" {
			t.Fatal(name + ": " + diff)
		}
	}
	t.Logf("%d exact Go refusals and 4 valid controls, %d identical control bytes", len(paths), len(want))
}

func TestImportSourceConstructorMutant(t *testing.T) {
	main := mutantPort(t, "sourceParser.ts", "this.suffix(this.primary(false), false)", "this.suffix(this.primary(), false)")
	binary, script := build(t, main, true)
	path := importSourceWitness(t)
	list := filepath.Join(t.TempDir(), "manifest")
	if err := os.WriteFile(list, []byte(path+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var record struct{ Status, Error string }
	if err := json.Unmarshal(execute(t, "", goOracle(t), "--audit", list, t.TempDir()), &record); err != nil {
		t.Fatal(err)
	}
	if record.Status != "error" || record.Error != "parse error at 351: Expression expected." {
		t.Fatalf("Go witness refusal changed: %+v", record)
	}
	for name, got := range map[string][]byte{"Node": onNode(t, main, path), "sanitized native": execute(t, "", binary, path), "emitted JavaScript": onNode(t, script, path)} {
		if !strings.HasPrefix(string(got), "0 Program ") || len(got) != 9552 {
			t.Fatal(fmt.Sprintf("%s revert mutant did not reproduce the original 9552-byte acceptance: %d bytes", name, len(got)))
		}
		t.Log(name + ": revert mutant exits successfully and accepts the Go-refused witness; refusal check catches it")
	}
}
