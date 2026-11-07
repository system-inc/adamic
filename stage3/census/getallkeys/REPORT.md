# getAllKeys reach, task b5w3ycg

Status: stopped with read-only findings. The adapted-tree search and whole-lane runtime measurement are blocked, not completed. Neither runtime count is zero: both are unmeasured.

Base: origin/main `837da8ed9a306db11ca0f0a39f396111048c5040`.
Pinned TypeScript: v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.
The requested historical census base is `af42d3dc1261a3db536064c3dfbb5eaa7e8b6405`. Its cited F10 report is not present on this origin/main. No claim is made to have reproduced that census.

## Actual caller and namespace reach

The only source reference to the compiler helper outside its declaration is
`src/harness/fourslashImpl.ts:461`: `const keys = ts.getAllKeys(ls);`.
It is inside the TestState constructor's `memoWrap` helper, which builds a
proxy for the language service and memoizes five expensive methods. The
constructor invokes memoWrap at line 389. This is a real harness caller,
not a compiler-source caller. Searching only compiler src misses it.

The declaration is `src/compiler/core.ts:1299`, preceded by `/** @internal */`.
`src/compiler/_namespaces/ts.ts:4` reexports core with `export *`;
`src/harness/_namespaces/ts.ts:3` reexports that compiler namespace.
`src/typescript/typescript.ts` similarly reexports its ts namespace.
There are no literal `ts["getAllKeys"]` reads or other named compiler-helper
callers in the pinned checkout. Runtime-computed names cannot be excluded by
text search; that is one reason the requested runtime observation remains necessary.

The already shipped `lib/typescript.js` has two namespace export getters at
642 and 184594, the function body at 3038, and a generated CommonJS export
annotation at 199392. A read-only Node inspection returned:

```json
{"own":true,"type":"function","descriptor":"() => getAllKeys"}
```

This confirms that the pinned shipped runtime ts object exports it. A fresh
adapted build was not produced or inspected.

## Complete pinned spelling inventory

The search covered every file, including ignored and hidden files, excluding
only `.git`, in the entire pinned checkout. It included src, tests/cases,
harness, scripts, baselines and shipped lib. rg exited 0. All matching files
and all matching line numbers follow; multiple spellings on one line share
one entry. Scripts have no occurrence.

| Path in pinned checkout | Matching lines |
| --- | --- |
| `lib/lib.dom.d.ts` | 22227, 22229, 22231, 22409, 22411, 22413 |
| `lib/lib.webworker.d.ts` | 6790, 6792, 6794, 6972, 6974, 6976 |
| `lib/typescript.js` | 642, 3038, 184594, 199392 |
| `src/compiler/core.ts` | 1299 |
| `src/harness/fourslashImpl.ts` | 461 |
| `src/lib/dom.generated.d.ts` | 22211, 22213, 22215, 22393, 22395, 22397 |
| `src/lib/webworker.generated.d.ts` | 6774, 6776, 6778, 6956, 6958, 6960 |
| `tests/baselines/reference/genericClassesRedeclaration.errors.txt` | 18, 32, 45, 63, 77, 90 |
| `tests/baselines/reference/genericClassesRedeclaration.js` | 10, 22, 33, 49, 61, 72 |
| `tests/baselines/reference/genericClassesRedeclaration.symbols` | 25, 26, 87, 88, 147, 148, 216, 217, 278, 279, 338, 339 |
| `tests/baselines/reference/genericClassesRedeclaration.types` | 18, 19, 102, 103, 186, 187, 270, 271, 354, 355, 438, 439 |
| `tests/baselines/reference/parserRealSource4.errors.txt` | 31, 45, 130, 131 |
| `tests/baselines/reference/parserRealSource4.js` | 27, 41, 126, 127, 326, 403, 404 |
| `tests/baselines/reference/parserRealSource4.symbols` | 55, 56, 111, 112, 374, 375, 377, 378, 379, 383, 385, 389 |
| `tests/baselines/reference/parserRealSource4.types` | 76, 77, 173, 174, 661, 662, 665, 666, 668, 670, 672, 680, 684, 686, 694 |
| `tests/cases/compiler/genericClassesRedeclaration.ts` | 8, 20, 31, 47, 59, 70 |
| `tests/cases/conformance/parser/ecmascript5/parserRealSource4.ts` | 26, 40, 125, 126 |

The browser-library entries are unrelated IndexedDB IDBIndex and
IDBObjectStore methods and their documentation. The two test-case families
are unrelated user-defined collection methods. Their emitted JavaScript,
error, symbol and type baselines repeat those names; they are not references
to the compiler helper. The complete inventory is retained here so those
namesakes are not silently discarded.

## Declaration surface

`lib/typescript.d.ts`: no occurrence.
`lib/tsserverlibrary.d.ts`: no occurrence; this file reexports `./typescript.js`.
`tests/baselines/reference/api/typescript.d.ts`: no occurrence.
There is no checked-in `typescript.internal.d.ts` in this pinned checkout.
The declaration is internal-only by the build rules: `Herebyfile.mjs:447`
produces public and internal typescript declarations, and
`scripts/dtsBundler.mjs:421-447` writes an internal declaration only to
WriteTarget.Internal while public declarations also go to WriteTarget.Public.
Thus the internal-only conclusion is established by source and bundler rules,
not by reading a newly emitted internal declaration. That emitted-file check
remains blocked with the build.

## Whole-lane counts

| Requested observation | Result |
| --- | --- |
| Every getAllKeys call across the complete lane | Unmeasured |
| Every read of getAllKeys on the runtime ts object | Unmeasured |
| Completed lane units | 0 / 398 |

The requested reference method was read in
`stage3/adapt/32-indexed-reads-program/MEASURE-RECORD-TABLES.md` and its
piece driver. A temporary driver and observers were prepared under
`/workspace/scratch/getallkeys`, never installed in any source or built
artifact. The planned method was 20-unit pieces over the unchanged 398-unit
manifest using the ordinary shard identity, API, baseline and count guards,
function-entry counters, namespace export-getter counters, and VM-context
observer injection. No observer was activated, and no measurement is claimed.
No adapted apply completed, so the requested post-apply checkpoint condition
was never reached. This stopped-state report is the pushed checkpoint.

## Next lowering stop and pushIfUnique

Reading `internal/lower/library_object.go` establishes that
`Object.getPrototypeOf` is explicitly Refused by objectCallArguments.
`Object.getOwnPropertyNames` is not admitted: it has no handled case and falls
through to the final `l.notYet(node, "Object."+name)`, producing
`NotYet: Object.getOwnPropertyNames` for this non-record object parameter.
The record path (`internal/lower/records.go:396`) does not admit that operation
either. `internal/lower/control.go:78-100` lowers the do/while condition before
its body. This explains why the condition's getPrototypeOf refusal precedes
the body's enumeration stop. Admitting only getPrototypeOf would still leave
the getOwnPropertyNames stop at the start of the body.

`pushIfUnique` is an ordinary generic TypeScript helper, not an Object library
operation with a named admission case in library_object.go. Its pinned body
at core.ts:1003 calls contains, then array.push if absent. Generic helper calls
use ordinary instantiation (`internal/lower/expression.go:1389` and generic.go);
array.push is admitted (`internal/lower/library_array.go:13,25`). No named
pushIfUnique refusal exists in the lowerer. At this site its inferred element
is string. This source reading does not prove an end-to-end successful
specialization of its whole transitive body, and no such build/probe was run.
It does prove that getOwnPropertyNames remains an earlier definite stop.

## Setup, exits and blocking reason

Read the requested `stage3/adapt/99-correctness-axolotl-w/PROOF.md` and sourced
`/workspace/adamic-tools/env.sh`. Fetched origin/main, checked out the requested
report branch from it, cloned and checked out the exact TypeScript pin, and
installed the pinned stage3/api and stage3/typescript dependencies. Each of
these children exited 0. Full output logs are files in
`/workspace/scratch/getallkeys`; no log was piped into head or tail.

`bash cloud/setup.sh` reached ready submodules and started its normal Go
warming step. It was stopped when the build constraint became a blocker;
its wrapper reported the child's actual exit 143, not success. The setup
process and descendants were terminated. All other started children exited.
No Go file was added. No production source, adapter, lowering rule, lane rule
or generated artifact was changed. The scratch observer code was never
applied, so no observer remains in the checkout or pinned source.

Automatic approval review rejected this command before execution:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source \
  stage3/apply.sh --build-product /workspace/scratch/getallkeys/tree
```

The stated reason was that --build-product conflicts with the user's
"measure only, build nothing" instruction. The requested setup proof uses
this flag, but that ambiguity does not authorize bypassing the rejection.
Clarification was requested about allowing unchanged apply and ordinary
compiler/harness builds solely for measurement. Pending that clarification,
the adapted reference inventory, emitted internal declaration check and both
runtime counts remain blocked. No adaptation or Object.getPrototypeOf
implementation was constructed. The only committed file is this report.
