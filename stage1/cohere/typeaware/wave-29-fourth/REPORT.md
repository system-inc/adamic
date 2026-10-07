Built: three numeric-node JSX partial ports, per-node fragment decisions, raw declaration/tag/construction kernels and rule.json subscriptions; old React SSA claims parked.
Commits: claim 991a851a pushed before code, based on current main f8013f0b; implementation/evidence sha reported in handoff.
Commands: setup 27s, nproc 5; 74 parsed TSX cases, 79 rows, 7359 identical Go/native/Node/JS/sanitizer bytes; three positive Go source witnesses reproduce native JSX blockers.
Mutants: fragment import module (3 rows), intrinsic dash predicate (2 rows), conditional branch preference (2 rows) all compile, exit zero with empty stderr; only Go comparison catches them.
Uncovered: full source-rule findings/corpora/fixes/suggestions and timing, checker lifetimes, binding acquisition and context memo/escape/capture analysis remain incomplete; no shared files edited.

## Claims and landing

The previous three React hooks/SSA claims are parked with their blockers named
in claims/wave-29.md, per Ahra's instruction. Their pushed evidence remains in
wave-29-third. The only pushed branch is codex/typeaware-wave-29, already rebased
and validated on f8013f0baac41ddc340d76f83bddde38536a8f07 through 18ff19c8.
Main stayed unchanged throughout this continuation. No main/area push or PR.

An all-origin scan covers 529 refs, 33 unique Markdown claim blobs, the complete
197-rule combined-volume ranking and the 26 documented baseline ports. It leaves
18 candidates. The first three unclaimed names are react/jsx-fragments,
react/jsx-no-constructed-context-values and react/jsx-no-undef, all default
volume zero. Main and the bridge branch contain only their inventory/count
mentions, not ports. The complete scan/provenance is in selection/. They do not
call the parked React high-level IR/SSA validator pipeline. The context rule's
separate AST memo/escape analysis still needs an adapter; it is not implemented
by the direct construction kernel.

## Numeric node contracts and implementation

Each rule owns rules/<name>/rule.a and rule.json with numeric kinds. Exact live
production Go registrations are fragments [285,286,289], context values and
undefined tags [286,287]. syntax_kinds.a constants are checked against the pinned
Go parser on every kernel run. RawNode carries numeric kind and raw text/parent/
child-role/operator facts. It is an owned probe model, not a newly published
shared node protocol, bridge question or source parser. No Go lint decision is
exposed to a native runtime. The raw facts come from independent Go parsing for
the kernel tests, not from a runtime Go rule callback.

The fragment handler receives the current node and related raw declarations;
it decides both modes, checker availability, attributes, exact React.Fragment
shape and imported/variable/destructured fragment bindings natively. The shared
source adapter to supply those declarations has not landed. Declaration tests
cover react/preact, aliases, other exports, qualified initializers, destructuring,
require callee/module, empty require and parenthesis declines. The Go helper's
counterintuitive acceptance of const {Other}=React is preserved, not corrected.

Undefined-tag logic follows the leftmost numeric property-access root, preserves
lowercase member roots, declines this/namespaced tags, applies ASCII/non-ASCII,
underscore/dollar and dash semantics to bare identifiers. Checker resolution,
current-file declaration origin, allowGlobals/commonjs policy and final reporting
are not integrated. No untested default-silent scope predicate is shipped.

The direct context kernel covers ten construction spellings, parenthesis/as
unwrapping, left-first conditional/logical recursion, assignment relabeling,
property-access usage anchoring and the production recursion limit. Messages
preserve IDs, remedy choice, construction kind, node/usage line interpolation
and identifier usage. Variable quoting currently accepts printable ASCII without
quote/backslash only, and refuses other names explicitly. Identifier construction
lookup refuses explicitly; provider resolution, component detection, memo factory/
dependency/callee/primitive inference, escaped holders and full reporting are not
integrated. This is a direct-syntax kernel, not a full rule entry point.

No rule tests its kind as a string or refetches the handed primary node. Related
child/declaration nodes are followed by numeric IDs. The probe hands a node to
one requested helper/handler; it is not an all-rules-on-every-node driver. Missing
source integration is explicit; no factory/generator registration claims these
partial kernels are ready production rules.

## Independent comparisons and mutants

The Go overlay calls unchanged private production helpers and, for both fragment
modes, actual production Run callbacks. It parses 74 real TSX inputs and emits raw
node facts separately from expected rows. Native produces 79 rows/7359 bytes,
sha256 f4d1fe2af24838bc1101de970acf6c8edc35baaba71208411cd4103c4499da55,
identical to Go, source Node, emitted JavaScript and ASan/UBSan/LeakSanitizer;
native/Node/JS/sanitizer stderr is empty. This includes helper decisions and
context message bytes; it is not the canonical full-source findings stream.

One mutation per owned rule compiles and executes with exit zero and empty
stderr: require preact rather than react on import declarations (3 changed rows),
replace the intrinsic dash exclusion with underscore (2), reverse conditional
branch preference (2). Only independent Go bytes catch them. These are kernel
mutants, not the required completed full-rule mutants. Commands and full outputs
are in validation/. No checker archive is used by these kernels, so no released-
handle or bridge sanitizer claim is made for this batch. The completed six-rule
landing gate already checked released handles on this same main revision.

## Source blockers and pinned Go failure

A separate independent production oracle runs actual configured Run callbacks
on three JSX sources. It emits three findings, zero fixes/suggestions. Native
shared-parser probes compile, but fragments exits zero after producing only a
TypeAssertionExpression and no JSX nodes. The context value fails with exit 70,
expected GreaterThanToken/got Identifier at offset 42; undefined-tag JSX fails
with exit 70, expected GreaterThanToken/got SlashToken at offset 22. The shared
parser/driver, numeric node/fact preparation and checker-backed declarations are
not changed here. JSX support is landing on area/stage1-lint. The existing
numeric arrays/manifests do not make the shared parser produce numeric JSX nodes.

The initial direct-construction oracle also observed a production Go panic on
({} satisfies object): ConstructionOf calls AsAsExpression on a SatisfiesExpression.
The complete panic is preserved in validation/pinned-go-satisfies-panic.log. The
fixture is excluded from supported comparisons, and the native kernel refuses
that shape explicitly. Matching a Go runtime stack trace byte for byte is not
claimed, and this worker did not fix the pinned Go rule or its shared accessor.

No full compiler/repository lint comparisons, full-rule mutants, full-source
emitted-JavaScript checker adapter, full gate, bridge lifetime checks or complete
lint timing were run for this batch. There is no native-versus-Go performance
claim for partial kernels. Claim status remains partial/blocked, not ported.

```sh
source /workspace/adamic-tools/env.sh
python3 -u stage1/cohere/typeaware/wave-29-fourth/check.py /workspace/wave29-fourth-final > /tmp/wave29-fourth-final.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-fourth/prove_gaps.py /workspace/wave29-fourth-gaps > /tmp/wave29-fourth-gaps.log 2>&1
```

Setup: Go/clang/Node/submodules ready 0s, cache warm 27s, total 27s; nproc 5,
CPU quota four cores and 17.6 GB. Native sources are .a. No new bridge question
was added: current tests consume raw parser facts in an owned probe, and a
source adapter is still required before checker acquisition can be integrated.
