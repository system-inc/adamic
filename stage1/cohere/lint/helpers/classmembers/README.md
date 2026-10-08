# Class member helpers

Cohere pin: `7945d102a6c18dd36adf9114a758ce646e8b2359`.
Fresh ports; no partial implementation existed on the audited origin branches.
The owning claim is [ecmascript-classmembers.md](../claims/ecmascript-classmembers.md).

- [MemberName](member_name.a): accepted node kinds and raw Name() identity; -1 means nil.
- [IsOverloadSignature](is_overload_signature.a): methods/accessors with no body, including abstract ones.
- [IsAccessorKind](is_accessor_kind.a): getter or setter kind only.
- [KeyOf](key_of.a): shared NameTagged plus static/private parser facts.
- [ForEachDuplicate](for_each_duplicate.a): first-seen member collision and accessor pairing.

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

## Dependency and rule proof

[NameTagged](../property/name_tagged.a) is now provided by the shared property
package, on branch lint-helpers/property-name-tagged at
`ecfd7ffc1e4d37c8b1933c8b339913e2cd3c9a59`. It is imported rather than copied.
Both core and TypeScript no-dupe-class-members consumers are covered by the
live capture. The [core rule](../../rules/no-dupe-class-members/rule.a) proves
this package against 37 unique captured upstream cases and two witnesses.
The TypeScript extension is now helper-ready; its earlier claim remains with
its owner. No shared dependency or language gap remains for this package.

The duplicate walker deliberately retains the first member kind. Consequently,
getter/setter/setter and setter/getter/getter do not report a duplicate in Go;
the port preserves that behavior. Numeric and string keys remain distinct,
and the Go tagged name is preserved in diagnostic text.
