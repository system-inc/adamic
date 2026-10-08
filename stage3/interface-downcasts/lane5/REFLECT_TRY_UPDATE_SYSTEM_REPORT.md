Built: original reflect try update and System callable reads.
Commits: follows pushed d5d8a5eb25014ad7d3b4ee9c98aad4e50dc9ce51 on codex/views-callables.
Checks: 46 fixtures pass original declaration/read verification, Node, release/sanitized native, JavaScript and leaks; 46 counts rows added with existing rows unchanged.
Mutants: 7 applicable runs caught and restored, 47 pair-level checks, with early member-read and reached payload-read assertions.
Uncovered: 151/2818 pairs and 2669/11063 candidate reads certified; 2667 pairs and 8394 reads remain.

Reporting date October 12. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Ranks 296,298,301,303,307,315,316 add 42 conservative candidate reads. Adjacent carriers/helpers are reduced; exact production reachability remains unmeasured.

Ranks 296,298,301,303,307,315,316 retain complete original reflect-set, try-statement, element-access/type-argument/declaration-list update and System declarations and read expressions. context.factory, factory, sys and system are retained as original member read receivers. Adjacent AST carriers and producer bodies are reduced. Optional receiver and catch/finally arguments are exercised, as are absent/supplied type arguments and empty/supplied declaration arrays. System string parameters exercise zero/nonzero lengths and boolean results. No recursive original AST or original complete System execution is claimed.

Rank 294 uses the original createNewExpression: factoryCreateNewExpression destructuring method read; it remains pending preservation/proof of that calling context. Factory 288,290,291,292,302,304,305,306 have original larger signatures or tagged aliases and remain pending. 308 through 312 are collection intrinsics; 313 uses original builder-context carriers; 314 is generic; 317 carries original ResolutionMode; 318 through 324 need intrinsic or generic predicate signatures. No new refusal is claimed for untested families. Original IdentifierNameMap.toKey at 264 is a static class method, rather than an exported namespace function, and needs that receiver contract preserved. DebugType 268 is an original intersection alias with a debug method; it remains pending that complete receiver contract. Info 278 and performance 325 declarations are recorded for the next group and excluded here.

Set and arrow remain deferred under their earlier requirements. Existing callback calling-convention guards remain intact. No integrator Union exception decision was received.


Each named mutant ran the corresponding wrong-arity, wrong-result, wrong-members or wrong-parameter-payload fixture. Every mutant was caught by the assertion for the affected member or descendant read; payload mutations include sanitizer evidence. No compilation failure counted. Recorded runs:

```text
native-arity: caught for reflect-set, try-statement, element-access-update, type-arguments-update, declaration-list-update, system-directory-exists, system-file-exists
javascript-arity: caught for reflect-set, try-statement, element-access-update, type-arguments-update, declaration-list-update, system-directory-exists, system-file-exists
native-result: caught for reflect-set, try-statement, element-access-update, type-arguments-update, declaration-list-update, system-directory-exists, system-file-exists
javascript-result: caught for reflect-set, try-statement, element-access-update, type-arguments-update, declaration-list-update, system-directory-exists, system-file-exists
native-parameters: caught for reflect-set, try-statement, element-access-update, type-arguments-update, declaration-list-update, system-directory-exists, system-file-exists
javascript-parameters: caught for reflect-set, try-statement, element-access-update, type-arguments-update, declaration-list-update, system-directory-exists, system-file-exists
payload-reads: caught for reflect-set, try-statement, element-access-update, type-arguments-update, declaration-list-update
```

Commands used source /workspace/adamic-tools/env.sh and direct log redirection, with durable evidence under logs/reflect-try-update-system:

```sh
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original 296,298,301,303,307,315,316 > /tmp/lane5-reflect-try-update-system-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(reflect-set|try-statement|element-access-update|type-arguments-update|declaration-list-update|system-directory-exists|system-file-exists)$' -count=1 -timeout 5m > /tmp/lane5-reflect-try-update-system-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py reflect-set try-statement element-access-update type-arguments-update declaration-list-update system-directory-exists system-file-exists > /tmp/lane5-reflect-try-update-system-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$/^(reflect-set|try-statement|element-access-update|type-arguments-update|declaration-list-update|system-directory-exists|system-file-exists)$' -count=1 -timeout 5m > /tmp/lane5-reflect-try-update-system-restored.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-reflect-try-update-system-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-reflect-try-update-system-counts.log 2>&1
```

oracle: ok  	github.com/system-inc/adamic/internal/oracle	21.099s
restored: ok  	github.com/system-inc/adamic/internal/oracle	21.031s
counts: ok  	github.com/system-inc/adamic/internal/oracle	117.777s

The required whole-table updater exits one on outside-lane cases; its complete output is preserved. No whole-package tests or full gate. Existing callback and intrinsic-member calling conventions remain unchanged. Set needs collection-to-view conversion and recorded intrinsic callable signatures, as described in HELPER_HOIST_MODIFIERS_REPORT.md. Static totals stay 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs.
