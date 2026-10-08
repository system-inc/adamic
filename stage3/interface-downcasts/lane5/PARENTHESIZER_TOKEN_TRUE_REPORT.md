Built: original ranks 68, 69 and 70 callable members certified with 14 fixtures and 68 candidate reads.
Commits: continued from f04e467b005ac0e59d1501dff244151c97a9c2e8; this report belongs to the next commit on codex/views-callables.
Checks: original declarations/read spans, Node, release/sanitized native, JavaScript, leaks and lane counts pass; whole-table counts updater fails outside this group.
Mutants: native/JavaScript arity and result checks caught for all three pairs; native/JavaScript parameter checks and deferred payload registration caught for rank 68; all seven restored.
Uncovered: 32/2818 pairs and 1326/11063 candidate reads certified; 2786 pairs and 9737 reads remain; exact production reachability and the pending Union exception remain unmeasured.

Reporting date October 12; execution environment date October 8. The starting tip was resolved in full before work. Earlier reports already certify ranks 6, 7, 22 and 26. This group continues the ranked fixtures without changing compiler/runtime code or callback calling-convention refusals. The rank 61 original object-union refusal remains pinned.

The independent original source checkout is TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. The original verifier reproduced later-ranked-original-witnesses.json with no diff. All 14 new fixtures preserve the complete declarations and original member reads. Adjacent structural carriers are reduced, as in the prior groups. These certify member contracts, not the original compiler program or measured production demand.

| Rank | Member | Candidate reads | Fixtures |
|---|---|---:|---:|
| 68 | ParenthesizerRules.parenthesizeExpressionForDisallowedComma | 23 | 6 |
| 69 | Scanner.getTokenStart | 23 | 4 |
| 70 | NodeFactory.createTrue | 22 | 4 |

The parenthesizer keeps parenthesizerRules().parenthesizeExpressionForDisallowedComma(expression), including a reached expression.value producer read. Compatible object parameter/result conventions pass. A string-only parameter producer is refused at the callable read. A producer declaring expression.value as string receives the lazy field check when its body reads the actual number. Scanner returns its original number result; the true-literal factory returns an aggregate whose value is read. Every member has valid, noncallable, wrong-arity and wrong-result fixtures.

| Mutant | Check that caught it |
|---|---|
| native-arity | Pinned callable member refusal for all three pairs versus execution |
| javascript-arity | Same three member-read pins versus execution |
| native-result | Same three member-read pins versus later result handling |
| javascript-result | Same three member-read pins versus later result handling |
| native-parameters | Parenthesizer member-read parameter refusal versus execution |
| javascript-parameters | Same parenthesizer parameter refusal versus execution |
| payload-reads | Deferred expression.value string-read pin versus sanitizer output |

All mutations failed executable oracle assertions, with no build or clang-warning failures counted. The existing runner restores each source in finally. Four mutations select all three pairs; three select rank 68. This is 15 pair-level mutant checks across seven mutation runs. The restored compiler/runtime files have no diff.

Commands below use source /workspace/adamic-tools/env.sh; output goes directly to logs. Durable raw logs are under logs/parenthesizer-token-true.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/lane5-setup.log 2>&1
node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-original 61,67,68,69,70,73,74,75,76 later-ranked-original-witnesses.json > /tmp/lane5-next-original.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 68,69,70 >> /tmp/lane5-next-original.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(disallowed-comma|token-start|true)$' -count=1 -timeout 5m > /tmp/lane5-next-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py disallowed-comma token-start true > /tmp/lane5-next-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-next-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-next-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-next-counts.log 2>&1
go vet ./internal/oracle > /tmp/lane5-next-vet.log 2>&1
```

Setup completed successfully: Node ready 0.060s, Go ready 0.062s, clang ready 0.498s, markdown dependencies ready 0.885s, submodules ready 177.790s, Go build ready 417.304s, build cache warm 417.400s, done 417.425s. nproc is 5; cgroup quota is 400000/100000. Go 1.27.1, clang 20.1.8 and Node 24.19.0. The initial test attempt before submodule completion could not load cohere/TypeScript/tsc/go.mod; it was rerun after setup and is not certification evidence.

Original verification passed: nine original declaration/read spans and 202 candidate reads in the evidence file, then 14 new fixtures retaining complete declarations/reads. The focused oracle passed 22.246s; restored later-ranked oracle passed 8.762s, including the prior local-name family and rank 61 refusal. Mutant runner exited zero after catching all seven. Lane counts updater passed 26.193s and adds exactly 14 measured rows; existing rows are unchanged. Oracle vet passed with an empty log.

The required whole-table updater exited one in 37.014s. Its raw log records predicate/overload/process and Node-library lowering refusals, nominal-write refusals and native count executions outside these new fixtures. Previous reports record this branch's whole-table updater limitation. No whole-table pass is claimed. No full package tests or full repository gate were run.

Static inventory remains 4/308 pairs and 34/1503 reads certified, with 304 pairs and 1469 reads remaining. Conservative counts are candidate inventory counts, not production execution counts. Next original fixtures already identified in later-ranked-original-witnesses.json include ranks 73 through 76. Generic, predicate, overload, rest and unresolved intrinsic signatures remain uncertified; no pending integrator decision was inferred or implemented.
