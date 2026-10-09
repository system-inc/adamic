package suppression

// coversAFindingAboutADirective is ESLint's coverage of a column -1 finding, in this package's
// line-based terms, and it applies only to rules registered through `directives.RegisterSubject`.
//
// Such a rule reports on a directive comment itself, anchored at the comment's first byte, and the
// directive must not be able to silence the finding about itself. Upstream gets that inside ESLint
// by reporting at column -1, which sorts before every directive starting on the same line. Measured
// through the installed ESLint 10.8.1 driving the cloned `require-description`
// (scratchpad `clone-suppression/oracle.cjs`):
//
//	x; // eslint-disable-line                                         reports
//	x; // eslint-disable-line eslint-comments/require-description     reports
//	// eslint-disable-next-line -- why          then  // eslint-disable-line   reports on line 2
//	/* eslint-disable -- why */                 then  // eslint-disable-line   silent
//	/* eslint-disable -- why */ /* eslint-disable-line */             reports
//	/* eslint-disable */                        then  /* eslint-enable */      one finding, line 1 only
//	/* eslint-disable -- why */, /* eslint-enable */, // eslint-disable-line  reports on line 3
//
// So a file-scope directive opened on an earlier line covers it, through its closing enable's own
// line inclusive (an enable on line N sits at column 0 or later, after a column -1 finding on line
// N). A same-line directive never does, and neither does a next-line directive on the line above,
// since both start their coverage at column 0 of this line. Every other rule is unaffected.
func coversAFindingAboutADirective(candidate *Directive, line int) bool {
	return candidate.Kind == KindFile && candidate.Line < line
}
