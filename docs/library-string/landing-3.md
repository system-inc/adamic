# String third unit landing

This landing preserves `codex/library-string-3` at f298cee5db18029dc82b8606ffdf347c6264636a. Only `codex/library-string-3-land` is pushed. Main is never pushed and no branch is force-pushed. String second unit remains assigned to integration library seat 1; its prerequisites remain in this branch until main contains them.

The rebase retains both current main borrow and call-target analyses and the String/RegExp protocol implementations. Counts conflicts are resolved by regenerating the complete table, not selecting historical rows. The exception merge retains RegExp callback tracking and String code point failure tracking; obsolete protocol refusal guards are removed because the imported protocol implements catchable exceptions.

Two integration repairs were necessary. Freshness analysis admits the protocol's intrinsic and symbol methods while replacement callbacks use ordinary conservative call effects, including operand escape and state clobbering. Unknown future methods still refuse. Borrow analysis admits builtin error allocation and identity reads, with their operands still checked for array mutation.

Two additional mutants validate these repairs using Go overlays, leaving the tested compiler unchanged. Replacing callback effects with an empty value fails TestKnownRegexCallbacksUseOrdinaryCallEffects for all three callback forms. Removing builtin error admission fails TestBuiltinErrorBorrowEffects and TestThrowElementBorrowPlan. Both mutants exit 1; controls pass. Original method-family Node comparison mutants and before/after TypeScript-validity survey remain documented in report-3.md; no new library work was undertaken for landing.

Validation commands and outputs are recorded in the adjacent landing logs. Counts are regenerated, all affected package tests and the full uncached oracle suite are rerun, and go vet and diff whitespace checks are run. The final landing SHA is reported to integration externally; integration owns the merge into area/library.
