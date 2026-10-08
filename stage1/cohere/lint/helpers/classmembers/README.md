# Class member helpers

Cohere pin: `7945d102a6c18dd36adf9114a758ce646e8b2359`.
Fresh ports; no partial implementation existed on the audited origin branches.
The owning claim is [ecmascript-classmembers.md](../claims/ecmascript-classmembers.md).

- [MemberName](member_name.a): accepted node kinds and raw Name() identity; -1 means nil.
- [IsOverloadSignature](is_overload_signature.a): methods/accessors with no body, including abstract ones.
- [IsAccessorKind](is_accessor_kind.a): getter or setter kind only.

Adapters pass immutable parser facts: the canonical kind name (without Go's
`Kind` prefix), raw numeric name-node identity, and body presence. These helpers
perform the upstream judgment; callers must supply actual parser facts.

[testdata/capture.py](testdata/capture.py) instruments calls in unmodified Go
helper bodies through an overlay, and runs both consuming upstream suites.
Every invocation and its original Go result is retained. The MemberName
observation checks pointer identity against member.Name(), then transports that
identity as 0/-1; the port preserves any supplied numeric identity.

[helpers_test.go](helpers_test.go) regenerates observations against the current
pin, compares source Node, emitted JavaScript and ASan/UBSan native byte-for-byte,
and builds and runs one semantic mutant per helper on all three backends.
The capture refuses upstream test failures and skips.

## Stopped helpers and rule proof

`KeyOf`: `cohere/internal/lint/ecmascript/classmembers/duplicates.go:128` calls
`property.NameTagged(name, property.Static)`. The shared NameTagged port is absent
from the lint area and origin/lint-helpers/property. This package is being landed
by another worker tonight. No private name/tag implementation is supplied.
`ForEachDuplicate`, at line 61, depends on KeyOf and therefore stops with it.
The two consuming rules (`no-dupe-class-members` and
`@typescript-eslint/no-dupe-class-members`) both require that duplicate walker,
so neither can be proved or credited as unblocked by this partial package.
This is a shared dependency stop, not a language gap. There is no compiler
refusal to report. Resume those two helpers and the rule proof when NameTagged
lands. The three independent helpers have complete upstream-call coverage.
