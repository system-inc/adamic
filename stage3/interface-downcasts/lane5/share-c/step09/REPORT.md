Built: step 09 adds 10 certified pairs / 27 ranked reads; cumulative share c is 256 pairs / 700 reads. Fixtures, tests, evidence and own counts only.
Commits: prior delivery 6e2e4878; this batch is committed and pushed only to codex/views-callables-c. Delivery SHA is in the final response.
Commands and outputs: scoped oracle PASS 10.987s, source verifier PASS 256 pairs / 512 fixtures, context verifier PASS five headers, own counts PASS. Required global updater still fails 60 inherited fixtures.
Mutants: ten Array.push write-certificate omissions caught in native and JavaScript, 20 executed backend comparisons; every mutant exits 0 with stdout 2 instead of the pinned exit 70.
Uncovered: six new code-needing pairs, ranks 2, 8, 56, 65, 80, 389; 31 Set pairs / 107 reads remain handed off under 058635b9. No compiler/runtime code changed.

New certified ranks, with ledger read weights: 161 (11), 326 (5), 704 (2), 707 (2), 884 (2), 1235 (1), 1238 (1), 1250 (1), 1307 (1), 2798 (1). Each has its own family directory. Original receiver types, Array.push declaration, member read paths, original alias text, and original numeric tags are retained. The seed and inserted value use the actual original concrete union member; assigning an anonymous record to a union array produced uncertified JavaScript source contracts in an earlier preparation run. The final typed-carrier fixtures pass unchanged compiler/runtime code.

For EACH listed rank, Node prints 2 and exits 0 in both variants. The positive is held to Node in native release, ASan/UBSan, LSan with balanced allocation counts, and JavaScript. The negative stops with the exact manifest diagnostic and exit 70 in release, sanitizer and JavaScript. The mutant changes precisely one ArrayPush.DictionaryProduction flag to bypass its write certificate; both backends execute the push and print 2, so the required exit/message assertion catches it. Ten ranks times two backends gives 20 mutants; exact negative messages are in step09/certified.json and array-certified.json.

New observed code boundaries:

- Rank 2, Debug.checkDefined, 313 reads: stage 0 cannot lower a generic function as a value. Needs generic function values through checked views.
- Rank 8, Debug.fail, 110 reads: refuses the original unsupported callable contract. Needs its optional AnyFunction callback and never-return contract.
- Rank 56, deduplicated.push, 29 reads: a producer explicitly declared 1[] admits push(3) through a number[] view in native and JavaScript. This negative does NOT stop. Needs literal producer array write-contract preservation and enforcement; excluded from certificates and counts.
- Rank 65, Debug.failBadSyntaxKind, 24 reads: refuses the original callable contract, including never and optional AnyFunction.
- Rank 80, Debug.assertEqual, 21 reads: stage 0 cannot lower a generic function as a value.
- Rank 389, Debug.assertGreaterThanOrEqual, 5 reads: refuses the original callable contract with optional AnyFunction.

The five namespace controls preserve original function headers and member reads; Node prints completed, and lowering refusals are pinned in code-blocked.json. Their bodies and adjacent interfaces are reduced explicitly. Rank 56 has an observation-only reproduction test which records current failure to reject the forbidden write; passing that reproduction does not certify sound behavior. Code-worker details are in BLOCKERS.md, code-blocked.json and step09/runtime-blocked.json. Together with the prior 15 boundaries there are 21 observed code-needing ranks; 629 share ranks remain uncertified, including 608 delegated/context/preparation cases. Set declarations and source witnesses remain listed in step09/SET_HANDOFF_058635b9.json for the combined worker.

Commands run from /workspace/adamic, after sourcing /workspace/adamic-tools/env.sh, with test output redirected (no whole package or full gate):

- `go test ./internal/oracle -run '^TestCheckedViewCallableShareC(Array(Intrinsics|Mutants)|CodeBoundaries)/(rank-(161|326|704|707|884|1235|1238|1250|1307|2798|2|8|65|80|389))$|^TestCheckedViewCallableShareCStep09LiteralArrayBoundary$' -count=1 -timeout=30m -v > /tmp/lane5-c-step09-final.log 2>&1`: PASS 10.987s.
- `node stage3/interface-downcasts/lane5/share-c/verify.cjs /tmp/lane5-c-original > /tmp/lane5-c-step09-source-verify.log 2>&1`: PASS 256 distinct original pairs / 512 fixtures.
- `node stage3/interface-downcasts/lane5/share-c/step09/verify-contexts.cjs /tmp/lane5-c-original > /tmp/lane5-c-step09-context-verify.log 2>&1`: PASS five original function headers and reads.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/lane5-c-step09-required-counts.log 2>&1`: FAIL 45.387s, same 60 inherited fixture failures. No inherited count rows changed.
- `go test ./internal/oracle -run '^TestCheckedViewCallableShareCCounts$' -count=1 -timeout=30m -args -update-counts`: PASS 2.320s, log /tmp/lane5-c-step09-counts-update.log. A try/finally wrapper temporarily selected only these ten array rows, then restored both full manifests.
- The same scoped count command without update args: PASS 2.381s, /tmp/lane5-c-step09-counts-verify.log. Exactly 20 own rows appended.

Preparation logs /tmp/lane5-c-step09-validation.log and validation2.log are failed exploratory fixtures, not certification evidence; validation3.log passed the then-expanded selection, and final.log validates the final deduplicated ten new ranks. Existing ranks 2414 and 2420 were recognized as already certified and retained byte-for-byte, excluded from new totals. No claim rests on failed runs. A missing clang PATH on the first build probe was corrected by sourcing the installed toolchain environment; it is not a compiler boundary.

Source pin remains independent upstream 050880ce59e30b356b686bd3144efe24f875ebc8. No cohere implementation copied, no other lane merged, no optional-host or bind/condition families claimed. This step proceeds from the delivery branch rather than main as authorized. Setup already completed earlier: Go 0.076s; Node 0.074s; clang 0.520s; markdownlint 0.981s; submodules 165.671s; go build 353.266s; warm 353.389s; done 353.416s; nproc 5.
