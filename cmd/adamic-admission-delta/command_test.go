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
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Not parallel: the child helper exits the test process after running the command.
func TestCommandHelper(test *testing.T) {
	if os.Getenv("ADMISSION_COMMAND_HELPER") == "" {
		test.Parallel()
		return
	}
	var args []string
	if decodeError := json.Unmarshal([]byte(os.Getenv("ADMISSION_COMMAND_ARGS")), &args); decodeError != nil {
		os.Exit(2)
	}
	os.Exit(exitCode(run(args)))
}

// Exclude forks until the writable descriptor is closed; rename alone preserves its inode.
func writeExecutable(test *testing.T, path, script string) {
	test.Helper()
	var temporary string
	defer func() {
		if temporary != "" {
			os.Remove(temporary)
		}
	}()
	func() {
		syscall.ForkLock.Lock()
		defer syscall.ForkLock.Unlock()
		file, temporaryFileError := os.CreateTemp(filepath.Dir(path), ".executable-*")
		if temporaryFileError != nil {
			test.Fatal(temporaryFileError)
		}
		temporary = file.Name()
		if _, writeStringError := file.WriteString(script); writeStringError != nil {
			file.Close()
			test.Fatal(writeStringError)
		}
		if chmodError := file.Chmod(0700); chmodError != nil {
			file.Close()
			test.Fatal(chmodError)
		}
		if closeError := file.Close(); closeError != nil {
			test.Fatal(closeError)
		}
	}()
	if renameError := os.Rename(temporary, path); renameError != nil {
		test.Fatal(renameError)
	}
}

func commandFixture(test *testing.T) (string, string, string, string) {
	test.Helper()
	dir := test.TempDir()
	write := func(path, text string) {
		test.Helper()
		if createDirectoryError := os.MkdirAll(
			filepath.Dir(filepath.Join(dir, path)),
			0755,
		); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(filepath.Join(dir, path), []byte(text), 0755); writeError != nil {
			test.Fatal(writeError)
		}
	}
	write(
		"cloud/admission-corpus/negative-witnesses.json",
		`{ "_comment": "Adding an entry is a ruling routed to @system_adamic", "witnesses": {} }`,
	)
	write(
		"cloud/admission-corpus/runtime-notyet-stops.json",
		`{ "_comment": "Adding an entry is a ruling routed to @system_adamic", "stops": {} }`,
	)
	write("corpus/input.a", "console.log('42');\n")
	write(
		"oracle/node.mjs",
		"import {readFileSync} from 'node:fs';console.log(readFileSync(process.argv[2],'utf8').includes('42')?'42':'changed');\n",
	)
	write("generator.py", "# pinned generator\n")
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.invalid"},
		{"config", "user.name", "Test"},
		{"add", "."},
		{"commit", "-qm", "fixture"},
	} {
		if _, gitError := git(dir, args...); gitError != nil {
			test.Fatal(gitError)
		}
	}
	sha, gitError := git(dir, "rev-parse", "HEAD")
	if gitError != nil {
		test.Fatal(gitError)
	}
	blob, gitError := git(dir, "rev-parse", "HEAD:corpus/input.a")
	if gitError != nil {
		test.Fatal(gitError)
	}
	manifest := manifest{SHA: sha, Generator: "generator.py", Corpora: []corpus{
		{Name: "witnesses", Programs: []program{{Path: "corpus/input.a", Blob: blob}}},
		{Name: "fuzz", Programs: []program{}},
	}}
	data, encodeError := json.Marshal(manifest)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	write("manifest.json", string(data))
	writeExecutable(
		test,
		filepath.Join(dir, "base"),
		"#!/bin/sh\necho \"adamic: stage 0 can't lower witness yet\" >&2\nexit 1\n",
	)
	writeExecutable(
		test,
		filepath.Join(dir, "head"),
		"#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\nc) echo C;;\njs) echo \"console.log('42');\";;\nbuild) printf '#!/bin/sh\\nprintf \"41\\\\n\"\\n' > \"$4.tmp\" && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n",
	)
	return dir, sha, filepath.Join(dir, "base"), filepath.Join(dir, "head")
}
func invokeCommand(test *testing.T, dir string, args ...string) observation {
	test.Helper()
	return invokeCommandWith(test, dir, nil, args...)
}

// invokeCommandWith runs the command with environment appended to the test's, returning its own exit code.
func invokeCommandWith(test *testing.T, dir string, environment []string, args ...string) observation {
	test.Helper()
	exe, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	// Fake compilers implement the lowering adapter contract explicitly.
	for index := 0; index+1 < len(args); index++ {
		if args[index] == "--base-binary" {
			args = append(args, "--base-lower-binary", args[index+1])
		}
		if args[index] == "--head-binary" {
			args = append(args, "--head-lower-binary", args[index+1])
		}
	}
	data, encodeError := json.Marshal(args)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(timeoutContext, exe, "-test.run=^TestCommandHelper$")
	command.Dir = dir
	command.Env = append(
		append(os.Environ(), "ADMISSION_COMMAND_HELPER=1", "ADMISSION_COMMAND_ARGS="+string(data)),
		environment...,
	)
	var out, errors bytes.Buffer
	command.Stdout = &out
	command.Stderr = &errors
	runError := command.Run()
	code := 0
	if exit, ok := runError.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if runError != nil {
		test.Fatal(runError)
	}
	return observation{Exit: code, Stdout: out.String(), Stderr: errors.String()}
}
func TestNewAdmissionMismatch(test *testing.T) {
	test.Parallel()
	dir, sha, base, head := commandFixture(test)
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		base,
		"--head-binary",
		head,
		"--manifest",
		filepath.Join(dir, "manifest.json"),
		"--json",
		"--timeout",
		"2s",
		"--budget",
		"1",
	)
	var report report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatalf("%v: %s %s", decodeError, observation.Stdout, observation.Stderr)
	}
	if observation.Exit != 1 ||
		report.Verdict != "fail" ||
		report.Void ||
		report.Admitted != 1 ||
		report.SamplingSize != 1 ||
		report.Omitted != 0 {
		test.Fatalf(
			"missed mismatch, or charged it as other than a red: exit %d %+v %s",
			observation.Exit,
			report,
			observation.Stderr,
		)
	}
	program := report.Programs[0]
	if program.Agree == nil ||
		*program.Agree ||
		program.Node.Stdout != "42\n" ||
		program.JavaScript.Stdout != "42\n" ||
		program.Native.Stdout != "41\n" {
		test.Fatalf("observations: %+v", program)
	}
	if report.GeneratorBlob == "" || report.ManifestBlob == "" || len(report.Corpora[2].Programs) != 0 {
		test.Fatal("missing provenance or empty corpus")
	}
}
func TestHeadInputIsolation(test *testing.T) {
	test.Parallel()
	dir, sha, base, head := commandFixture(test)
	if writeError := os.WriteFile(filepath.Join(dir, "corpus/input.a"), []byte("changed"), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		head,
		"--head-binary",
		head,
		"--manifest",
		filepath.Join(dir, "manifest.json"),
		"--json",
	)
	var report report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr)
	}
	if report.CompileTimeoutSeconds != 45 ||
		report.RuntimeTimeoutSeconds != 10 ||
		report.Programs[0].Head.WallSeconds <= 0 {
		test.Fatal("missing timeout limits or measured command time", report)
	}
	if observation.Exit != 0 || report.Admitted != 0 || report.Programs[0].Class != "accepted-by-both" {
		test.Fatalf("equal revisions failed: %+v %s", report, observation.Stderr)
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
	writeExecutable(test, head, script)
	observation = invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		head,
		"--head-binary",
		head,
		"--manifest",
		filepath.Join(dir, "manifest.json"),
		"--json",
	)
	if observation.Exit != 0 {
		test.Fatal("working-tree edits changed inputs", observation.Stderr, observation.Stdout)
	}
	// A newly admitted program must also use head's bytes during runtime comparison.
	observation = invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		base,
		"--head-binary",
		head,
		"--manifest",
		filepath.Join(dir, "manifest.json"),
		"--json",
	)
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr)
	}
	if observation.Exit != 0 ||
		report.Verdict != "pass" ||
		report.Admitted != 1 ||
		report.SamplingSize != 1 ||
		report.NodeAgreements != 1 ||
		len(report.Programs) != 1 {
		test.Fatalf("working-tree edits changed runtime inputs: %+v %s", report, observation.Stderr)
	}
	program := report.Programs[0]
	if program.Agree == nil ||
		!*program.Agree ||
		program.Node == nil ||
		program.JavaScript == nil ||
		program.Native == nil ||
		program.Node.Stdout != "42\n" ||
		program.JavaScript.Stdout != "42\n" ||
		program.Native.Stdout != "42\n" {
		test.Fatalf("runtime comparison did not observe head's source: %+v", program)
	}
}
func TestManifestMismatch(test *testing.T) {
	test.Parallel()
	dir, sha, base, head := commandFixture(test)
	data, readError := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var manifest manifest
	if decodeError := json.Unmarshal(data, &manifest); decodeError != nil {
		test.Fatal(decodeError)
	}
	manifest.Corpora[0].Programs[0].Blob = "bad"
	data, encodeError := json.Marshal(manifest)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		base,
		"--head-binary",
		head,
		"--manifest",
		filepath.Join(dir, "manifest.json"),
		"--json",
	)
	if observation.Exit == 0 || !strings.Contains(observation.Stderr, "blob mismatch") {
		test.Fatal("blob mismatch admitted", observation)
	}
}

func TestNewlyRefusedIsInformational(test *testing.T) {
	test.Parallel()
	dir, sha, refusing, accepting := commandFixture(test)
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		accepting,
		"--head-binary",
		refusing,
		"--manifest",
		filepath.Join(dir, "manifest.json"),
		"--json",
	)
	var report report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr)
	}
	if observation.Exit != 0 ||
		report.Admitted != 0 ||
		report.Verdict != "pass" ||
		report.Programs[0].Class != "newly-refused" {
		test.Fatalf("new refusal failed gate: %+v %s", report, observation.Stderr)
	}
}
func TestEmptyCorpus(test *testing.T) {
	test.Parallel()
	dir, sha, base, head := commandFixture(test)
	manifest := manifest{SHA: sha, Generator: "generator.py", Corpora: []corpus{{Name: "fuzz", Programs: []program{}}}}
	data, encodeError := json.Marshal(manifest)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		base,
		"--head-binary",
		head,
		"--manifest",
		filepath.Join(dir, "manifest.json"),
		"--json",
	)
	var report report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr)
	}
	if observation.Exit != 0 ||
		report.Admitted != 0 ||
		report.Verdict != "pass" ||
		len(report.Corpora) != 2 ||
		len(report.Programs) != 0 {
		test.Fatalf("empty corpus failed: %+v %s", report, observation.Stderr)
	}
}
func TestCompilerTimeoutIsError(test *testing.T) {
	test.Parallel()
	observation := execute(test.TempDir(), 100*time.Millisecond, "sh", "-c", "while :; do :; done")
	if observation.Error != "timeout" || compileClass(observation) != "error" {
		test.Fatalf("timeout became refusal: %+v", observation)
	}
}

func compilerFixture(test *testing.T) (string, string) {
	test.Helper()
	dir, _, _, _ := commandFixture(test)
	if createDirectoryError := os.MkdirAll(filepath.Join(dir, "cmd/adamic"), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	for path, text := range map[string]string{
		"go.mod":             "module github.com/system-inc/adamic\n\ngo 1.27\n",
		"cmd/adamic/main.go": "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"C\")}\nfunc compile(string)(*int,int){n:=0;return &n,0}\n",
	} {
		if writeError := os.WriteFile(filepath.Join(dir, path), []byte(text), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if _, gitError := git(dir, "add", "go.mod", "cmd/adamic/main.go"); gitError != nil {
		test.Fatal(gitError)
	}
	// A root commit with pinned metadata gives Workshop and every runner the same revision.
	tree, gitError := git(dir, "write-tree")
	if gitError != nil {
		test.Fatal(gitError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(
		timeoutContext,
		"git",
		"-c",
		"commit.gpgsign=false",
		"commit-tree",
		tree,
		"-m",
		"built compiler fixture",
	)
	command.Dir = dir
	command.Env = toolchain.GitEnv(
		"TZ=UTC",
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@example.invalid",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@example.invalid",
		"GIT_AUTHOR_DATE=2000-01-01T00:00:00Z",
		"GIT_COMMITTER_DATE=2000-01-01T00:00:00Z",
	)
	out, outputError := command.Output()
	if outputError != nil {
		test.Fatal(outputError)
	}
	sha := strings.TrimSpace(string(out))
	if _, gitError := git(dir, "update-ref", "HEAD", sha); gitError != nil {
		test.Fatal(gitError)
	}
	return dir, sha
}

const admissionCompilerOwner = "cmd/adamic-admission-delta TestProduct_AdmissionCompiler"

// Not parallel: build the owner before this package's parallel readers in a developer run.
func TestProduct_AdmissionCompiler(test *testing.T) {
	_, sha := compilerFixture(test)
	buildcache.Product(test, compilerInputs(sha), func(product string) error {
		// Generate the build's inputs inside its trace, rather than reading another test's temporary files.
		root, builtSHA := compilerFixture(test)
		if builtSHA != sha {
			return fmt.Errorf("fixture revision changed: %s != %s", builtSHA, sha)
		}
		return buildCompiler(root, test.TempDir(), sha, "compiler", product)
	})
}

func findAdmissionCompiler(test *testing.T, sha string) string {
	test.Helper()
	directory := buildcache.BuiltElsewhere(test, "the admission compiler", compilerInputs(
		sha,
	), admissionCompilerOwner, func(string) error {
		return fmt.Errorf("admission compiler is %s's to build", admissionCompilerOwner)
	})
	return filepath.Join(directory, "adamic")
}

func TestProduct_AdmissionCompilerReady(test *testing.T) {
	test.Parallel()
	_, sha := compilerFixture(test)
	findAdmissionCompiler(test, sha)
}

func TestBuildRevisions(test *testing.T) {
	test.Parallel()
	dir, sha := compilerFixture(test)
	compiler := findAdmissionCompiler(test, sha)
	// Resolve in the real tree before the helper enters its temporary revision repository.
	observation := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		compiler,
		"--head-binary",
		compiler,
		"--corpus",
		"corpus",
		"--json",
	)
	var report report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr)
	}
	if observation.Exit != 0 ||
		report.Admitted != 0 ||
		report.Verdict != "pass" ||
		report.Programs[0].Class != "accepted-by-both" {
		test.Fatalf("revision build failed: %+v %s", report, observation.Stderr)
	}
}
