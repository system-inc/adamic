package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: the child helper exits the test process after running the command.
func TestCommandHelper(t *testing.T) {
	if os.Getenv("ADMISSION_COMMAND_HELPER") == "" {
		t.Parallel()
		return
	}
	var args []string
	if err := json.Unmarshal([]byte(os.Getenv("ADMISSION_COMMAND_ARGS")), &args); err != nil {
		os.Exit(2)
	}
	if err := run(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func commandFixture(t *testing.T) (string, string, string, string) {
	t.Helper()
	dir := t.TempDir()
	write := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, path), []byte(text), 0755); err != nil {
			t.Fatal(err)
		}
	}
	write("corpus/input.a", "console.log('42');\n")
	write("oracle/node.mjs", "import {readFileSync} from 'node:fs';console.log(readFileSync(process.argv[2],'utf8').includes('42')?'42':'changed');\n")
	write("generator.py", "# pinned generator\n")
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "Test"}, {"add", "."}, {"commit", "-qm", "fixture"}} {
		if _, err := git(dir, args...); err != nil {
			t.Fatal(err)
		}
	}
	sha, err := git(dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	blob, err := git(dir, "rev-parse", "HEAD:corpus/input.a")
	if err != nil {
		t.Fatal(err)
	}
	m := manifest{SHA: sha, Generator: "generator.py", Corpora: []corpus{{Name: "witnesses", Programs: []program{{Path: "corpus/input.a", Blob: blob}}}, {Name: "fuzz", Programs: []program{}}}}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	write("manifest.json", string(data))
	write("base", "#!/bin/sh\necho \"adamic: stage 0 can't lower witness yet\" >&2\nexit 1\n")
	write("head", "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\nc) echo C;;\njs) echo \"console.log('42');\";;\nbuild) cp \"$0\" \"$4\"; printf '#!/bin/sh\\nprintf \"41\\\\n\"\\n' > \"$4\";;\nesac\n")
	return dir, sha, filepath.Join(dir, "base"), filepath.Join(dir, "head")
}
func invokeCommand(t *testing.T, dir string, args ...string) observation {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	// Fake compilers implement the lowering adapter contract explicitly.
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--base-binary" {
			args = append(args, "--base-lower-binary", args[i+1])
		}
		if args[i] == "--head-binary" {
			args = append(args, "--head-lower-binary", args[i+1])
		}
	}
	data, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, exe, "-test.run=^TestCommandHelper$")
	command.Dir = dir
	command.Env = append(os.Environ(), "ADMISSION_COMMAND_HELPER=1", "ADMISSION_COMMAND_ARGS="+string(data))
	var out, errors bytes.Buffer
	command.Stdout = &out
	command.Stderr = &errors
	err = command.Run()
	code := 0
	if err != nil {
		code = 1
	}
	return observation{Exit: code, Stdout: out.String(), Stderr: errors.String()}
}
func TestNewAdmissionMismatch(t *testing.T) {
	t.Parallel()
	dir, sha, base, head := commandFixture(t)
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json", "--timeout", "2s", "--budget", "1")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatalf("%v: %s %s", err, o.Stdout, o.Stderr)
	}
	if o.Exit == 0 || r.Verdict != "fail" || r.Admitted != 1 || r.SamplingSize != 1 || r.Omitted != 0 {
		t.Fatalf("missed mismatch: %+v %s", r, o.Stderr)
	}
	p := r.Programs[0]
	if p.Agree == nil || *p.Agree || p.Node.Stdout != "42\n" || p.JavaScript.Stdout != "42\n" || p.Native.Stdout != "41\n" {
		t.Fatalf("observations: %+v", p)
	}
	if r.GeneratorBlob == "" || r.ManifestBlob == "" || len(r.Corpora[2].Programs) != 0 {
		t.Fatal("missing provenance or empty corpus")
	}
}
func TestHeadInputIsolation(t *testing.T) {
	t.Parallel()
	dir, sha, _, head := commandFixture(t)
	if err := os.WriteFile(filepath.Join(dir, "corpus/input.a"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", head, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr)
	}
	if r.CompileTimeoutSeconds != 45 || r.RuntimeTimeoutSeconds != 10 || r.Programs[0].Head.WallSeconds <= 0 {
		t.Fatal("missing timeout limits or measured command time", r)
	}
	if o.Exit != 0 || r.Admitted != 0 || r.Programs[0].Class != "accepted-by-both" {
		t.Fatalf("equal revisions failed: %+v %s", r, o.Stderr)
	}
	// Make the compiler inspect source bytes, so this check separates snapshot use from working-tree use.
	script := "#!/bin/sh\nIFS= read -r line < \"$2\"; case \"$line\" in *changed*) echo \"panic: dirty source\" >&2; exit 1;; esac\necho C\n"
	if err := os.WriteFile(head, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	o = invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", head, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json")
	if o.Exit != 0 {
		t.Fatal("working-tree edits changed inputs", o.Stderr, o.Stdout)
	}
}
func TestManifestMismatch(t *testing.T) {
	t.Parallel()
	dir, sha, base, head := commandFixture(t)
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err = json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	m.Corpora[0].Programs[0].Blob = "bad"
	data, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json")
	if o.Exit == 0 || !strings.Contains(o.Stderr, "blob mismatch") {
		t.Fatal("blob mismatch admitted", o)
	}
}

func TestNewlyRefusedIsInformational(t *testing.T) {
	t.Parallel()
	dir, sha, refusing, accepting := commandFixture(t)
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", accepting, "--head-binary", refusing, "--manifest", filepath.Join(dir, "manifest.json"), "--json")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr)
	}
	if o.Exit != 0 || r.Admitted != 0 || r.Verdict != "pass" || r.Programs[0].Class != "newly-refused" {
		t.Fatalf("new refusal failed gate: %+v %s", r, o.Stderr)
	}
}
func TestEmptyCorpus(t *testing.T) {
	t.Parallel()
	dir, sha, base, head := commandFixture(t)
	m := manifest{SHA: sha, Generator: "generator.py", Corpora: []corpus{{Name: "fuzz", Programs: []program{}}}}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json")
	var r report
	if err = json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr)
	}
	if o.Exit != 0 || r.Admitted != 0 || r.Verdict != "pass" || len(r.Corpora) != 2 || len(r.Programs) != 0 {
		t.Fatalf("empty corpus failed: %+v %s", r, o.Stderr)
	}
}
func TestCompilerTimeoutIsError(t *testing.T) {
	t.Parallel()
	o := execute(t.TempDir(), 100*time.Millisecond, "sh", "-c", "while :; do :; done")
	if o.Error != "timeout" || compileClass(o) != "error" {
		t.Fatalf("timeout became refusal: %+v", o)
	}
}

func TestBuildRevisions(t *testing.T) {
	t.Parallel()
	dir, _, _, _ := commandFixture(t)
	if err := os.MkdirAll(filepath.Join(dir, "cmd/adamic"), 0755); err != nil {
		t.Fatal(err)
	}
	for path, text := range map[string]string{"go.mod": "module github.com/system-inc/adamic\n\ngo 1.27\n", "cmd/adamic/main.go": "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"C\")}\nfunc compile(string)(*int,int){n:=0;return &n,0}\n"} {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := git(dir, "add", "go.mod", "cmd/adamic/main.go"); err != nil {
		t.Fatal(err)
	}
	if _, err := git(dir, "commit", "-qm", "built compiler fixture"); err != nil {
		t.Fatal(err)
	}
	sha, err := git(dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--corpus", "corpus", "--json")
	var r report
	if err = json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr)
	}
	if o.Exit != 0 || r.Admitted != 0 || r.Verdict != "pass" || r.Programs[0].Class != "accepted-by-both" {
		t.Fatalf("revision build failed: %+v %s", r, o.Stderr)
	}
}
