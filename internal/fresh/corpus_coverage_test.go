package fresh_test

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func beginFreshUnit(t *testing.T) {
	t.Helper()
	began := time.Now()
	t.Cleanup(func() {
		if elapsed := time.Since(began); elapsed >= 30*time.Second {
			t.Errorf("test unit exceeded 30 seconds: %s", elapsed)
		}
	})
}

func TestFreshCorpusUnitsCoverEveryProgram(t *testing.T) {
	t.Parallel()
	expected := freshPrograms(t)
	if len(freshCorpusPaths) != len(expected) || len(freshCorpusTests) != 1 || len(freshCorpusTests[0]) != len(expected) {
		t.Fatalf("corpus count: %d paths, %d expected, %d families", len(freshCorpusPaths), len(expected), len(freshCorpusTests))
	}
	seen := map[string]bool{}
	for index, path := range freshCorpusPaths {
		if path != filepath.ToSlash(expected[index]) || seen[path] {
			t.Fatalf("corpus coverage at %d: %q, want unique %q", index, path, expected[index])
		}
		seen[path] = true
		var label strings.Builder
		for _, character := range path {
			if character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
				label.WriteRune(character)
			} else {
				label.WriteByte('_')
			}
		}
		digest := fmt.Sprintf("%x", sha256.Sum256([]byte(path)))
		want := "TestFreshWrites_" + label.String() + "_" + digest[:12]
		name := runtime.FuncForPC(reflect.ValueOf(freshCorpusTests[0][index]).Pointer()).Name()
		if got := name[strings.LastIndex(name, ".")+1:]; got != want {
			t.Fatalf("unit for %s: %s, want %s", path, got, want)
		}
	}
	t.Logf("%d programs, %d selectable write-proof units", len(expected), len(freshCorpusTests[0]))
}
