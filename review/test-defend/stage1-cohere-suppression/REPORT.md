# Suppression gap-row defense

Defended: D01 uniquely fails TestEachGapStandsWhereGapsMdSaysItDoes in the complete current two-row package. TestThePortParsesAsGoCohereDoes passes its positive comparison and all twelve built-in mutant witnesses. No skipped rows, Go panics, timed-out commands, bounded fallback or unknown outside-unit package rows. Stop after this unique catch; the brief permits up to three attempts, not three after a successful defense.

Starting origin/main: bfe0553300773c0b37db2c10df97adeb909705f8. Both top-level names still exist, unchanged since the audit. list.log records the current list. Copies of prior REPORT.md, rows.json, frozen-plan.json and matrix.json are preserved with prior- prefixes. No oracle, harness, test or fixture was edited. Compiler source was restored after the matrix.

## Code under test and oracle

CODE UNDER TEST: Adamic's Go lowering and native compiler for the two closed-gap source programs. In particular, lowering converts boolean template operands to ir.BooleanToString at internal/lower/expression.go:951, and native evaluates that node at internal/native/emit_expressions.go:135. The suppression port itself is not the code under test for this row; its fixture inputs are not mutated.

ORACLE: Node actually executes each unchanged gap source. Node's success and stdout must match the recorded expectations, true false newline and true newline respectively. Native success and stdout are compared to the same values, with leak checks. This combines external-run Node execution with self-written expected stdout. The parser row's oracle is externally executed Go cohere, compared byte for byte with native, source Node and emitted JavaScript. It is not a twin of the gap row: the rows run different programs; executor sides inside each row are not separate rows.

## Coverage and targeted difference

Commands for each NAME (the requested row and its named subsumer): timeout 120 go test -count=1 -timeout 90s ./stage1/cohere/suppression/ -run '^NAME$' -coverpkg=github.com/system-inc/adamic/internal/lower,github.com/system-inc/adamic/internal/native -coverprofile=review/test-defend/stage1-cohere-suppression/NAME.cover > NAME-coverage.log 2>&1. Both passed. Gap binary: 0.395s. Parser binary: 18.032s. compare_coverage.py and coverage-diffs.json preserve the nonzero-block covered-line difference calculation.

Four exclusive covered-line leads: internal/lower/expression.go:951, internal/lower/namespaces_call_graph.go:43 and :44, internal/native/emit_expressions.go:135. These are leads rather than completeness claims. The gap fixtures directly interpolate boolean expressions into text. The suppression driver prints booleans using conditionals selecting literal strings 1/0, so it does not reach native BooleanToString even when the twelve built-in port mutants run.

D01 changes one C format-string constant: ((%s) ? &adamic_string_true : &adamic_string_false) becomes ((%s) ? &adamic_string_false : &adamic_string_true). This is a permitted constant change in production emission, not an inserted statement, fixture-name trigger or harness edit. It emits valid C that maps boolean values to the opposite text; the input program still executes normally.

## Mutation evidence and replay

D01.diff is a standalone unified diff against the starting origin/main, with no selector. It passed git apply --check and go vet ./internal/native/. The full matrix used ADAMIC_BUILD_CACHE_DIR=/workspace/suppression-defend-cache/D01, preventing stale native products after the compiler change. The altered compiler built the gap programs and the suppression port's native products. Both gap programs exited 0 with empty stderr; their native stdout differed from unchanged Node output. No verdict depends on a compile failure, runtime panic or failed setup.

Exact failing lines:

gaps_test.go:77: natively: exit 0, stdout "false\n", stderr ""; Node prints "true\n"
gaps_test.go:77: natively: exit 0, stdout "false true\n", stderr ""; Node prints "true false\n"

D01-passed-rows.txt lists the other package row. matrix.json also lists all thirteen passing parser subcases (one positive comparison and twelve built-in mutant witnesses), complete commands, failure lines and timings. D01.log is the complete unpiped JSON test output. Both gap subcases fail, but they count as one requested top-level row. No survivor remains from the one planted mutant.

## Costs, brief friction and limits

Warm /workspace/adamic-tools/env.sh worked; setup skipped. npm ci ran in stage3/api before baseline. nproc: 5. Clean baseline: 17.623 binary seconds. D01 command: 24.261 wall seconds, including compiler build and execution; binary: 17.743s. Native phase-only rebuild time was not separated from the matrix. No run exceeded 90s and no input corpus was narrowed. The default fixed seed and 2,500 generated parser cases remained enabled. No tests from other packages were run; go vet validated the mutated compiler package.

The brief's package path identifies a stage1 port, but the requested gap row's actual target is the compiler, as the audit also documented. Mutating the port cannot test these separate gap programs. Its proper -coverpkg therefore instruments internal/lower and internal/native, rather than the Go test-only stage1 package. The four exclusive lines provided a direct lead, so no V8 workaround was needed. The prior audit report was named REPORT.md, not report.md; discovering its actual file name resolved the first read attempt. The audit's original subsumption rested on one generic undefined-comparison mutant, whereas this defense changes an exclusive boolean-text conversion path. No twins exception changes this result.

No requested row remains undefended. No missing name/assertion promise was identified in this row's two current closed cases: it runs Node, checks recorded output, compiles and compares native output, and checks leaks. The experiment does not prove every compiler defect or future gap state. The defense establishes package uniqueness for this production mutant; other packages and repo-wide uniqueness remain outside scope. No tests were deleted, rewritten or weakened, and no pull request was opened.

Final restored whole-package baseline passed in 18.078 binary seconds, with no skips.
