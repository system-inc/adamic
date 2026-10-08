Built: original declaration-name, left-side-access parenthesizer and returned source-file array member contracts certified; original liftToBlock method read retained as a refusal witness.
Commits: follows pushed ccdf894e; this report belongs to the next commit on codex/views-callables.
Checks: 19 supported fixtures, original declarations/read spans, Node, release/sanitized native, JavaScript, leaks, lane counts and oracle vet pass; whole-table counts updater fails outside this group.
Mutants: seven native/JavaScript callable and deferred payload mutations caught for applicable pairs and restored.
Uncovered: 42/2818 pairs and 1534/11063 candidate reads certified; 2776 pairs and 9529 reads remain; original binding, unbound-method and Union conventions remain uncertified.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 81, 83 and 84 each add 20 candidate reads. Complete declarations and original member reads are retained; adjacent structural carriers and helpers are reduced. Member-contract certification does not claim whole original compiler execution or exact production reachability.

getDeclarationName takes its original Declaration | undefined and two optional Boolean parameters. Undefined and supplied declarations, omitted/undefined/false/true flags and a reached node.value producer read agree with Node or stop at the pinned lazy read. parenthesizeLeftSideOfAccess keeps parenthesizerRules().parenthesizeLeftSideOfAccess and the optional-chain flag's observed behavior. getSourceFiles returns its original readonly SourceFile array; its element is read, then sourceFile.value. Numeric data behind a string field is checked at that descendant read, without eager aggregate admission.

Rank 82 is not certified. lift-block/good.a preserves context.factory.liftToBlock and its complete original method declaration. Node prints 3; lowering refuses unbound-method when the method is read as a value. TestCheckedViewCallableLaterRankedMethodRefusal pins that result. The initial broader group logged this refusal for all proposed variants; only the valid original refusal witness remains. No arrow replacement or property-signature rewrite is used to claim the original pair. Rank 80's generic Debug signature remains outside these supported fixtures.

Mutants: native-arity/javascript-arity and native-result/javascript-result were caught for all three supported pairs by early member-read pins versus execution or later result handling. Native-parameters/javascript-parameters were caught for declaration-name and left-access by parameter pins versus execution. Payload-reads was caught for all three pairs by the node.value, expression.value and sourceFile.value string-read pins versus sanitizer output. Seven runs, 19 pair-level mutation checks, all restored; no build failures counted.

Commands use source /workspace/adamic-tools/env.sh and direct log redirection:

```sh
node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-original 61,67,68,69,70,72,73,74,75,76,77,78,79,81,82,83,84 later-ranked-original-witnesses.json > /tmp/lane5-declaration-original.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 81,82,83,84 > /tmp/lane5-declaration-original-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(declaration-name|left-access|source-files)$|^TestCheckedViewCallableLaterRankedMethodRefusal$' -count=1 -timeout 5m > /tmp/lane5-declaration-supported.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py declaration-name left-access source-files > /tmp/lane5-declaration-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-declaration-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-declaration-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-declaration-counts.log 2>&1
go vet ./internal/oracle > /tmp/lane5-declaration-vet.log 2>&1
```

Original evidence covers 17 members/367 candidate reads. Final fixture verification confirms 20 declarations/reads: 19 certified fixtures and the method refusal witness. Supported oracle passed 8.411s; restored later-ranked oracle passed 32.950s, including previous members and all three refusal witnesses. Lane counts updater passed 32.114s and adds exactly 19 measured rows; existing rows are unchanged. Whole-table updater exited one in 34.867s on previously recorded outside-group cases. Oracle vet passed with an empty log. Durable logs are under logs/declaration-access-source-files.

Setup reused from the previous batch: done 417.425s, nproc 5, cgroup quota 4 CPUs; Go 1.27.1, clang 20.1.8 and Node 24.19.0. Static totals remain 4/308 pairs and 34/1503 reads certified. No compiler/runtime changes or calling-convention relaxation, no whole-package tests, full gate or PR.
