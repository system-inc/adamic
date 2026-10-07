# ORM column declarations

A `.a` port of pinned Go cohere's `base/correctness-require-orm-column-declare`.
The property listener preserves all eight decorators, bare and called identifiers,
first matching decorator order, the `declare` exemption, computed-key unwrapping
and private/non-identifier message names. There are no fixes or suggestions.
The message is a subset of the helper's Go-generated policy catalog and is
rendered by the shared PolicyMessage helper.

Status: validated candidate, awaiting `.a` registration support. Default registry
and package commands currently fail. The proposed compatibility patch changes
shared Go files and is deliberately unapplied pending the territory exception.
See [the shared validation runner](../nexus-consistency-no-boolean-outcome/validate.py)
and [evidence](../nexus-consistency-no-boolean-outcome/evidence/).

The owned mutant ignores a bare decorator while keeping call decorators. It
compiles and exits successfully on source Node, emitted JavaScript and sanitized
native, then fails only the comparison with Go because the bare finding is gone.
The witness retains a called decorator finding, so the mutation cannot hide
behind a completely silent listener.

Parity covers cohere's own captured tests, TypeScript v6.0.3's compiler sources,
stage1 sources, and synthetic spans/order cases. Ordinary `.a` self-lint through
the pinned cohere CLI is blocked: that CLI refuses the extension before checking.
The compiler's typecheck and both lowering backends do accept the port.
