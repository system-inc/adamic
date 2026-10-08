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
			folder := "refused"
			if probe.name == "local" {
				folder = "accepted"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/"+folder+"/task_parser_"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				t.Fatalf("Node witness: %+v", observed)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if probe.name == "local" {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.As(err, &refused) {
				t.Fatalf("shared parser must remain refused, got %v", err)
			}
			// A shared receiver write or publication of a private receiver into
			// an outer capture must continue to be refused.
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mutant := strings.Replace(string(source), "return this.scanner.text", "this.scanner.pos++; return this.scanner.text", 1)
			if probe.name == "local" {
				mutant = strings.Replace(string(source), "const items:", "let leaked: Parser | undefined;\nconst items:", 1)
				mutant = strings.Replace(mutant, "const parser = new Parser(text);", "const parser = new Parser(text); leaked = parser;", 1)
			}
			mutated := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(mutated, []byte(mutant), 0644); err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, mutated)
			if !errors.As(err, &refused) {
				t.Fatalf("receiver ownership mutant must remain refused, got %v", err)
			}
		})
	}
}

func TestTaskWorkerPrivateClassesThreadSanitizer(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux TSan")
	}
	for _, fixture := range []string{"task_parser_local", "task_private_parser"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/"+fixture+".a"))
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
							t.Fatalf("%s stderr=%s", difference, observed.stderr)
						}
					}
				})
			}
		})
	}
}

func TestTaskPrivateReceiverSharedMutant(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux TSan")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/task_private_parser.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	constructor := regexp.MustCompile(`static adamic_object \* (adamic_function_[0-9]+_Parser_new)\(adamic_string \* adamic_local_[0-9]+_text\) \{`)
	match := constructor.FindStringSubmatch(source)
	if len(match) != 2 {
		t.Fatal("lost private constructor")
	}
	calls := regexp.MustCompile(match[1] + `\(adamic_local_[0-9]+_text\)`)
	if !calls.MatchString(source) {
		t.Fatal("lost per-task allocation")
	}
	source = calls.ReplaceAllString(source, "adamic_retain(mutant_shared_parser)")
	source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\nstatic adamic_object *mutant_shared_parser;", 1)
	source = strings.Replace(source, "adamic_start(argc, argv);", "adamic_start(argc, argv);\nmutant_shared_parser = "+match[1]+"(&adamic_string_0);\nadamic_share(mutant_shared_parser);", 1)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{ThreadSanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, threads := range []string{"4", "16"} {
		for attempt := 0; attempt < 3; attempt++ {
			observed := executeParallel(t, threads, false, binary)
			if observed.exitCode != 0 && strings.Contains(string(observed.stderr), "ThreadSanitizer: data race") {
				t.Logf("shared receiver mutation caught at %s threads: %s", threads, observed.stderr)
				return
			}
		}
	}
	t.Fatal("TSan did not catch sharing the mutable parser across tasks")
}

func TestTaskActualScannerShareableBoundary(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "notes/runtime-step37-classes/private_scanner.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "ConstKeyword|LetKeyword\n" || len(observed.stderr) != 0 {
		t.Fatalf("actual scanner Node witness: %+v", observed)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.What != "task capture 'keywords' is not shareable: keywords is a mutable Map" {
		t.Fatalf("want the actual next scanner boundary, got %v", err)
	}
}
