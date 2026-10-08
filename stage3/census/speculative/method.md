The mode is part of the existing latent overlay. Set `LATENT_SPECULATIVE=1`;
`LATENT_FULL=1` selects the existing full census when speculation is disabled.
`LATENT_MUTANT_NO_STUBS=1` disables speculation and restores that full path.
Without a Go overlay, none of these environment variables adds compiler behavior.

Expression recovery records the actual failure and substitutes an opaque census
value with the checker's type name and its native representation when available.
An unrepresented type uses the census's opaque Union carrier. This carrier is not
proof of a native representation. Statement recovery restores its typed state
snapshot, visits the failed construct's children, and substitutes an empty block.
Signature recovery restores state, supplies parameters and the checked return
representation, and continues through the body. Local-declaration recovery records
the real local-lowering failure and supplies a checked census slot. The walk also
attempts parameter and variable bindings which a failed parent never reached;
their names remain binding positions, not invented expression reads. Generic declarations are examined
as written. The AST walk also reaches children omitted by otherwise successful
lowering and bodies with checker diagnostics. Structural, binding, and type syntax
is traversed; binding names are not invented as value-position reads.

A boundary is the AST construct replaced after a failed statement, expression,
signature, or body attempt. Multiple errors on the same construct do not add
multiple boundaries. Exact visitor identities and byte spans are retained; printed
line/column positions alone are ambiguous. Depth is computed after all files have
been walked, so dependency attempts and parents which fail after visiting a child
cannot make the result depend on file order. Each site counts its failed AST
ancestors, excluding its own boundary. `root_kind` is the exact existing reason
text, not a normalization or the name of the enclosing declaration. Headline
identities remain `(kind, where, reason, text)`; raw records preserve attempt and
phase. Project registration can expose binding failures before a unit attempt;
those records retain the current file context and their actual source site identity. The top twenty ranks those identities by `(kind, exact reason)`, across all
depths; 4+ groups all depths of four or more.

Coverage means speculative examination of every AST child in a parsed TypeScript
file, checked by a second traversal and an independent file inventory. The
per-file visited/attempted sets also include dependency recovery nodes reached
during that file's attempts, so visited_nodes can exceed total_nodes;
unvisited_nodes counts missing nodes in the source file itself. It does not
mean successful lowering or reproduce the hidden census's successful-unit exposure
arithmetic. Parsed source bytes include comments and whitespace. Non-TypeScript
files remain explicitly unexamined. Stock TypeScript 6.0.3 supplies independent
UTF-8 spans, syntax kinds (including enum aliases), and ancestor chains for the
depth audit. All checker diagnostics stay
in the raw header; the no-stubs comparison requires identical diagnostics and
raw diagnostic spans.

The compiler's production Load and Lower are still disabled in every measurement
binary; the latter returns nil IR and the explicit measurement error. The driver
imports no backend and runs both guards with `LATENT_ASSERT_NO_OUTPUT=1`.
This unit edits only the census territory. Its base includes the authorized
mapper-fix merge `69501280`; production comparison uses that merged base. The normal
compiler is separately built from the clean candidate and the working tree; C and
JavaScript output on the unchanged functions.a fixture must match byte for byte,
including with the speculative flag set on the production binary.

Limits: these are findings in hypothetical continuation contexts, not proof that
the original program can compile. An opaque type carrier, an isolated function or
method context, and registrations after a failure can change downstream reasons.
The fallback also probes expression children omitted by specialized lowering,
such as a direct call's callee; those probes can produce first-class-value findings
that a supported direct call avoids. Recovered non-NotYet/non-Refused errors are retained in the raw observations and
reported separately, rather than promoted to compiler blockers. Expressions keep
their partial census state; statement and signature failures roll it back. The
final module-order, ownership, freshness and backend passes are outside this census.
No native semantic oracle or whole-package confirmation is claimed. The controls
are `.a` census witnesses materialized only in scratch directories, not additions
to the native oracle's fixture corpus; this unit adds no native oracle fixtures. Its control ledger is counts.md beside this file.

The measurement base is `a5630a90d05abe85670ba1e00bff3643a2742e53`, the merge of
`69501280a81259fb512edbb8dd0e52c6eb0d88c8` into `ed6e2975`. The checker-owned
TypeMapper alias keeps its identity through snapshots. Checked placeholder types
use the current concrete instantiation. Signature recovery wraps signatureReturn,
including its resolved-return generic path. Earlier measurements on the opaque
mapper copy are superseded and are not included in these results.

Only speculative snapshots optimize immutable scalar aggregates: values and scalar struct fields copy by
value, while scalar slices receive fresh backing storage. Cached mutable field
indices select only the fields that need deep copying. Mutable graphs retain
the existing deep copier and its private-field and foreign-pointer guards. A
focused overlay-only test compares optimized values to the legacy copier and
mutates scalar slices and nested bodies to verify source ownership; a shared-slice
mutant fails it. Full and no-stubs modes use the original copier.
