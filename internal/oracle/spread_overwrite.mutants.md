# Mutant for #xy74m5x's spread-overwrite stop

**S1, the stop dropped.** `spreadOverwrite` (internal/lower/spread_overwrite.go) fits a union slot as before and
returns the value without judging it, so an overwrite after a spread lowers whatever the source's slot holds.

Must turn red:
- `TestSpreadOverwriteStops/number_source.ts`, `number_present.ts`, `string_source.ts`, `optional_string.ts`,
  `optional_number.ts` and `width_method.ts`: each lowers where it must stop, and its failure carries the sanitized native run.

Must stay green: `TestSpreadOverwriteNeighborsAgree`, and `TestSpreadOverwriteStops/width.ts`, which object.go stops before
the helper.
