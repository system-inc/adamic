package oracle

import "testing"

func environmentRecordedCount(t *testing.T, name string) {
	t.Helper()
	key := "internal/oracle/testdata/environment_" + name + ".a"
	for _, row := range countCases() {
		if row.key != key {
			continue
		}
		_, measure := row.prepare(t)
		actual := measure()
		if *updateCounts {
			measuredCounts.Store(key, actual)
		} else {
			recorded, err := recordedCounts()
			if err != nil {
				t.Fatal(err)
			}
			if recorded[key] != actual {
				t.Fatalf("count row changed: %s; recorded %s", actual, recorded[key])
			}
		}
		t.Log(actual)
		return
	}
	t.Fatalf("count fixture undeclared: %s", key)
}
func TestEnvironmentRecordedCountArray(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "array")
}
func TestEnvironmentRecordedCountAsync(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "async")
}
func TestEnvironmentRecordedCountCallbackEscape(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "callback_escape")
}
func TestEnvironmentRecordedCountCallbacks(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "callbacks")
}
func TestEnvironmentRecordedCountCapture(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "capture")
}
func TestEnvironmentRecordedCountCapturedParameters(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "captured_parameters")
}
func TestEnvironmentRecordedCountCoverageAliasJoin(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_alias_join")
}
func TestEnvironmentRecordedCountCoverageBoundary64(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_boundary64")
}
func TestEnvironmentRecordedCountCoverageBoundedCall(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_bounded_call")
}
func TestEnvironmentRecordedCountCoverageCallbackExits(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_callback_exits")
}
func TestEnvironmentRecordedCountCoverageCallbackResult(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_callback_result")
}
func TestEnvironmentRecordedCountCoverageCallbacksExtra(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_callbacks_extra")
}
func TestEnvironmentRecordedCountCoverageCapturedParameter(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_captured_parameter")
}
func TestEnvironmentRecordedCountCoverageClosureSelf(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_closure_self")
}
func TestEnvironmentRecordedCountCoverageCollections(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_collections")
}
func TestEnvironmentRecordedCountCoverageInitializerThrow(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_initializer_throw")
}
func TestEnvironmentRecordedCountCoverageLocalContainers(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_local_containers")
}
func TestEnvironmentRecordedCountCoveragePrimitives(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_primitives")
}
func TestEnvironmentRecordedCountCoverageRegionValues(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_region_values")
}
func TestEnvironmentRecordedCountCoverageReturnedValues(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_returned_values")
}
func TestEnvironmentRecordedCountCoverageValuesExits(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "coverage_values_exits")
}
func TestEnvironmentRecordedCountDirect(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "direct")
}
func TestEnvironmentRecordedCountExits(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "exits")
}
func TestEnvironmentRecordedCountField(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "field")
}
func TestEnvironmentRecordedCountGlobal(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "global")
}
func TestEnvironmentRecordedCountKeepingCall(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "keeping_call")
}
func TestEnvironmentRecordedCountLarge(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "large")
}
func TestEnvironmentRecordedCountLocalCall(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "local_call")
}
func TestEnvironmentRecordedCountLoop(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "loop")
}
func TestEnvironmentRecordedCountMap(t *testing.T) { t.Parallel(); environmentRecordedCount(t, "map") }
func TestEnvironmentRecordedCountReturned(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "returned")
}
func TestEnvironmentRecordedCountSet(t *testing.T) { t.Parallel(); environmentRecordedCount(t, "set") }
func TestEnvironmentRecordedCountSiblings(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "siblings")
}
func TestEnvironmentRecordedCountUnknownCall(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "unknown_call")
}
func TestEnvironmentRecordedCountUnknownCallback(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "unknown_callback")
}
func TestEnvironmentRecordedCountVirtualCall(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "virtual_call")
}

func TestEnvironmentRecordedCountCapturedParameterEscape(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "captured_parameter_escape")
}
func TestEnvironmentRecordedCountCapturedParameterReassign(t *testing.T) {
	t.Parallel()
	environmentRecordedCount(t, "captured_parameter_reassign")
}
