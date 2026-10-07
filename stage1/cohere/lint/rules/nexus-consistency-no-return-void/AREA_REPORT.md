Built: eight retained rule ports rebased onto landed area d3a37422c; return-void uses JS RegExp literals and raw regex/JSX controls.
Commits: this latest-area evidence follows the owned rebase; helper branch is pushed at 8d48f3f81 on the same area and main b6b1538b0.
Commands: fresh 411-file corpus PASS (20,649,216 Go bytes), shared witnesses PASS (116,148 bytes), eight active mutants and vet PASS; full upstream TestRulesAgree FAIL.
Mutants: all eight active rule mutants and twelve helper mutants compile, finish cleanly and are caught on source Node, emitted JavaScript and sanitized native on area d3a37422c.
Not covered: complete upstream-fixture parity, the other required correctness packages, new throughput or full gate; top-level await and JSX fixed-source reparsing block landing, so no new helper claim.

## Landed harness and ledger

Fetched all origin heads. The landed area is 7481e0324e34a2537aafa9db7eeacda50405611b, containing the green harness merge 50a5f105 and main 39638d9e278d38bb5aeae887f46d55a70e47aaad. Before rebasing, compared its complete DEDUP_LEDGER.md with the previously read and pinned 41eb6eab2 ledger: identical SHA-256 4c39ec0cb129b05a3971ff257c26296d0d4ce06545b42e53dc14bbf0828a526f. The nine losing copies remain removed; only the eight winners/unique ports listed in DEDUP_REPORT.md remain. No new batch-only assignment to slot 13 is present. Both owned branches rebased cleanly. Upstream shared parser, harness, helper and allocator changes were retained without editing shared files.

The actual shared oracle now loads .a modules. No compatibility overlay is used for this area's rule checks. The local cohere and nested TypeScript submodules were cloned from existing pinned checkouts after the first test attempt identified a missing nested go.mod. This environment failure is not an oracle comparison or a mutant catch.

## Regex translation

Removed return-void's hand-written whitespace and keyword regex matchers. The fixed unsafe-leading-word literal is the shared regex table's nexus/consistency_no_return_void.go:100 translation, used as a fresh /gu literal to avoid persistent global state. The dynamic canonical Go regexp quotes the exact AST operand; its prefix and suffix are now matched by two fixed JS RegExp literals around that same operand slice. They use Go's ASCII whitespace and an absolute-end assertion. No option pattern, runtime-pattern fallback or finding-offset conversion is introduced. Source ranges still pass through the shared finding model.

Raw testdata/regex-boundaries.ts.txt covers form feed, NBSP, vertical tab, unsafe function declarations, identifier prefixes, a dollar after async, Unicode operands, parentheses and comments. Go decides findings and full fixed bytes. The active return-void mutant descriptor now broadens the ASCII prefix whitespace to JS \s; it is caught because NBSP incorrectly gains a fix. The former braces descriptor is retained inert as evidence/parking/brace-mutant.json.txt and was also executed successfully in this run.

## Actual shared tests

With source /workspace/adamic-tools/env.sh, from the rule worktree:

```
go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=20m > /tmp/wave13-area-final-rules.log 2>&1
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m > /tmp/wave13-area-supported-witnesses.log 2>&1
go test ./stage1/cohere/lint -run '^TestMutants$/(delay-resolve-name-ignored|return-void-braces-omitted|default-case-last-semantic-mutant|for-direction-semantic-mutant|guard-for-in-semantic-mutant|no-constructor-return-semantic-mutant|no-delete-var-semantic-mutant|no-eq-null-semantic-mutant)$' -count=1 -v -timeout=20m > /tmp/wave13-area-mutants.log 2>&1
go test ./stage1/cohere/lint -run '^TestMutants$/^return-void-unicode-whitespace-mutant$' -count=1 -v -timeout=10m > /tmp/wave13-area-regex-mutant.log 2>&1
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry > /tmp/wave13-area-rule-vet.log 2>&1
```

Supported raw witnesses, including the regex controls and a JSX callback with an unfixable return-void sibling, match actual Go, source Node, emitted JavaScript and ASan/UBSan native: 71,769 bytes, 20.538s. This covers all registered shared witnesses, not only this unit. Vet passes with empty output. The original eight-mutant run passes in 151.178s; the extra whitespace mutant passes in 20.286s. Each mutant's log records clean execution and its independent Go-output mismatch on all three modes. These are fresh unified-harness results, unlike the historical compatibility checks.

## Confirmed blockers

1. Final TestRulesAgree captures 2,238 unique actual Go source/rule/options cases, then exits 1 in 43.377s. Source Node refuses the Go-valid delay fixture `await new Promise<void>(function(resolve) { setTimeout(resolve, 500); }); export {};` with parser slice expected semicolon at 6. No complete corpus comparison succeeds. The unchanged input is gaps/top-level-await.ts.txt under the delay rule.
2. The original JSX callback now parses correctly and its clean behavior agrees with Go on all three modes. A fixable sibling `function stopEarly() { return void handleEnroll(); }` causes shared fixed-source reparsing to lose JSX mode: source Node exits 70 with parser slice expected GreaterThanToken, got Identifier at 24 in fixed source. The failed TestOwnedWitnesses run lasts 20.568s and is retained as area-jsx-fix-gap.log. The complete failing source is gaps/jsx-fix-reparse.tsx.txt. The active JSX witness instead has an unfixable numeric operand, keeping its successfully tested finding path distinct from this named fix-reparse gap.

The top-level-await frontend gap extends beyond the only-harness parking exception. Neither failure is silently counted as Go parity. This rule branch is pushed for review but not declared landing-ready or fully green. Shared parser or repair files are outside this unit, so no workaround or shared edit is made. No new helper is claimed.

The helper branch's three direct-helper packages passed again: 4,470 cases, 820,676 matching Go bytes and twelve semantic mutants; helper vet is empty and passing. Their same six Tailwind consumers remain documented in their reports: eighteen dependency occurrences removed, zero final blockers removed. Full consuming-rule parity is not claimed. Toolchain setup is reused from the logged 128s setup; nproc 5. No main or area branch is pushed and no PR is opened.

## Latest area runtime rebase

The final fetch found area d65a8f931c98655936ae04c6899f38f14862b73e, containing the native runtime profile landing. Main remains 39638d9e. Both owned branches rebased cleanly again and retained the native heap, string-build and string-search changes. No shared compiler, runtime, parser or harness file was edited by this unit. Earlier sections above describe the initial 7481e032 run.

The latest shared command is go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants)$/(delay-resolve-name-ignored|return-void-unicode-whitespace-mutant|default-case-last-semantic-mutant|for-direction-semantic-mutant|guard-for-in-semantic-mutant|no-constructor-return-semantic-mutant|no-delete-var-semantic-mutant|no-eq-null-semantic-mutant)$' -count=1 -v -timeout=20m. Its complete output and fresh vet are retained as area-latest-rules.log and area-latest-vet.log. The former braces mutant is historical evidence on the preceding area; the active return-void descriptor now carries the stronger whitespace translation mutant.

The fresh uncached TestRuntimeLastIndexOfMatchesNode passed in 5.963s: Node, emitted JavaScript, native release and ASan/UBSan agree on 758 output bytes; native cache hits=0 misses=3, Node hits=0 misses=2. This is an inherited runtime check, not a newly claimed mutant. All twelve helper mutants and 4,470 cases were rerun on this runtime and passed. The final full upstream rule gate and failing JSX fix control above were run on 7481e032; the subsequent area change modifies native runtime code, not the frontend or shared fixed-source parser. Those source-Node failures therefore remain blockers rather than evidence of a new complete green run.

Latest combined shared run exited zero in 172.166s: active mutants 150.98s and supported witness comparison 21.17s, with 71,769 matching Go bytes. Fresh rule/registry vet exited zero with empty output. This is bounded supported coverage; the full upstream failure is not converted into a passing claim. Both branches are on current fetched area d65a8f931 and main 39638d9e. No new work is claimed.

## Required correctness inputs follow-up

Fetched every origin head again. Main 39638d9e and area d65a8f931 are unchanged; both owned branches contain them and both worktrees are clean. No new mandatory-skip gate implementation has appeared in these refs. The instruction to supply correctness inputs is honored directly for the lint compiler/stage1 check, whose existing test otherwise skips without ADAMIC_TYPESCRIPT_SOURCE. No test or shared file is edited, disabled, relaxed or deleted.

Executed from the rebased rule worktree, with the logged setup environment sourced:

```
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -v -timeout=20m > /tmp/wave13-required-compiler-stage1.log 2>&1
```

The actual TypeScript checkout pin is 050880ce59e30b356b686bd3144efe24f875ebc8. The test ran, with no skip: 364 compiler and stage1 files, Go/source Node/emitted JavaScript/ASan-UBSan native identical on 20,648,441 output bytes, PASS in 84.933s. Complete output is retained as evidence/parking/required-compiler-stage1.log. This supersedes the earlier limitation about lacking a fresh broad corpus run on this area. No rule implementation changed, so the latest eight active rule mutant and twelve helper mutant results remain applicable.

This check's all-rule source corpus is distinct from the failing upstream-fixture corpus. It does not cover or repair the Go-valid top-level-await delay fixture or JSX fixed-source reparsing. Those remain named with exact inputs and failure logs above, and the branch is not declared fully green or landing-ready. The other input-dependent stage1 correctness packages and full gate were not run; they are not claimed passing or silently skipped. No new helper claim follows the landing cap.

## Current main proof and record landing

Main advanced to c7991b900362796aefd111474e65eb5398e91953, and area to b84a9d9314b65d3d0261ee017e233287b4f071da, which contains current main. The complete dedup ledger has the same SHA-256 as the pinned, already-read ledger. Both owned branches rebased cleanly, keeping upstream proven-predicate/lowering and record-runtime changes; no shared file was edited or reverted. Helper branch is published at cdb36d199 after all three packages, 4,470 direct-helper cases, twelve compiling semantic mutants and owned vet passed again.

The fresh shared rule command supplies the actual pinned TypeScript checkout and selects the upstream-fixture test, compiler/stage1 test, eight active mutants and supported witness test together:

```
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -run '^(TestCompilerAndStage1Agree|TestRulesAgree|TestOwnedWitnesses|TestMutants)$/(delay-resolve-name-ignored|return-void-unicode-whitespace-mutant|default-case-last-semantic-mutant|for-direction-semantic-mutant|guard-for-in-semantic-mutant|no-constructor-return-semantic-mutant|no-delete-var-semantic-mutant|no-eq-null-semantic-mutant)$' -count=1 -v -timeout=20m > /tmp/wave13-b84-rules.log 2>&1
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry > /tmp/wave13-b84-rule-vet.log 2>&1
```

No selected correctness test skips and no failed assertion is weakened. TestRulesAgree still fails: source Node exits 70 at ConsistencyNoHandRolledDelay.ts with parser slice expected semicolon at 6; duration 59.13s. The exact Go-valid reproducer remains under the delay rule's gaps directory. TestCompilerAndStage1Agree passes: 364 files, 20,648,441 matching Go bytes across source Node, emitted JavaScript and sanitized native, duration 91.84s. The full upstream failure is distinct from that broad supported-source success. Other required correctness packages and the complete gate were not run or claimed passing.

Rule implementations and test inputs have not changed in this rebase. JSX fixed-source reparsing is still documented with its exact failing input; this landing changes lowering/native records, not that shared source-Node reparse path. New claims remain blocked while the branch lacks full upstream parity. The new fresh suite output and vet log are retained under evidence/parking/b84-*.log when the run completes.

Fresh results on b84a9d931: combined suite exits 1 in 344.900s because TestRulesAgree fails on the named top-level-await input. TestCompilerAndStage1Agree PASS in 91.84s with 364 real-input files and 20,648,441 matching Go bytes; TestMutants PASS in 168.54s with all eight active semantic mutants; TestOwnedWitnesses PASS in 25.37s with 71,653 matching Go bytes. Every mutant is caught separately on source Node, emitted JavaScript and sanitized native after clean compilation/execution. Fresh rule/registry vet exits zero with empty output. No selected correctness test skips. Complete output is evidence/parking/b84-rules.log and b84-vet.log.

The shared frontend block is before rule dispatch: stage1/typescript/parser/parser.ts:1478 recognizes AwaitKeyword only when awaitContext is true or the following token is Identifier; awaitContext starts false, and this fixture next has NewKeyword. These code observations explain why changing the rule listener cannot fix its missing AwaitExpression. No parser change is made by this unit.

The JSX bug is also shared: stage1/cohere/lint/lint.ts:1412 rebuilds the repaired input with new Parser(result, 'fixed source'), while stage1/typescript/parser/parser.ts:23 derives JSX mode from the source path suffix. A fresh source-Node reproduction on this area emits the expected return-void finding and fix, then exits 70 while reparsing, with expected GreaterThanToken, got Identifier at 24 in fixed source. The exact source remains gaps/jsx-fix-reparse.tsx.txt, and complete fresh output is evidence/parking/b84-jsx-reparse.log. Reproduce with the configured Node environment:

```
cp stage1/cohere/lint/rules/nexus-consistency-no-return-void/gaps/jsx-fix-reparse.tsx.txt /tmp/wave13-b84-jsx-reparse.tsx
printf '/tmp/wave13-b84-jsx-reparse.tsx\tnexus/consistency-no-return-void\n' > /tmp/wave13-b84-jsx-reparse.manifest
node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/cohere/lint/main.ts --manifest /tmp/wave13-b84-jsx-reparse.manifest > /tmp/wave13-b84-jsx-reparse.log 2>&1
```

Both issues remain explicit blockers, not passing parity or newly skipped checks. No shared file, finding model, registry generator, oracle or comparison is edited. This branch is rebased and pushed with bounded green checks but is not declared fully green or landing-ready; no new helper claim follows.

## Directory-only legacy dispatcher landing

Area advanced to b46914832d70e00847d82d5d221ab7bb24040c53, carrying the remaining legacy and volume rules in registered rule directories. Main stays c7991b900. Its complete dedup ledger remains byte-identical to the pinned, already-read ledger. Both owned branches rebased cleanly; incoming shared context, dispatcher, oracle, legacy rule directories and tests were retained without edits. The eight owned implementations and mutant descriptors are unchanged. Helper branch is pushed at 39d68709a after the same three helper suites, 4,470 cases, 820,676 matching Go bytes, all twelve semantic mutants and vet pass on this area.

Reran the same four-test selected shared command recorded above, supplying ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 (actual pin 050880ce59e30b356b686bd3144efe24f875ebc8), with output directly to /tmp/wave13-b469-rules.log. TestRulesAgree still fails at the delay's Go-valid top-level await: source Node exits 70 with parser slice expected semicolon at 6, duration 46.27s. The captured file is ConsistencyNoHandRolledDelay.ts. TestCompilerAndStage1Agree passes on the migrated sources: 411 files and 20,649,216 matching Go bytes across source Node, emitted JavaScript and sanitized native, duration 87.41s. No selected correctness test skips or is weakened. Other required packages and the complete gate remain outside this filtered run.

A fresh explicit .tsx manifest reproduces the JSX fixed-source issue on the migrated driver: the expected return-void finding and fix are emitted, then source Node exits 70 with expected GreaterThanToken, got Identifier at 24 in fixed source. Complete output is retained as b469-jsx-reparse.log. The original failing source stays under the owned gaps directory; the migrated Linter still rebuilds new Parser(result, 'fixed source'). Shared code is not edited.

This fresh source corpus supersedes the earlier 364-file observations as current-area coverage. The full upstream failure remains distinct from those passing supported sources. No new rule or helper is claimed while the recorded blockers keep the branch from complete landing parity.

Final migrated-dispatch results: combined selected suite exits 1 in 322.032s solely because TestRulesAgree fails at the named top-level-await fixture. TestCompilerAndStage1Agree passes in 87.41s (411 files, 20,649,216 matching Go bytes); all eight active TestMutants pass in 163.84s; TestOwnedWitnesses passes in 24.49s (116,148 matching Go bytes, now including the additional shared legacy witnesses). Fresh rule/registry vet passes with empty output. No selected correctness test skips. All fresh stdout/stderr is retained under evidence/parking/b469-rules.log, b469-vet.log and b469-jsx-reparse.log. The supported witness count reflects the migrated shared rules and does not withdraw either unsupported fixture.

The branch remains a pushed, rebased work product with explicit full-parity blockers, not a fully green or landing-ready branch. The shared parser and fixed-source harness are outside the unit's files. No new helper is claimed under the landing cap, and no main or area branch is pushed.

## Current typeof compiler landing

Fetched all origin heads. Main advanced to b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and area to d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, which contains main. The complete dedup ledger is unchanged (SHA-256 4c39ec0cb129b05a3971ff257c26296d0d4ce06545b42e53dc14bbf0828a526f). Both own branches rebased cleanly and retained upstream typeof lowering/native slot changes. No shared source was edited.

Reran the same four-test selected command above with ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 and -count=1 -v -timeout=20m, output directly to /tmp/wave13-d3a-rules.log. TestRulesAgree fails in 55.53s at the unchanged Go-valid top-level-await delay (source Node expected semicolon at 6). TestCompilerAndStage1Agree passes in 84.21s: 411 real-input files and 20,649,216 identical Go bytes across source Node, emitted JavaScript and sanitized native. All eight active semantic mutants pass in 156.03s, independently caught by Go-output comparison after clean compilation/execution in all three modes. TestOwnedWitnesses passes on the registered supported fixtures; rule/registry vet passes with empty output. No selected correctness test skips. Complete results are evidence/parking/d3a-rules.log and d3a-vet.log.

A fresh explicit .tsx manifest also reproduces fixed-source JSX mode loss: source Node emits the expected fix and then exits 70, expected GreaterThanToken, got Identifier at 24 in fixed source. Its exact source remains gaps/jsx-fix-reparse.tsx.txt; fresh output is evidence/parking/d3a-jsx-reparse.log. The frontend top-level-await blocker extends beyond the only-harness parking exception; this branch is not fully green or landing-ready. No new helper claim follows.

All three helper packages and vet pass freshly on this base: 4,470 direct-helper cases, 820,676 matching Go bytes and twelve mutants caught in all three modes. Helper evidence is pushed at 8d48f3f81. Full consuming-rule parity, other required correctness packages, new throughput and the full gate are not claimed. All eight owned upstream rules ignore their options argument and declare no options type; their nil adapters remain unchanged, and the shared options guard is retained.
