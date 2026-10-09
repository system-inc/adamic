e0a9d0258e44863235fdcc49376713bda3503912: initialization implementation; the pushed report commit is the branch tip.
Built one memoized, cycle-safe initialization reachability analysis for namespaces and regular enums, admitting unrelated calls and constructions.
Base main f4efdd23, namespace tip 47a6fabe merged in e1478264; full lower PASS 14.828s, uncached filtered oracle PASS 4.817s, full counts update PASS 20.099s.
Mutants: erased enum reach loses pinned refusals; dropped memo expands 8191 bodies instead of 13; removed readiness causes UBSan NULL reads; wrong callable branch disagrees with Node stdout.
Not covered: the full repository gate, complete parser/compiler compilation, runtime Date.now support or first-class class constructor values; no census total is claimed.

## Rule and shared implementation

The original enum preflight rejects any unresolved call or construction while
any regular enum is pending. It separately traverses call paths without a
completed-function memo. The namespace preflight already owns a Tarjan call
graph whose completed components share their transitive reach sets.

The enum pass now validates declarations only. The existing orchestration then
runs namespaceInitialization once. Its graph includes both namespace and
module enum declarations, comparing the reached declarations with those already
initialized at each module statement. No second graph, body walker or cache was
introduced. Mixed namespace/enum cycles share one component and expand each body
once. Known callbacks, parameter defaults and immutable aliases participate.

Known class construction includes instance field initializers, constructor
parameter defaults and bodies, and statically known base construction. Static
members run at the declaration, so they are not executed again by construction.
An enum's own object is marked available during its member assignments.

Known reaching reads retain the existing initialization diagnostic and its
NotYet classification: `reading an enum before its runtime initialization; move
the call after the enum declaration`. This is a conservative diagnostic policy,
not a new claim that the unsafe read can be compiled. Mutable aliases, parameters
and unresolved callable properties are admitted by initialization preflight.
Their eventual enum reads retain the existing checked module-read IR. Neither
backend reads native zero storage as though the enum had initialized.

The original timestamp probe has a literal `true ? () => 0 : Date.now` initializer.
After removal of the broad enum refusal, its dead Date.now value exposed the
unbound-method refusal. A shared helper now selects callable conditional branches
only when the condition is syntactically true or false. The same selection is
used by refusal scanning, expression lowering and the initialization graph.
The selected source branch still passes ordinary refusal and lowering checks.
No runtime clock intrinsic or general method detachment was admitted.

## Independent observations

The two original probes are copied unchanged from
`codex/stage3-parser-proof` e74c38e43228af3faba5eebd3927019625f5c12d,
`stage3/drivers/parser/native-enum-map.a` and `native-call-before-enum.a`.
They are preserved under stage3/fixtures/enum-init-reach.

| Probe | Source Node | Native sanitized/release and JavaScript |
| --- | --- | --- |
| native-enum-map.a | Exit 0, `0\n` | Same output and exit; clean leaks |
| native-call-before-enum.a | Exit 0, `0:0\n` | Same output and exit; clean leaks |

A scratch Go overlay reinstating the original enum_initialization.go reproduces
both old indirect-call/construction refusals. It changes only that preflight and
adds a temporary assertion test; it is not a claimed full baseline compiler build.
Log /tmp/enum-init-baseline.log, PASS 0.074s.

Node's TypeScript transform folds a literal `Pending.Zero` read to a constant,
including in a function called before the enum declaration. Therefore the new
unsafe witnesses use a computed `Pending[key]` read, with key typed as 'Zero'.
Those sources actually observe the pending runtime object. This distinction was
found by running Node first; the initial constant-member witnesses were replaced.

Direct, helper, recursive-cycle, instance-initializer and derived-constructor
witnesses fail on source Node with TypeError. Adamic keeps the named refusal,
pinned to the actual enum read on line 2. The reaching-call mutant confirms that
this refusal is caused by reachability rather than an unrelated syntax boundary.

Unresolved parameter-call, mutable callback and callable-property witnesses print
`before` and fail on source Node with TypeError. Native sanitized/release and
JavaScript instead stop at the shared readiness boundary, exit 70, with stdout
`before\n` and exactly:

```text
adamic: panic: ReferenceError: Cannot access 'Pending' before initialization
```

This checked boundary is intentionally pinned separately from Node's hoisted-var
TypeError. It is not presented as byte-equivalent error text. The same unresolved
call after enum initialization matches Node and finishes normally.

Seven successful fixtures match source Node in both backends and pass leak checks.
Three checked fixtures have exact runtime stop pins. Five reaching witnesses are
refused and never enter code generation. The full counts table adds ten rows and
changes no previous numeric row. Programs stopped by panic retain live values;
programs that finish have clean leak checks.

## Speed and mutants

The 24-level fixture uses two edges to the preceding function at each level,
inside dead branches, and calls f24 before an unrelated pending enum. Node runs
no inner calls. Its source, native and JavaScript outputs match. Separate graph
pins count body expansions, independent of machine speed:

| Depth | Expanded bodies | Repeated reach queries |
| --- | ---: | --- |
| 12 | 13 | No additional expansions |
| 24 | 25 | No additional expansions |

Both depth pins took about 0.03s individually in the focused run. This is an
observation on the test machine, not a timing speedup claim against the old full
compiler. The mixed cycle pin reaches one enum and one namespace with two body
expansions in the same graph.

Run `source /workspace/adamic-tools/env.sh` then
`python3 stage3/fixtures/enum-init-reach/mutants.py`. Every mutation restores its
subject in a finally block. All four were run and caught without a build failure:

| Mutant | Independent catcher and observed failure |
| --- | --- |
| Treat reaching enum calls as unreaching | TestEnumInitializationReach/reaching-* loses all five required refusals |
| Drop completed-function memo, keep active-cycle guard | TestEnumInitializationGraphMemo/12 counts 8191 expansions, required 13 |
| Drop checked enum reads | TestEnumInitializationUnknownPinned finds UBSan NULL object member accesses and exit 1, required readiness stop 70 |
| Choose wrong true callable branch | TestEnumInitializationNode/callable-selection finds different stdout against source Node |

Logs: /tmp/enum-init-mutants.log and /tmp/enum-init-{unreaching,no-memo,no-readiness,wrong-callable-branch}.log.

## Verification and scope

Setup succeeded with GOPROXY=https://proxy.golang.org|direct. Timing lines:
Node 0.024s, Go 0.025s, submodules 0.075s, markdown 0.080s, clang 0.179s,
Go build 30.718s, deferred test binaries 30.897s, warm 30.898s, total 30.925s.
nproc=5, CPU quota 4, Go 1.27.1, clang 20.1.8, Node 24.19.0. The environment
/workspace/adamic-tools/env.sh was sourced for builds and tests.

Final restored commands and logs:

```text
go test ./internal/lower -count=1 -timeout 30m
  PASS 14.828s, /tmp/enum-init-lower-final.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestEnumInitialization|TestNativeAgreesWithNode/.*(namespaces|enum-init-reach)' -count=1
  PASS 4.817s, /tmp/enum-init-oracle-restored.log
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
  PASS 20.099s, /tmp/enum-init-counts-final.log
go test ./internal/lower -run 'TestEnumInitialization|TestEnumNamespaceSharedCycle' -count=1 -v
  PASS 0.503s, /tmp/enum-init-shared-graph.log
go vet ./internal/lower ./internal/oracle
  PASS, /tmp/enum-init-vet.log empty
gofmt -l internal/lower internal/oracle
  /tmp/enum-init-format.log empty
git diff --check
  /tmp/enum-init-whitespace.log empty
```

Every test command redirected output to its log file. The full repository gate
was not run. Final fetch confirms main f4efdd23 is an ancestor of this branch.
Only codex/enum-init-reach is pushed; no main/area push or pull request.

Manual files: internal/lower/enum_initialization.go, enums_test.go, expression.go,
namespaces.go, namespaces_call_graph.go, refusals.go;
new literal_callable_branch.go and enum_initialization_reach_test.go;
new internal/oracle/enum_initialization_reach_test.go;
internal/oracle/counts.md; docs/enums.md; this report;
fifteen .a files and mutants.py under stage3/fixtures/enum-init-reach.
The four prohibited files were untouched manually. Their namespace integration
changes are imported by the requested merge. No source was copied from cohere.
