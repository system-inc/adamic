package tailwind

var adamicRecording bool
var adamicRecordTypes bool
var adamicRecordFamily bool
var adamicTrace string

func AdamicPredicates(value string) []bool {
	result := []bool{}
	for _, predicate := range []func(string) bool{IsColor, IsLength, isPercentage, isFraction, isNumber, IsPositiveInteger, isURL, isBackgroundPosition, isBackgroundSize, isLineWidth, isImage, isFamilyName, isGenericName, isAbsoluteSize, isRelativeSize, isAngle, isVector} {
		result = append(result, predicate(value))
	}
	return result
}
func AdamicParts(value string) []string { return segment(value, ',') }
func AdamicMatch(value, kind string) (bool, string) {
	adamicRecording = true
	adamicRecordTypes = false
	adamicTrace = ""
	result := matchesDataType(value, DataType(kind))
	adamicRecording = false
	return result, adamicTrace
}
func AdamicInfer(value string, types []string) (string, string) {
	kinds := []DataType{}
	for _, kind := range types {
		kinds = append(kinds, DataType(kind))
	}
	adamicRecording = true
	adamicRecordTypes = true
	adamicTrace = ""
	result := InferDataType(value, kinds)
	adamicRecording = false
	adamicRecordTypes = false
	return string(result), adamicTrace
}
func AdamicFamily(value string) (bool, string) {
	adamicRecordFamily = true
	adamicTrace = ""
	result := isFamilyName(value)
	adamicRecordFamily = false
	return result, adamicTrace
}
