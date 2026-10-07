# Upstream note

TypeScript 6.0.3 constructs type objects without their constraint/index caches,
then initializes those caches lazily. Declare the three caches as optional on
`Type`, and make the parallel union/intersection declarations optional too.
The present-value contracts stay `Type` and `IndexType`; undefined represents
an unfilled cache, as construction and the existing lazy getters already do.

Also let `SymbolVisibilityResult` declare the optional string module-error name
already carried by `SymbolAccessibilityResult`, so a visibility view does not
hide that metadata. These are internal declaration changes. The published API
snapshot is unchanged by this adaptation, and emitted JavaScript must remain
byte identical. No initialization or runtime computation changes are proposed.

Adamic needs these contracts because viewing a structural object through a type
that hides a target optional field is unsound for native field reads. The source
must declare the optional property it actually can carry. This proposal adds
truthful contracts at owners rather than casting individual compiler uses.
