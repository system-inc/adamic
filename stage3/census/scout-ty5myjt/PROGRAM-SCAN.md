# program.ts — scan only

**No body lowering.** All runs load the same 79 compiler roots, select one exact file, use CommonJSBundle and no-output guards, and end in relation_report with zero units. Counts below include every emitted refusal-scan finding, not just the two targets.

| Candidate | Mode | Before → after | New | Gone | Changed | Both target sites clear? |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| diag-readonly | scan only | 81 → 83 | 4 | 2 | 0 | No — numeric-field refusals replace file refusals |
| diag-readonly-sorter | scan only | 81 → 81 | 2 | 2 | 0 | No — numeric-field refusals remain |
| diag-intersection | scan only | 81 → 81 | 2 | 2 | 0 | No — same writable file, different constraint spelling |
| diag-exact | scan only | 81 → 81 | 2 | 2 | 0 | Yes at callers; two sorter-handoff refusals replace them; stock-invalid |
| diag-overloads | scan only | 81 → 79 | 0 | 2 | 0 | Yes; no new finding |

The exact original two findings disappear by identity in every candidate, but that is **not** clearance of their sites when replacement refusals remain. A reason/constraint spelling change counts as gone plus new; changed requires the same (phase, kind, mapped where, reason) but different text. All locations below are baseline-mapped; raw archives retain actual candidate line numbers.

## diag-readonly — scan only

81 → 83 raw; 81 → 83 unique; new 4, gone 2, changed 0.

### new

- `src/compiler/program.ts:2785:20` — refusal_scan / Refused: a value of type T seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace

- `src/compiler/program.ts:2787:16` — refusal_scan / Refused: a value of type T seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace

- `src/compiler/program.ts:2796:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as DiagnosticReadView, which can write number | undefined where number is read

- `src/compiler/program.ts:2829:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as DiagnosticReadView, which can write number | undefined where number is read

### gone

- `src/compiler/program.ts:2796:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

- `src/compiler/program.ts:2829:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

### changed

None.

## diag-readonly-sorter — scan only

81 → 81 raw; 81 → 81 unique; new 2, gone 2, changed 0.

### new

- `src/compiler/program.ts:2796:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as DiagnosticReadView, which can write number | undefined where number is read

- `src/compiler/program.ts:2829:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as DiagnosticReadView, which can write number | undefined where number is read

### gone

- `src/compiler/program.ts:2796:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

- `src/compiler/program.ts:2829:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

### changed

None.

## diag-intersection — scan only

81 → 81 raw; 81 → 81 unique; new 2, gone 2, changed 0.

### new

- `src/compiler/program.ts:2796:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Readonly<Pick<Diagnostic, "file">> & Diagnostic, which can write SourceFile | undefined where SourceFile is read

- `src/compiler/program.ts:2829:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Readonly<Pick<Diagnostic, "file">> & Diagnostic, which can write SourceFile | undefined where SourceFile is read

### gone

- `src/compiler/program.ts:2796:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

- `src/compiler/program.ts:2829:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

### changed

None.

## diag-exact — scan only

81 → 81 raw; 81 → 81 unique; new 2, gone 2, changed 0.

### new

- `src/compiler/program.ts:2785:20` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

- `src/compiler/program.ts:2787:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

### gone

- `src/compiler/program.ts:2796:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

- `src/compiler/program.ts:2829:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

### changed

None.

## diag-overloads — scan only

81 → 79 raw; 81 → 79 unique; new 0, gone 2, changed 0.

### new

None.

### gone

- `src/compiler/program.ts:2796:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

- `src/compiler/program.ts:2829:16` — refusal_scan / Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

### changed

None.

## diag-readonly-sorter-utilitiesPublic — scan only

55 → 57 raw; 34 → 36 unique; new 2, gone 0, changed 0.

### new

- `src/compiler/utilitiesPublic.ts:304:47` — refusal_scan / Refused: a value of type (d1: Diagnostic, d2: Diagnostic) => Comparison seen as Comparer<T>, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace

- `src/compiler/utilitiesPublic.ts:304:67` — refusal_scan / Refused: a value of type (d1: Diagnostic, d2: Diagnostic) => boolean seen as EqualityComparer<T> | undefined, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace

### gone

None.

### changed

None.

## Readonly field and sorter follow-up

The helper-only named view makes file readonly, but DiagnosticWithLocation also narrows start and length from number | undefined to number. The new emitted numeric reason does not name which of those two fields was reached first. Neither was made readonly in this candidate; both call sites still refuse.

The same file-only named view on the sorter constraint removes the helper→sorter refusals. Its own file scan then exposes **two new comparator-handoff refusals**, utilitiesPublic.ts baseline-mapped 304:47 and 304:67 (candidate 306:47 and 306:67), for compareDiagnostics and diagnosticsEqualityComparer. Baseline utilitiesPublic.ts **34 → 36 unique (55 → 57 raw), scan only**, 2 new, 0 gone, 0 changed. The file-only view therefore does not produce a zero-new repair by moving the contract into the sorter.

The exact-type negative control remains rejected by stock tsc (1 TS2322). The overload candidate has stock 0, byte-identical JavaScript, and program.ts **scan only 81 → 79, 0 new, 2 gone, 0 changed**. This is the best measured program candidate; it is not a full-lowering or adaptation-oracle result.
