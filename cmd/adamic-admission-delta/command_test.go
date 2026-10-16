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

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
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

// Publish only a closed executable, so parallel forks cannot see its write in progress.
func writeExecutable(t *testing.T, path, script string) {
	t.Helper()
	file, err := os.CreateTemp(filepath.Dir(path), ".executable-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if _, err := file.WriteString(script); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Chmod(0700); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		t.Fatal(err)
	}
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
	write("cloud/admission-corpus/negative-witnesses.json", `{ "_comment": "Adding an entry is a ruling routed to @system_adamic", "witnesses": {} }`)
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
	writeExecutable(t, filepath.Join(dir, "base"), "#!/bin/sh\necho \"adamic: stage 0 can't lower witness yet\" >&2\nexit 1\n")
	writeExecutable(t, filepath.Join(dir, "head"), "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\nc) echo C;;\njs) echo \"console.log('42');\";;\nbuild) printf '#!/bin/sh\\nprintf \"41\\\\n\"\\n' > \"$4.tmp\" && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n")
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
	dir, sha, base, head := commandFixture(t)
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
	script := `#!/bin/sh
IFS= read -r line < "$2"
case "$line" in *changed*) echo "panic: dirty source" >&2; exit 1;; esac
case "$1" in
admission-lower) exit 0;;
js) printf "console.log('42');\n";;
build) printf '#!/bin/sh\nprintf "42\\n"\n' > "$4.tmp" && chmod +x "$4.tmp" && mv "$4.tmp" "$4";;
esac
`
	writeExecutable(t, head, script)
	o = invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", head, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json")
	if o.Exit != 0 {
		t.Fatal("working-tree edits changed inputs", o.Stderr, o.Stdout)
	}
	// A newly admitted program must also use head's bytes during runtime comparison.
	o = invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json")
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr)
	}
	if o.Exit != 0 || r.Verdict != "pass" || r.Admitted != 1 || r.SamplingSize != 1 || r.NodeAgreements != 1 || len(r.Programs) != 1 {
		t.Fatalf("working-tree edits changed runtime inputs: %+v %s", r, o.Stderr)
	}
	p := r.Programs[0]
	if p.Agree == nil || !*p.Agree || p.Node == nil || p.JavaScript == nil || p.Native == nil || p.Node.Stdout != "42\n" || p.JavaScript.Stdout != "42\n" || p.Native.Stdout != "42\n" {
		t.Fatalf("runtime comparison did not observe head's source: %+v", p)
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

func compilerFixture(t *testing.T) (string, string) {
	t.Helper()
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
	// A root commit with pinned metadata gives Workshop and every runner the same revision.
	tree, err := git(dir, "write-tree")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "git", "-c", "commit.gpgsign=false", "commit-tree", tree, "-m", "built compiler fixture")
	command.Dir = dir
	command.Env = toolchain.GitEnv("TZ=UTC", "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.invalid", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.invalid", "GIT_AUTHOR_DATE=2000-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2000-01-01T00:00:00Z")
	out, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.TrimSpace(string(out))
	if _, err := git(dir, "update-ref", "HEAD", sha); err != nil {
		t.Fatal(err)
	}
	return dir, sha
}

const admissionCompilerOwner = "cmd/adamic-admission-delta TestProduct_AdmissionCompiler"

// Not parallel: build the owner before this package's parallel readers in a developer run.
func TestProduct_AdmissionCompiler(t *testing.T) {
	_, sha := compilerFixture(t)
	buildcache.Product(t, compilerInputs(sha), func(product string) error {
		// Generate the build's inputs inside its trace, rather than reading another test's temporary files.
		root, builtSHA := compilerFixture(t)
		if builtSHA != sha {
			return fmt.Errorf("fixture revision changed: %s != %s", builtSHA, sha)
		}
		return buildCompiler(root, t.TempDir(), sha, "compiler", product)
	})
}

func findAdmissionCompiler(t *testing.T, sha string) string {
	t.Helper()
	directory := buildcache.BuiltElsewhere(t, "the admission compiler", compilerInputs(sha), admissionCompilerOwner, func(string) error {
		return fmt.Errorf("admission compiler is %s's to build", admissionCompilerOwner)
	})
	return filepath.Join(directory, "adamic")
}

func TestProduct_AdmissionCompilerReady(t *testing.T) {
	t.Parallel()
	_, sha := compilerFixture(t)
	findAdmissionCompiler(t, sha)
}

func TestBuildRevisions(t *testing.T) {
	t.Parallel()
	dir, sha := compilerFixture(t)
	compiler := findAdmissionCompiler(t, sha)
	// Resolve in the real tree before the helper enters its temporary revision repository.
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", compiler, "--head-binary", compiler, "--corpus", "corpus", "--json")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr)
	}
	if o.Exit != 0 || r.Admitted != 0 || r.Verdict != "pass" || r.Programs[0].Class != "accepted-by-both" {
		t.Fatalf("revision build failed: %+v %s", r, o.Stderr)
	}
}
