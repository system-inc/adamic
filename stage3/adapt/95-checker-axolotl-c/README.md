# 95c: three internal declaration contracts and a blocked host repair

lane: run as part of a Correctness wave

Base: origin/typescript/axolotl-checker-a at
`ef4d26a4a38b7890317e1ba71f8ad928fd8588f6` (including `d5c15cbf`).
Only this adaptation directory belongs to this branch. The separate public
TypeChecker repair at checker.ts:1683 belongs to 95e.

Three shipping repairs describe absence that the implementation already permits:

- checker.ts:5311:45: the private `getMergedSymbol` temporary becomes
  `Symbol | undefined`. A missing merged-table entry still returns `symbol`.
  The public TypeChecker method is unchanged; only its implementation's local
  declaration changes.
- checker.ts:17890:21: `SymbolLinks.tupleLabelDeclaration` admits a present
  undefined for unlabeled tuples. `SymbolLinks` is marked `@internal`.
- checker.ts:38430:64: the four-argument private `getTupleElementLabel`
  overload admits undefined for `elementFlags`. Its implementation's existing
  `ElementFlags.Fixed` default stays exactly as written.

There are no added assertions, defaults, casts, or optional chains.
`SymbolLinks`, its tuple-label slot, `ModuleSpecifierResolutionHost`, and
`getTupleElementLabel` are absent from the pinned public API baseline.
`getMergedSymbol` appears there only as a TypeChecker method; its local
`merged` is private. The upstream baseline bytes are not edited.

The adapter uses stock TypeScript 6.0.3 through 47-host-errors' `compiler()`.
It addresses parsed declarations within inventoried owners, checks declaration
symbols and before/after types, and pins normalized owner text rather than
checker.ts as a whole. Both output files are parsed and stock-transpiled in
both comment modes before either is written. Edits are source slices; CRLF
and unrelated bytes are preserved, with no printer. Partial or moved owners
are refused. `--check` never writes.

`verify.cjs` independently reconstructs the exact three declaration edits and
compares every source file, CRLF counts, and complete emitted JavaScript.
`prove.py BEFORE AFTER CENSUS OUTPUT` replays the exact census, every repair
in isolation, idempotence, and the one-edit undo mutant. Build CENSUS with
`go build -o CENSUS ./stage3/census/tool`, sourcing the setup environment.
The census normalization is the same entry-root deduplication as 95a's
compare-census.py: root and aggregate observations must agree and count once.



## Measured scope and limitation

The literal four-edit candidate fails the no-new-diagnostics rule. Its census
removes the four original messages but adds a different TS2375 at 54261:5:
`getSymlinkCache` can be explicitly undefined, while its optional method slot
excludes that value. The returned host also explicitly stores an optional
`getGlobalTypingsCacheLocation` callback. Those two further slot widenings were
not among the four requested declarations. The shipping adapter leaves out the
host repair and retains its original diagnostic. This is three removals, not
the requested four. [proofs/initial-scope.json](proofs/initial-scope.json) keeps
the exact failed candidate difference; [proofs/deferred-host.json](proofs/deferred-host.json)
records the attempted readFile repair, not a shipping edit.

The shipping census is **131 to 128 whole-program, 17 to 14 checker own-file**.
Exactly TS2322 at 5311:45, TS2412 at 17890:21 and TS2345 at 38430:64 disappear.
There is no added diagnostic anywhere. Each of the three repairs alone measures
131 to 130 and 17 to 16, with exactly its own diagnostic removed and none added.
The control's entire 131-diagnostic set matches 95a's committed reproduction
minus its 13 removed findings. Root and aggregate observations agree exactly.

The undo mutant restores only `merged: Symbol`, recovering the original
TS2322 at 5311:45: **129 whole-program, 15 checker own-file**. The independent
verifier exits 1 (`every declaration repair must be present`); adapter `--check`
exits 1 (`missing declaration repair: merged`). A second unmutated run reports
`files: 0, declarations: 0`. No mutant enters the oracle tree.

The independent verifier compares all **746 files under src**. Exactly two
files change. CRLF counts stay **54,434** for checker.ts and **10,681** for
types.ts. Complete stock transpileModule output is byte-identical in both modes:

| File | Comments retained bytes / SHA-256 | Comments removed bytes / SHA-256 |
| --- | --- | --- |
| checker.ts | 2,892,419 / `786ee138ff00765ab7c7ea7651de48df32c5981ec2c0810c68e302570d566f6e` | 2,354,716 / `3664a2f5fb2aba0e839a3153e6ff4d093b530db890b5b2433be6cd38e6e91aac` |
| types.ts | 139,659 / `4839bedc439cad06969bad2a3d805a7c0b480f06df73fcff62d2cd0957009fc3` | 121,398 / `f041d720cfda4ccb137ffc32198ec3d2fce46dbf13404df737dc637e21290639` |

Toolchain setup passed. Timings: Node 0.023s, Go 0.031s, clang 0.258s,
tsc 0.547s, Markdown 1.431s, submodules 163.991s, Go build 436.508s,
cache warm 436.625s, done 436.662s. `nproc=5`, CPU quota 4.
Every build/test shell sources `/workspace/adamic-tools/env.sh`.

The first control apply failed because the scratch worktree lacked
`stage3/api/node_modules/.package-lock.json`; linking the verified lockfile
install fixed it. The first generic fetch only fetched main; explicitly fetching
the 95a branch fixed the missing remote ref. Neither required changing existing
adaptations, checker options, or the compiler.

## End-to-end apply and oracle

Fresh `bash stage3/apply.sh /workspace/scratch/95c-final` exits 0. The generated
patch table's incremental row is `95-checker-axolotl-c | 2 | 3 | 3`.
The independent source verifier passes on that output, its fresh census has
exactly the same 128 diagnostics as the standalone proof, and a second adapter
run makes zero edits. No patch table or lane source is edited in the repository.
At the remote branch check, 95b and 95d do not exist; 95c applies on 95a alone.

The original four-edit investigation tree's standalone default oracle completed
in 432.065s: install 0 (5.195s), build 0 (23.207s), tests 1 (403.573s).
Counts are 106,366 passing, 1 failing, 0 pending. Only the standing Public APIs
acknowledgement test fails. Its built API hash is main's
`b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`;
the accepted reference is main's
`f87d6c30da287293aae741711fa94594b75d49e0977cea6bfd5982f90e2bafcf`.
There is no new public API delta. The exact three-repair shipping tree's oracle
is running separately; its completed evidence will be added before the final report.

Commands (all test output is logged):

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH="$PWD/stage3/api/node_modules"
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/95-upstream
go build -o /workspace/scratch/95-census ./stage3/census/tool > /tmp/95-census-build.log 2>&1
bash stage3/apply.sh /workspace/scratch/95c-final > /tmp/95c-final-apply.log 2>&1
# BEFORE is the independent 95a-only end-to-end apply tree.
python3 stage3/adapt/95-checker-axolotl-c/prove.py /workspace/scratch/95a-tree /workspace/scratch/95c-three /workspace/scratch/95-census /workspace/scratch/95c-three-proofs > /tmp/95c-three-prove.log 2>&1
node stage3/adapt/95-checker-axolotl-c/verify.cjs /workspace/scratch/95a-tree /workspace/scratch/95c-final > /tmp/95c-final-source.json 2>&1
node stage3/adapt/95-checker-axolotl-c/adapt.cjs /workspace/scratch/95c-final > /tmp/95c-final-idempotence.json 2>&1
/workspace/scratch/95-census /workspace/scratch/95c-final/src/tsc/tsc.ts /workspace/scratch/95c-final-census.jsonl > /tmp/95c-final-census.log 2>&1
timeout --kill-after=20s 880 bash stage3/oracle/run.sh /workspace/scratch/95c-final /workspace/scratch/95c-final-oracle --limit-seconds 720 > /tmp/95c-final-oracle.log 2>&1
```

Not covered: the blocked host repair, native tsc lowering/compilation, browser
or ESLint integration, lane predicates, and the whole repository Go gate.
