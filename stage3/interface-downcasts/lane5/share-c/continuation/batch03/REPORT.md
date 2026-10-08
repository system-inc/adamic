Built: 17 further Map pairs / 22 ranked candidate reads; cumulative share c 324 pairs / 931 reads. Fixtures, tests, evidence and own counts only.
Commits: continues f4d80337 on codex/views-callables-c; delivery SHA is reported after push.
Commands and outputs: focused oracle PASS 15.585s; verifier PASS 324 pairs / 648 fixtures; changed mutant harness and boundary suite PASS 8.831s; scoped counts update PASS 12.602s and verification PASS 12.052s.
Mutants: seventeen new Map receiver-certificate omissions caught in native and JavaScript, 34 executed comparisons; all 47 Map pairs rechecked after the test selector gained generic Target support.
Uncovered: no whole-tsc lowering or arbitrary generic-instantiation claim; three new Map source-schema boundaries remain excluded; delegated families and remaining ranked contexts stay listed; no compiler/runtime changes.

Certified ranks and ledger weights:

| Rank | Reads | Receiver |
| --- | ---: | --- |
| 419 | 4 | Map<__String, DeclarationMeaning> |
| 851 | 2 | Map<__String, Symbol> |
| 854 | 2 | Map<__String, TransientSymbol> |
| 1637 | 1 | Map<RegularExpressionFlags, ScriptTarget> |
| 1664 | 1 | Map<__String, InheritanceInfoMap> |
| 1667 | 1 | Map<__String, Node> |
| 1673 | 1 | Map<__String, boolean> |
| 1697 | 1 | Map<number, Symbol[]> |
| 1724 | 1 | Map<string, CommandLineOption[]> |
| 1796 | 1 | Map<string, TempFlags> |
| 1805 | 1 | Map<string, Type[]> |
| 1823 | 1 | Map<string, false | MutableFileSystemEntries> |
| 1835 | 1 | Map<string, true> |
| 1508 | 1 | Map<K, V> |
| 1511 | 1 | Map<K, V> |
| 1514 | 1 | Map<K2, V2> |
| 1793 | 1 | Map<string, T> |

Original member declarations, Map receiver types, aliases, enum values and exact read paths remain intact. Adjacent interfaces are reduced explicitly. A separate original-carrier audit finds no reduced interface with an index signature or direct Array/ReadonlyArray/Map/ReadonlyMap/Set/ReadonlySet ancestry in these Map certificate fixtures. Unread payload fields are not certified by this reduction.

The four generic Map witnesses preserve their original K/V, K2/V2 or T receiver spelling and use explicit string/number specializations. Their original enclosing parameters are unconstrained at those positions. The earlier nine generic Array.push witnesses also have unconstrained T, U or V at the relevant position; other original key parameters' constraints are retained in the evidence without claiming their contexts executed. These are bounded specializations, not universal generic proofs.

Source Node completes in both variants. Positives match Node in release native, ASan/UBSan, leak/count checks and JavaScript. Wrong Map value schemas stop at exit 70 with exact field, expected and found text. Each mutant clears exactly one receiver Property.ViewContract and both backends complete at exit 0. The test selector now identifies the single named items read instead of hardcoding Target without type arguments; its exactly-one-change assertion remains. All 47 Map mutant pairs pass after this test-only change. No production guard changed.

Three new code-needing pairs:

- Rank 1706, Map<number, { type: Type; declarations: IndexSignatureDeclaration[]; }>.get, indexSignatureMap.get: the valid empty source Map stops in both backends at `(value as Target).items`, expected and found Map names identical. Source schema emission/matching for the original nested object/array payload is missing. Sanitized/release native and JavaScript pins retain the exact stop; source Node prints completed.
- Rank 1763, Map<string, Map<string, ImportSpecifier>>.get, currentFileState.utilizedImplicitRuntimeImports.get: both lowerings refuse the holder field's unsupported Map key/value certificate contract. Needs complete nested Map source/target schemas and propagation of their child certificates. Node prints completed; the lowering refusal is pinned.
- Rank 1769, Map<string, RegExp>.has, namedArgRegExCache.has: the valid empty source Map stops in both backends, expected and found Map<string, RegExp>. Needs source-certified RegExp schema emission/matching in Map producers. Sanitized/release native and JavaScript pins retain the exact stop; Node prints completed.

These are observations of the reduced receiver witnesses, not native successes or certificate-count additions. Initial fixture preparation missed the __String enum inhabitant and inline payload declarations; those generator errors were corrected before validation. A direct Node invocation of generated JavaScript lacked the adamic loader; the oracle wrapper and final Go oracle supply the recorded runtime observations. Failed preparation and the initial mutant-selector mismatch provide no completion credit.

Exactly 34 new own allocation rows are appended, no old row changes. The mandatory global updater again retains sixty inherited fixture failures; the complete raw log is compressed. The scoped updater and verification pass. Compiler refusal pins introduced by this continuation now retain only the diagnostic message, making them independent of workspace path; their recheck passes. Original source observations retain full locations separately.

Commands, with output redirected:

```
node stage3/interface-downcasts/lane5/share-c/continuation/batch03/prepare.cjs /tmp/lane5-c-original
go test ./internal/oracle -run '^TestCheckedViewCallableShareCMap(Mutants)?/rank-(419|851|854|1637|1664|1667|1673|1697|1724|1796|1805|1823|1835|1508|1511|1514|1793)$' -count=1 -timeout=10m -v
node stage3/interface-downcasts/lane5/share-c/verify.cjs /tmp/lane5-c-original
go test ./internal/oracle -run '^TestCheckedViewCallableShareCMapMutants$|^TestCheckedViewCallableShareCMapProducerBoundaries$|^TestCheckedViewCallableShareCCodeBoundaries/rank-(11|47|23|224|233|425|1763)$' -count=1 -timeout=10m -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewCallableShareCContinuationCounts$' -count=1 -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewCallableShareCContinuationCounts$' -count=1
```

No whole package or full gate run. Setup is reused with timings and nproc in continuation/REPORT.md. Gofmt and diff checks pass; pushes use only codex/views-callables-c.
