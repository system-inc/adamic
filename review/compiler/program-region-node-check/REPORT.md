Built a tsc declaration projection with parent, declaration-symbol and symbol-declarations links; the selector selects Node and NodeArray<Node> and their allocation sites.
Delivery branch compiler/program-region-node-check extends bbc3eb356b7de27ee86f18841e76137fee6ee67d; runtime base remains f2de9280c19f9b427d0cc5142accb0ba940365be.
The new source matches Node in both backends with ASan/UBSan and leak checks; the host and empty parse.a controls still pass; counts has one new row.
The Node and NodeArray drops now fail TestProgramRegionTscNodeMemberSites; applicable host/control and literal-array-adoption mutants also fail.
Full NodeArray metadata construction, inherited-array indexing and the full tsc.ts build are not certified; the smallest cycle-preserving subset and its boundaries are recorded below.

Source and subset

The source is cohere/TypeScript/tsc/testdata/fixtures/compiler/types.ts at TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f, cohere 7945d102a6c18dd36adf9114a758ce646e8b2359. tsc-declaration-provenance.json records the file hash and exact declaration excerpts. Node.parent is at 948; symbol belongs to Declaration extends Node at 1758, not directly to Node in this source. Symbol.declarations is optional Declaration[] at 6040. NodeArray<T extends Node> extends ReadonlyArray<T> and carries position and transform metadata at 1589-1592.

internal/oracle/testdata/program_region/tsc_node_membership.a preserves readonly parent links, mandatory Declaration.symbol, optional mutable Symbol.declarations, and readonly NodeArray elements. Non-cycle fields and any brands are omitted. Parent admits undefined for the honest unbound root: no undefined value is advertised as initialized Node. The factory initializes each represented field in its constructor, constructs a root and two children, then binds the declarations array back to the symbol. All three nodes hold that symbol, creating the actual strong cycle Symbol -> declarations -> Declaration -> Symbol. Child parent links reach the root. A separate NodeArray<Node> holds the same nodes and keeps array identity. No unchecked representation or name-based membership rule was added.

Two source boundaries justify the supported projection. tsc-node-array-full.a keeps NodeArray position and transform metadata, but its array Object.assign construction stops at 26:46 with `stage 0 can't lower Object.assign on a shape not proven by a plain literal or its const binding yet`. Removing metadata while retaining `interface NodeArray<T extends Node> extends ReadonlyArray<T> {}` stops at its element read, 47:18: `stage 0 can't lower a computed key without an own data field yet`. The positive witness therefore uses the equivalent element-only `type NodeArray<T extends Node> = readonly T[]`. This certifies cycle membership of the element-array projection, not metadata layout. Both stopped witnesses and diagnostics are under review/. No compiler or runtime implementation changed.

Selector and allocation sites

TestProgramRegionTscNodeMemberSites obtains the actual checker identity for the Node declaration and the concrete NodeArray<Node> variable, then requires both identities in programPlan.types. It checks the source allocation types and contextual types against that same member set, prints locations, and checks emitted IR allocation metadata. This uses the selector's own output, not counts and not a name-based override. Names only identify the declarations being asserted.

| Allocation | Source site in tsc_node_membership.a | Selection |
|---|---|---|
| new BaseNode factory, executed three times | 34:12 | member |
| empty Symbol.declarations array | 40:43 | member via Declaration[] context |
| populated declarations array | 44:41 | member |
| NodeArray<Node> element-array allocation | 46:41 | member |

The symbol object at 40:28 is also a region member. Runtime execution adopts seven values: the symbol, three nodes, and three arrays. Counts are allocations 11, frees 4, retains 19, releases 26, peak 11, regions 7; allocations = frees + regions. Four temporary output values are freed ordinarily. Source Node stdout is `3|0|3`, then `true`, `true`, `true`, each on its own line. Native, JavaScript, sanitizers and leak checks agree.

Controls

TestProgramRegionHost14MemberSites and TestProgramRegionHost14AgreesWithNode still hold fixture 14's two captured cells and two closures, with its unchanged 11 allocations, 7 frees and 4 region values. TestProgramRegionScoutParseSelectionEmpty still reads the pinned parse.a source bundle and requires empty type, cell, function and class selections. The acyclic parse tree stays counted. host-only-report.md and host-only-mutants-results.json preserve the previous unit's missing-witness finding; it is no longer the current parser-node witness.

Tests, mutants and counts

Focused new tests passed: TestProgramRegionTscNodeMemberSites 0.04s; TestProgramRegionTscNodesAgreeWithNode 14.02s including native setup. Combined witness/control rerun passed: named selector test 0.10s; host member sites 0.46s; parse selection 8.26s; source graph oracle 0.98s; host oracle 1.41s. All new leaves are below 60s and use t.Parallel. Commands used timeout 90 and go test -timeout 60s; focused vet of ./internal/lower and ./internal/oracle passed. Logs are tsc-witness-tests.log, combined-witness-controls.log and tsc-vet.log.

run-mutants.py runs the same requested selector drops against the new Node witness. drop-Node fails with `selector left Node (Node) counted`; drop-NodeArray fails with `selector left NodeArray<Node> (NodeArray<Node>) counted`. Both execute the test and fail its explicit member-set assertion; neither is caught by a build error, lowering refusal or sanitizer crash. drop-host-callback-cell fails the host member-site assertion. select-acyclic-type fails the empty parse selector assertion. Mutant commands and results are in mutants-results.json; per-test logs record their failing leaves.

The independent drop-witness-array-adoption overlay changes the literal-array emitter's adoption path, leaving the three arrays counted. Both backend behavior and sanitizer/leak checks pass; the new oracle then fails its member assertion: regions 4 instead of 7 (allocations 11, frees 7). The first attempt targeted programArray, a helper not used by literal arrays, and survived. That attempt is preserved as unused-array-helper-mutant.log; the corrected emitter mutation is the caught proof, not a retired live check. No changed runtime/compiler source is delivered.

Counts regenerated once with timeout 240 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 180s -args -update-counts. Existing aggregate counts driver passed in 129.644s; it is not a new test leaf. Exactly one row was added, `Program region: internal/oracle/testdata/program_region/tsc_node_membership.a | 11 | 4 | 19 | 26 | 11 | 7`. It records the seven members and four freed output allocations above. No existing row changed. The fixture is registered in programRegionCountFixtures. Off-mode lowering refuses Symbol.declarations with adamic/cycle-capable, confirming the fixture's a-check header. This is the existing intended off-mode behavior, not a change to selection.

Setup remains the same authorized runtime base: GOPROXY=https://proxy.golang.org|direct; Go 0.023s, Node 0.023s, markdown 0.066s, submodules 0.068s, clang 0.162s, Go build 47.800s, cache warm 47.995s, done 48.028s; nproc=5, CPU quota=4. Every command had a hard limit and running jobs were checked. No whole-package gate ran. Evidence lives under review/compiler/program-region-node-check, with mutant sources ending in .go.txt.

This delivers the type-based membership check for the permitted faithful cycle subset. It does not claim a native full tsc.ts build, NodeArray metadata layout or Program-region coverage of every parser allocation. Those require the corresponding lowering/build work. The actual host cell checks and empty acyclic control remain in place.

Integration lane checks exited 0: lane checks 11.9 s: gofmt and tools on 259 Go files, t.Parallel on 33 test packages; no t.Parallel analyzer on this tree; vet skipped, over 10 s. Focused vet of both touched packages passed separately, and both new tests call t.Parallel. No absent analyzer or skipped lane vet is claimed as run.
