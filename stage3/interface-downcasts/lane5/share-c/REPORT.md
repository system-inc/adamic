Built: step 09 adds 10 certified pairs / 27 ranked reads; cumulative share c is 256 pairs / 700 reads. Fixtures, tests, evidence and own counts only.
Commits: prior delivery 6e2e4878; this batch is committed and pushed only to codex/views-callables-c. Delivery SHA is in the final response.
Commands and outputs: scoped oracle PASS 10.987s, source verifier PASS 256 pairs / 512 fixtures, context verifier PASS five headers, own counts PASS. Required global updater still fails 60 inherited fixtures.
Mutants: ten Array.push write-certificate omissions caught in native and JavaScript, 20 executed backend comparisons; every mutant exits 0 with stdout 2 instead of the pinned exit 70.
Uncovered: six new code-needing pairs, ranks 2, 8, 56, 65, 80, 389; 31 Set pairs / 107 reads remain handed off under 058635b9. No compiler/runtime code changed.

Current step: see step09/REPORT.md for full commands, mutants and new boundary list. The following is the historical prior-delivery report and validation evidence.

Built: 246 pairs / 673 ranked candidate reads certified: 229 member families and 17 native Array.push receiver families. Fixtures/tests/evidence only.
Commits: base 926a1d39; delivery codex/views-callables-c. Certificate commits and final evidence commit are listed below.
Commands and outputs: final scoped oracle suite PASS in 343.197s; independent source verifier PASS for 246 pairs / 492 fixtures; scoped counts PASS. Required global counts updater failed on 60 inherited fixtures.
Mutants: all 229 callable-certificate omissions and 17 Array.push write-certificate omissions caught in both native and JavaScript, 492 executed mutant backend comparisons.
Uncovered: 639 share ranks remain uncertified: 15 observed code boundaries and 624 delegated/original-context/preparation cases. Full rank-by-rank disposition is in disposition.json.

Candidate reads are ledger weights, not measured production execution. The additive share manifest does not modify any shared ranked ledger or other worker's fixtures. Including the branch baseline, the lane arithmetic is 402/2818 pairs and 3372/11063 candidate reads, subject to integration reconciliation.

Original member declarations and read paths remain intact. Adjacent interface carriers and implementation bodies are reduced as in the lane harness; original alias declarations and numeric tags remain intact. Enum type carriers retain every original name and numeric value and use const enums because these fixtures need no enum runtime objects. Positive Array.push seed elements retain their declared interface types. All new Adamic files use .a.

Independent upstream source/checker pin: 050880ce59e30b356b686bd3144efe24f875ebc8. No cohere implementation code is copied. Reproduction: run prepare.cjs with that checkout root, then verify.cjs with the same root. Generated preparation artifacts remain outside certification until the oracle and mutants pass. Inferred implementation members, intrinsic receivers outside the certified Array.push families, generic/rest/overload signatures, and binding/effectful reads do not receive certificates merely by replacing their original contexts.

Validation commands, with all output redirected to files:

- `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh > /tmp/lane5-c-setup.log 2>&1`; PASS. Source /workspace/adamic-tools/env.sh.
- `go build -o /tmp/lane5-c-adamic ./cmd/adamic > /tmp/lane5-c-build.log 2>&1`; PASS.
- `node stage3/interface-downcasts/lane5/share-c/prepare.cjs /tmp/lane5-c-original > /tmp/lane5-c-original-extraction-final.log 2>&1`; extracted 885 share ranks.
- `node stage3/interface-downcasts/lane5/share-c/verify.cjs /tmp/lane5-c-original > /tmp/lane5-c-final-verify.log 2>&1`; PASS, 246 original pairs / 492 certification fixtures, with aliases and numeric tags verified.
- `go test ./internal/oracle -run '^TestCheckedViewCallableShareC($|Mutants$|Array(Intrinsics|Mutants)$|CodeBoundaries$|Counts$)' -count=1 -timeout=30m -v > /tmp/lane5-c-final-validation.log 2>&1`; PASS in 343.197s. Native release and ASan/UBSan, successful-fixture LSan/count balance, JavaScript backend, Node controls, byte-for-byte negative diagnostics, all mutants and recorded counts run.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/lane5-c-required-counts.log 2>&1`; FAIL in 41.403s on 60 inherited fixture subtests (62 failure lines including two parents). No baseline rows were rewritten. Scoped `TestCheckedViewCallableShareCCounts -args -update-counts` records only this share's fixture rows; final non-update comparison PASS.

No full package or full gate was run. The initial ordinary-enum exploration hit the default ten-minute test timeout. That run does not establish any certificate; const-enum reruns and the final complete scoped suite supply the final evidence. The temporary certificate claims from batch 12 were revalidated with all mutants in batch 14. A runner must treat timeout as an unconditional failure even if some subtests completed.

Each member mutant clears only its original read's ViewContract in lowered IR; both backends execute completed and exit 0 instead of the exact negative exit 70. Each Array.push mutant sets DictionaryProduction on the one push, bypassing its required write certificate; both backends print 2 and exit 0 instead of the pinned element-write failure. Changed-read/push counts are asserted to be exactly one. Every rank and diagnostic is pinned in certified.json or array-certified.json; the final log names every caught mutant.

The 15 code boundaries, their original declarations, exact stops, and required changes are in code-blocked.json and BLOCKERS.md. Their Node controls and lowering refusal assertions pass. They cannot provide native executed negatives or allocation counts until the compiler admits the complete original callable or carrier contract; they are excluded from certification/count totals. Three positives work while the wrong-arity variant is refused during lowering. No compiler/runtime fix is made.

Setup timing lines: Go 0.076s; Node 0.074s; clang 0.520s; markdownlint 0.981s; submodules 165.671s; go build 353.266s; warm 353.389s; done 353.416s. nproc=5; cpu.max=400000 100000. The initial generic origin fetch lacked the lane tracking ref; an explicit lane fetch established origin/codex/views-callables at the required 926a1d39 tip. No other lane is merged.

Pushes used only codex/views-callables-c, in batches of 10-20 new certificates. An HTTPS credential failure and its retry occurred after the Array.push batch; a later push succeeded and included that commit. Final delivery status is recorded in the final response.

Certificate commits:

```
80352491 Certify first 20 lane 5 share c callable pairs
3466a086 Certify lane 5 share c callable batch 2
6174e76e Certify lane 5 share c callable batch 3
07620501 Certify lane 5 share c callable batch 2
e0257790 Certify lane 5 share c callable batch 3
c99da2a6 Certify lane 5 share c callable batch 4
23cabbb4 Certify lane 5 share c callable batch 5
726008ab Certify lane 5 share c callable batch 6
a2418ec3 Certify lane 5 share c callable batch 7
572f7fa3 Certify lane 5 share c callable batch 8
cd22130d Certify 17 original Array push callable pairs through views
4d018fcf Certify lane 5 share c callable batch 12
1ca05684 Certify lane 5 share c callable batch 14
8283591a Certify lane 5 share c callable batch 15
```
