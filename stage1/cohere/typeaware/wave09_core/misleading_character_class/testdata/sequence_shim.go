package core

import (
	"fmt"
	"strings"
	"unicode/utf16"
)

type Wave09ClassCharacter struct {
	Value                    uint32
	Start, End               int
	CodePointEscape, Escaped bool
}

func wave09Written(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != 92 {
			out.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			h, l := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, h, l)
		}
	}
	return out.String()
}
func Wave09Sequence(rows []Wave09ClassCharacter) string {
	var sequence []misleadingCharacter
	for _, row := range rows {
		sequence = append(sequence, misleadingCharacter{value: row.Value, start: row.Start, end: row.End, isCodePointEscape: row.CodePointEscape, escaped: row.Escaped})
	}
	var out strings.Builder
	for _, found := range misleadingSequenceFindings(sequence) {
		suggest := 0
		if found.suggest {
			suggest = 1
		}
		fmt.Fprintf(&out, "%d\t%d\t%s\t%s\t%d\n", found.start, found.end, found.message.Id, wave09Written(found.message.Description), suggest)
	}
	return out.String()
}

func Wave09MeaningChanges(pattern string) bool { return patternMeaningChangesUnderUnicodeFlag(pattern) }
