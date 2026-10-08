Built: six groups certify 23 additional callable pairs, 284 candidate reads and 141 fixtures, including original tagged name unions and a stored function property.
Commits: 1a3c524a, fe969d26, ad2e41d5, de5826ab, 00407523 and this commit, pushed separately to codex/views-callables.
Checks: original tsc declarations/read spans, original aliases/tags, Node, release/sanitized native, JavaScript, leaks, measured lane counts, complete later-ranked tests and oracle vet pass; whole-table counts fails outside the lane.
Mutants: 40 runtime runs caught and restored, 142 pair-level checks; three additional original-source verifier mutants caught and restored.
Uncovered: 85/2818 pairs and 2159/11063 candidate reads certified; 2733 pairs and 8904 reads remain; arrow and broad ForInitializer union contracts, method-value and intrinsic families remain uncertified.

Reporting date October 12. Continued from 85d2852044c0efa2452ca104427c591408956703. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Counts are conservative inventory reads, not exact production execution coverage.

| Group | Commit | Fixtures | Runtime runs / pair checks | Restored oracle | Lane counts |
|---|---|---:|---:|---|---|
| helpers-case-newline | 1a3c524a | 15 | 5 / 13 | 7.182s | 52.786s |
| tagged-variable-conditional-binary-diagnostic-text | fe969d26 | 31 | 7 / 30 | 15.051s | 52.938s |
| lexical-module-false-type | ad2e41d5 | 22 | 7 / 23 | 9.780s | 54.030s |
| export-property-type-tags | de5826ab | 21 | 7 / 21 | 10.893s | 60.933s |
| prefix-syntax-static-unary | 00407523 | 25 | 7 / 27 | 12.447s | 62.382s |
| substitution-internal-call-class | this commit | 27 | 7 / 28 | 12.725s | 66.183s |

Certified ranks: 122,124,127,128,129,131,132,133,135,136,137,141,142,143,144,145,148,151,154,155,156,157,158. Each retains complete original declarations and member read expressions. Adjacent carriers and helper/producer bodies are reduced as specified in each group report. The certificates cover represented contracts and reached reads; they do not claim whole original compiler execution or every narrower literal-subtype producer.

BindingName preserves Identifier | BindingPattern and BindingPattern's two original alternatives, including kind tags 80,207,208. ModuleExportName preserves Identifier | StringLiteral and both original tags. PropertyName preserves all seven original alternatives and their tags. TypeOfTag preserves all nine original strings, each tested against Node. The new carrier verifier resolves aliases and kind values through the original TypeScript checker. The earlier untagged BindingName reduction remains a separate refusal probe; it is not proof that the original tagged pair is blocked.

The stored onSubstituteNode read uses tsc's original function-property declaration and previousOnSubstituteNode variable. Its arity/result/parameter guards and lazy node.value check are held to both backends. The source verifier now accepts and records property-function declarations while requiring the complete original spelling/style; no method declaration is rewritten to allow a detached read. Original createComma passed to reduceLeft remains an unbound-method refusal, excluded from certification and counts.

Arrow rank 120 is skipped as instructed. Its original ConciseBody witness needs a proved union representation and descendant-read contract. ForInitializer rank 146 also remains uncertified: the original VariableDeclarationList kind is 262, while Expression's kind remains broad; the reached initializer.value read refuses the untagged object-union contract. Its Node output 15 and lowering refusal are pinned. Neither alias is narrowed or given a synthetic tag to obtain certification. Set remains skipped with collection-to-view conversion and intrinsic callable-signature requirements in HELPERS_CASE_NEWLINE_REPORT.md. Other intrinsic, rest, generic, predicate and overload families remain uncertified; no plain-object substitute or single overload is counted.

Every runtime mutant and its affected pairs are listed in the six group reports and direct logs. Arity/result/parameter mutations are caught by early viewed member-read assertions; payload mutations by reached string field-read assertions with sanitizer evidence. All 40 runs were restored; no compilation failure was counted. Additional source mutants independently change Identifier's tag, BindingPattern's alias and onSubstituteNode's property style. Each is caught by its corresponding original-source assertion and restored. The three source-mutant logs are in tagged-variable-conditional-binary-diagnostic-text and substitution-internal-call-class.

Exact per-group commands and outputs are recorded in the group reports. Final checks also ran:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-updates-all-restored.log 2>&1
go vet ./internal/oracle > /tmp/lane5-updates-vet.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original "$(cat /tmp/lane5-next-ranks.txt)" > /tmp/lane5-updates-all-fixtures.log 2>&1
node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 122,141,144,145 > /tmp/lane5-updates-all-carriers.log 2>&1
gofmt -l internal/oracle/checked_views_callable_later_ranked_test.go internal/oracle/checked_views_callable_counts_test.go
git diff --check
```

ok  	github.com/system-inc/adamic/internal/oracle	152.853s
Verified 333 fixtures retain complete original declarations and reads.
Verified 28 fixtures retain original aliases and numeric kind discriminators.

The complete later-ranked harness includes earlier families and retained refusal pins. Vet and formatting pass with no output. All 141 counts rows are appended with the entire previous table unchanged. Each required TestCountsAreRecorded update was run and failed on existing outside-lane cases; each lane-specific updater passed. No whole-package tests or full gate were run. Source evidence covers 65 original members and 1064 candidate reads. No production compiler/runtime code or calling-convention guard changed, and no new integrator decision on the Union exception was received.

Static totals stay 4/308 pairs and 34/1503 reads, with 304 pairs and 1469 reads remaining. Setup reused: Node .060s, Go .062s, clang .498s, markdown .885s, submodules 177.790s, Go build 417.304s, cache 417.400s, done 417.425s. nproc 5, quota 4 CPUs; Go 1.27.1, clang 20.1.8, Node 24.19.0. Only codex/views-callables was pushed; no PR opened.
