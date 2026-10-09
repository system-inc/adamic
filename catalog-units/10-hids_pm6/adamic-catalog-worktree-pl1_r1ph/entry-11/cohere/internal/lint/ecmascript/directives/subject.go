package directives

// subjectRules are the rules whose findings are about a directive comment itself.
//
// Written only from `init` through RegisterSubject and read afterwards, so it needs no lock: every
// package `init` finishes before the first file is walked.
var subjectRules = map[string]bool{}

// RegisterSubject declares that a rule reports on suppression comments themselves, so the
// suppression index must not let a directive silence the finding about itself.
//
// The case this exists for is `@eslint-community/eslint-comments/require-description`. Its finding
// sits on the directive's own line, and an ordinary rule's finding there is covered by that same
// directive: `x; // eslint-disable-line` would silence the report that it carries no reason, and a
// bare `/* eslint-disable */` at the top of a file would silence its own report and every later
// one. The rule would then pass exactly the comments it exists to catch, and count each one as a
// suppression that did its job.
//
// Upstream avoids this inside ESLint by reporting at column -1 (`utils.toForceLocation`), which
// sorts before every directive that starts on the same line. What that means in line terms, and the
// measurements behind it, are on `coversAFindingAboutADirective` in `internal/lint/suppression`,
// which is the only reader of this list.
//
// A registration rather than a field on the rule, because the suppression index owns the question
// of what a directive covers and the rule only has to say that the question is different for it.
func RegisterSubject(ruleName string) {
	subjectRules[ruleName] = true
}

// IsSubject is whether a rule registered itself through RegisterSubject.
func IsSubject(ruleName string) bool {
	return subjectRules[ruleName]
}
