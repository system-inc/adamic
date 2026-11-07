package fresh_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// beginFreshUnit holds a unit to the 30-second budget where it's measured: on the reference
// box (one Codex instance, 4 CPUs, cold), which sets ADAMIC_UNIT_BUDGET=1. Elsewhere a loaded
// machine only logs it, so the gate's correctness verdict never depends on its load.
func beginFreshUnit(test *testing.T) {
	test.Helper()
	began := time.Now()
	test.Cleanup(func() {
		if elapsed := time.Since(began); elapsed >= 30*time.Second {
			if os.Getenv("ADAMIC_UNIT_BUDGET") == "1" {
				test.Errorf("test unit exceeded 30 seconds: %s", elapsed)
			} else {
				test.Logf("test unit took %s, over the 30-second budget measured on the reference box", elapsed)
			}
		}
	})
}

func TestFreshCorpusUnitsCoverEveryProgram(test *testing.T) {
	test.Parallel()
	expected := freshPrograms(test)
	if len(freshCorpusTests) != 1 || len(freshCorpusTests[0]) != len(freshCorpusPaths) {
		test.Fatalf(
			"corpus count: %d paths, %d expected, %d families",
			len(freshCorpusPaths),
			len(expected),
			len(freshCorpusTests),
		)
	}
	live := map[string]bool{}
	for _, path := range expected {
		live[filepath.ToSlash(path)] = true
	}
	seen := map[string]bool{}
	for index, path := range freshCorpusPaths {
		if !live[path] || seen[path] {
			test.Fatalf("corpus coverage at %d: %q is missing or duplicated", index, path)
		}
		seen[path] = true
		var label strings.Builder
		for _, character := range path {
			if character >= 'A' && character <= 'Z' ||
				character >= 'a' && character <= 'z' ||
				character >= '0' && character <= '9' {
				label.WriteRune(character)
			} else {
				label.WriteByte('_')
			}
		}
		digest := fmt.Sprintf("%x", sha256.Sum256([]byte(path)))
		want := "TestFreshWrites_" + label.String() + "_" + digest[:12]
		name := runtime.FuncForPC(reflect.ValueOf(freshCorpusTests[0][index]).Pointer()).Name()
		if got := name[strings.LastIndex(name, ".")+1:]; got != want {
			test.Fatalf("unit for %s: %s, want %s", path, got, want)
		}
	}
	test.Logf("%d programs, %d selectable write-proof units", len(expected), len(freshCorpusTests[0]))
}

// A long remainder is the signal to regenerate the selectable corpus units.
// Until then it keeps new programs covered without requiring regeneration.
func TestFreshCorpusRemainder(test *testing.T) {
	test.Parallel()
	beginFreshUnit(test)
	generated := map[string]bool{}
	for _, path := range freshCorpusPaths {
		generated[path] = true
	}
	count := 0
	for _, path := range freshPrograms(test) {
		if generated[filepath.ToSlash(path)] {
			continue
		}
		count++
		test.Run(filepath.ToSlash(path), func(test *testing.T) {
			checkFreshProgram(test, path)
		})
	}
	test.Logf("remainder ran %d programs", count)
}
