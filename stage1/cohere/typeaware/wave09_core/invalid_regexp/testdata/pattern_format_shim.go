package core

import (
	"errors"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"strings"
)

type Wave09PatternInput struct {
	Pattern, Flags, Canonical, Error string
	Unsupported                      bool
}

func Wave09PatternCompileInput(pattern, flags string) Wave09PatternInput {
	var canonical strings.Builder
	for _, flag := range "imsu" {
		if strings.ContainsRune(flags, flag) {
			canonical.WriteRune(flag)
		}
	}
	compileFlags := canonical.String()
	if !strings.Contains(compileFlags, "u") && strings.Contains(flags, "v") {
		compileFlags += "u"
	}
	input := Wave09PatternInput{Pattern: pattern, Flags: flags, Canonical: compileFlags}
	if _, err := esregexp.Compile(pattern, compileFlags); err != nil {
		input.Error = err.Error()
		input.Unsupported = errors.Is(err, esregexp.ErrUnsupportedSyntax) || errors.Is(err, esregexp.ErrUnsupportedFlag)
	}
	return input
}
func Wave09PatternMessage(pattern, flags string) string { return invalidPatternMessage(pattern, flags) }
