Starting commit: 09fe4b54913753188a9357982bfd47cdf36ef97c

Code under test: regexp Parse and reached parser helpers; unicodeproperties CanonicalizeUnicode and CanonicalizeLegacy. Oracle: handwritten parser admission and AST expectations; regexp generated tables for shared folding. All four are self oracles. Runtime canonicalization is the implementation under test in TestSharedCanonicalize; the regexp tables are never mutated. Package initializers also call set, which is preparation, not an entry.

Reached functions (clean coverage; callback closures execute within their enclosing functions):

```
github.com/system-inc/adamic/internal/regexp/parser.go:23:				Parse				90.0%
github.com/system-inc/adamic/internal/regexp/parser.go:72:				patternRune			75.0%
github.com/system-inc/adamic/internal/regexp/parser.go:80:				parseFlags			80.0%
github.com/system-inc/adamic/internal/regexp/parser.go:127:				disjunction			100.0%
github.com/system-inc/adamic/internal/regexp/parser.go:143:				alternative			88.2%
github.com/system-inc/adamic/internal/regexp/parser.go:170:				term				45.0%
github.com/system-inc/adamic/internal/regexp/parser.go:206:				group				68.6%
github.com/system-inc/adamic/internal/regexp/parser.go:295:				quantifier			72.0%
github.com/system-inc/adamic/internal/regexp/parser.go:333:				escape				51.6%
github.com/system-inc/adamic/internal/regexp/parser.go:460:				characterClass			93.3%
github.com/system-inc/adamic/internal/regexp/parser.go:504:				finishClass			80.0%
github.com/system-inc/adamic/internal/regexp/parser.go:513:				classSetOperand			79.5%
github.com/system-inc/adamic/internal/regexp/parser.go:568:				classCharacter			50.0%
github.com/system-inc/adamic/internal/regexp/parser.go:595:				classMayContainStrings		78.6%
github.com/system-inc/adamic/internal/regexp/parser.go:628:				validateDuplicateNames		93.3%
github.com/system-inc/adamic/internal/regexp/parser.go:676:				literal				71.4%
github.com/system-inc/adamic/internal/regexp/parser.go:744:				legacyOctal			33.3%
github.com/system-inc/adamic/internal/regexp/parser.go:762:				groupName			80.0%
github.com/system-inc/adamic/internal/regexp/parser.go:778:				groupNameRune			19.2%
github.com/system-inc/adamic/internal/regexp/parser.go:815:				isRegExpIdentifierStart		100.0%
github.com/system-inc/adamic/internal/regexp/parser.go:819:				isRegExpIdentifierPart		100.0%
github.com/system-inc/adamic/internal/regexp/parser.go:822:				decimal				85.7%
github.com/system-inc/adamic/internal/regexp/parser.go:857:				done				100.0%
github.com/system-inc/adamic/internal/regexp/parser.go:858:				peek				100.0%
github.com/system-inc/adamic/internal/regexp/parser.go:868:				take				100.0%
github.com/system-inc/adamic/internal/regexp/parser.go:875:				match				100.0%
github.com/system-inc/adamic/internal/regexp/parser.go:876:				fail				100.0%
github.com/system-inc/adamic/internal/regexp/parser.go:877:				failAt				100.0%
github.com/system-inc/adamic/internal/regexp/parser.go:878:				quantifierAhead			93.3%
github.com/system-inc/adamic/internal/regexp/parser.go:904:				countCaptures			100.0%
github.com/system-inc/adamic/internal/regexp/property.go:13:				validProperty			100.0%
github.com/system-inc/adamic/internal/regexp/property.go:22:				set				100.0%
github.com/system-inc/adamic/internal/unicodeproperties/canonicalize.go:10:		CanonicalizeUnicode		87.5%
github.com/system-inc/adamic/internal/unicodeproperties/canonicalize.go:51:		CanonicalizeLegacy		100.0%
```

The fixed 12-mutant menu is menu.json, recorded before any mutant outcome. No family wrappers in the four scoped rows. Distinct package checks are retained; TestMatcherOct6Mutants is a witness, so its production failures are excluded from uniqueness/subsumption. Its preconditions are recorded separately.
