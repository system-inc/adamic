package oracle

import "testing"

func environmentRecordedCount(test *testing.T, name string) {
	test.Helper()
	key := "internal/oracle/testdata/environment_" + name + ".a"
	for _, row := range countCases() {
		if row.key != key {
			continue
		}
		_, measure := row.prepare(test)
		actual := measure()
		if *updateCounts {
			measuredCounts.Store(key, actual)
		} else {
			recorded, err := recordedCounts()
			if err != nil {
				test.Fatal(err)
			}
			if recorded[key] != actual {
				test.Fatalf("count row changed: %s; recorded %s", actual, recorded[key])
			}
		}
		test.Log(actual)
		return
	}
	test.Fatalf("count fixture undeclared: %s", key)
}
func TestEnvironmentRecordedCountArray(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "array")
}
func TestEnvironmentRecordedCountAsync(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "async")
}
func TestEnvironmentRecordedCountCallbackEscape(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "callback_escape")
}
func TestEnvironmentRecordedCountCallbacks(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "callbacks")
}
func TestEnvironmentRecordedCountCapture(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "capture")
}
func TestEnvironmentRecordedCountCapturedParameters(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "captured_parameters")
}
func TestEnvironmentRecordedCountCoverageAliasJoin(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_alias_join")
}
func TestEnvironmentRecordedCountCoverageBoundary64(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_boundary64")
}
func TestEnvironmentRecordedCountCoverageBoundedCall(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_bounded_call")
}
func TestEnvironmentRecordedCountCoverageCallbackExits(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_callback_exits")
}
func TestEnvironmentRecordedCountCoverageCallbackResult(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_callback_result")
}
func TestEnvironmentRecordedCountCoverageCallbacksExtra(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_callbacks_extra")
}
func TestEnvironmentRecordedCountCoverageCapturedParameter(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_captured_parameter")
}
func TestEnvironmentRecordedCountCoverageClosureSelf(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_closure_self")
}
func TestEnvironmentRecordedCountCoverageCollections(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_collections")
}
func TestEnvironmentRecordedCountCoverageInitializerThrow(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_initializer_throw")
}
func TestEnvironmentRecordedCountCoverageLocalContainers(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_local_containers")
}
func TestEnvironmentRecordedCountCoveragePrimitives(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_primitives")
}
func TestEnvironmentRecordedCountCoverageRegionValues(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_region_values")
}
func TestEnvironmentRecordedCountCoverageReturnedValues(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_returned_values")
}
func TestEnvironmentRecordedCountCoverageValuesExits(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "coverage_values_exits")
}
func TestEnvironmentRecordedCountDirect(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "direct")
}
func TestEnvironmentRecordedCountExits(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "exits")
}
func TestEnvironmentRecordedCountField(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "field")
}
func TestEnvironmentRecordedCountGlobal(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "global")
}
func TestEnvironmentRecordedCountKeepingCall(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "keeping_call")
}
func TestEnvironmentRecordedCountLarge(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "large")
}
func TestEnvironmentRecordedCountLocalCall(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "local_call")
}
func TestEnvironmentRecordedCountLoop(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "loop")
}
func TestEnvironmentRecordedCountMap(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "map")
}
func TestEnvironmentRecordedCountReturned(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "returned")
}
func TestEnvironmentRecordedCountSet(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "set")
}
func TestEnvironmentRecordedCountSiblings(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "siblings")
}
func TestEnvironmentRecordedCountUnknownCall(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "unknown_call")
}
func TestEnvironmentRecordedCountUnknownCallback(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "unknown_callback")
}
func TestEnvironmentRecordedCountVirtualCall(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "virtual_call")
}

func TestEnvironmentRecordedCountCapturedParameterEscape(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "captured_parameter_escape")
}
func TestEnvironmentRecordedCountCapturedParameterReassign(test *testing.T) {
	test.Parallel()
	environmentRecordedCount(test, "captured_parameter_reassign")
}
