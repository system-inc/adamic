package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
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
		if refusal.Where != probe+want[0] || refusal.What != want[1] || refusal.Fix != want[2] {
			t.Fatalf("diagnostic changed: %+v; want location, reason, fix %q", refusal, want)
		}
	}
	return false
}
func recordRefusalPin(name string) [3]string {
	readFix := "check Object.hasOwn(record, key) immediately before reading its own value, or use a Map"
	scalar := "a narrowed record scalar read after a call mutates that record: its current slot type cannot be proven"
	scalarFix := "snapshot the value before the call, or read and recheck its type after the call"
	switch name {
	case "prototype_read":
		return [3]string{":2:54", "record read of toString: the record's prototype chain isn't modeled", readFix}
	case "prototype_in":
		return [3]string{":2:53", "record membership for constructor: the record's prototype chain isn't modeled", "use Object.hasOwn(record, key) for own membership, or a Map"}
	case "prototype_set":
		return [3]string{":2:34", "record assignment to __proto__: its inherited setter isn't modeled", "define an explicit computed own __proto__ data entry before assigning, or use a Map"}
	case "compare_properties_left":
		return [3]string{":9:9", "record read of toString: the record's prototype chain isn't modeled", readFix}
	case "compare_properties_right":
		return [3]string{":9:20", "record read of constructor: the record's prototype chain isn't modeled", readFix}
	case "environment_boundary":
		return [3]string{":3:12", "record read of toString: the record's prototype chain isn't modeled", readFix}
	case "named_invalidated":
		return [3]string{":9:20", scalar, scalarFix}
	case "narrowed_number":
		return [3]string{":3:58", scalar, scalarFix}
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
