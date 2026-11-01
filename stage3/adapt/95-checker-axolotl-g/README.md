# 95g: return no symbol only from the throwing index-parent fallback

JavaScript-changing fork fix, based on 95a
`ef4d26a4a38b7890317e1ba71f8ad928fd8588f6`. Correctness's final ruling scopes
the missing-declaration return to the third parent-selection arm. The first
two arms retain today's successful results, including zero-declaration index
symbols. No service consumer is edited and no public API declaration changes.

lane: run as part of a Correctness wave

## Scope and reference implementations

The adapter expands the original ordered conditional into the same ordered
aliasSymbol, symbol and fallback branches. Only the third arm saves
copy.declarations[0], returns undefined when absent, and otherwise uses its
parent. The read remains after both original symbol tests; there is no hoist,
optional chain, cast, invented default, broad empty-list guard or assertion.
An absent fallback used to throw immediately. The temporary copy is not cached
or returned on the new exit, so no consumer observes a partially built symbol.

Official microsoft/typescript-go at
`89d5d5b2849a0db0957065889ca58536fa6d2e4a`,
[internal/checker/checker.go:32209](https://github.com/microsoft/typescript-go/blob/89d5d5b2849a0db0957065889ca58536fa6d2e4a/internal/checker/checker.go#L32209),
returns nil when the filtered declaration list is empty. TypeScript main at
`6161772c1370bdd4a82f808e6823bd755eb1870d` has its checker in Go;
[tsc/internal/checker/checker.go:32895](https://github.com/microsoft/TypeScript/blob/6161772c1370bdd4a82f808e6823bd755eb1870d/tsc/internal/checker/checker.go#L32895)
has the same behavior. Both complete functions and the computed-index producer
were read before building; no typescript-go execution is claimed.

G follows those references at the previously throwing arm only. Go also returns
no symbol on a zero-declaration index whose type has a symbol. G deliberately
keeps 6.0.3's __index result on that successful path, per the final ruling. The
prior broad guard and its reader regression are historical failed candidates,
not this patch; their original evidence remains labeled as such.

## Adapter contract

Stock TypeScript 6.0.3 through 47-host-errors/compiler parses and binds the
private getApplicableIndexSymbol owner. sites.json pins only its normalized
owner text, not checker.ts's whole-file hash. The adapter binds the Type
parameter and TransientSymbol copy, checks the Declaration | undefined indexed
read, the original getSymbolAtLocation callee, the ordered type-property tests,
and the narrowed Declaration inside the final fallback. Exact before/after
parsed shapes are inventoried. All planning and parsing precede one write,
with the original bytes reread before writing. No printer is used.

verify.cjs independently compares all source files and requires exactly the
inventoried statement replacement inside the pinned owner. The patch changes
JavaScript intentionally. Existing CRLFs are preserved; added branch lines
use CRLF. Idempotence, diagnostic, Node and oracle results follow below.

## Baseline witness

The existing 95a built compiler was queried, without rebuilding or modifying it.
A JavaScript class Computed has a broad computed static string key. A separate
empty class expression is cast through JSDoc to typeof Computed and exported
via module.exports; an extra export member removes the original type symbol.
A TypeScript consumer imports it and queries v.xyz with getSymbolAtLocation.
The fallback has one index info but no index declaration. At both noImplicitAny
settings the program has zero diagnostics and the query throws TypeError:
Cannot read properties of undefined (reading 'parent'). Both complete stacks
and all diagnostic output are retained with the witness.


## Caller audit and proofs

Evidence and complete before/after query results are stored only on the
separate typescript/axolotl-checker-g-evidence branch under
stage3/evidence/95-checker-axolotl-g/. Pinned evidence:
[cd1761dab0ca7a7ae6b4e5a3bf7b27b3462ff5e5](https://github.com/system-inc/adamic/tree/cd1761dab0ca7a7ae6b4e5a3bf7b27b3462ff5e5/stage3/evidence/95-checker-axolotl-g/).
The actual built results use names ending scoped-built.json; older guard
results retained for comparison are superseded, including guard-named files
copied into scoped-final/.
The adapter directory contains only adapt.cjs, verify.cjs, sites.json and README.md.
F and G both append at the same 95a ledger endpoint; whoever merges second must
resolve the append conflict. The existing G ledger entry is not renumbered.

### Readers of the newly absent result

The parsed inventory covers 123 getSymbolAtLocation calls, including 122 outside
testRunner, with their owner and source context in caller-audit-source.json.
The private checker wrapper forwards undefined; getSymbolOfNameOrPropertyAccessExpression
already has an optional result and its cache accepts undefined. The separate
JSX getApplicableIndexSymbol caller checks indexSymbol before use.

The original and imported-super witnesses trace the following service readers.
The trace uses an in-memory baseline bundle; all final query results are checked
again against the real apply/oracle build, not the instrumented module.

| Reader | Handling on these inputs |
| --- | --- |
| services.ts getSymbolAtLocationForQuickInfo:3568 | Missing-symbol fallback; quick info succeeds. |
| goToDefinition.ts getSymbol:559 | Optional symbol; definition, definition-and-span and type definition complete. |
| findAllReferences.ts Core.getReferencedSymbolsForNode:1015 | Missing-target exit; references, highlights and rename locations complete. |
| rename.ts getRenameInfoForNode:74 | Missing symbol returns cannot-rename info. |
| classifier2020.ts visit:151 | Missing symbol skipped; classifications complete. |
| callHierarchy.ts resolveCallHierarchyDeclaration:400 | Missing declaration returns undefined; call hierarchy queries complete. |
| findAllReferences.ts getImplementationReferenceEntries:580 | Imported-super variant moves its existing checker crash to valueDeclaration. |

One additional assertion consumer, harness/fourslashImpl.ts:1296
verifySymbolAtLocation, now raises "Could not get symbol at location" on the
original witness. Its actual parsed method body was run with the real checker
and range/error-service stubs; this is a method-level harness probe. Before it
threw the original parent TypeError. This is also a moved existing exception.
These are the two readers found that throw after the newly absent result.

Other asserted call sites in the inventory do not establish a new crash on
these inputs: findAllReferences:793 needs reference entries collected from a
successfully resolved target, while this target exits at :1015; :1757 and :2761
require declaration/shorthand names. importTracker's assertions require
import/export/binding declarations or module.exports assignment symbols;
navigateTo and convertExport require declaration names; moveToFile's asserted
binding and assignment paths do not query this unresolved property access.
The referenceSymbol use at findAllReferences:1712 is equality/optional-helper
handling on property accesses; its export-specifier branch has a different
node precondition. Other calls guard, forward an optional value, or use equality
or Set.has without dereferencing it. This audit concerns the newly absent
fallback result, not a claim that every API call always returns a symbol.

### Node results and diagnostic census

The two original noImplicitAny variants change parent TypeErrors to undefined,
with full diagnostic output empty before and after. The direct Computed.xyz
control still returns __index, declarations=0, parent=Computed, diagnostics=[].
The original successful super.xyz getImplementationAtPosition fixture is
byte-identical as JSON before/after: xyz=undefined, diagnostics=[], and the
super query still identifies Computed. Explicit index/alias controls also match.

The services matrix runs 19 operations on each of the original, imported-super
and successful-super programs: 57 rows, all with zero diagnostics. Of 28
formerly crashing rows, 27 complete and one moves as described above. All 29
previously successful rows are unchanged, including every successful-super
operation. There are zero newly crashing rows. The separate 14-program control
matrix preserves 232 queries and every diagnostic byte; only the original two
crashing queries now complete. Eight unrelated 42849 failures remain unchanged.

The production-meter census on 95a, using both root and aggregate records,
changes whole program 131 -> 130 and checker.ts own-file 17 -> 16. The complete
set comparison removes exactly TS2532 at 50153:121 and adds zero diagnostics
anywhere. Later checker coordinates are aligned by the ten added lines.
These are dependency-free census trees, matching the meter; installed oracle
dependencies are not allowed to silently alter this census.

### Apply, source, oracle and mutant

stage3/apply.sh completed end to end, exit 0, against pinned TypeScript 6.0.3.
The verifier compares all 711 source files: only checker.ts changes, exactly
the inventoried statement, with CRLF count 54434 -> 54444. No public API line
changes. The owner SHA-256 is
35b9308f3edbb3ae332535bd198677d3cfba0044bd1588e327cb98d9b2a4eaeb.
Second application: files=0, edits=0. JavaScript intentionally changes:
with comments 2892419 -> 2892803 bytes, without comments 2354716 -> 2355100.

Full unfiltered stage3/oracle/run.sh (all runners), under an 880-second hard
limit and 720-second runner limit, completed in 357.316 seconds: 106366 passing,
1 failing, 0 pending. Install/build exit 0. The sole failure is the standing
Public APIs case and the sole baseline diff is api/typescript.d.ts. No baseline
was accepted or rewritten. The baseline diff is 51123 bytes, SHA-256
6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031,
identical to the recorded standing main diff. No other baseline moves.

The mutant restores the exact original parent statement. The verifier exits 1,
the required-fixed Node witness exits 1 with the restored parent crashes, and
the meter returns to 131 whole-program / 17 checker diagnostics, including
50153:121. The reverted mutant's complete source set equals the built 95a
control, so that existing unmodified built compiler is its Node executable.

Prepared shared F/G toolchain setup timings: Go/Node 0.026s, submodule 0.074s,
clang 0.199s, tsc 0.194s, markdown 0.260s, build 14.839s, cache 14.963s,
total 14.990s. nproc=5 (cpu.max=400000 100000); Node 24.19.0, Go 1.27.1,
clang 20.1.8. Environment was sourced from /workspace/adamic-tools/env.sh.
99e's complete README and adapter were inspected at 5957341ec6b275035552c5da5c77fd124418d051.
No native compilation or typescript-go execution is claimed; reference behavior
is based on the pinned source. No exhaustive arbitrary-input service proof is
claimed beyond the source audit, concrete witnesses and unchanged full oracle.

Composition on 95a was also exercised with current B
408f8cb16c4e97ca51997d5a340a0b1826b66ba9 and C
7b1e14c56e54683180aa0cf7819e50ad086486eb. Orders B,C,G,D,F and G,C,B,F,D
both apply with zero failures and produce byte-identical source trees. G's
--check on the fully composed tree returns files=0, edits=0. The primary
single-head census/oracle above remains on 95a plus G only.

Reproduce from the G checkout, with NODE_PATH pointing at stock 6.0.3:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/95d-upstream bash stage3/apply.sh /workspace/scratch/95g-fresh > apply.log 2>&1
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/95-checker-axolotl-g/verify.cjs /workspace/scratch/95d-control /workspace/scratch/95g-fresh > source.json 2>&1
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/95-checker-axolotl-g/adapt.cjs /workspace/scratch/95g-fresh > idempotence.json 2>&1
timeout --kill-after=20s 880 bash stage3/oracle/run.sh /workspace/scratch/95g-fresh /workspace/scratch/95g-oracle --limit-seconds 720 > oracle.log 2>&1
```

For the meter, run the production-meter reproduction from 95a against the
dependency-free 95a and G trees, then the evidence compare-census.py with
before.jsonl, after.jsonl and mutant.jsonl. Run compare-witnesses.py from the
evidence checkout to assert the saved full before/after results. The witness
scripts accept the actual built typescript.js; controls/fixtures.cjs instead
accepts the built tree root. All test output is redirected to files.
