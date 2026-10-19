package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Not parallel: the opt-in service oracle measures rates and builds independent mutants.
func TestWASIService(t *testing.T) {
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		t.Skip("set ADAMIC_TEST_WASI=1, source cloud/setup.sh --wasi-sdk's environment, and use Node 24")
	}
	if os.Getenv("WASI_SYSROOT") == "" {
		t.Fatal("WASI_SYSROOT is required once the service oracle is opted in")
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	compiler := filepath.Join(directory, "adamic")
	wasiCommand(t, repository, "go", "build", "-o", compiler, "./cmd/adamic")
	source := filepath.Join(repository, "internal/native/wasm/service/service.a")
	host := filepath.Join(repository, "internal/native/wasm/service/host.mjs")
	contents, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	code := wasiCommand(t, repository, compiler, "c", source)
	regions := "0"
	if strings.Contains(string(code), "adamic_region_end(") {
		regions = "1"
	}
	requests := filepath.Join(directory, "requests.jsonl")
	responses := filepath.Join(directory, "responses.jsonl")
	runHost := func(module, counted string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", host,
			module, source, requests, responses, counted, regions)
		command.Dir = repository
		return command.CombinedOutput()
	}
	for _, counted := range []string{"0", "1"} {
		module := filepath.Join(directory, "service-"+counted+".wasm")
		arguments := []string{"build", "--target", "wasm32-wasi", source, "-o", module}
		if counted == "1" {
			arguments = append(arguments, "--count")
		}
		wasiCommand(t, repository, compiler, arguments...)
		output, err := runHost(module, counted)
		if err != nil {
			t.Fatalf("service counted=%s: %v\n%s", counted, err, output)
		}
		t.Logf("service counted=%s: %s", counted, output)
	}
	// The command imports the very same handler and uses the exact generated request file.
	commandSource := filepath.Join(directory, "command.a")
	writeWASIFile(t, commandSource, "import { handleRequest } from "+strconv.Quote(filepath.ToSlash(source))+";\n"+
		"import { programArguments, readTextFile } from 'adamic';\n"+
		"const input = readTextFile(programArguments()[0] ?? '');\n"+
		"if (input.kind === 'Ok') { for (const request of input.text.split('\\n').slice(0, -1)) { console.log(handleRequest(request)); } } else { console.error(input.message); }\n")
	binary := filepath.Join(directory, "service-native")
	wasiCommand(t, repository, compiler, "build", commandSource, "-o", binary, "--sanitize")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, binary, requests)
	output, err := command.Output()
	if err != nil {
		if failed, ok := err.(*exec.ExitError); ok {
			t.Fatalf("native service: %v\n%s", err, failed.Stderr)
		}
		t.Fatalf("native service: %v", err)
	}
	expected, err := os.ReadFile(responses)
	if err != nil {
		t.Fatal(err)
	}
	if string(output) != string(expected) {
		actualLines := strings.Split(string(output), "\n")
		expectedLines := strings.Split(string(expected), "\n")
		for index, line := range expectedLines {
			if index >= len(actualLines) {
				t.Fatalf("native command ended before response %d", index)
			}
			if actualLines[index] != line {
				t.Fatalf("native response %d: got %q, want %q", index, actualLines[index], line)
			}
		}
		t.Fatalf("native command emitted %d lines, want %d", len(actualLines), len(expectedLines))
	}
	t.Log("native sanitized command: all 100000 responses agree with Node")
	mutants := []struct {
		name, before, after, caught string
	}{
		{"response-byte", `service":"orders`, `service":"ordert`, "wasm semantic pin"},
		{"retained-object", "const parser = new Parser(request);", "const parser = new Parser(request); retained.push(parser);", "retained a live allocation"},
		{"unicode-escape", "String.fromCharCode(code)", "String.fromCharCode(code + 1)", "wasm semantic pin"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			if strings.Count(string(contents), mutant.before) != 1 {
				t.Fatal("mutant must change exactly one source location")
			}
			changed := strings.Replace(string(contents), mutant.before, mutant.after, 1)
			if mutant.name == "retained-object" {
				changed += "\nconst retained: Parser[] = [];\n"
			}
			mutantSource := filepath.Join(directory, mutant.name+".a")
			writeWASIFile(t, mutantSource, changed)
			module := filepath.Join(directory, mutant.name+".wasm")
			wasiCommand(t, repository, compiler, "build", "--target", "wasm32-wasi", mutantSource, "-o", module, "--count")
			output, err := runHost(module, "1")
			if err == nil || !strings.Contains(string(output), mutant.caught) {
				t.Fatalf("mutant not caught by %q: %v\n%s", mutant.caught, err, output)
			}
			t.Logf("mutant compiled and was caught by %q\n%s", mutant.caught, output)
		})
	}
}
