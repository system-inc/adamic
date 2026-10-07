package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestMovesAgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/moves/accepted/objects.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 {
		t.Fatalf("Node: %+v", source)
	}
	if difference := disagreement(source, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	checkParallelVariants(t, program, source)
}

func TestMovesRefusals(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/moves/refused/*.a"))
	if err != nil || len(paths) < 30 {
		t.Fatalf("fixtures: %d %v", len(paths), err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			want, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".what")
			if err != nil {
				t.Fatal(err)
			}
			fix, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".fix")
			if err != nil {
				t.Fatal(err)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want Refused %q, got %v", strings.TrimSpace(string(want)), err)
			}
			if refused.Fix != "return it through the results" && refused.Fix != "don't use it after the parallelMap" {
				t.Fatalf("unapproved move fix: %q", refused.Fix)
			}
			if refused.What != strings.TrimSpace(string(want)) || refused.Where == "" || refused.Fix != strings.TrimSpace(string(fix)) {
				t.Fatalf("want %q; got %+v", strings.TrimSpace(string(want)), refused)
			}
		})
	}
}

// Ordinary retain totals cannot distinguish plain from atomic count traffic.
// Count actual graph headers at join, independently of the lowering's move flag.
func TestMovesPlainCounts(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/moves/accepted/objects.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	generated := native.C(program)
	if !strings.Contains(generated, "adamic_parallel_map_move(") {
		t.Fatal("no move ABI")
	}
	for _, mutant := range []bool{false, true} {
		name := "control"
		abi := "adamic_parallel_map_move"
		want := "move graph counts: plain 2050 shared 0\n"
		if mutant {
			name = "shared_mutant"
			abi = "adamic_parallel_map"
			want = "move graph counts: plain 1 shared 2049\n"
		}
		t.Run(name, func(t *testing.T) {
			source := fmt.Sprintf(`#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
adamic_array *%s(adamic_array *, adamic_closure *, bool);
adamic_array *moves_count_map(adamic_array *items, adamic_closure *work, bool references) {
 adamic_array *results = %s(items, work, references);
 if (results == NULL) { abort(); }
 size_t shared = adamic_is_shared(&items->heap) + adamic_is_shared(&results->heap);
 for (size_t index = 0; index < items->length; index++) {
  shared += adamic_is_shared(items->elements[index].reference);
 }
 fprintf(stderr, "move graph counts: plain %%zu shared %%zu\n", items->length + 2 - shared, shared);
 return results;
}
`, abi, abi) + strings.ReplaceAll(generated, "adamic_parallel_map_move(", "moves_count_map(")
			binary := filepath.Join(t.TempDir(), "program")
			if err := native.Build(source, binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			for _, threads := range []string{"1", ""} {
				observed := executeParallel(t, threads, false, binary)
				if observed.exitCode != 0 || !strings.HasPrefix(string(observed.stderr), want) {
					t.Fatalf("want %q: exit %d stderr %s", want, observed.exitCode, observed.stderr)
				}
				if mutant && strings.HasPrefix(string(observed.stderr), "move graph counts: plain 2050 shared 0\n") {
					t.Fatal("sharing mutant was not detected")
				}
				t.Logf("threads=%q %s", threads, observed.stderr)
			}
		})
	}
}
