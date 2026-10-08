Built: merged the fixture base and replayed namespace prerequisites, enum reachability and namespace initialization; reduced the parser assertion to an extension-policy fixture.
Commits: fixture-base merge 9c80835dd; duplicate refusal entry correction d78e4438b; remote remains 2c7d9fd2 until the group is green.
Commands and outputs: compiler packages pass lower and IR; flow fails the two preserved namespace refusals below; focused enum and checked-assertion oracle passes in 17.293 seconds.
Mutants: census continuation, eligibility and attribution mutants are caught on both comparator binaries; new parser fixture and policy mutants have separate logs.
Not covered: a complete native parser build, checked views, runtime graph-region ownership, and unresolved namespace admissions.

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
and undefined panic controls. This refusal fixture adds no runnable count row.

The enum reduction from ec71eb48 is byte-identical to
`stage3/fixtures/enum-init-reach/native-enum-map.a`, already carried by the
authorized enum topic. It constructs `new Map<never, never>()` while an enum
is pending. The local candidate lowers it, and the focused oracle compares
sanitized native, release native and JavaScript with Node. The namespace
prerequisite and enum topic are local; they have not yet reached the remote.

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

The 8cb5e7c1 fixture base is merged. Its TypeScript helper and fixture-path
conflicts preserve the stack's checked .ts controls and .a refusals. Its new
presence and logical-reference refusal witnesses remain. An identical duplicate
non-null map entry introduced by the merge was removed; the remaining entry
and its source-extension guard are unchanged.

Group 4 additions are under review: array-literal-never-element's own
94c85da88 and b3578751d, debugger-statement's own 0fe2588fe and 7d2cbc895,
and Source boundary proof 19bcb8ee. The array resolution preserves the stack's
nonempty contextual representation while adding the topic's empty-literal
destination search. The debugger resolution retains namespace and labeled
control flow; it removes only the debugger refusal required by that topic.
19bcb8ee changes no production admission and conflicts only in counts rows.

Memoize-regions depends on runtime's synchronous captured-cell graph regions
and environment adoption. This compiler stack still refuses those cycles and
has no graph-region runtime. The topic's optional-closure admission alone cannot
establish its ownership proof, so it is skipped with that runtime dependency.
Checked-view-dependent topics remain skipped as ruled.
