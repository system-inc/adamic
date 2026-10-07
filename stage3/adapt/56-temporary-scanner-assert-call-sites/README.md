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
