# String third unit landing

This landing preserves `codex/library-string-3` at f298cee5db18029dc82b8606ffdf347c6264636a. Only `codex/library-string-3-land` is pushed. Main is never pushed and no branch is force-pushed. String second unit remains assigned to integration library seat 1; its prerequisites remain in this branch until main contains them.

The rebase retains both current main borrow and call-target analyses and the String/RegExp protocol implementations. Counts conflicts are resolved by regenerating the complete table, not selecting historical rows. The exception merge retains RegExp callback tracking and String code point failure tracking; obsolete protocol refusal guards are removed because the imported protocol implements catchable exceptions.

Three integration repairs were necessary. Freshness analysis admits the protocol's intrinsic and symbol methods while replacement callbacks use ordinary conservative call effects, including operand escape and state clobbering. Unknown future methods still refuse. Borrow analysis admits builtin error allocation and identity reads, with their operands still checked for array mutation.

Two additional mutants validate these repairs using Go overlays, leaving the tested compiler unchanged. Replacing callback effects with an empty value fails TestKnownRegexCallbacksUseOrdinaryCallEffects for all three callback forms. Removing builtin error admission fails TestBuiltinErrorBorrowEffects and TestThrowElementBorrowPlan. Both mutants exit 1; controls pass. Original method-family Node comparison mutants and before/after TypeScript-validity survey remain documented in report-3.md; no new library work was undertaken for landing.

Validation commands and outputs are recorded in the adjacent landing logs. Counts are regenerated, all affected package tests and the full uncached oracle suite are rerun, and go vet and diff whitespace checks are run. The final landing SHA is reported to integration externally; integration owns the merge into area/library.

## Final base and commands

Final rebased main: e8ba3d5d81de4d3773c723914fccd4c76248b965. The earlier gate against e011f8f passed all affected packages and the full oracle (179.957s); main then advanced, so validation was repeated after rebasing. The new call-target borrow test and the String builtin error borrow test are both retained.

Each command sources `/workspace/adamic-tools/env.sh`. Linux gate commands (stdout and stderr redirected to the corresponding log):

```
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test ./internal/load ./internal/ir ./internal/lower ./internal/javascript ./internal/native ./internal/flow ./internal/fresh ./internal/regexp ./cmd/adamic-test262 -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m
go vet ./internal/fresh ./internal/native ./internal/ir
git diff --check
```

Mutant commands use `go test -overlay` with replacement files outside the repository. Callback mutant selects TestKnownRegexCallbacksUseOrdinaryCallEffects in internal/fresh; borrow mutant selects TestBuiltinErrorBorrowEffects and TestThrowElementBorrowPlan in internal/native. Expected exit is 1 for each; unmodified controls are covered by the full package gates. No compiler runtime mutation was left in the landing branch.

The new main call-target reader guard rejected the imported RegExp diagnostic sort-comparator field read. Diagnostic analysis now uses ClosureTargets: bounded targets consult their diagnostic escape summaries, and unknown targets retain the conservative closure summary. TestCallTargetReaders supplied the failing pre-change control; no guard allowlist exception was added. The lower and IR packages are rerun after this repair, followed by the String/RegExp oracle and count checks.

The initial current-main package command exposed two integration controls: the direct comparator read and a test using MakeClosure while expecting an unknown target. The latter test now uses a Read of an unbounded function local; bounded harmless closures correctly remain safe under main's CallTargets implementation. This was a test-only correction, with no weakening of operand checks. The corrected borrow controls pass, and the native package is rerun in full. The initial failure output is kept alongside the final passing runs.

Current-main full oracle passes (231.378s), final filtered String/RegExp/count oracle passes (59.691s), final IR and lower packages pass (45.496s and 53.424s), and corrected borrow controls pass (0.075s). Load, flow, fresh, regexp and runner packages pass in packages-current.log. Vet and whitespace checks pass with empty logs.

Final full native gate: ok  	github.com/system-inc/adamic/internal/native	96.944s
