package worker

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOraclePanicPolicy(test *testing.T) {
	test.Parallel()
	runtime, err := os.ReadFile("../../oracle/adamic.mjs")
	if err != nil {
		test.Fatal(err)
	}
	source := string(runtime)
	begin := strings.Index(source, "export function panic(")
	if begin < 0 {
		test.Fatal("oracle fixture has no panic definition")
	}
	closing := strings.Index(source[begin:], "\n}")
	if closing < 0 {
		test.Fatal("oracle fixture has no panic closing brace")
	}
	end := begin + closing + 2
	for _, fixture := range []struct{ name, source, witness string }{
		{"duplicated", source + "\n" + source[begin:end], `expected exactly one "export function panic(" definition`},
		{"removed", source[:begin] + source[end:], `expected exactly one "export function panic(" definition`},
		{"unbalanced", source[:end-1], `unbalanced body braces for "export function panic("`},
		{
			"missing body",
			source[:begin] + "export function panic(message)",
			`unbalanced body braces for "export function panic("`,
		},
	} {
		test.Run(fixture.name, func(test *testing.T) {
			test.Parallel()
			directory := test.TempDir()
			harness := filepath.Join(directory, "internal", "worker")
			if err := os.MkdirAll(harness, 0755); err != nil {
				test.Fatal(err)
			}
			oracle := filepath.Join(directory, "oracle")
			if err := os.Mkdir(oracle, 0755); err != nil {
				test.Fatal(err)
			}
			for _, name := range []string{"run.mjs", "wasm-node.mjs"} {
				contents, err := os.ReadFile(name)
				if err != nil {
					test.Fatal(err)
				}
				write(test, filepath.Join(harness, name), string(contents))
			}
			write(test, filepath.Join(oracle, "adamic.mjs"), fixture.source)
			bridge := filepath.Join(directory, "worker.mjs")
			write(test, bridge, "import { panic } from 'adamic'; export default { fetch() { panic('test'); } };\n")
			requests := filepath.Join(directory, "requests.jsonl")
			write(test, requests, "")
			command := exec.Command("node", filepath.Join(harness, "run.mjs"), bridge, requests, "--oracle-runtime")
			output, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "oracle panic policy: "+fixture.witness) {
				test.Fatalf("expected named panic definition error %q: %v\n%s", fixture.witness, err, output)
			}
		})
	}
}
