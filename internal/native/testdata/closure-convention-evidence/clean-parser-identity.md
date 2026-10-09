# Convention-only parser identity

Measured compiler commit: be3926dcb5c0fffdff21e63cff9607e30131f26d.
Both parser C and release binary compare byte for byte: **cmp exits 0**.
This is the same merged tree on both sides, not a comparison with main.
The baseline is that exact commit plus count-transport-revert.patch in the
never-pushed detached worktree /tmp/closure-no-count-baseline.

The counterfactual patch changes exactly three files. It removes counted
pointer shapes, counted closure constructors, counted dispatch and count tags
from runtime/adamic.h and runtime/closure.c. Its IR selector rejects a program
which needs counted transport rather than silently accepting it with the wrong
ABI. The original observer predicate is retained solely for that rejection.
The exact reader probe exits 2 before C is emitted, with the explicit baseline
rejection. This control proves counted support really is absent. It is a
measurement tree, not a compiler intended to ship.

All other features are retained unchanged: front-3, host-blockers, census
optional admission and caller padding, checked views, canonical/receiver support,
no-cast enforcement, object state, signal reset and both carried owner modules.
The same cohere SDK pin is referenced locally; no cohere source is copied.
The directory's local submodule symlink changes its Git path mode only; the SDK
commit and contents are identical. The saved patch excludes that path adaptation.

Both compiler binaries use go build -buildvcs=false. Both compile the same
canonical /workspace/adamic/stage1/typescript/parser/main.ts entry, with the
same default release flags (-O2) and clang20.1.8. C and binary hashes, raw cmp
results and parser stdout/stderr are adjacent. Each whole-parser run prints
10523, exits 0 and has empty stderr. --count is the parser driver's AST count
mode, not a counted heap build. No hardware instruction/cycle measurement is
claimed; the CPU PMU remains unavailable.

## Four optional method thunks

They originate in f69bf608's optional/default function-value admission on the
census branch: it permits MaybeBoolean method parameters and originally adds
an always-count callee convention. Reconciliation replaces compatible missing
slots with caller padding. The extra parser entries were left behind because
function-signature assignability alone admitted same-signature methods even
when the actual interface receiver could not hold those classes.

Removed parser thunks: Parser.type, Parser.assignment, Parser.allowInAssignment,
and Statements.type. The no-reader parser emits none of those four thunks.
Optional method admission now also checks compatible source structural receivers;
count analysis still uses its original closed-world function-type candidate set.
Derived instances using an inherited implementation are included. Generic and
computed receivers retain conservative admission. Hand-built IR with no receiver
proof retains the previous conservative fallback.

The new optional_method_thunks.a fixture keeps a required optional method, rejects
an unrelated same-signature class lacking a required facade field, and exercises
an inherited optional method through a wider interface. Release and ASan/UBSan
native execution match source Node. Its required optional slots use caller
padding and no count. Removing required thunks still compiles, then exits 70
for a missing method; Node succeeds, so the oracle catches the mutant. Forcing
unneeded thunks back into the parser fails TestParserHasNoUnusedOptionalMethodThunks
with unused optional method thunk: Parser_type. Raw proving logs are adjacent.

## Owner sites and validation

All four former external sites are now carried, from Adamic snapshot 132d9919:
node_process.c:442/447 (originally 436/441), parallel.c:197/300. Builtin definitions
have prior declarations through adamic_code_function, and their static closures
use designated initializers for every feature-dependent layout. Every call uses
adamic_closure_call, with actual counts, without a function-pointer cast.

Both complete owner translation units compile under -Werror in plain, counted,
and counted plus receiver/canonical configurations. Dropping the count at each
of the four sites fails compilation: expected 3, have 2. The six runtime sites
and all indirect pointer invocations are listed in runtime-sites.json and
final-runtime-call-audit.log. Broad host/shared-heap activation remains owned by
those lanes; this compiler admits neither new operation. Unused translation
units stay unselected, including direct cache harnesses, and no dependency stub
was invented. C object compilation is claimed, not a full threaded scheduler or
performance-host execution proof.

Full touched suites: native 193.860s, lower 53.986s, IR 17.792s pass. The first
native attempt found direct cache harnesses bypassing owner-module selection;
be3926dc moves that selection into the shared cache entry point. Its complete
native rerun passes. Final uncached reader, optional, nested, inherited and host24
Node gate passes in 16.961s, with release/sanitized readers and both backends.
Touched-package go vet passes. No complete repository gate was rerun.

Taste21 remains the recursive checked objectType view gap. Its refusal is not
attributed to the convention; its source and expected-success fixture remain
unchanged for the checked-views lanes.
