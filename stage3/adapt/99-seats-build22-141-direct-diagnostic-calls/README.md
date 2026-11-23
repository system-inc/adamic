# Direct error-summary diagnostic calls

Adamic's Nov 22 Build ruling on #bbssdcy accepts family 141. Adamic refuses
`watch.ts:294:55` in the unit at `watch.ts:275:1` because the contextual
`DiagnosticAndArguments` tuple's rest counts as one element in its literal
arity check. The edit is confined to the diagnostic selector in
`getErrorSummaryText`.

The zero-error path returns the empty string before selection. The five other
branches call `createCompilerDiagnostic` with the arguments below.

| Condition after the zero-error return | Member of `Diagnostics` | Arguments in order |
| --- | --- | --- |
| `errorCount === 1` and `filesInError[0] !== undefined` | `Found_1_error_in_0` | `firstFileReference` |
| `errorCount === 1` and `filesInError[0] === undefined` | `Found_1_error` | None |
| `errorCount !== 1` and distinct length is 0 | `Found_0_errors` | `errorCount` |
| `errorCount !== 1` and distinct length is 1 | `Found_0_errors_in_the_same_file_starting_at_Colon_1` | `errorCount`, `firstFileReference` |
| `errorCount !== 1` and distinct length exceeds 1 | `Found_0_errors_in_1_files` | `errorCount`, `distinctFileNamesWithLines.length` |

The old tuple local is written by the selector and read only by the spread
into `createCompilerDiagnostic`. The selected branch now creates the same
diagnostic directly for the same formatting tail. The conditions still
short-circuit and evaluate only the selected call's arguments.
`firstFileReference` is still computed before the selector.

Rulings rows 214 and 334 in [docs/rulings.md](../../../docs/rulings.md) govern
absent arguments and tuple operations. Every call uses the exact number of
arguments supplied by its old tuple, including the message-only call. The edit
removes the intermediate tuple from this selector. The stock 6.0.3 syntax tree
identifies the function and the selector's statement boundaries. Both the
original and replacement text are pinned, including line endings.
