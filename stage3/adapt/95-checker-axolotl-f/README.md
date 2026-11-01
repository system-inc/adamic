# 95f: return no constraint for an excess JavaScript type argument

JavaScript-changing fork fix, separate from the soundness assertions. Based on
95a, ef4d26a4a38b7890317e1ba71f8ad928fd8588f6. Correctness's November 1 ruling
authorizes returning undefined for the absent parameter; no TypeError message
is copied into the compiler. No public API declaration changes.

lane: run as part of a Correctness wave

## Reference gate, checked before building

Official microsoft/typescript-go at
`89d5d5b2849a0db0957065889ca58536fa6d2e4a`,
[internal/checker/services.go](https://github.com/microsoft/typescript-go/blob/89d5d5b2849a0db0957065889ca58536fa6d2e4a/internal/checker/services.go#L793),
checks typeArgumentPosition >= len(typeParameters) and returns nil before
indexing. This agrees with returning undefined here. Reference comparison is
by reading the complete function, not by claiming a typescript-go runtime run.

TypeScript main at `6161772c1370bdd4a82f808e6823bd755eb1870d` has moved the
checker to Go: src/compiler/checker.ts is absent. Its equivalent
[tsc/internal/checker/services.go](https://github.com/microsoft/TypeScript/blob/6161772c1370bdd4a82f808e6823bd755eb1870d/tsc/internal/checker/services.go#L793)
already has the same bounds check and nil return. This commit-pinned current
implementation agrees; an introducing commit is not inferred. The optional
model branch typescript/axolotl-99e did not exist at the initial ref check.

## Every in-repository caller

The 95a complete src scan finds these uses and no others:

- checker.ts:1901-1903: the internal TypeChecker object wrapper obtains a parse
  TypeNode and returns the private function's result unchanged. Its existing
  types.ts:5464 API declaration already returns Type | undefined. There are no
  other checker-internal calls to this private function.
- services/codefixes/fixAddMissingConstraint.ts:141, tryGetConstraintType:
  forwards the result. Its sole caller getInfo at line 108 uses
  `tryGetConstraintType(checker, token) || tryGetConstraintFromDiagnosticMessage(...)`,
  so a missing constraint selects the existing diagnostic-message fallback.
- services/completions.ts:5776, getConstraintOfTypeArgumentProperty: returns
  a present constraint; otherwise walks to the parent and returns undefined
  when that ancestor provides no constraint. Its result is explicitly optional.
- services/stringCompletions.ts:531, fromUnionableLiteralType: passes the
  result to getStringLiteralTypes. That helper at line 638 accepts
  Type | undefined and immediately returns emptyArray when absent.

External consumers are governed by the existing optional API contract, not
claimed to be exhaustively audited. The service consumers already distinguish
no constraint from a constraint, so the new return does not require a caller edit.

## Adapter and scope

Only the private getTypeArgumentConstraint owner changes. Immediately after
`const relevantTypeParameter = typeParameters[typeArgumentPosition];`, insert:

```ts
if (relevantTypeParameter === undefined) return undefined;
```

The indexed read stays at its original evaluation point. A present parameter
continues into the original getConstraintOfTypeParameter call; an absent one
previously threw in getResolvedBaseConstraint. Earlier call-like, decorator,
instantiation and error-type branches remain byte-for-byte intact.

Stock 6.0.3 through 47-host-errors/compiler binds the array, index and local
symbols, checks the TypeParameter | undefined indexed type, the original
callee, the narrowed TypeParameter argument after the guard, and the guard's
parsed position. sites.json pins the owner's normalized full text, not the
whole checker file. Every edit is planned before one write; no printer, cast,
optional chain or unrelated default is introduced. CRLF is preserved, with
one intentional new CRLF for the guard. A second run makes no edits.

verify.cjs independently compares all src TypeScript files and requires exactly
that one insertion in the pinned owner. JavaScript is deliberately different
in both transpileModule comments modes; this is not an erasure claim.

## Evidence and reproduction

Fixtures, stacks, complete diagnostic/query output, census records and oracle
logs live only on branch `typescript/axolotl-checker-f-evidence`, under
`stage3/evidence/95-checker-axolotl-f/`. They are not apply inputs. The adapter
directory contains only adapt.cjs, verify.cjs, sites.json and this README.

The upstream issue draft is appended to stage3/upstream/LEDGER.md. F and G
both start from 95a and append at its end; G records its reference-gate stop.
Whoever merges second must resolve the append conflict. No issue has been sent upstream.

Pinned evidence commit: `c136950ba70522ca7211de5a10e01c99806b5e6a`.
[Browse all proof artifacts](https://github.com/system-inc/adamic/tree/c136950ba70522ca7211de5a10e01c99806b5e6a/stage3/evidence/95-checker-axolotl-f/).
[Full before/after witness output](https://github.com/system-inc/adamic/blob/c136950ba70522ca7211de5a10e01c99806b5e6a/stage3/evidence/95-checker-axolotl-f/node-comparison.json) summarizes the eight changed queries and unchanged controls.
The full stacks and diagnostic text are in witness-before.json and witness-after.json
on that branch. The source and oracle results follow below.

## Completed source, diagnostic and Node proofs

Fresh apply exited 0 and reported one changed checker file, one added line,
zero removed lines for F. Both direct and fresh-tree zero-edit reruns return
files 0, edits 0. The independent verifier compared all 711 src TypeScript
files and found only the single inserted guard. Original 54,434 CRLFs remain;
the intentional new guard line makes 54,435. No public source declaration changed.

The real production census, rooted only at src/tsc/tsc.ts with loader options
unchanged, emits a root and identical aggregate observation; count one, as 95a
and the meter do. Before: 131 whole-program, 17 checker own-file. After: 130
and 16. Reverting the guard: 131 and 17. Removed only the original
checker.ts:42849:65 TS2345 for TypeParameter | undefined; added anywhere: 0.
The full-set comparator maps later checker.ts locations back by one for the
single inserted line, rather than calling coordinate shifts new diagnostics.
Raw observations and that explicit mapping are retained on the evidence branch.

The built-compiler witness covers four fixture programs: excess.js and
excess.ts with noImplicitAny false and true. Eight JSDoc excess C arguments
previously threw TypeError reading resolvedBaseConstraint; all eight now return
undefined. All other 116 queried results and all 27 compiler diagnostics remain
identical. Complete formatted diagnostic output, results, and all pre-fix
stacks are retained, not only diagnostic codes. The exact issue-draft program
also exits 1 with that TypeError before and prints undefined, exit 0, after.

One mutant removes only the new guard. All 711 mutant src files are then
byte-identical to 95a's build input, so its existing built artifact is reused
as the reverted mutant without runtime patching. Running witness.cjs with
--expect-fixed turns red, exit 1, on the restored TypeError. The source verifier
also exits 1 and the census restores precisely the removed diagnostic. No
mutant is used for the full oracle.

The JavaScript change is visible with comments retained (2,892,419 to
2,892,514 bytes) and removed (2,354,716 to 2,354,811 bytes). The source verifier
proves no other source edit; the guard is reached only when the old callee
would immediately read resolvedBaseConstraint on undefined. This argument,
the Node controls and the unchanged full-suite baselines bound the behavior
claim; arbitrary external consumers are not exhaustively tested.

Setup completed on nproc=5, cpu.max=400000 100000. Go and Node ready 0.026s,
submodules 0.074s, clang 0.199s, tsc 0.194s, markdown 0.260s, Go build 14.839s,
cache warm 14.963s, total 14.990s. Go 1.27.1, Node v24.19.0, clang 20.1.8.
Every build/test shell sources /workspace/adamic-tools/env.sh. Test output goes
to log files. Native compiler execution and the full repository Go gate were
not run; only this fork fix's specified proofs are claimed.

Reproduction commands from an adapter checkout, with an evidence checkout at
/workspace/scratch/95f-evidence-work and a pinned upstream TypeScript tree:

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH=/workspace/adamic/stage3/api/node_modules
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/95d-upstream
bash stage3/apply.sh /workspace/scratch/95f-fresh > /tmp/95f-apply-fresh.log 2>&1
node stage3/adapt/95-checker-axolotl-f/verify.cjs CONTROL /workspace/scratch/95f-fresh > /tmp/95f-final-source.json 2>&1
node stage3/adapt/95-checker-axolotl-f/adapt.cjs /workspace/scratch/95f-fresh > /tmp/95f-final-idempotence.json 2>&1
node /workspace/scratch/95f-evidence-work/stage3/evidence/95-checker-axolotl-f/witness.cjs BEFORE/built/local/typescript.js > /tmp/95f-witness-before.json 2>&1
node /workspace/scratch/95f-evidence-work/stage3/evidence/95-checker-axolotl-f/witness.cjs /workspace/scratch/95f-fresh/built/local/typescript.js --expect-fixed > /tmp/95f-witness-after.json 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout --kill-after=20s 880 bash stage3/oracle/run.sh /workspace/scratch/95f-fresh /workspace/scratch/95f-oracle --limit-seconds 720 > /tmp/95f-oracle.log 2>&1
```

CONTROL is a dependency-free 95a tree. To reproduce it from fresh F output,
copy the tree without node_modules/built and restore checker.ts from
refs/stage3/95-checker-axolotl-f-before. For census, use the production tool
built with `go build -o /workspace/scratch/95d-census ./stage3/census/tool`.
Run it on CONTROL, AFTER, and the isolated one-guard reversion MUTANT, writing
the three JSONL files, then run evidence's compare-census.py BEFORE.jsonl
AFTER.jsonl MUTANT.jsonl. None of these support scripts is an adapter input.

## Full oracle result

The unchanged complete upstream suite finished within its 880-second hard bound
in 360.963 seconds: install 2.733s (exit 0), build 26.761s (exit 0), tests
331.361s (exit 1). Counts: **106,366 passing, 1 failing, 0 pending**. Raw status
is fail solely for the standing `unittests:: Public APIs for typescript.d.ts
should be acknowledged when they change`. There are **0 new baseline moves**;
the only difference remains api/typescript.d.ts. No tests or baselines were
modified, added, removed or accepted by the adapter.

Built/local API is 590,996 bytes with SHA-256
b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d;
accepted reference API is 588,626 bytes with SHA-256
f87d6c30da287293aae741711fa94594b75d49e0977cea6bfd5982f90e2bafcf.
Both match main's recorded artifacts in adaptation 40. Local API equals the
built declaration. The complete baseline diff is 51,123 bytes, SHA-256
6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031,
exactly the existing 95a/main difference. Raw oracle report and compressed
install/build/test logs and baseline diff are on the evidence branch.
