package regexp

import (
	"fmt"
	"github.com/system-inc/adamic/internal/unicodeproperties"
)

// UnicodeProperties resolves every ECMAScript property using generated UCD data.
// Syntax checking has already restricted properties of strings to the v flag.
type UnicodeProperties struct{}

func (UnicodeProperties) Lookup(expression string) (PropertySet, error) {
	property, ok := unicodeproperties.Lookup(expression, true)
	if !ok {
		return PropertySet{}, fmt.Errorf("regexp: unknown Unicode property %q", expression)
	}
	var result PropertySet
	if property.Set != nil {
		for _, interval := range property.Set.Ranges {
			result.Ranges = append(result.Ranges, RuneRange{rune(interval.Start), rune(interval.End)})
		}
	}
	for _, sequence := range property.Sequences {
		result.Strings = append(result.Strings, []rune(sequence))
	}
	return result, nil
}
