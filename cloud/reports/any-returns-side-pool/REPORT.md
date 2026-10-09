Built: merge current main and prove simple generic field writes against the concrete field; preserve the unsupported tagged-write boundary when proof fails.
Commits: parent e79f12eb; exact comparison base 3d1095b3 (parent c79c7572); merged main 8afb1ee4, which contains requested 745dc0bb. Delivery is the merge commit containing this report.
Commands and outputs: all 18 named cases ran independently with -timeout 90s on both comparison trees and the final merge; own focused checks, build, vet, stage3 records, counts and mutants pass.
Mutants: all nine previous any-returns mutants and the new removed-field-proof mutant are caught by their assertions or pinned refusals, without build failures.
Not covered: inherited stack failures were classified and left for their owners; no full gate, census replay or additional a-check was run. No new fixture files.

The original e79f12eb run reproduces all 18 reported failures. The exact stack base has 14 deterministic failures in its first run, plus the signal mutant's failing repeat. Three failures are any-returns regressions: the lower TNode mutation test and its native/WASI NotYet entries. None is class c. The final main merge has those three green and the signal mutant green, with the same fourteen inherited failures still red. Every selected run has matching test events; none is classified from a no-tests run or skip. WASI is enabled with the installed SDK. Raw JSON events, exact selectors and exits are retained beside this report.

| Check | Stack base | e79f12eb | Main merge and repair | Class and owner |
| --- | --- | --- | --- | --- |
| skipcensus | red | red | red | a: compiler/test-split-native ed05fde6; compiler/test-split-bridge 93cd7227; stage1-split/printers 16fc3639. Main merge also exposes the stage3 fixture split inventory drift. |
| flow | red | red | red | a: compiler/test-split-flow e6a6174e. Base misses 38 entries; any-returns adds two fixture entries to that inherited drift. |
| fresh | red | red | red | a: compiler/test-split-flow e6a6174e. Base misses 38 entries; any-returns adds two fixture entries to that inherited drift. |
| fuzz | red | red | red | a: integration/runtime-stack-on-c1 5aa673ed, through runtime merge 234e2eb0; coordinate with codex/fuzz-flaked-under-load. Leak death is reported as a finding, not confirmed as a flake. |
| ir | red | red | red | a: compiler/hidden-boundaries c9beed41. The already-delivered reader fix 573bd3e7 is absent from this stack base; not cherry-picked here. |
| tnode | green | red | green | b: codex/any-returns-concrete. Exact substitution bypassed the constrained mutable-field refusal; fixed here. |
| representation | red | red | red | a: compiler/hidden-boundaries TNode storage 702c3ecb/8fc4be59 versus the Source representation test at 963e7c53. Tagged Union storage is now known; test still expects unknown. |
| signal | green, then red on repeat | red | green | a: integration/runtime-stack-on-c1 5aa673ed, runtime/slice-3-concurrency. Schedule-sensitive mutant witness: base pass then fail; identical standalone runtime sources across base and any-returns. |
| uint16 | red | red | red | a: runtime Uint16Array 277e3937 carried by integration/runtime-stack-on-c1; stale new-expression NotYet expectations. |
| optional | red | red | red | a: compiler/optional-presence a774d316 with views-v1. Exact checked-view refusal text changed; source-slot integration still pending. |
| native-uint16 | red | red | red | a: runtime Uint16Array 277e3937 carried by integration/runtime-stack-on-c1; stale fixture admission record. |
| native-uint16-notyet | red | red | red | a: runtime Uint16Array 277e3937 carried by integration/runtime-stack-on-c1; stale fixture admission record. |
| native-tnode-mutation | green | red | green | b: codex/any-returns-concrete; same concrete field-write regression. |
| native-binder | red | red | red | a: compiler/optional-presence a774d316. Binder-flow lowers, but oracle still registers NotYet. |
| wasi-uint16 | red | red | red | a: runtime Uint16Array 277e3937 carried by integration/runtime-stack-on-c1; stale fixture admission record. |
| wasi-uint16-notyet | red | red | red | a: runtime Uint16Array 277e3937 carried by integration/runtime-stack-on-c1; stale fixture admission record. |
| wasi-tnode-mutation | green | red | green | b: codex/any-returns-concrete; same concrete field-write regression. |
| wasi-binder | red | red | red | a: compiler/optional-presence a774d316. Binder-flow lowers, but oracle still registers NotYet. |

The signal mutant is schedule-sensitive, rather than an any-returns dependency: its standalone runtime sources are identical on the comparison trees, and it first passed then failed on the stack base without any code change. The repeat has the same signal-terminated-without-TSan-witness failure as Loom and the any-returns run. No change or retry was made to this runtime test. Its green post-merge observation is reported as an observation, not a repair.

Flow/fresh coverage already misses 38 paths on the base (826/864 and 801/839). Any-returns adds any_returns_concrete.a and node_namespace_timeout.a, producing 826/866 and 801/841. This is an inherited failing unit with an additional two-path contribution from this branch, not a green-base regression. Per the routing instruction, the generated manifests are unchanged; compiler/test-split-flow must refresh the complete merged corpus, including those two paths. Main's stage3 test split also adds stale metadata to the already-red skip inventory; its first failure remains the native checker split metadata.

The own repair reuses refuseInstantiatedMutation, rather than weakening the fixture. Node prints cat when rename<TNode extends Node> writes holder.node.name = 'cat' into Dog.name: 'dog'. Exact checker-owned substitution makes TNode concrete and bypassed the old Union-receiver stop. For a simple field assignment through a tagged generic object constraint, the repair requires one concrete receiver, a known field, a non-any right-hand type, and assignability to that concrete field's type. Otherwise it preserves the existing NotYet text at the write. Valid numeric setTextRange writes remain admitted and agree with Node. Compound assignments and union receivers remain conservatively stopped under this rule; this unit does not claim their mutation proof. The new mutant disables only this proof guard: the TNode refusal test then fails with want constrained mutation refusal, got <nil>. This is an assertion witness before emission, not a clang kill. The native and WASI NotYet entries now pass too.

The merge has eleven conflicts, each resolved keeping both meanings:

- internal/javascript/javascript.go: keep undefined for uninitialized definitions and use main's viewFieldRepresentation for semantic tags.
- internal/native/emit_objects.go: keep actual-slot indexing and optional presence for packed caches; use main's fieldInitialRepresentation and tuple marker; keep metadata for reserved missing fields.
- internal/native/emit_statements.go: use main's fieldRepresentation with the actual-slot index, not the removed cache.index.
- internal/native/reuse.go: combine presence publication and actual-slot indexing with main's initial semantic representation and plain-object tuple reset.
- internal/native/runtime/adamic.h: keep tuple, dynamic shape/type state, order accessor, readiness and representation tails; use main's shared adamic_object_size formula, including the order tail.
- internal/native/runtime/object.c: retain dynamic shape, presence/order and packed-cache operations; use the shared allocation size and initialize tuple false, including dynamic spread copies; keep main's undefined-reference write admission.
- internal/native/runtime/region.c: use shared allocation size, initialize tuple and dynamic state, and retain insertion ranks and both byte tails.
- internal/oracle/counts.md: regenerate from the merged compiler, never hand-merge rows. Adds 50 main view-fixture rows, changes no existing row and removes none; every added path is in count-changes.json. The TNode constraints row remains allocations/frees 8/8, retains/releases 4/12.
- stage3/fixtures/assertions/status.json: keep main's views-v3 array-arm NotYet for 09_interface_kind and validate the merged record through stage3 tests.
- stage3/fixtures/predicates/status.json: keep main's Compiles record for 03_void_zero and validate it.
- stage3/fixtures/taste/status.json: keep main's Compiles record for refused/21_truthy_loops and validate it.

Main was merged with a merge commit, never rebased or forced. The any-returns generic mapper and overload inference were retained by the clean merge. Stack-owned reader fixes, stale admission records, corpus manifests and runtime tests were not patched or cherry-picked here. The source guard is the only new lowering change.

Additional validation:

- go test ./internal/lower -run '^TestHiddenTNodeConstraintMutationRemainsNotYet$|^TestGenericResolvedReturnContracts$|^TestGenericReturnFixtureLowers$' -count=1 -timeout 90s -v: pass.
- Node executes the unsafe TNode fixture directly: exit 0, stdout cat, empty stderr.
- go test ./internal/oracle -run '^(TestNativeAgreesWithNode|TestWASIAgreesWithNode)$/^internal$/^oracle$/^testdata$/^(any_returns_concrete|node_namespace_timeout|overload_sort_deduplicate|hidden_boundary_generic_tnode_constraints)[.]a$' -count=1 -timeout 90s -v: pass, uncached. Source Node, JavaScript, release and sanitized native with leak checks, and real WASI all agree.
- go build ./... and go vet ./internal/...: pass.
- go test ./stage3/fixtures -count=1 -timeout 90s -v: pass.
- GOMEMLIMIT=2GiB go test ./internal/oracle -run '^TestCountsAreRecorded$' -parallel=2 -count=1 -timeout 30m -args -update-counts: pass, table regenerated.
- GOMEMLIMIT=2GiB go test ./internal/oracle -run '^TestCountsAreRecorded$' -parallel=5 -count=1 -timeout 30m: pass, final table checked after the concrete union-receiver guard was added. The longer timeout applies only to the table-wide record sweep, not the 18 classified checks.
- Overlay mutant runner: unsubstituted-return and trust-any caught by TestGenericResolvedReturnContracts; opaque-mapper by TestGenericMapperKeepsCheckerIdentity; copier-loses-alias by TestCheckerMapperAliasKeepsIdentity; ignore-namespace and ignore-unresolved-alias by TestNodeNamespaceAnnotationUsesPinnedTimeout; replace-local-namespace by TestNodeNamespaceLocalDeclarationIsNotReplaced; accept-unresolved-type by TestNodeNamespaceUnresolvedTypesStayRejected; drop-overload-inference by the Node-held overload_sort_deduplicate oracle; drop-generic-field-proof by TestHiddenTNodeConstraintMutationRemainsNotYet. Each Go invocation uses -timeout 90s. Ten caught; runner exit 0.

Assumption: a concrete read/return mapper alone supplies no mutation permission; an unproven constrained field write retains the pre-existing tagged-write refusal. No broader generic mutation admission is claimed. This preserves roadmap step 16's concrete return contracts while repairing their mutable-field interaction.

Setup used GOPROXY=https://proxy.golang.org|direct and succeeded after local submodule Git worktree metadata was repaired to the actual pinned cohere and TypeScript locations. The base references the same pinned dependency checkout; no cohere code was copied. nproc 5, CPU quota 4. Setup timing lines: Node 0.023s; markdown dependencies 0.085s; submodules 0.095s; clang 0.182s; go build 37.310s; build cache warm 37.441s; done 37.467s. Setup log retained. Go 1.27.1, Node 24.19.0, clang 20.1.8, Linux on the shared EPYC workspace.
