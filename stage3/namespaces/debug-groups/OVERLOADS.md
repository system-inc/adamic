# Debug bodyless overloads

Real site debug.ts:281:5 is assertEachNode's overloaded signature. A reduced
checker-accepted overload violation is in overload_contract.a: Node prints
wrong:string although the number overload promises number. Its source mutant
returns the actual numeric input, prints 1:number and is caught by Node.

The differing-contract production diagnostic remains:

```text
adamic: /workspace/adamic/stage3/namespaces/debug-groups/overload_contract.a:2:2: stage 0 can't lower a function without a body yet
```

Narrow sound acceptance implemented: every runtime parameter type and return
type exactly matches a nongeneric implementation. Required plain parameters
only; no optional/default/rest/this parameters, and no type predicates.
Unused signature type parameters may remain when all those types still match.
Generic implementations and genuinely different overload contracts remain
NotYet. Real assertion overloads require their postconditions to be proved;
skipping them cannot turn conditional/disabled runtime checks into proof.

identical-overloads.a prints 3:8, module:okay, 6 on three lines in source
Node, sanitized native and emitted JavaScript. It covers namespaces, ordinary
module functions and explicitly supplied unused signature type arguments.
The normalized Parser, IncrementalParser and Debug shapes now lower, bringing
the isolated declaration-shape count to nine of ten. tracingEnabled remains
NotYet for its escaping namespace object. Normalized shapes erase real body
and external type obligations; no whole-parser success is claimed.

Actual mutants: disabling the equality-of-contracts guard admits the invalid
overload and fails TestDebugOverloadContracts with boundary lost: <nil>.
Changing the accepted implementation to return 99 is caught against Node
by both backends, with clean native sanitizers. Both checks ran; temporary
production mutation was restored. The IR semantic mutant is retained as
TestNamespaceIdenticalOverloadMutant. Full internal/lower passed (20.384s),
namespace oracle plus mutants passed (6.190s), counts refreshed (37.685s).
The original real Debug census site remains unsupported; only the proven
identical subset moves.
