Built: process-exit-after-output, uncleared-race-timeout and require-blocking-standard-streams as separate .a rules, bringing wave 02 to nine completed ports.
Commits: claim 958c5c2fa02758920b4c949a48f0f5f325dd7b09; implementation 42724911d92d05416c20d0e5c83fa1299f2c1d43; compatibility guards 774a422e13fb20f52747b3c2a9dcc71e1b29362a, all pushed.
Checks: standalone agreement PASS 108.608s; all bridge packages PASS; ASan/UBSan/LSan and released-handle checks PASS; source gate 276 rules, 14 modules, zero findings; Go vet and filtered Node oracle PASS.
Mutants: three rule judgments, namespace identity, signature scope, all four new raw questions and the released registry are killed; all seven existing bridge foundation mutants also fail as expected.
Limits: concurrent claims on 17 other branches; default options and pinned 77 compiler plus 287 repository roots; filtered Node gate rather than the full repository test suite; native slower than Go.

The six earlier claimed rules were implemented, tested and pushed through
84ecd45c before this continuation. Selection fetched 325 origin refs, inspected
92 distinct claim blobs, and excluded main/base implementations and every rule
name in the claim directories. The reconstructed initial tip audit has zero
matches for these three names. VOLUME_REPORT.md gives each zero compiler and
repository volume. The reservation was pushed before code at 00:59:43 UTC.

The final fetch revealed overlapping reservations on waves 03, 04, 05, 07, 08,
09, 10, 12, 13, 19, 20, 21, 24, 25, 27, 28 and 30. Their observed claim commits
have later timestamps; this is not a proof of push-arrival order. The full tips,
claim matches and commits are recorded under validation-wave-02-continuation-2.
No additional rules were claimed, no history was rewritten and no PR was opened.

The native rules decide symbol ownership, handle reachability, writes and exits,
catch-state invalidation, callee eligibility, import closure and execution order.
Four isolated Go/.a question pairs supply raw AST contexts, resolved signature
declarations/return flags, program module references, and unlabeled syntax events
and execution-graph edges. The raw graph builder is pinned from cohere's generic
MIT graph package, with its attribution under bridge/tsgo/checker/executiongraph;
its bridge wrapper emits no write/exit/block judgments or diagnostics. This graph
construction runs in the Go C library. The independent oracle invokes cohere's
unmodified production rules and program loader. Eight Go switch lines register
four questions. Existing registration generators, existing test harnesses and
protected compiler files were not edited.

The AST-context selector resolves exact foreign nodes through their program
source identities. It therefore handles virtual bundled-library paths without
changing the shared path normalizer. It checks canonical positions and exact
node kinds and preserves declaration-name kinds, so the NodeJS namespace cannot
be confused with an equally named ambient-module string.

Normal and sanitized outputs compare every diagnostic byte: rule and message
IDs, UTF-8 ranges, message text, complete fix records and suggestion records.
These production rules intentionally emit no fixes or suggestions; their zero
counts are also compared. Positive controls are necessary because both corpus
populations have zero findings for this trio.

| Population | Roots | Findings | Identical bytes | Sanitizers |
| --- | ---: | ---: | ---: | --- |
| Controls with Node declaration layout | 90 | 95 (51 exit, 32 blocking, 12 timeout) | 58,644 | PASS |
| Separate DOM timer program | 1 | 3 | 1,712 | PASS |
| TypeScript src/compiler | 77 | 0 | 5,318 | PASS |
| Frozen repository corpus | 287 | 0 | 18,485 | PASS |

Controls exercise imported process members, shadowing, Unicode/CRLF, branches,
loops, labels, switches, return/finally, catch bindings, callback and generator
roots, direct and one-level imported callees, async/await, blocking import order,
type-only/computed imports, entry reachability, await using, for await, default
DOM declarations, const timeout aliases, handle reads, shorthand references,
assignment-only writes, nested scopes and type-only signatures. Parse-valid
negative controls can deliberately contain semantic errors, as lint input can;
they are not claimed to be valid Adamic programs.

Every listed native or question mutant compiles and exits 0 with empty stderr.
Only comparison with independent Go finding bytes kills it. Timeout and blocking
message mutants keep the total finding count, showing why count equality alone
would miss the defect.

| Mutant | Change | First differing byte | Findings |
| --- | --- | ---: | ---: |
| Timeout rule | Invert unread initializer-handle test | 51,759 | 95 |
| Exit rule | Require empty write state at exit | 82 | 52 |
| Blocking rule | Suppress count phrase for exactly two exits | 2,067 | 95 |
| Namespace identity | Invert namespace identifier-kind guard | 58 | 14 |
| Signature scope | Traverse a function-type signature | 58,632 | 96 |
| Raw graph | Mark unreachable blocks reachable | 11,611 | 98 |
| Raw AST context | Remove global-augmentation flag | 58 | 6 |
| Resolved call | Mark every return type never | 15,467 | 92 |
| Module references | Erase type-only import flag | 39,110 | 94 |
| Released registry | Retain released program | exits 0 instead of required panic 70 | n/a |

The normal released-program probe panics with exit 70 and exactly
`adamic: panic: invalid or released checker handle`. The foundation bridge gate
also verifies 100 C ABI queries and 1,600 positions over four compiler files:
54,982 bytes match Go under ASan/UBSan/LSan. Its input/output-length mutants hit
ASan; output-free and region-ownership mutants hit LSan; stale-registry, source
position and link-opt-in mutants hit their dedicated assertions.

Three alternating complete count-only runs measure the final linked binary and
the Go production-rule oracle. Process times include loading, native parsing and
judgments, raw bridge work, output and teardown. All runs agree on zero findings.
They ran after the validation builds completed, without CPU pinning.

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 2.403913s | 0.351204s | 6.84x |
| Repository | 0.296339s | 0.112873s | 2.63x |

All rounds and load/query/run phases are preserved in measurements.json. Go's
load/run includes the same checker and corpus configuration; no speedup is
claimed. The TypeScript corpus is 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3),
cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db, and typescript-go is
8d550c837c90bd1805b047b7eeccc2baac2d5e7a. The branch remains based on 0d540f4.

The setup run recorded Go 1.27.1 ready (0s), clang 20.1.8 ready (0s), Node
v24.19.0 ready (0s), submodules (0s), build cache (76s), done (76s).
`source /workspace/adamic-tools/env.sh` supplies the environment. `nproc` is 5;
the cgroup quota is four CPUs. Logs capture complete test output; none was piped.

Commands and results, with output saved to the evidence directory:

```text
go test -v -count=1 -timeout 20m ./stage1/cohere/typeaware -run '^TestWave02Continuation2AgreementAndMutants$'
PASS 108.608s, with ADAMIC_WAVE02_CONTINUATION2_ARTIFACTS, COMPILER_MANIFEST,
REPOSITORY_MANIFEST and ADAMIC_TYPESCRIPT_SOURCE set to the recorded populations.

go test -v -count=1 -timeout 15m ./bridge/tsgo/...
PASS: bridge 81.263s, checker 0.335s; final updated checker rerun PASS 0.270s.

go test -v -count=1 -timeout 10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures)\.a$'
PASS 10.747s; matching generic-function and method fixtures also ran.

go vet ./...
PASS, empty output.

git diff --check
PASS, empty output.
```

The standalone test builds stage 0, normal/sanitized C archives and native
binaries, an independent Go oracle overlay, the compiling mutants, and stale
handle probes. It writes every subprocess stream separately. The unchanged
pinned cohere CLI does not scan .a for this source gate, so temporary copies
rename .a imports to .ts; the 14 actual implementation modules remain .a and
compile natively with their types checked. The full 276-rule source gate passes.

Earlier failed builds are preserved separately: unsupported bare construction,
optional calls/element access and untyped array fallbacks were rewritten locally;
a self-referential helper cache was removed after an ownership-analysis stack
capture. The DOM control exposed the virtual-path issue, fixed in the new
question. The byte oracle caught the await-using flag mismatch before release.
No shared compiler/harness change or deferred rule blocker remains.

The full repository gate was not run. The required populations are the frozen
manifests, not every repository file or every TypeScript program. Default rule
options, Linux and the pinned compiler versions are what this evidence covers.
The complete canonical streams, hashes, generated TypeScript input fixtures,
portable manifests, claim audit, timing phases and final logs are committed in
validation-wave-02-continuation-2. Binary archives are reproducible from the test
and are not committed.
