package nextjs

import "strconv"

var adamicTrace string
var adamicRecording bool

func lastSeparator(path string) int {
	result := adamicRawLastSeparator(path)
	if adamicRecording {
		adamicTrace += "last:"
		for _, b := range []byte(path) {
			adamicTrace += strconv.Itoa(int(b)) + ","
		}
		adamicTrace += ";"
	}
	return result
}
func AdamicSplitPath(path string) (string, string, string) {
	adamicRecording = true
	adamicTrace = ""
	base, parent := splitPath(path)
	adamicRecording = false
	return base, parent, adamicTrace
}
