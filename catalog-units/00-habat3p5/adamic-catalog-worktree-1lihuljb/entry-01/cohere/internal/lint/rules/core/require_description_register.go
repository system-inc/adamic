package core

import (
	"github.com/system-inc/cohere/internal/lint/ecmascript/directives"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// init registers @eslint-community/eslint-comments/require-description.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once touch
// nothing in common.
//
// Two registrations, and the second is the one that is easy to lose. The rule reports on suppression
// comments themselves, so it also tells the suppression index that a directive must not silence the
// finding about itself; without that, `x; // eslint-disable-line` hides its own report. See
// `directives.RegisterSubject`, and `coversAFindingAboutADirective` in the suppression package for
// the measurements.
func init() {
	rule.Register(rule.Registration{
		Rule:   RequireDescription,
		Decode: DecodeRequireDescriptionOptions,
	})
	directives.RegisterSubject(RequireDescription.Name)
}
