Built: explicit read-demand inventory, shared lane 3 receiver queries, and both-backend helper guards; lazy admission remains unfinished.
Commits: ce2545e4 pushed the first table before further work; the receiver/oracle follow-up is recorded in branch history.
Commands: 66,960 exact spans/12,181 pairs audit passes; helper oracle passes in 3.142s; oracle vet and diff checks pass.
Mutants: removing the helper read guard runs valid native/JavaScript code and is caught; trusting an Unknown helper receiver is caught by the measurement audit.
Uncovered: exact view propagation, implicit spread/rest/destructuring-assignment reads, missing-support attribution, lazy contracts, nullish/optional completion date and current-main landing.

The first table was pushed at 21:14 UTC October 7, before the 23:00 UTC
(17:00 MDT) deadline. It deliberately labels its result as conservative
Unknown-fallback demand, not completed closed-world reachability or missing
support. The old transitive 2,936-site family columns must not drive priority.
The full table is in docs/checked-views-blockers.md; compressed site and pair
artifacts preserve type identity, field, declared type family, count and witness.

| Family | Distinct pairs | Explicit read sites |
| --- | ---: | ---: |
| Nullish | 2,011 | 9,084 |
| Optional | 1,681 | 7,863 |
| Callable | 2,818 | 11,063 |
| Array/tuple | 389 | 2,117 |
| Mixed primitives | 63 | 690 |
| Objects/primitives | 73 | 267 |
| Untagged objects | 84 | 186 |

Deduplicated lane totals: lane 1 2,370 pairs/11,321 reads; lane 2 3,207/13,180;
lane 4 213/1,111; unassigned 780/3,721. Counts overlap between lanes. Immediate
field families are counted, not recursively reached descendants. This is not
proof that every read in a supported family still lacks compiler support.

Lane 4 priorities now start with __String (543 reads), string|number|undefined
(66), string|NodeArray<JSDocComment>|undefined (63), the CommandLineOption.type
string-literal/Map union (33), then an untagged node/name union (26). Exact names
and witnesses are in the table/summary. __String still needs intersection/brand
support; it is not silently admitted as an unbranded string. The old false-array
and false-VersionPaths shapes no longer lead read demand.

The receiver overlay is built in a detached worktree at lane 3 c1f4c5a7, with
its exact cohere and TypeScript pins checked out as submodule worktrees. No
cohere implementation was copied, and no producer solver was duplicated.
The small extension asks the original source adapter for receiver producers,
then calls assignAllocationSites/newAllocationFlowGraph/ReachingAllocations.
Its loader/lowering cannot emit an executable rejected program.

The unadapted corpus has zero stock TypeScript diagnostics but 6,037 native
checker diagnostics. The source adapter records 412 allocation schemas and
69,621 read-receiver queries, 67,141 with unknown allocation frontiers. It joins
all 66,148 explicit property/element inventory reads; 812 destructuring bindings
retain Unknown without a graph query. 2,457 joined receivers have known
allocations and 40 overlap allocations reaching a cast. Unknown cast frontiers
can still route those or other allocations through unsupported load/callback/
generic edges, so zero receivers are certified non-viewed. All 66,960 read
checks remain. These are conservative may-flow facts, not execution counts.
Raw query/diagnostic evidence and source hashes are committed.

The helper fixtures pass ordinary and viewed values to one interface-taking
helper. Source Node prints true/true for the valid fixture and true/42 for the
malformed fixture. Native and JavaScript stop after true with exit 70:

    adamic: panic: field read failed: value.unsupported is not a boolean; expected boolean, found number

Mutating only the helper Property.View marker produces true/false in release C
and true/42 in JavaScript, both exit 0. The ordinary pinned exit/message assertion
catches them. The unsupported string|number helper is refused by shared lowering
at its read location, with the exact type message pinned. This is conservative
compile refusal and current scalar helper-check evidence, not a lazy union pass.
The artifact mutant sets that shared helper receiver's retain_check to false;
the independent audit rejects it without relying on a changed aggregate count.

Current admission is eager in this branch and lane 1 9b85eada: viewObjectFields
walks every descendant, viewContract builds all children, and a global field-name
scan also rejects uses. Lazy supported runtime payload checks do not make that
compile-time admission lazy. The exact deferred-descriptor/read-query/dispatch
handoff is in docs/checked-views-plan.md. Lane 4 did not modify lane 1 compiler
files. Without those hooks it cannot claim unread unsupported fields are allowed
or commit an honest completion date for lane 1 nullish plus optional. This is an
owner/integration dependency, not an estimate of a finished implementation.

Lane 2 0b141c26 remains merged. Fetching lane 1 9b85eada and attempting its merge
conflicted in shared cast.go and the plan; the merge was aborted without editing
another lane's files. Lane 3 is a pinned read-only dependency in the isolated
measurement worktree. No PR was opened and only codex/views-mixed-unions is pushed.
Per-cast only-lane-4 and lane-4-plus-other counts under the new rule are unmeasured.

Commands and outputs are in lane4/logs (all runs redirected, never piped):

    node read-demand.cjs ...                 PASS, stock diagnostics 0
    python3 audit-read-demand.py ...         PASS, exact spans/pairs/families/source hashes/helper witness
    python3 make-read-flow-overlay.py ...    PASS, shared flow overlay
    go build -overlay=... ./.../latent/tool PASS
    /tmp/views-read-flow-meter ...           PASS, analysis-only rejected program
    python3 export-read-flow.py ...          PASS, exact property/element joins
    python3 read-demand-mutants.py ...       PASS, trust-Unknown helper mutation caught
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewLane4' -count=1 -v -timeout 10m
                                            PASS, 3.142s, both backend runtime mutants caught
    go vet ./internal/oracle                PASS
    git diff --check                        PASS

The full gate was not rerun. Earlier TestSharedArrayContractAdapter fails on the
unchanged baseline as recorded in REPORT.md. Initial setup timing remains
68.321s total, nproc=5, quota=4, env /workspace/adamic-tools/env.sh; Node .087s,
Go .113s, submodules .120s, clang .737s, markdown 1.388s, build 68.078s and
cache 68.206s. GOPROXY was set to https://proxy.golang.org|direct.
