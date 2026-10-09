package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The overlay runner removes only recordDivergenceRefusal. The same witnesses
// must then compile and diverge in both backends; no runtime guard is removed.
func recordRefusalMutant(t *testing.T, name string) bool {
	t.Helper()
	path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_"+name+".a"))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	if os.Getenv("ADAMIC_RECORD_REFUSAL_MUTANT") == "1" {
		p, err := lowered(t, path)
		if err != nil {
			t.Fatalf("refusal mutant must admit the witness: %v", err)
		}
		truth := onNode(t, path)
		native, _ := nativelyUncached(t, p)
		for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, p)} {
			if truth.exitCode != 0 || got.exitCode != 70 || disagreement(truth, got) == "" {
				t.Fatalf("%s refusal mutant survived: Node %+v; mutant %+v", backend, truth, got)
			}
			t.Logf("%s refusal removal admits divergent checked stop", backend)
		}
		return true
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, extension := range []string{".a", ".ts"} {
		probe := filepath.Join(t.TempDir(), "records_"+name+extension)
		if err := os.WriteFile(probe, source, 0600); err != nil {
			t.Fatal(err)
		}
		_, err := lowered(t, probe)
		var refusal *lower.Refused
		if !errors.As(err, &refusal) {
			t.Fatalf("%s must be refused: %v", extension, err)
		}
		want := recordRefusalPin(name)
		if !strings.HasSuffix(refusal.Where, want[0]) || !strings.Contains(refusal.What, want[1]) || !strings.Contains(refusal.Fix, want[2]) {
			t.Fatalf("diagnostic changed: %+v; want location, reason, fix %q", refusal, want)
		}
	}
	return false
}
func recordRefusalPin(name string) [3]string {
	switch name {
	case "prototype_read":
		return [3]string{":2:54", "prototype chain isn't modeled", "Object.hasOwn"}
	case "prototype_in":
		return [3]string{":2:53", "prototype chain isn't modeled", "Object.hasOwn"}
	case "prototype_set":
		return [3]string{":2:34", "inherited setter isn't modeled", "computed own"}
	case "compare_properties_left":
		return [3]string{":9:9", "prototype chain isn't modeled", "Object.hasOwn"}
	case "compare_properties_right":
		return [3]string{":9:20", "prototype chain isn't modeled", "Object.hasOwn"}
	case "environment_boundary":
		return [3]string{":3:12", "prototype chain isn't modeled", "Object.hasOwn"}
	case "named_invalidated":
		return [3]string{":9:20", "current slot type cannot be proven", "recheck"}
	case "narrowed_number":
		return [3]string{":3:58", "current slot type cannot be proven", "recheck"}
	}
	panic("unknown record refusal fixture")
}
func TestRecordRefusalPrototypeRead(t *testing.T) {
	t.Parallel()
	recordRefusalMutant(t, "prototype_read")
}
func TestRecordRefusalPrototypeIn(t *testing.T) { t.Parallel(); recordRefusalMutant(t, "prototype_in") }
func TestRecordRefusalPrototypeSet(t *testing.T) {
	t.Parallel()
	recordRefusalMutant(t, "prototype_set")
}
func TestRecordRefusalCompareLeft(t *testing.T) {
	t.Parallel()
	recordRefusalMutant(t, "compare_properties_left")
}
func TestRecordRefusalCompareRight(t *testing.T) {
	t.Parallel()
	recordRefusalMutant(t, "compare_properties_right")
}
func TestRecordRefusalEnvironment(t *testing.T) {
	t.Parallel()
	recordRefusalMutant(t, "environment_boundary")
}
func TestRecordRefusalNamedInvalidated(t *testing.T) {
	t.Parallel()
	recordRefusalMutant(t, "named_invalidated")
}
func TestRecordRefusalNarrowedNumber(t *testing.T) {
	t.Parallel()
	recordRefusalMutant(t, "narrowed_number")
}
