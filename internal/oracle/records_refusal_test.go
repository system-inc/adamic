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
func recordRefusalMutant(test *testing.T, name string) bool {
	test.Helper()
	path, pathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/records_refused/records_"+name+".a"),
	)
	if pathError != nil {
		test.Fatal(pathError)
	}
	if os.Getenv("ADAMIC_RECORD_REFUSAL_MUTANT") == "1" {
		p, loweredError := lowered(test, path)
		if loweredError != nil {
			test.Fatalf("refusal mutant must admit the witness: %v", loweredError)
		}
		truth := onNode(test, path)
		native, _ := nativelyUncached(test, p)
		for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(test, p)} {
			if truth.exitCode != 0 || got.exitCode != 70 || disagreement(truth, got) == "" {
				test.Fatalf("%s refusal mutant survived: Node %+v; mutant %+v", backend, truth, got)
			}
			test.Logf("%s refusal removal admits divergent checked stop", backend)
		}
		return true
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	for _, extension := range []string{".a", ".ts"} {
		probe := filepath.Join(test.TempDir(), "records_"+name+extension)
		if writeError := os.WriteFile(probe, source, 0600); writeError != nil {
			test.Fatal(writeError)
		}
		_, loweredError := lowered(test, probe)
		var refusal *lower.Refused
		if !errors.As(loweredError, &refusal) {
			test.Fatalf("%s must be refused: %v", extension, loweredError)
		}
		want := recordRefusalPin(name)
		if refusal.Where != probe+want[0] || refusal.What != want[1] || refusal.Fix != want[2] {
			test.Fatalf("diagnostic changed: %+v; want location, reason, fix %q", refusal, want)
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
		return [3]string{":3:54", "record read of toString: the record's prototype chain isn't modeled", readFix}
	case "prototype_in":
		return [3]string{
			":3:53",
			"record membership for constructor: the record's prototype chain isn't modeled",
			"use Object.hasOwn(record, key) for own membership, or a Map",
		}
	case "prototype_set":
		return [3]string{
			":3:34",
			"record assignment to __proto__: its inherited setter isn't modeled",
			"define an explicit computed own __proto__ data entry before assigning, or use a Map",
		}
	case "compare_properties_left":
		return [3]string{":10:9", "record read of toString: the record's prototype chain isn't modeled", readFix}
	case "compare_properties_right":
		return [3]string{":10:20", "record read of constructor: the record's prototype chain isn't modeled", readFix}
	case "environment_boundary":
		return [3]string{":4:12", "record read of toString: the record's prototype chain isn't modeled", readFix}
	case "named_invalidated":
		return [3]string{":10:20", scalar, scalarFix}
	case "narrowed_number":
		return [3]string{":4:58", scalar, scalarFix}
	}
	panic("unknown record refusal fixture")
}

func TestRecordRefusalPrototypeRead(test *testing.T) {
	test.Parallel()
	recordRefusalMutant(test, "prototype_read")
}
func TestRecordRefusalPrototypeIn(test *testing.T) {
	test.Parallel()
	recordRefusalMutant(test, "prototype_in")
}
func TestRecordRefusalPrototypeSet(test *testing.T) {
	test.Parallel()
	recordRefusalMutant(test, "prototype_set")
}
func TestRecordRefusalCompareLeft(test *testing.T) {
	test.Parallel()
	recordRefusalMutant(test, "compare_properties_left")
}
func TestRecordRefusalCompareRight(test *testing.T) {
	test.Parallel()
	recordRefusalMutant(test, "compare_properties_right")
}
func TestRecordRefusalEnvironment(test *testing.T) {
	test.Parallel()
	recordRefusalMutant(test, "environment_boundary")
}
func TestRecordRefusalNamedInvalidated(test *testing.T) {
	test.Parallel()
	recordRefusalMutant(test, "named_invalidated")
}
func TestRecordRefusalNarrowedNumber(test *testing.T) {
	test.Parallel()
	recordRefusalMutant(test, "narrowed_number")
}
