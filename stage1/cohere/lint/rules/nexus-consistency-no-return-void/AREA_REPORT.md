Built: eight retained rule ports rebased onto landed area 7481e032; return-void now uses JS RegExp literals, with raw regex and JSX controls.
Commits: this area evidence follows the owned rule rebase; helper branch is pushed at 78002cef4 on the same area and main 39638d9e.
Commands: shared TestOwnedWitnesses PASS (71,769 matching Go bytes); shared mutants and vet PASS; final TestRulesAgree FAIL on top-level await.
Mutants: eight original retained rule mutants plus one regex translation mutant compile and are caught on source Node, emitted JavaScript and sanitized native; twelve helper mutants also pass after rebase.
Not covered: complete upstream parity, current broad compiler/stage1 corpus, new throughput or full gate; top-level await and JSX fix reparsing block landing, so no new helper claim.

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
