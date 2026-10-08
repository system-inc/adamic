Built: 20 further Map pairs / 136 ranked candidate reads; cumulative share c 307 pairs / 909 reads. Fixtures, tests, evidence and own counts only.
Commits: continues 85381238 on codex/views-callables-c; delivery SHA is reported after push.
Commands and outputs: focused Map oracle PASS 15.771s; alias/enum Map oracle PASS 3.750s; verifier PASS 307 pairs / 614 fixtures; scoped counts update and verification PASS.
Mutants: twenty Map receiver-certificate omissions caught in both native and JavaScript, forty executed comparisons; each changes exactly one read.
Uncovered: whole original tsc call contexts and remaining ranked families; Set, optional-host and bind/condition stay delegated or excluded; no compiler/runtime changes.

Certified ranks and ranked read weights:

| Rank | Reads | Receiver |
| --- | ---: | --- |
| 20 | 61 | SymbolTable |
| 29 | 53 | SymbolTable |
| 1097 | 2 | SymbolTable |
| 1685 | 1 | Map<number, NumberLiteralType> |
| 1700 | 1 | Map<number, Type> |
| 1739 | 1 | Map<string, ExtendedConfigCacheEntry> |
| 1748 | 1 | Map<string, ImportSpecifier> |
| 1754 | 1 | Map<string, IndexedAccessType> |
| 1760 | 1 | Map<string, LiteralType> |
| 1778 | 1 | Map<string, Signature> |
| 1787 | 1 | Map<string, StringMappingType> |
| 1790 | 1 | Map<string, SubstitutionType> |
| 1799 | 1 | Map<string, TemplateLiteralType> |
| 1802 | 1 | Map<string, Type> |
| 1808 | 1 | Map<string, UnionType> |
| 1832 | 1 | Map<string, string> |
| 422 | 4 | Map<string, HostFileInfo> |
| 1718 | 1 | Map<string, CommandLineOption> |
| 1745 | 1 | Map<string, HostFileInfo> |
| 1775 | 1 | Map<string, SeenPackageName> |

Complete original member declarations, aliases, enum values and read paths are preserved. Adjacent payload interfaces and implementation bodies are reduced explicitly as in the existing lane harness. SymbolTable retains the complete original __String and InternalSymbolName declarations. Native/JavaScript immutable Map producer certificates govern these known get/set/has/delete intrinsics. This is a controlled receiver/member-contract certificate, not whole-tsc execution or proof of unread payload fields.

Each source Node control prints completed. Positives match Node in release native, ASan/UBSan, leak/count checks and JavaScript. A Map with an incompatible boolean value schema stops at exit 70 before the original call, naming `(value as Target).items`, the expected Map or SymbolTable, and the actual schema. Deleting exactly one Property.ViewContract makes both mutant backends execute completed at exit 0. No mutant is credited for a clang or sanitizer failure. Exact messages are in next-map-candidates.json and next-alias-candidates.json.

New observed compiler boundaries (source Node completes; pinned lowering tests pass):

- Rank 23, NodeFactory.createTempVariable, factory.createTempVariable, 58 reads: unsupported callable contract. Both original overloads are retained. Needs a producer/invocation certificate for the complete overloaded signature, including its callback and optional payloads.
- Rank 224, DiagnosticCollection.getDiagnostics, suggestionDiagnostics.getDiagnostics, 7 reads: unsupported callable contract. Both original overloads are retained. Needs checked overload selection and compatible producer metadata.
- Rank 233, Math.floor, Math.floor, 7 reads: inherited library member floor read as an own field. Needs the actual intrinsic Math receiver path/certificate.
- Rank 425, Math.abs, Math.abs, 4 reads: inherited library member abs read as an own field. Needs the same receiver support.

Rank 11's generic callable and rank 47's Math.min receiver remain the prior batch's code dependencies. All prior code boundaries remain recorded, not silently treated as certified. The corrected disposition list records every new certificate and observed boundary. Set and optional-host worker branches were not merged.

Exactly forty new own allocation rows are appended, no existing row changes. The required global count updater again fails sixty inherited fixture subtests; its raw log is compressed. Scoped update and verification pass. The first batch's raw global log is also compressed to preserve its bytes without a text trailing-whitespace warning. Its staged diff check warned before that earlier commit; this batch corrects the artifact packaging. No compiler behavior changed.

Commands, with output redirected to evidence logs:

```
go test ./internal/oracle -run '^TestCheckedViewCallableShareCMap(Mutants)?/rank-(20|29|1097|1685|1700|1739|1748|1754|1760|1778|1787|1790|1799|1802|1808|1832)$' -count=1 -timeout=10m -v
go test ./internal/oracle -run '^TestCheckedViewCallableShareCMap(Mutants)?/rank-(422|1718|1745|1775)$' -count=1 -timeout=10m -v
go test ./internal/oracle -run '^TestCheckedViewCallableShareCCodeBoundaries/rank-(23|224|233|425)$' -count=1 -v
node stage3/interface-downcasts/lane5/share-c/verify.cjs /tmp/lane5-c-original
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewCallableShareCContinuationCounts$' -count=1 -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewCallableShareCContinuationCounts$' -count=1
```

Setup is reused from the preceding batch; timings and nproc are in continuation/REPORT.md and its raw setup log. No whole package or full gate run. Only codex/views-callables-c is pushed.
