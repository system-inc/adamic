# Parser build-ahead rehearsal

Native red. This evidence may land; the scratch merge must never land whole. Any source reached only behind a placeholder is pending on the dependency named below. The parser has no native output to compare.

## Pinned stack

Started on the fetched train tip, origin/main `54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8`. Scratch HEAD is `d1b329ae2d4effac8fa6ce73a7510c591e8a3b99`, branch `codex/stage3-parser-rehearsal-scratch`, worktree `/workspace/scratch/parser-rehearsal-main`. Only the proof branch receives this evidence. All conflicts were aborted and skipped; none were resolved.

| Requested topic | Pinned SHA | Result |
|---|---|---|
| 64c59784 | 64c59784b612 | merged |
| origin/codex/records-maplike-next | 8c013f1cf3ec | skipped conflict (3 paths) |
| origin/area/library | 7778b3606a22 | skipped conflict (61 paths) |
| origin/codex/placeholder-nonnull | 711312292569 | merged |
| origin/codex/entries-provenance | 05a6006690f4 | skipped conflict (4 paths) |
| origin/codex/assignment-proofs | 4d17f1d4d820 | skipped conflict (2 paths) |
| origin/codex/scanner-cast-checks | 77ed44642155 | skipped conflict (60 paths) |
| origin/codex/stricter-options-next | 8f32e51e8fc4 | skipped conflict (12 paths) |

`library/area-on-next-2` and `compiler/stricter-options-next` did not exist in the remote snapshot. Library fallback was `area/library`; the available stricter topic was `codex/stricter-options-next`. Both conflicted and were skipped. [merges/report.json](merges/report.json) gives every conflict path and merge exit; the raw fetch and merge logs are retained. Conflicts include lowering and native emission/runtime paths. `placeholder-nonnull` 71131229 adds a scout read-order report only; it is not a compiler non-null implementation.

Go build ran in the scratch with the pinned cohere SDK redirected locally in go.mod, `GOWORK=off`, `go build -p 1 -mod=mod -o /workspace/scratch/parser-rehearsal-adamic ./cmd/adamic`. It exited zero. No compiler sources changed. [scratch-sdk-go-mod.patch.gz](scratch-sdk-go-mod.patch.gz) records SDK paths and Go module ordering. `nproc=5`, Node v24.19.0. Existing toolchain setup was reused. An initial Go build accidentally ran in the proof checkout; its binary was replaced by this scratch build before any probe. The excluded log and restored go.mod diff are retained.

## Validated source and first stop

Fresh scratch stage3/apply.sh completed, then 65 was applied idempotently and 76 checked. Both report identical emitted JavaScript; sameMap remains unapplied. No source adaptation or public API sanction changes in this unit. Whole upstream stage3/oracle and full stage3/lane were not rerun; 32 local lane tests pass.

The shared slicer gathered createSourceFile and the six existing driver support roots. The slice has **27 declaration files, 1,994 code declarations, 41,861 copied lines**, 2,083 audited source spans and 79 ordered module import lists. verify.cjs passes. The native build uses the unchanged validated slice, not the discovery copy.

Both `ADAMIC_NATIVE_SPLIT=0` and `1` stop before C with:

```text
src/compiler/debug.ts:113:19: error TS2339: Property 'captureStackTrace' does not exist on type 'ErrorConstructor'.
src/compiler/debug.ts:114:19: error TS2339: Property 'captureStackTrace' does not exist on type 'ErrorConstructor'.
```

Build-attempt wall times are 0.516s unsplit and 0.524s split, before clang. These are not clang times or warm-cache measurements. No parser C, binary, native dump, native recovery-corpus run or parser performance measurement was reached.

## Ordered stops

Row 1 is the real stop. Rows 2 through 16 are pending, found behind the saved uncommitted throwing placeholders. Owners below are routing suggestions; coordinates, diagnostics and order are observations. No placeholder is an adaptation or an acceptance implementation.

| Order | Slice location | Against BLOCKERS.md | Owner | Minimal witness | Pending on |
|---|---|---|---|---|
| 1 | src/compiler/debug.ts:113:19 | still there | library | [native-error-capture-stack.a](../../native-error-capture-stack.a) | compatible area/library ErrorConstructor facts; attempted library merge conflicted |
| 2 | src/compiler/corePublic.ts:9:5 | still there | compiler / records | [native-maplike-index.a](../../native-maplike-index.a) | records-maplike-next 8c013f1c reconciled on the train; merge conflicted |
| 3 | src/compiler/core.ts:195:37 | still there | stage3 source contract | [native-same-map-return-cast.a](../../native-same-map-return-cast.a) | truthful sameMap/builder/diagnostic contract; union proposal remains unapplied |
| 4 | src/compiler/core.ts:230:22 | still there | compiler / generic arrays | [native-generic-empty-array.a](../../native-generic-empty-array.a) | generic empty-array fallback admitted on train |
| 5 | src/compiler/core.ts:421:12 | still there | compiler / array brands | [native-sorted-array-brand.a](../../native-sorted-array-brand.a) | SortedReadonlyArray brand cast/view on train |
| 6 | src/compiler/core.ts:616:97 | still there | compiler / predicates | [native-predicate-callback-parameter.a](../../native-predicate-callback-parameter.a) | proven generic predicate callback parameter |
| 7 | src/compiler/core.ts:890:14 | still there | stage3 source types / compiler casts | [native-process-any-cast.a](../../native-process-any-cast.a) | truthful removal of the browser cast, or supported checked any view; no edit in this unit |
| 8 | src/compiler/debug.ts:80:31 | still there | compiler / entries provenance | [native-key-array-cast.a](../../native-key-array-cast.a) | finite key-array provenance; entries-provenance 05a60066 conflicted; 76 already supplies truthful key type |
| 9 | src/compiler/debug.ts:102:56 | new | compiler / namespace containers | [native-debug-container-read.a](minimals/native-debug-container-read.a) | dynamic Debug namespace value container; initialization/declaration probes do not cover this |
| 10 | src/compiler/debug.ts:161:131 | still there | compiler / assertion predicates | [native-generic-assert-non-nullable.a](../../native-generic-assert-non-nullable.a) | generic NonNullable assertion proof on train |
| 11 | src/compiler/debug.ts:191:105 | new | compiler / predicate callbacks | [native-predicate-callback-parameter.a](../../native-predicate-callback-parameter.a) | assertEachNode predicate callback and assertion overload proof |
| 12 | src/compiler/debug.ts:207:101 | new | compiler / predicate callbacks | [native-predicate-callback-parameter.a](../../native-predicate-callback-parameter.a) | assertNode predicate callback and assertion overload proof |
| 13 | src/compiler/debug.ts:220:107 | new | compiler / predicate callbacks | [native-predicate-callback-parameter.a](../../native-predicate-callback-parameter.a) | assertNotNode predicate callback and assertion overload proof |
| 14 | src/compiler/debug.ts:233:97 | new | compiler / predicate callbacks | [native-predicate-callback-parameter.a](../../native-predicate-callback-parameter.a) | assertOptionalNode predicate callback and assertion overload proof |
| 15 | src/compiler/debug.ts:289:26 | new | compiler / detached callable methods | [native-function-to-string-call.a](minimals/native-function-to-string-call.a) | Function.prototype.toString.call receiver/call support |
| 16 | src/compiler/debug.ts:358:34 | new | compiler / enum record views | [native-enum-module-record-cast.a](minimals/native-enum-module-record-cast.a) | namespace-import enum dictionary view without invariant mutable writes |

[stops.json](stops.json) includes each exact message, placeholder and raw attempt number. [discovery3/report.json](discovery3/report.json) and compressed per-attempt output retain the complete diagnostics; [discovery-only.patch.gz](discovery-only.patch.gz) is the final line-normalized discovery diff. Assertion-overload discovery removes predicate contracts and throws. This loses narrowing, so three subsequent TS2322/TS2345 errors in Debug.checkDefined and nodeConverters are artifacts, excluded from the sixteen source stops. Their consumer bodies are replaced with throws only to continue discovery. No declaration or caller admitted this way counts as native acceptance.

The first two discovery attempts changed overload signatures inconsistently and produced additional artifacts; they are retained under excluded/, not used for the ordered list. The final attempt replaces each overload group together. Rows 11 to 14 share the minimal generic callback-signature failure; that witness has a throwing implementation and only tests signature admission, so it cannot establish the real assertion bodies. It is pending on the corresponding predicate/overload implementation and the unmodified parser build.

## Changes against the prior list

[comparison.json](comparison.json) compares all 21 prior focused probes with ec71eb48. Twelve now build and match Node stdout, stderr and exit; all twelve one-byte output mutants are caught. The enum initialization probes, runtime namespace-initialization probe and mutable Debug export probe changed from refused to matching. The three namespace declaration probes remain admitted, but their loaded messages do not prove invocation or construction semantics. Predicate overload/result probes and arguments.length witnesses remain matching; generic predicate callback parameters remain refused. The function-value arguments witness still prints 1. No accepted probe has wrong output.

The standalone `.a` non-null probe remains refused. The slice `.ts` indexed-read assertions are admitted by the existing extension-specific checked-assertion path, so their earlier slice refusal is gone; this does not show that `.a` assertions are accepted. SameMap, generic-empty fallback, SortedReadonlyArray cast, the key-array view and MapLike remain blocked. The finite key type from 76 is present, but does not implement array provenance.

Seven supplemental witnesses include a matching Partial<Record> argument and a matching predicate-to-AnyFunction declaration. Their two byte mutants are caught. The former does not test the explicit key-array cast, and the latter only prints a declaration-loaded message. Process-browser any cast, generic assertion proof, detached toString.call, enum-module dictionary cast and the Debug dynamic container remain refused. The four new fixture files are reduced witnesses, not extra source adaptations. All results, including declaration-only matches, are **pending rehearsal evidence**, awaiting compatible landing plus full unmodified parser acceptance.

## Node reference and local checks

Full-tree and slice Node dumps compare with empty cmp output: **36,429,231 bytes**, SHA-256 `686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615`. Both use the same pinned 81-file corpus at `/tmp/parser-adapted10`, preserving the existing acceptance input. They contain 4,149 JSDoc nodes, 3,846 tag nodes and five type expressions. The Node-end mutant changes one row; dropped tags and the directed JSDoc diagnostic mutant are caught on both trees. This is Node identity only, pending native parser output. The 10,406-case recovery manifest was not rerun in this unit because native compilation stopped in the checker.

Fresh apply and slice commands, adapter checks, source-span audit, split builds, all 28 focused/supplemental probes, 14 actual native-output byte mutants, Node-end/JSDoc mutants, 32 lane tests and evidence checks are retained. No whole compiler gate is claimed for this conflicted rehearsal. The evidence checker also rejects a removed stop, an erased pending dependency and a false native-green claim.

To inspect the finished evidence:

```sh
python3 stage3/drivers/parser/evidence/rehearsal/check.py
```

Rebuild commands are in the saved scripts, with the exact local scratch paths and pinned compiler SHA. Build mode commands were:

```sh
ADAMIC_NATIVE_SPLIT=0 ADAMIC_NATIVE_JOBS=5 /workspace/scratch/parser-rehearsal-adamic build /workspace/scratch/parser-rehearsal-slice/parser-proof-main.a -o /workspace/scratch/parser-rehearsal-build-modes/parser-split0
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 /workspace/scratch/parser-rehearsal-adamic build /workspace/scratch/parser-rehearsal-slice/parser-proof-main.a -o /workspace/scratch/parser-rehearsal-build-modes/parser-split1
```

Only evidence under this directory is committed. Neither scratch, discovered source edits, upstream clones, native binaries nor full dumps are committed.
