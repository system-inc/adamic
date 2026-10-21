package worker

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

type workerMutation struct{ name, file, before, after, check, witness string }

var workerMutants = []workerMutation{
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

func runWorkerMutantsShard(t *testing.T, shard int) {
	requireMutants(t, false)
	for _, mutant := range workerMutants {
		if familyOwns("TestWorkerMutants", shard, mutant.name) {
			t.Run(mutant.name, func(t *testing.T) { runWorkerMutant(t, mutant) })
		}
	}
}

func requireMutants(t *testing.T, wasm bool) {
	t.Helper()
	if os.Getenv("ADAMIC_WORKER_MUTANTS") != "1" {
		t.Skip("set ADAMIC_WORKER_MUTANTS=1")
	}
	if wasm {
		requireWasm(t)
	}
}

func mutationOf(mutant *workerMutation) []buildcache.Mutation {
	if mutant == nil {
		return nil
	}
	return []buildcache.Mutation{{File: mutant.file, Before: mutant.before, After: mutant.after}}
}

func runWorkerMutant(t *testing.T, mutant workerMutation) {
	t.Helper()
	check := mutant.check
	if check == "TestWasmWorkers/transform" {
		check = familyCaseName("TestWasmWorkers", "transform") + "/transform"
	}
	if check == "TestWasmWorkerMemory" {
		check = "TestWasmWorkerMemory_000/transform"
	}
	if report, err := mutantCheck(t, nil, check); err != nil {
		t.Fatalf("unmutated check failed: %v\n%s", err, report)
	}
	report, err := mutantCheck(t, &mutant, check)
	assertMutantCaught(t, report, err, check, mutant.witness)
}

func mutantCheck(t *testing.T, mutant *workerMutation, check string) ([]byte, error) {
	t.Helper()
	binary := workerTestProduct(t, mutant, false)
	command := exec.CommandContext(t.Context(), "go", "tool", "test2json", "-t", "-p", "github.com/system-inc/adamic/internal/worker", binary,
		"-test.v=test2json", "-test.count=1", "-test.parallel=1", "-test.timeout=20s", "-test.run=^"+strings.ReplaceAll(check, "/", "$/^")+"$")
	name := ""
	if mutant != nil {
		name = mutant.name
	}
	command.Env = append(os.Environ(), "ADAMIC_WORKER_MUTANTS=0", "ADAMIC_WORKER_MUTANT_CHILD=1", "ADAMIC_WORKER_TEST_MUTATION="+name)
	return command.CombinedOutput()
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
