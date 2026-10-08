Built: merged the fixture base and reduced the parser assertion to an extension-policy fixture; the namespace and enum prerequisite replay is retained on a local held branch.
Commits: delivery fixture-base merge 4f03163bc; parser proof d5c9a8f8c; held namespace candidate c1e42e612; delivery validated at 8bf45dadc; final report-only commit carries the push.
Commands and outputs: delivery lower 51.294s, IR 31.647s, flow 163.247s, JavaScript has no tests; filtered oracle 2.470s; counts refresh 80.483s, all pass. The held namespace candidate fails flow and counts on the two refusals below.
Mutants: three parser fixture/policy mutants, six debugger mutants, two empty-array checks, three Source boundary/output mutants, and three census audit mutants on each comparator are caught; individual logs are retained.
Not covered: a complete native parser or scanner build, checked views, runtime graph-region ownership, and unresolved namespace admissions.

The parser proof ec71eb48 contains 172 non-null expression sites: 168 in its
archived TypeScript slice and four in three standalone probes. Evidence under
`evidence/parser-ec71` records an outcome for each site. Current main 3ffb1a835
refuses all 168 archived TypeScript sites. On the delivered 2c7d9fd2, the guarded
census reaches 68 distinct sites and records checked unwraps. The other 100
have earlier lowering or eligibility boundaries, recorded with their unit
findings. This is measurement on a checker-rejected program, not a successful
parser build: debug.ts:113 and 114 still have Error.captureStackTrace checker
errors. Both census binaries pass the measurement-only audit.

The alleged regression is `native-nonnull.a:5:16`, containing `array[index]!`.
Main, the delivered stack and the local candidate all refuse that .a source.
The same source with a .ts extension builds on the delivered stack and local
candidate; native output is `1`, matching Node. Main refuses the .ts version.
The overload-result probe also matches Node when run as .ts on the stack.
The same-map-side-effect .ts probe reaches its independent unchecked-cast
refusal at 9:37. Its later assertion is not claimed executed. The earlier
scratch compiler's admission of a .a assertion does not establish a main
regression. No production change admitting .a assertions is justified.

The permanent reduction is
`internal/oracle/testdata/non_null_refused/parser_index.a`.
`TestParserNonNullSourceExtension` checks its refusal, writes the unchanged
source to a temporary .ts control, requires a checked assertion, and compares
both backends with Node. Existing checked-assertion fixtures retain the null
and undefined panic controls. The reduction passes with a Node stdout mutant
in both backends. Production overlays which refuse .ts or admit .a fail this
test at the intended extension check. This refusal fixture adds no runnable
count row.

The enum reduction from ec71eb48 is byte-identical to
`stage3/fixtures/enum-init-reach/native-enum-map.a`, already carried by the
authorized enum topic. It constructs `new Map<never, never>()` while an enum
is pending. The local candidate lowers it, and the focused oracle compares
sanitized native, release native and JavaScript with Node. The namespace
prerequisite and enum topic are retained at `compiler/area-stack-namespace-held`
commit c1e42e612; they have not reached the delivery or remote. Enum-init is
named as blocked by its prerequisite judgments. No namespace refusal was
removed to make the group green.

Two prerequisite judgments still prevent a green group:

* `internal/lower/namespaces.go`, witnessed by
  `namespaces_observed_narrowing.a:14:57`: the stack refuses a qualified narrowed
  union read; the namespace topic admits it. Proposal: use the existing local
  primitive-union tag proof at the qualified read, retaining readiness and
  refusal of object tags that cannot be checked.
* `internal/lower/locals.go`, witnessed by
  `namespaces_parser_state.a:3:6`: the stack refuses repeated var assertion
  declarations; the namespace topic shares a hoisted var slot. Proposal: allow
  only repeated namespace vars to share one slot, preserve initializer order,
  never reset it for a declaration without an initializer, and retain the
  repeated-var refusal elsewhere.

The 8cb5e7c1 fixture base is merged in the delivery. Its TypeScript helper and fixture-path
conflicts preserve the stack's checked .ts controls and .a refusals. Its new
presence and logical-reference refusal witnesses remain. An identical duplicate
non-null map entry introduced by the merge was removed; the remaining entry
and its source-extension guard are unchanged.

Group 4 additions are prepared in a review worktree: array-literal-never-element's own
94c85da88 and b3578751d, debugger-statement's own 0fe2588fe and 7d2cbc895,
and Source boundary proof 19bcb8ee. The array resolution preserves the stack's
nonempty contextual representation while adding the topic's empty-literal
destination search. The debugger resolution retains namespace and labeled
control flow; it removes only the debugger refusal required by that topic.
19bcb8ee changes no production admission and conflicts only in counts rows.
Their focused tests pass: lower 0.592 seconds, flow 0.010 seconds, oracle
3.524 seconds; Source lower 0.177 seconds and oracle 1.057 seconds. All their
mutants are caught, including the empty-array wrong-kind leak at native
destruction. These topics still require group-wide checks before delivery.

A review-only conflict script truncated file tails; complete stack files were
restored and only the inspected topic deltas applied before successful tests.
Discarding older build-cache files during active builds also interrupted a
review build. The successful retry log replaces that infrastructure failure.
The premature held-candidate catalog run was stopped with its logs retained
after flow and counts established that its prerequisite group was not green.

The scanner's `debug.ts:8:5` mutable export is reduced in
`stage3/namespace-live-export/live.a`. Node observes `false:false`,
`true:true`, and `false:false`: external reads and a namespace function agree
after each outside write. The delivery stops exactly at `live.a:8:5`, in
`internal/lower/namespaces.go`, with `a mutable namespace export; use a module
or export functions around private state`. `TestNamespaceLiveExportBoundary`
pins both Node's observation and that named stop. A production overlay which
changes the boundary is caught by this test.

The held namespace candidate produces Node's same output in release native
and JavaScript for this reduction. Its namespace own commits supply the live
slot; those commits remain held by the two unrelated namespace prerequisite
judgments above. No whole scanner success is claimed. This fixture is a named
NotYet boundary in the delivery and adds no runnable count row.

The delivery count refresh adds the incoming logical-reference fixture in its
normal position and removes `taste/17_binder_flow.a` from runnable counts:
8cb5e7c1 records that fixture as NotYet at its optional own-field write. The
fixture and refusal record remain; this is the authorized baseline change.

Memoize-regions depends on runtime's synchronous captured-cell graph regions
and environment adoption. This compiler stack still refuses those cycles and
has no graph-region runtime. The topic's optional-closure admission alone cannot
establish its ownership proof, so it is skipped with that runtime dependency.
Checked-view-dependent topics remain skipped as ruled.

Delivery catalog: `bash verify/catalog/check.sh 8bf45dadc --jobs 2` validates 11 applicable entries; five remain the catalog's recorded nonapplicable entries. Entry 2's initial control build ran out of disk, then `bash verify/catalog/check.sh 8bf45dadc --entry 2 --jobs 1` passes its control and mutant. No undo patch drifted. All compiler changes in the delivery were checked at 8bf45dadc; the final commit adds reports and logs only.

Exact delivery commands: `go test ./internal/lower ./internal/ir ./internal/flow ./internal/javascript`; filtered oracle patterns in the retained delivery-oracle log; `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts`; the two catalog commands above. Additional scanner-boundary test and its production overlay mutant pass and fail respectively, as expected.

The method-values resolution remains blocked by automatic approval review, which rejected a multi-file native closure ABI rewrite as unverified. No rejected code was applied. See METHOD-VALUES-PROPOSAL.md for the concrete counted-dispatch, packed-slot and ownership resolution. Other accepted topics beyond the prepared candidates still require replay and validation; they are not claimed delivered here.
