Built 14 additional original-member callable certifications covering 84 candidate reads, serving step 09.
Commit: this checkpoint on codex/views-callables-b; previous checkpoints 473904f4 and 8086754d.
Validation: 68 fixtures checked against Node, native release, ASan/UBSan, JavaScript, and successful-run leak checks; exact commands and final outputs below.
Mutants: all 14 arity obligation mutants killed by exit-70/message pins; both-backend valid-execution controls pass; original-declaration mutant killed.
Uncovered: 10 signature-admitted pairs still need full receiver certification; other share ranks remain open; optional-host and bind/condition families excluded.

This checkpoint adds ranks 61, 277, 313, 385, 427, 430, 433, 436, 445, 451, 454, 457, 478, 484. The share total is now 34 pairs / 169 candidate reads in 158 certified fixtures. Added counts are this share's rows only. Including the original lane ledger, this branch carries 190 certified pairs / 2,868 reads. No compiler or runtime changes, other lane merges, or PRs.

Complete original member declarations and original read expressions are retained. Evidence pins the independent TypeScript source at 050880ce59e30b356b686bd3144efe24f875ebc8. Adjacent carrier interfaces are reduced, following the lane's existing convention; this is not certification of all fields of those carrier types. The verifier checks member declarations, source hashes, and exact original reads. Rank 427 preserves the nested context.factory.converters read.

Higher-ranked admission probes retain overloads, generic parameters, and rest parameters. Fifteen now have pinned refusal tests; rank 205 has a pinned positive-fixture lowering refusal. Supported admission by itself is not certification, and the 10 admitted pending pairs are listed in batch-02-admitted-pending.json. The ranked list is not exhausted; original Debug namespace functions and tracing signatures still need witnesses, and remaining higher-read admitted pairs take priority in the next batch.

The full code-needed list is code-dependencies.json: 42 pairs / 1,112 candidate reads including the previous four blockers. Every row names the pair, original read, stop, and required change. Set rows are user-provided dependencies, not observations against the fixed branch: 22 assigned pairs / 48 reads name codex/views-set-receiver at 058635b9. This includes ReadonlySet and the Set arm of the mixed ReadonlyMap/ReadonlySet receiver; other obligations of that union remain open. That branch was not merged.

New observed dependencies:

| Rank | Pair | Reads | Needed |
| --- | --- | ---: | --- |
| 16 | ObjectConstructor.entries | 75 | original overload-set callable descriptors and overload selection checks |
| 25 | NodeFactory.getGeneratedNameForNode | 58 | original overload-set callable descriptors and overload selection checks |
| 28 | NodeFactory.createKeywordTypeNode | 53 | generic callable descriptors with proven instantiation |
| 40 | NodeFactory.createToken | 41 | original overload-set callable descriptors and overload selection checks |
| 55 | Math.max | 30 | rest-parameter callable descriptors and intrinsic signature checks |
| 130 | NodeFactory.mergeLexicalEnvironment | 13 | original overload-set callable descriptors and overload selection checks |
| 172 | NodeFactory.copyPrologue | 10 | callable descriptors for optional callback parameters and their return contracts |
| 181 | ArrayConstructor.isArray | 9 | proven type-predicate callable admission |
| 190 | NodeFactory.restoreEnclosingLabel | 9 | callable descriptors for optional callback parameters and their return contracts |
| 205 | ModuleResolutionCache  /  undefined.getPackageJsonInfoCache | 8 | lower the original optional-chain member call and its following object field read |
| 226 | ElementFlags[].push | 7 | rest-parameter callable descriptors and intrinsic signature checks |
| 244 | NodeFactory.getGeneratedPrivateNameForNode | 7 | original overload-set callable descriptors and overload selection checks |
| 250 | ObjectConstructor.defineProperties | 7 | generic callable descriptors with proven instantiation |
| 253 | Signature[].push | 7 | rest-parameter callable descriptors and intrinsic signature checks |
| 319 | Type[].map | 6 | generic callable descriptors with proven instantiation |
| 322 | string[].forEach | 6 | callable descriptors for optional callback parameters and their return contracts |

Commands use the existing successful toolchain setup and /workspace/adamic-tools/env.sh. GOPROXY=https://proxy.golang.org|direct; nproc=5. Original setup timing lines were Go 0.288s, Node 0.297s, clang 0.886s, markdown 2.163s, submodules 490.545s, go build 784.336s, cache 784.462s, done 784.490s. No setup failure.

Executed checks (output files are in logs/batch-02):

- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-02-families.json ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareB' -count=1 -v -timeout 10m
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-02-families.json python3 stage3/interface-downcasts/lane5/share-b/run-mutants.py: 14 arity mutants caught; both-backend valid-execution controls pass. Under the mutant, all 14 wrong-arity subtests fail their original exit/message pins while valid code runs successfully. No clang or sanitizer failure is credited as a kill.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-02-families.json node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: Verified 14 pairs / 84 candidate reads in 68 original-member fixtures.
- node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original: Verified 34 pairs / 169 candidate reads in 158 original-member fixtures.
- Declaration mutant: changed only rank 61 Target optional boolean annotation to number; verifier exited nonzero naming original declaration changed; restored before final checks.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: required global attempt; see global-counts.log for existing unsupported-fixture failures. The scoped updater appends only the 68 new rows.
- ADAMIC_CALLABLE_SHARE_B_FAMILIES=batch-02-families.json go test ./internal/oracle -run '^TestCheckedViewCallableShareBCounts$' -count=1 -args -update-counts
- go vet ./internal/oracle

Exploratory corrections: the first numeric-parameter wrong-members fixture retained number and correctly executed successfully. It was changed to string, then the negative passed. The mutant runner initially read the first checkpoint's family list while executing batch 02; its bookkeeping assertion failed. It now selects the same family file as the Go tests, and the rerun proved all 14 kills. Earlier admission probes accidentally tested unbound reads; they were changed to receiver-preserving calls before any code dependency was recorded. These initial results are not credited as blocker evidence or mutant kills.

No full package tests or full gate were run. Push goes only to codex/views-callables-b after the scoped final check and counts verification.

Final observed outputs: scoped final PASS (48.544s); scoped counts PASS (8.622s); go vet exit 0; combined verifier 34 pairs / 169 reads / 158 fixtures; required global counts FAIL (40.241s), 60 pre-existing fixture subtests. Counts diff is 68 added rows / 0 removed rows.
