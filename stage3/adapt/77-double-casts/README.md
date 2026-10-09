# Double-cast review

TypeScript 6.0.3, upstream pin `050880ce59e30b356b686bd3144efe24f875ebc8`.
Main base: `031a1259bc7973934792dc6cb1bd4074fc2204b9`.
Scout: `origin/codex/step09-double-casts`, `8a7ab17e`.
Sibling pattern read in full: `76-truthful-casts` (adapter, check, README and
sameMap proposal). Like 76, this adapter reads the **current** tree after every
lower-numbered adapter; it does not reconstruct upstream text or alter compiler
implementations. Stock TypeScript 6.0.3 supplies its AST and transpilation.

Four of the eighteen bridges become explicit **single downcasts**. Fourteen are
left unresolved, with their original bridge text guarded and reasons below.
This is deliberately incomplete contract work. It does not claim all eighteen
have been admitted by Adamic or that the remaining single casts already compile
natively. The single downcasts belong to checked-cast work #b5w3ycg.

| Site | Scout coordinate | Outcome | Final expression |
|---|---|---|---|
| C01 | `src/compiler/checker.ts:10950:32` | unresolved | `results as unknown as T[]` |
| C02 | `src/compiler/checker.ts:49445:40` | checked-downcast | `paramTag.parent.parent as JSDocCallbackTag` |
| C03 | `src/compiler/core.ts:736:36` | unresolved | `emptyArray as any as SortedReadonlyArray<T>` |
| C04 | `src/compiler/core.ts:759:12` | unresolved | `deduplicated as any as SortedReadonlyArray<T>` |
| C05 | `src/compiler/core.ts:764:12` | unresolved | `[] as any as SortedArray<T>` |
| C06 | `src/compiler/core.ts:810:89` | unresolved | `compareStringsCaseSensitive as any as Comparer<T>` |
| C07 | `src/compiler/debug.ts:849:36` | checked-downcast | `this.mapper1 as TypeMapper & DebugTypeMapper` |
| C08 | `src/compiler/debug.ts:850:8` | checked-downcast | `this.mapper2 as TypeMapper & DebugTypeMapper` |
| C09 | `src/compiler/program.ts:3284:116` | unresolved | `emptyArray as any as SortedReadonlyArray<Diagnostic>` |
| C10 | `src/compiler/transformer.ts:338:108` | checked-downcast | `node as T & SourceFile` |
| C11 | `src/compiler/tsbuildPublic.ts:992:23` | unresolved | `program as any as SemanticDiagnosticsBuilderProgram` |
| C12 | `src/compiler/tsbuildPublic.ts:993:22` | unresolved | `program as any as SemanticDiagnosticsBuilderProgram` |
| C13 | `src/compiler/tsbuildPublic.ts:1355:12` | unresolved | `readBuilderProgram(parsed.options, compilerHost) as any as T` |
| C14 | `src/compiler/watch.ts:862:41` | unresolved | `createEmitAndSemanticDiagnosticsBuilderProgram as any as CreateProgram<T>` |
| C15 | `src/compiler/watchPublic.ts:150:38` | unresolved | `createEmitAndSemanticDiagnosticsBuilderProgram as any as CreateProgram<T>` |
| C16 | `src/compiler/watchPublic.ts:151:24` | unresolved | `readBuilderProgram(options, host) as any as T` |
| C17 | `src/compiler/watchPublic.ts:555:22` | unresolved | `readBuilderProgram(compilerOptions, compilerHost) as any as T` |
| C18 | `src/compiler/watchUtilities.ts:839:13` | unresolved | `fallbackPolling as unknown as WatchFileKind` |

The coordinates are the scout's adapted main, not pristine upstream coordinates.
The adapter locates each expression by AST text and occurrence within the current
file. Duplicate builder expressions are counted individually. Missing, duplicate,
changed or additional bridge sites are rejected before any write. The complete
current compiler directory, including generated code, is scanned for unreviewed
bridges. Idempotence and LF/CRLF operation are checked.

## Four single downcasts

- C02: the existing `isJSDocCallbackTag(paramTag.parent.parent)` test selects the
  callback branch. The declared parent ancestry narrows this expression to never;
  replacing the public parent declaration is outside this sanction. The single
  `as JSDocCallbackTag` makes the actual tagged-object downcast explicit.
- C07/C08: nested mappers are viewed as the class whose prototype the debugger
  installs with `attachDebugPrototypeIfDebug`. `TypeMapper & DebugTypeMapper`
  retains the source mapper union and explicitly requests the debug class view.
  A bare DebugTypeMapper replacement produces TS2352; its class declaration does
  not declare the mapper's variant fields. No public class declaration changes.
  A mapper created before debug mode can lack the prototype. The runtime test
  exercises that failure rather than promising every mapper has debug methods.
- C10: the SourceFile arm already tests `node.kind === SyntaxKind.SourceFile`.
  `T & SourceFile` retains the input generic and requests its SourceFile
  refinement. Bare `as SourceFile` produces TS2352 for this generic owner.

The intersections are **downcast targets**, not assertions that the unrefined
source always has the extra members. They replace no runtime checks and add no
new runtime statements, helpers or hidden bridge through any/unknown. Admission,
failure behavior, optional-field compatibility and native proof of these casts
remain the checked-cast owner's work. No native cast-proof success is claimed.

## Fourteen unresolved sites

`sites.json` has a per-site reason, including every repeated site.

- C01: the produced array contains specific function-like declarations, while
  the callback advertises arbitrary T[]. The kind/factory/result/caller relation
  needs an internal typed producer. A generic array downcast would not prove
  arbitrary element T and is not applied. This producer work is unfinished.
- C03/C04/C05/C09: sorted arrays have a required phantom brand which is not
  stored. Empty and sorted producers do not materialize that required field.
  A single checked cast cannot establish the current declared storage contract.
  Removing/widening the public brand is not authorized. Audited producer and
  sortedness contracts remain unfinished.
- C06: a string comparator is not contravariantly a comparator for arbitrary T.
  Caller-supplied comparators or a truthful constrained public generic are needed;
  another function assertion cannot prove that relationship.
- C11/C12: checking one semantic-diagnostics capability does not establish the
  whole SemanticDiagnosticsBuilderProgram interface. Requiring that entire view
  before the existing probe would also break the missing-capability fallback.
- C13/C14/C15/C16/C17: a builtin emit/semantic builder or builtin factory cannot
  satisfy every caller-specific T or CreateProgram<T>. The public generic API
  remains exact. Internal builtin/custom factory correlation or splitting is
  unfinished; a single arbitrary-T cast would only hide the mismatch.
- C18: different enums happen to share numeric values. An explicit mapping is
  truthful but changes emitted JavaScript; altering public enum declarations is
  not authorized. Intersecting disjoint enum types to obtain never would hide
  the mismatch and is not applied.

These are neither completed truthful upcasts nor established checked-downcast
replacements under this unit's constraints. They remain blockers. This is not a
proof that no future internal producer/caller redesign can resolve them.

## Validation and reproduction

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/adapt77-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# Isolated worktree at the recorded main base, then this branch, serially.
NODE_OPTIONS=--max-old-space-size=1400 bash /tmp/adapt77-main/stage3/lane/run.sh /tmp/adapt77-before > /tmp/adapt77-before.log 2>&1
NODE_OPTIONS=--max-old-space-size=1400 bash stage3/lane/run.sh /tmp/adapt77-recovered > /tmp/adapt77-recovered.log 2>&1
# Eight-worker recovered lane hit one upstream 40-second setup timeout.
# Retain its successful apply tree, apply.log and patch-set.md in /tmp/adapt77-final.
NODE_OPTIONS=--max-old-space-size=1400 bash stage3/oracle/run.sh /tmp/adapt77-final/adapted-tree /tmp/adapt77-final/oracle --workers=4 > /tmp/adapt77-final/oracle.log 2>&1
# execution.json records actual oracle exit, platform and retained apply exit.
python3 stage3/lane/check.py /tmp/adapt77-final > /tmp/adapt77-final-lane.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules NODE_OPTIONS=--max-old-space-size=1400 node stage3/adapt/77-double-casts/proof.cjs /tmp/adapt77-before /tmp/adapt77-final /tmp/adapt77-final-proof > /tmp/adapt77-final-proof.log 2>&1
# Fixed worker contract rejects that control; final eight-worker recovery:
bash stage3/adapt/77-double-casts/evidence/accepted-recovery.sh > /tmp/adapt77-accepted-recovery.log 2>&1
```

The four-worker control above matched counts but the fixed-worker lane rejected
it. evidence/recovery.sh records that control; evidence/accepted-recovery.sh
records the subsequent eight-worker recovery command sequence, including
execution provenance; neither expectations nor sanctions are altered.

The main lane invokes stage3/apply.sh and the complete stage3/oracle. The final
adapted measurement retains the recovered lane’s successful apply output, runs
the complete oracle at the required eight workers, and uses the unchanged lane checker. The accepted proof
compares all real emitted JS and public API bytes, lane verdict/counts/failure
identity and baseline.diff bytes, all src source files, and stock compiler
checking. Each of the eighteen sites gets its own real source-input guard
mutant, including unresolved sites. Additional mutants cover stock checking,
actual nested-debugger output, JS/API bytes and lane counts. Guard mutants
prove source review checks, not native type soundness.

The real built debugger executes composite and merged mapper formatting, using
both nested cast sites, and demonstrates the missing-prototype TypeError. The
runtime mutant changes its real built m1 output label to m0; only the exact
output comparison catches it. See evidence/report.md for observed commands,
counts and timing, and proof.json for all sites and hashes.

No new .a fixture was added, so there is no a-check header or counts.md change.
No whole Go test gate, native tsc build, new API sanction or compiler edit is
claimed. Setup's required cache warming is separate from the Node measurement.
