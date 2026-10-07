Temporary: comes out when proven assertion predicates land

Plan published before implementation. Slice call sites only: replace each
reached Debug.assert statement with if (!expression) Debug.fail(...), preserving
one evaluation and the exact failure message. These eleven calls have Boolean
expressions and only absent or literal messages, no verbose callback or explicit
stack marker. Remove the now-unreached assert member from the slice. Full-tree
validation retains that member, replacing only the same reached call sites.
This answers adamic/no-type-predicate until the admission seam is wired.

Preserve resultingToken's narrowing and scanCharacterClassEscape's side effects.
Validate Node token equality, failure messages and the upstream baseline.

Validation: eleven call sites expanded; one now-unreached slice member removed.
The same reached statements were mapped onto complete upstream declarations
for baseline validation; full-tree Debug.assert remains intact. Combined 55-57
baseline passes 106,367 tests with zero differences, failures or pending tests.
Node matches every token. The real reScanQuestionToken failure message matches;
disabling that assertion is caught by diff exit 1. Idempotence: zero edits.
On the four-feature scratch compiler, the next refusal is languageVersion!
at scanner.ts:291:12. Adaptation 57 handles legitimate optional versions.
