package suppression

import "github.com/system-inc/cohere/internal/lint/ecmascript/directives"

// parseDirective reads one comment's text and returns the directive it declares, or nil.
//
// The grammar lives in `internal/lint/ecmascript/directives`, where a rule reporting on directives
// can read the same parse this index builds from. The grammar, identical for every honored spelling:
//
//	<tool>-disable[-next-line|-line] [rule[, rule...]] [-- reason]
//
// Everything after the directive word is optional. No rules means blanket. No reason means the
// author did not say why, which is recorded rather than rejected — see Index.WithoutReason.
func parseDirective(commentText string) *Directive {
	parsed, found := directives.ParseDisable(commentText)
	if !found {
		return nil
	}

	kind := KindFile
	switch parsed.Scope {
	case directives.ScopeNextLine:
		kind = KindNextLine
	case directives.ScopeSameLine:
		kind = KindSameLine
	}

	return &Directive{
		Kind:   kind,
		Rules:  parsed.Rules,
		Reason: parsed.Reason,
	}
}
