package tailwind

import "strings"

var adamicRecording bool
var adamicTrace string

func scanNumber(value string) int {
	if adamicRecording {
		adamicTrace += "scan;"
	}
	return adamicRawScanNumber(value)
}
func hasMathFunction(value string) bool {
	if adamicRecording {
		adamicTrace += "math;"
	}
	return adamicRawMath(value)
}
func numberWithSuffix(value string, suffixes []string) bool {
	if adamicRecording {
		adamicTrace += "suffix:" + strings.Join(suffixes, ",") + ";"
	}
	saved := adamicRecording
	adamicRecording = false
	result := adamicRawSuffix(value, suffixes)
	adamicRecording = saved
	return result
}
func AdamicDependencies(value string) (int, bool, bool, bool) {
	return adamicRawScanNumber(value), adamicRawMath(value), adamicRawSuffix(value, angleUnits), adamicRawSuffix(value, percentSuffix)
}
func AdamicPredicates(value string) ([]bool, []string) {
	results := []bool{}
	traces := []string{}
	adamicRecording = true
	adamicTrace = ""
	results = append(results, isAngle(value))
	traces = append(traces, adamicTrace)
	adamicTrace = ""
	results = append(results, isNumber(value))
	traces = append(traces, adamicTrace)
	adamicTrace = ""
	results = append(results, isPercentage(value))
	traces = append(traces, adamicTrace)
	adamicRecording = false
	return results, traces
}
