package core

import "fmt"

func Wave09PatternFindings(pattern, flags string, allowEscape bool) string {
	var out string
	checkRegexPatternWithReporter(pattern, flags, patternSource(pattern, allowEscape), func(found pendingMisleadingFinding) {
		suggest := 0
		if found.suggest {
			suggest = 1
		}
		out += fmt.Sprintf("%d\t%d\t%s\t%s\t%d\n", found.start, found.end, found.message.Id, wave09Written(found.message.Description), suggest)
	})
	return out
}
