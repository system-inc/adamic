package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestTaskReadonlyClassesThreadSanitizer(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("this task proof's TSan evidence is Linux only")
	}
	for _, fixture := range []string{"fields", "items"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/task_class_"+fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			binary := filepath.Join(t.TempDir(), "task")
			if err := native.Build(native.C(program), binary, native.Options{ThreadSanitize: true}); err != nil {
				t.Fatal(err)
			}
			for _, threads := range []string{"1", "4", "16"} {
				t.Run(threads, func(t *testing.T) {
					for attempt := 0; attempt < 3; attempt++ {
						observed := executeParallel(t, threads, false, binary)
						if difference := disagreement(expected, observed); difference != "" {
							t.Fatalf("threads=%s: %s stderr=%s", threads, difference, observed.stderr)
						}
					}
				})
			}
		})
	}
}

// Mutate the C after the source proof: an unchecked write to the shared receiver
// models exactly what would happen if the task boundary admitted this effect.
func TestTaskClassWriteRaceMutant(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux TSan mutation")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/task_class_items.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	method := regexp.MustCompile(`static adamic_string \* adamic_function_[0-9]+_Position_read\(adamic_object \* (adamic_local_[0-9]+_this)\) \{`)
	match := method.FindStringSubmatch(source)
	if len(match) != 2 {
		t.Fatal("lost shared method receiver")
	}
	source = strings.Replace(source, match[0], match[0]+"\n"+match[1]+"->slots[0].number += 1;", 1)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{ThreadSanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, threads := range []string{"4", "16"} {
		for attempt := 0; attempt < 3; attempt++ {
			result := executeParallel(t, threads, false, binary)
			if result.exitCode != 0 && strings.Contains(string(result.stderr), "ThreadSanitizer: data race") {
				t.Logf("unchecked this write caught at %s threads: %s", threads, result.stderr)
				return
			}
		}
	}
	t.Fatal("TSan did not catch the shared receiver write")
}

func TestTaskParserShapeNodeWitnesses(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, output string }{{"shared", "source|source\n"}, {"local", "first:1|second:1\n"}} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/refused/task_parser_"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				t.Fatalf("Node witness: %+v", observed)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("parser shape must remain refused, got %v", err)
			}
			// Replace a read with a write through the same parser receiver. The
			// shared graph or local constructor boundary must continue to refuse it.
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mutant := strings.Replace(string(source), "return this.scanner.text", "this.scanner.pos++; return this.scanner.text", 1)
			if probe.name == "local" {
				mutant = strings.Replace(string(source), "this.pos++;", "this.pos += 2;", 1)
			}
			mutated := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(mutated, []byte(mutant), 0644); err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, mutated)
			if !errors.As(err, &refused) {
				t.Fatalf("write-through-this parser mutant must remain refused, got %v", err)
			}
		})
	}
}
