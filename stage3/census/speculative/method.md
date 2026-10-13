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
multiple boundaries. Every failed boundary is emitted in a separate complete catalogue, independent of finding deduplication. Boundary syntax kinds distinguish equal-span nodes such as a parameter and its identifier. Exact visitor identities and byte spans are retained; printed
line/column positions alone are ambiguous. Depth is computed after the completed-file stream has
been combined, so dependency attempts and parents which fail after visiting a child
cannot make the result depend on file order. Each site counts its failed AST
ancestors, excluding its own boundary. `root_kind` is the exact existing reason
text, not a normalization or the name of the enclosing declaration. Speculative finding identities include the actual visitor site: file, syntax kind and byte span, plus the original `(kind, where, reason, text)` identity. Distinct constructs at the same printed location remain distinct sites. Repeated attempts of the same physical site and reason retain one source identity and depth. Full/no-stubs retain the original identity; raw records preserve attempt and
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
This unit edits only the census territory. Its main base is
`946a8f095a7fa419a92117406314b7b3d44630f0`; production comparison uses that pinned base. The normal
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

The measurement uses the identical adapted source bytes archived by f6bb0b41,
with TypeScript 6.0.3 pinned to 050880ce59e30b356b686bd3144efe24f875ebc8.
The three census commits were ported without their obsolete compiler base.
PORT.md records the required signature, mapper and nil backend-cache adaptations.
The main compiler already passes the dispatched mapper fixture through generic
binder inference in 6b33ec61; its reverted-function mutant fails. No production
compiler file changes in this branch.

The published table uses one whole-compiler project. Each bounded process loads
all 79 roots and filters only its census walk to one file. Every completed record
is persisted immediately; incomplete files have no claimed complete-file coverage.
Foreign dependency findings remain observations and do not establish that a
whole dependency file was examined. The typed boundary union from completed
records resolves all retained site depths before the independent stock recount.
This is a partial set of whole-project observations, not four projects combined.
streaming.md explains resumability, resource bounds, checksums and mutants.

Only speculative snapshots optimize immutable scalar aggregates: values and
scalar struct fields copy by value, while scalar slices receive fresh storage.
Mutable graphs retain deep copying and private-field/schema guards. Function
slices copy Parameters, OptionalParameters, Body, Environment, FrameEnvironment
and ReferenceParents deeply. The opaque main mapper pointer retains identity.
The six focused snapshot mutants prove ownership and schema checks can fail.
Full and no-stubs retain the original copier.

Historical directory tables remain in directories/ and evidence/directories/.
Their old method and validation are available at f6bb0b41; their observations are
not included in the new whole-project table. The comparison ledger lists every
observable change on the 12 historical files that completed on main. Six other
historical files hit limits and have no complete measured comparison.
