# Parser proof: slice blockers pending

The proof now targets the declaration slice of createSourceFile, using the
scanner worker's stage3/slice tool. As of fetched scanner commit 46041de, that
tool has not been pushed. No second slice tool is being built here. Once it
lands, the required order is: slice createSourceFile, compare its Node dump
byte for byte with the full-tree dump, then measure checker and lowering
blockers on that slice. Temporary 60-69 adaptations are limited to slice files
outside the scanner slice, and the slice blocker list must be pushed first.

The measurements below are historical whole-file evidence, not the requested
slice blocker list. Neither closure.cjs nor value-closure.cjs emits a slice.

# Historical whole-file checker measurements

The Node driver is complete and its node-end mutant is caught. The native
parser is still blocked at the checker. The entries below are observations
from real parser.ts/scanner.ts loads; no checker-rejected program was lowered.
The earlier stopped integration report is retained below as historical evidence.

## Inputs and reproduction

Base main remains ef3d907ecdc4c771b016f7d9c52372def057a340 (confirmed against
origin on this continuation). Feature SHAs are in each integration JSON and
in the historical input table below. Adaptation source is exactly
 a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e, adaptation 10 only, followed by normal
upstream diagnostic regeneration. Upstream source pin remains
050880ce59e30b356b686bd3144efe24f875ebc8. The Node dump includes 81 files.

The measurement overlay only replaces the Adamic prelude with pinned upstream
Node declarations (@types/node 25.3.3, @types/source-map-support 0.5.10).
All compiler options remain those of each merged feature compiler, including
noUncheckedIndexedAccess, exactOptionalPropertyTypes, noImplicitReturns and
noFallthroughCasesInSwitch. No upstream-tsconfig relaxation is used. The sound
regex library adapter remains in force. Each overlay source is saved in
 evidence/*-load.go.txt. Go VCS stamping is disabled because the scratch
worktrees reuse the pinned cohere checkout through a symlink. Absolute module
replacement paths reference the same pinned source; they change no semantics.

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/drivers/parser/measure.py /tmp/new-parser-profiles /tmp/parser-adapted10 > /tmp/measure.log 2>&1
python3 stage3/drivers/parser/summarize.py /tmp/new-parser-profiles > /tmp/summary.log 2>&1
```

Scratch branches are scratch/parser-main, scratch/parser-taste,
scratch/parser-flags, scratch/parser-namespaces, scratch/parser-nested and
scratch/parser-combined. None was pushed. Initial incoming-whole-file conflict
resolutions dropped main fields and failed Go compilation. Those failures stay
in the logs. Final flag/namespace/combined compilers rebuild successfully after
three-way hunk reconstruction; merge.py records how both sides are retained.
The final scratch resolution diffs are committed as evidence/*-scratch-resolution.patch.
No production compiler file on codex/stage3-parser-proof was edited.

## Closure and ownership boundary

Following resolved imports and re-exports, including type-only edges, parser.ts
and scanner.ts each reach the same 78 TypeScript source files. Both enter the
large SCC through _namespaces/ts.ts. Therefore the literal file-closure
subtraction contains zero files and zero exclusive diagnostics. closure.json
saves every edge and both ordered member lists. This is the graph the current
loader checks, not a claim that scanner executes every compiler function.

A supplementary value-reference walk from createSourceFile/createScanner
reaches 25/12 files and 13 parser-only files. It groups declarations by top-level
statement and retains namespaces whole, so it is a conservative inventory,
not a proven source-slicing transformation or a replacement loader graph.
The definition and declaration spans are in value-closure.json. The 13 files are:

- src/compiler/checker.ts
- src/compiler/factory/baseNodeFactory.ts
- src/compiler/factory/emitNode.ts
- src/compiler/factory/nodeChildren.ts
- src/compiler/factory/nodeConverters.ts
- src/compiler/factory/nodeFactory.ts
- src/compiler/factory/parenthesizerRules.ts
- src/compiler/parser.ts
- src/compiler/path.ts
- src/compiler/performance.ts
- src/compiler/performanceCore.ts
- src/compiler/tracing.ts
- src/compiler/visitorPublic.ts

No temporary adaptation was made while this distinction is unresolved.
These historical inventories do not authorize 60-69 edits. The slice blocker
list will be pushed before any such adaptation.

## What each feature alone moves

Each entry is an actual load and a successful probe executable. All six final
compilers build. Each stops at Checker, so no corpus entry reaches Lower.

| Profile | Closure diagnostics | Removed from main | Added | parser.ts diagnostics |
| --- | ---: | ---: | ---: | ---: |
| main | 2673 | 0 | 0 | 68 |
| taste | 2673 | 0 | 0 | 68 |
| flags | 2493 | 180 | 0 | 58 |
| namespaces | 2493 | 180 | 0 | 58 |
| nested | 2673 | 0 | 0 | 68 |
| combined | 2493 | 180 | 0 | 58 |

Flags and namespaces each remove the same 180 TS1294 diagnostics by enabling
non-erasable syntax in their loader options. This gate movement covers enums,
namespaces and parameter properties; it does not prove flag-enums alone lowers
namespaces. Taste and nested functions move no checker diagnostics. Their
lowering effects are unobserved behind the remaining checker failures.

The combined scratch merge order was taste, flags, namespaces, nested.
Its final result is exactly the same ordered parser-entry diagnostic list as
flags alone: 2,493 diagnostics, no added diagnostics. Intermediate combined
source loads were not run; feature-alone loads and the final combined load are
the measured comparisons. The first reported blocker is shared:

```
src/compiler/binder.ts:1109:17: error TS2412: Type 'undefined' is not assignable to type 'FlowNode' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
```

Parser/scanner entry loads have identical site/code populations and counts,
with one exact-text difference at program.ts:3541:13 TS2375: checker type
printing changes object-field order and abbreviates a function signature.
The complete strings are in feature-summary.json; they were not normalized
away or credited as byte agreement.

## Exact current ordered gate list

The complete diagnostic chains, in the loader's returned order, are in
 evidence/<profile>-ordered-diagnostics.json.gz. Their uncompressed SHA-256s
are in feature-summary.json. This order is the actual checker report order,
including its textual-location sorting; it is not an inferred order of future
lowering refusals. The combined inventory has 58 diagnostics in parser.ts itself.
The full parser.ts lists are also uncompressed in evidence/*-parser-file.json.

The supplementary 13-file value inventory has 1,528 remaining diagnostics
(1,562 on main); all exact strings and their order are in
 evidence/*-value-exclusive-diagnostics.json.gz. Unreachable declarations in
those files are still checked by today's loader; the count is not restricted
to only reachable statement spans.

Combined checker-code counts, in first-reported occurrence order:

| Code | Diagnostics |
| --- | ---: |
| TS2412 | 617 |
| TS7029 | 83 |
| TS2322 | 105 |
| TS2532 | 225 |
| TS18048 | 359 |
| TS7030 | 252 |
| TS2345 | 688 |
| TS2375 | 77 |
| TS2379 | 31 |
| TS2769 | 6 |
| TS2722 | 2 |
| TS2556 | 2 |
| TS2488 | 1 |
| TS2339 | 28 |
| TS2420 | 1 |
| TS18046 | 6 |
| TS2538 | 7 |
| TS2366 | 1 |
| TS2684 | 2 |

## TS7030 and TS7029 remain blockers

noImplicitReturns and noFallthroughCasesInSwitch remain on in every profile.
The combined closure retains 252 TS7030 and 83 TS7029 diagnostics.
parser.ts itself retains the following 16 return and 9 fallthrough diagnostics
(in the loader's actual order). These are candidates for temporary 60-69
adaptations if parser ownership is defined by value dependencies; none was
silently suppressed pending codex/fallthrough-and-implicit-returns.

| Location in parser.ts | Code | Exact message |
| --- | --- | --- |
| 10367:18 | TS7030 | Not all code paths return a value. |
| 1272:199 | TS7030 | Not all code paths return a value. |
| 1682:21 | TS7029 | Fallthrough case in switch. |
| 2904:13 | TS7029 | Fallthrough case in switch. |
| 3791:14 | TS7030 | Not all code paths return a value. |
| 3987:37 | TS7030 | Not all code paths return a value. |
| 4095:14 | TS7030 | Not all code paths return a value. |
| 447:153 | TS7030 | Not all code paths return a value. |
| 4603:13 | TS7029 | Fallthrough case in switch. |
| 4609:13 | TS7029 | Fallthrough case in switch. |
| 4758:14 | TS7030 | Not all code paths return a value. |
| 4924:14 | TS7030 | Not all code paths return a value. |
| 5819:13 | TS7029 | Fallthrough case in switch. |
| 5852:13 | TS7029 | Fallthrough case in switch. |
| 6610:13 | TS7029 | Fallthrough case in switch. |
| 7492:53 | TS7030 | Not all code paths return a value. |
| 7753:14 | TS7030 | Not all code paths return a value. |
| 7766:25 | TS7030 | Not all code paths return a value. |
| 8441:14 | TS7030 | Not all code paths return a value. |
| 8984:25 | TS7029 | Fallthrough case in switch. |
| 9195:80 | TS7030 | Not all code paths return a value. |
| 9272:25 | TS7029 | Fallthrough case in switch. |
| 9346:22 | TS7030 | Not all code paths return a value. |
| 9446:22 | TS7030 | Not all code paths return a value. |
| 9707:44 | TS7030 | Not all code paths return a value. |

## Limit of the measured order

The current gate list is complete, but the exact latent lowering blocker order
is not measured: all actual source entries fail checking. I did not remove
2,493 checker diagnostics in a scratch source copy to reach lowering, did not
make temporary adaptations, and did not run the stage-3 suite or a native
parser dump. Moving enum/namespace syntax through the checker does not show
that their runtime or ownership cases compile. There is no native agreement
claim. The dump's successful mutant proves only that its byte comparison can
catch an actual one-node end change on Node.

The driver command, dump hash and mutant lines are in README.md and
 evidence/node-report.json. Setup from the first pass remains valid: 98 seconds,
nproc 5. Additional checks: bash -n run.sh, probe go vet, Python AST parsing,
and git diff --check excluding saved patch artifacts. The unfiltered check
reports whitespace in unified-diff context lines inside those artifacts;
they preserve the scratch changes verbatim. Test output stayed in logs. No complete repository gate
was run for these driver/probe/report changes.

## Historical preflight, superseded by the measurements above

This is an incomplete integration report, not the requested exact, ordered
parser lowering blocker list. No parser-specific lowering blocker was measured.
No temporary adaptation was made. No native parser proof is claimed.

## Pinned inputs

Started from origin/main ef3d907ecdc4c771b016f7d9c52372def057a340 on
codex/stage3-parser-proof. Fetched the required branches explicitly because the
checkout's default fetch brought down only main.

| Input | Commit |
| --- | --- |
| taste-not-soundness | aa896b5d5ccc82210184fd01b8fe4d0ce0730a50 |
| flag-enums | f7d62772fa8e52fcfae754e047ceb66dba88b782 |
| namespaces-tsc | ce8a2acf14e420a9c82345236845a377cd4c7a50 |
| nested-functions | b15216dabf65ffaa7152f6e64709b7b062ea01a9 |
| stage3-base | 8728405135d329efc12c837a7a6c293234abbe1c |
| tsc-census | 429c1177f0130f785c19cf590d1860513b2ddbfc |
| stage3-type-imports | a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e |
| stage3-optional-declarations | e3535e702e285a5dcb4d06364c4f889b8178ea0c |

## Observed integration sequence

The scratch worktree is /tmp/adamic-parser-scratch, branch
scratch/parser-proof. That branch has never been pushed.

1. An octopus merge of taste, flags, namespaces and nested functions failed
   with exit 2. Git reported "Should not be doing an octopus."
2. A sequential merge of taste fast-forwarded successfully.
3. The sequential flag-enums merge conflicted in four files:
   internal/lower/class_inheritance.go, internal/lower/lower.go,
   internal/lower/refusals.go and internal/oracle/counts.md.
4. I stopped without resolving these conflicts. Namespaces and nested
   functions were not subsequently merged. The scratch tree is still in its
   conflicted merge state; it is not a buildable four-feature compiler.

The shell command which printed the sequential merge logs returned zero
because its final command was git diff, but the flag-enums merge itself
failed. Its log explicitly says "Automatic merge failed". Do not count that
shell exit as a successful merge.

The conflict diff is preserved in evidence/merge-conflicts.diff. The conflicts
join nominal checks with enum checks, accessor discovery with enum
initialization, definite-assignment refusal with enum refusal, and independent
count rows. These are integration issues, not evidence about parser.ts.
They may be resolvable; I did not establish that they are impossible.

## Second witness

The stage1/typescript/parser implementation is a separate indexed-node port
following typescript-go's parser. Its --whole traversal has kind names,
positions, ends, cooked identifier/literal text and additional selected
metadata. main.ts maps its internal UTF-16 positions to UTF-8 byte offsets.
nodes.ts prints only the optional-chain bit as its node-flags column, plus
separate literal flags. Declaration semantics are another column. The node
table has no full context/NodeFlags field. expect and semicolon panic on
unsupported or malformed grammar; the driver does not emit parse diagnostics.

Consequently its current output cannot be reformatted into the requested
createSourceFile dump with full flags and parse diagnostics. Producing that
answer would require adding parser state and diagnostic behavior, beyond a
printer change. A kind/position/text projection could be compared after
normalizing positions and kind spellings, but that would be a weaker witness.
I did not run that projection or a Node corpus comparison. There is no measured
list of corpus differences in this report.

The historical WHOLE_REPORT.md records Go/Node/native agreement for its own
protocol over 77 compiler files. That is prior evidence, not a run tonight,
and does not establish this unit's full-flags dump agreement.

## Setup and limits

bash cloud/setup.sh completed successfully. Source the printed tool environment
at /workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8, Node v24.19.0.
Timing lines: Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warm 98s,
total 98s. nproc: 5; cgroup cpu.max: 400000 100000. The complete setup log is
in evidence/setup.log. Its warm run builds test binaries without running tests.

Not done: main.a, run.sh, adapted-tree creation, scanner/parser closure
subtraction, sequential source blocker removal, temporary adaptations, stage-3
baseline oracle, native comparison and the planted Node node.end mutant.
No dump comparison exists yet, so no claim that it catches that mutant is made.
No compiler or scanner-worker file was edited on the deliverable branch.
