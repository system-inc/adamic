# or predicate helpers

Implementation and proof result for task 1w5bve2. The adaptation is built and its stock
TypeScript checks pass. Backend witness acceptance remains blocked on main;
this branch must not be described as a completed proof or accepted adaptation.

The adapter adds orPredicate2 and orPredicate3 beside or, using P, R1, R2 and R3
and p1/p2/p3 in the original overloads' exact order. Their closures return the
ordinary short-circuit disjunction. Nineteen calls change only their callee:
thirteen use orPredicate2 and six use orPredicate3. The two predicate overloads
are removed from or. Its two rest declaration texts, implementation and four
call expressions remain unchanged.

This is a JavaScript-changing established-shape adaptation under Adamic and
Compiler, task 1w5bve2, Nov 7 19:51Z. The existing complete lane count is reused:

| Sites | Closure calls | Boolean-only observed | Value differences |
| --- | ---: | --- | ---: |
| 19 predicate sites | 2,018,000 | yes | 0 |
| 4 rest sites | 168,858 | no | 0 |
| Total | 2,186,858 | no | 0 |

The count is from typescript/or-calls commit
701181917b36db38528257ff37da1a4e64aaac16. No new lane count was run.

## Ownership and transaction

25 whole before/after pins cover the core declaration group, fourteen function
owners, the isImport variable statement and nine service imports. A changed,
missing, duplicate, partially rewritten or mixed-state tree refuses before any
write. LF/CRLF and foreign text outside the owned spans survive. Replay writes
nothing and preserves bytes and mtime. Ten contract checks pass.

49-honest-optionals does not pin the deleted overloads, so their deletion is
owned here. The whole checker import is pinned earlier: adding orPredicate2 and
orPredicate3 is an amendment to 99-correctness-core-multimap's after import pin.
Its before pin and every MultiMap function pin are unchanged. Checker retains
or for its six-predicate rest call. All other helper imports and call rewrites
are owned here. owner-search.json retains the required name search inventory.

71-writable-views' diagnostics = [] as DiagnosticWithLocation[] and
78-variance-marker's variance expression inside createTypeChecker are untouched.
71 reports a dynamic runtime fingerprint of the outer createTypeChecker body;
the seven owned callee substitutions account for its entire changed hash.
replay-fingerprint.cjs independently proves this with 71's exact algorithm.
Paired replay review also proves the unchanged-line mapping for 78's reported
coordinates. These are different expressions, not whole-owner pins.

## Witnesses: acceptance is blocked

Both original witnesses copy all four main or overloads and its implementation
verbatim from the completed adapted core.ts. Both JavaScript and native
compiler attempts exit 1 with:

```
overload 1 of or type parameter P cannot satisfy implementation parameter T
```

With the adapted full declaration group and rewritten calls, both arities move
to the retained rest overloads' independent stop:

```
stage 0 can't lower overload 1 of or with a non-array rest parameter yet
```

Companion programs contain the exact helpers without the independently blocked
or overload set. An instantiated helper still fails on both backends:

```
stage 0 can't lower a checked predicate claim of R1, over R1, which this body is lowered without knowing yet
```

Explicit type arguments do not remove that stop. Unused helpers lower. All
Node executions agree, but no successful backend runtime agreement is claimed.
The mutants restore the original overloads and calls while retaining the new
helpers; both arities restore the exact binder refusal on both backends.
No production compiler change or exemption was introduced.

The separate Node site witnesses supply the same actual predicate callbacks
and same node objects to the original and new constructors. All twelve
comparisons return equal booleans, with identical predicate call counts. The
first-true cases prove that neither the second nor the third predicate is
called after the first returns true.

Rest-form probes preserve the exact production declaration group. Their
recorded errors include a non-array-rest-argument stop and an invariant
[N]-to-unknown[] stop. An implementation-only diagnostic control also records
the latter. These are blockers before the body is executed, not evidence that
its remaining body lowers. Exact sources, errors and exit codes are retained.

## Completed checks and outstanding census

- Both full --build-product applies exit 0.
- Both own npm run build commands exit 0, covering compiler, services, server,
  harness and tests. The candidate adds zero build diagnostics.
- Adamic types over all 79 sorted roots exits 0 on both trees.
- Stock 6.0.3 resolves the same 23 sites before and after: 19 now resolve to the
  intended helper and all four rest calls retain their exact declaration text.
  Their new ordinal numbers are 1 and 2 because the old overloads 1 and 2 were
  deleted; these are the original overloads 3 and 4.
- The separate stock inventory has 23 diagnostics on each dependency-linked
  tree. DIAGNOSTICS.md explains all 25 from the initial unlinked inventory.
- Across 746 source files, only the eleven files and the exact pinned spans
  change. All ten emitted JavaScript files are compared. Only _tsc.js,
  typescript.js and run.js differ, by the two helpers, seven or nineteen callee
  changes, namespace export properties and dead CommonJS export annotations.
  Reversing precisely those syntax nodes restores each complete original AST.
  Every emitted hunk is retained without normalization in the proof archive.
- typescript.d.ts and tsserverlibrary.d.ts are byte-identical. None of or,
  orPredicate2 or orPredicate3 is a public declaration there.
- typescript.internal.d.ts replaces the predicate or declarations at lines
  7721 and 7723 with the helper declarations at the same lines. The nonempty
  rest declaration starts at 7725 instead of 7726; general rest starts at 7732
  instead of 7733. The exact declaration lines and complete diff are retained.
- The full core census has 270 units and 34 events including boundaries on
  each tree. Every event matches after proven unchanged-line coordinate
  mapping: zero new recorded sites. Generic or/helper bodies remain deferred.
- The TypeScript checkout source audit finds zero root ts-object enumerations.
  Nested reads are listed in TS-KEYS.md. The proof's own runtime export-key
  comparison observes exactly two added properties and no removed properties.
- Paired replay passes: 95 baseline and 96 candidate adapters. All common exit
  codes, writes and byte-change outcomes match. Only 71's proved runtime hash
  and 78's proved coordinates differ in their reports; the new adapter replays
  with zero writes. Fourteen pre-existing replay refusals remain identical.
- Both best-effort full checker censuses ended in SIGKILL while the 16 GiB
  cgroup recorded OOM kills. Baseline: 2,928.867 seconds, last completed unit
  checker.ts:30917:5. Candidate: 3,722.148 seconds, last completed unit
  checker.ts:36018:5. Both ended before the 90-minute wall cap. Neither emitted
  a complete finding stream, so findings for the four owners of the seven
  compiler sites are unavailable, not zero. CHECKER-CENSUS.md records every
  cutoff, owner range, process result and resource observation. The standalone
  positive witness remains blocked; this best-effort allowance does not waive
  the missing backend runtime agreement.

Base: origin/main 3214dd465e1155d7bf4d41653e95b29a3f1060a6. TypeScript source:
050880ce59e30b356b686bd3144efe24f875ebc8. d15620712's parameter rule stays.
The parked binder exemption is not on this base. The existing pushed history is
preserved; all updates use plain pushes to typescript/core-or-predicate.

The original baseline evidence archive remains intact. The new proof archive
retains both backend blockers and excluded startup probes, including the
preview accidentally applied to baseline. That baseline was restored and all
746 source hashes verified before its accepted checker-census attempt began.

BACKEND-BLOCKER.md is the focused compiler handoff. The complete attempted
witnesses remain failures; the successful stock checks do not waive them.

All task worker processes have finished. The working source transformation is
complete, but the branch is not proof-complete or ready to land until the
backend witness condition is resolved.
