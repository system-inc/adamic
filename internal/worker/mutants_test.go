package worker

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestWorkerMutants proves the named checks detect real compiler/runtime changes.
func TestWorkerMutants(t *testing.T) {
	if os.Getenv("ADAMIC_WORKER_MUTANTS") != "1" {
		t.Skip("set ADAMIC_WORKER_MUTANTS=1 to run Worker mutants")
	}
	// Not parallel: each check recompiles the compiler; bound the resource cost.
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	mutants := []struct {
		name, file, before, after, check, witness string
	}{
		{"drop-last-header", "internal/worker/worker.go", "for (const field of output.headers)", "for (const field of output.headers.slice(0, -1))", "TestWorkers/echo", "header append order"},
		{"reverse-headers", "internal/worker/worker.go", "for (const field of output.headers)", "for (const field of [...output.headers].reverse())", "TestWorkers/echo", "header append order"},
		{"decode-path", "internal/worker/worker.go", "path: url.pathname,", "path: decodeURIComponent(url.pathname),", "TestWorkers/echo", "echo 0"},
		{"wrong-handle-export", "internal/javascript/javascript.go", "name = functionName(program, exported.Function)", `name = functionName(program, exported.Function); if exported.Name == "handle" { for function, value := range program.Functions { if value.Name == "other" { name = functionName(program, function) } } }`, "TestWorkers/echo", "oracle mismatch"},
		{"panic-200", "internal/worker/worker.go", "new Response('internal error', { status: 500 })", "new Response('internal error', { status: 200 })", "TestWorkers/panic", "panic and recovery"},
		{"panic-kills-next", "internal/worker/runtime.mjs", "throw new AdamicPanic(message);", "globalThis.process.exit(70);", "TestWorkers/panic", "Node: exit status 70"},
		{"accept-string-result", "internal/worker/worker.go", "types.GetReturnTypeOfSignature(signature) != responseType", "(types.GetReturnTypeOfSignature(signature) != responseType && types.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsString == 0)", "TestSignature/string", "accepted=false: <nil>"},
		{"utf8-code-units", "internal/worker/utf8.mjs", "return encoded(text).length;", "return text.length;", "TestRuntime", "AssertionError"},
		{"accept-lookalike-request", "internal/worker/worker.go", "types.GetTypeOfSymbolAtLocation(signature.Parameters()[0], declaration) != requestType", "(types.GetTypeOfSymbolAtLocation(signature.Parameters()[0], declaration) != requestType && types.TypeToString(types.GetTypeOfSymbolAtLocation(signature.Parameters()[0], declaration)) != types.TypeToString(requestType))", "TestSignature/lookalike", "accepted=false: <nil>"},
		{"swallow-handler-bug", "internal/worker/worker.go", "if (!(error instanceof AdamicPanic)) throw error;", "if (false) throw error;", "TestBridgeRethrows", "Missing expected rejection"},
		{"drop-generic-export", "internal/worker/worker.go", "declaration.Kind == ast.KindFunctionDeclaration && len(declaration.TypeParameters()) != 0", "false && declaration.Kind == ast.KindFunctionDeclaration && len(declaration.TypeParameters()) != 0", "TestGenericExportRefused", "generic export was silently lost"},
	}
	baseline := mutantScratch(t, root)
	if report, err := mutantCheck(t, baseline, "Test(Workers|Signature|Runtime|BridgeRethrows|GenericExportRefused)"); err != nil {
		t.Fatalf("scratch baseline failed: %v\n%s", err, report)
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			scratch := mutantScratch(t, root)
			path := filepath.Join(scratch, mutant.file)
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(original), mutant.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			write(t, path, strings.Replace(string(original), mutant.before, mutant.after, 1))
			report, err := mutantCheck(t, scratch, mutant.check)
			assertMutantCaught(t, report, err, mutant.check, mutant.witness)
		})
	}
}

// Copy all compiler sources and embedded assets; the unchanged dependency tree
// is shared read-only. Mutations never touch the working checkout or dependencies.
func mutantScratch(t *testing.T, root string) string {
	t.Helper()
	scratch := t.TempDir()
	for _, name := range []string{"internal", "cmd", "oracle", "bridge"} {
		if err := os.CopyFS(filepath.Join(scratch, name), os.DirFS(filepath.Join(root, name))); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"go.mod", "go.work", "go.sum", "go.work.sum"} {
		contents, err := os.ReadFile(filepath.Join(root, name))
		if (name == "go.sum" || name == "go.work.sum") && os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if name == "go.work" {
			// Use the real dependency path so each scratch copy reuses its build cache.
			contents = []byte(strings.Replace(string(contents), "./cohere/TypeScript/tsc", strconv.Quote(filepath.Join(root, "cohere", "TypeScript", "tsc")), 1))
		}
		write(t, filepath.Join(scratch, name), string(contents))
	}
	if err := os.Symlink(filepath.Join(root, "cohere"), filepath.Join(scratch, "cohere")); err != nil {
		t.Fatal(err)
	}
	return scratch
}

func mutantCheck(t *testing.T, scratch, check string) ([]byte, error) {
	t.Helper()
	logPath := filepath.Join(scratch, "check.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(t.Context(), "go", "test", "-json", "-count=1", "-timeout=2m", "-run", "^"+strings.ReplaceAll(check, "/", "$/^")+"$", "./internal/worker")
	command.Dir = scratch
	command.Env = append(os.Environ(), "ADAMIC_WORKER_MUTANTS=0", "GOWORK="+filepath.Join(scratch, "go.work"))
	command.Stdout, command.Stderr = log, log
	runError := command.Run()
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	report, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	return report, runError
}

func assertMutantCaught(t *testing.T, report []byte, err error, check, witness string) {
	t.Helper()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("expected failed check, got %v\n%s", err, report)
	}
	decoder := json.NewDecoder(bytes.NewReader(report))
	failed := false
	var output strings.Builder
	for {
		var event struct{ Action, Test, Output string }
		if err := decoder.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("invalid test report: %v\n%s", err, report)
		}
		if event.Test == check {
			failed = failed || event.Action == "fail"
			output.WriteString(event.Output)
		}
	}
	if !failed || !strings.Contains(output.String(), witness) {
		t.Fatalf("named check did not fail with %q\n%s", witness, report)
	}
	t.Logf("CAUGHT by %s (%s)", check, witness)
}
