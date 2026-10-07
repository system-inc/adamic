package tailwind

import "strconv"

var adamicRecording bool
var adamicParts []string
var adamicValue string
var adamicTrace string

func adamicIndex(value string) int {
	for i, part := range adamicParts {
		if part == value {
			return i
		}
	}
	return -1
}
func adamicRecord(name, value string) {
	if adamicRecording {
		adamicTrace += name + ":" + strconv.Itoa(adamicIndex(value)) + ";"
	}
}
func IsLength(value string) bool { adamicRecord("length", value); return adamicRawLength(value) }
func isNumber(value string) bool { adamicRecord("number", value); return adamicRawNumber(value) }
func isPercentage(value string) bool {
	adamicRecord("percentage", value)
	return adamicRawPercentage(value)
}
func isURL(value string) bool { adamicRecord("url", value); return adamicRawURL(value) }
func segment(value string, separator byte) []string {
	if adamicRecording {
		adamicTrace += "segment:" + string(separator) + ";"
		if value != adamicValue {
			adamicTrace += "wrong-value;"
		}
	}
	return adamicRawSegment(value, separator)
}

type AdamicPart struct {
	Value      string `json:"value"`
	Length     bool   `json:"length"`
	Number     bool   `json:"number"`
	Percentage bool   `json:"percentage"`
	URL        bool   `json:"url"`
}

func AdamicParts(value string, separator byte) []AdamicPart {
	result := []AdamicPart{}
	for _, part := range adamicRawSegment(value, separator) {
		result = append(result, AdamicPart{part, adamicRawLength(part), adamicRawNumber(part), adamicRawPercentage(part), adamicRawURL(part)})
	}
	return result
}
func AdamicObserve(value string) ([]bool, []string) {
	results := []bool{}
	traces := []string{}
	adamicValue = value
	adamicRecording = true
	adamicParts = adamicRawSegment(value, ' ')
	adamicTrace = ""
	results = append(results, isLineWidth(value))
	traces = append(traces, adamicTrace)
	adamicParts = adamicRawSegment(value, ',')
	adamicTrace = ""
	results = append(results, isImage(value))
	traces = append(traces, adamicTrace)
	adamicParts = adamicRawSegment(value, ' ')
	adamicTrace = ""
	results = append(results, isBackgroundPosition(value))
	traces = append(traces, adamicTrace)
	adamicRecording = false
	return results, traces
}
