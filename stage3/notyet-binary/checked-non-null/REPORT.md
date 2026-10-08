# Checked TypeScript assignment targets

Merge requested c41c0e062e99da37820f822968d4df1b48cdaee7 at a32f1086. Resolve the one IR conflict by retaining both AssignmentValues and NonNullChecks. No existing incoming tests were edited. The merged oracle selection passes in 7.872s, including the owned assignment/logical/equality fixtures and incoming checked non-null controls.

## Raw coverage

Nine unique root sites (ten attempted raw contexts) used `assigning to a NonNullExpression`: three numeric array updates in debug, three numeric field updates in es2015, one string-array append in debug, and two Uint16Array updates in utilities. The new checked update rule covers the seven target shapes whose representations exist. This is a raw family count, not a measured net census reduction. The two Uint16 sites still depend on the unsupported representation at utilities.ts:10491:22 `typed array element type Uint16Array`; this unit does not manufacture another element width or change runtime C.

Evaluate receiver once, index once, and the old value once; route that held read through the imported nonNullValue before evaluating the RHS; fit the updated value to the original storage. Getter/setter dispatch uses the existing accessor finishing pass. New .a fixtures use explicit `?? panic` repairs and remain honest Adamic. Tests generate temporary .ts controls from those repairs to exercise the newly admitted checked syntax. Missing controls intentionally hold native to the checked JavaScript backend and the ruling's exit-70 behavior; source Node is independently shown to continue to the RHS. Plain `target! = value`, local assertion targets, static/private/super stores and unrepresented target types keep named stops.

Eight repair fixtures and matching checked controls cover numeric/string arrays, numeric/string fields, an optional-number accessor, Uint8Array, Int32Array and Float64Array. Nullable string RHS operands spell undefined before concat. Valid controls match Node in both backends, release native, ASan/UBSan and LeakSanitizer. Checked missing reads stop before RHS effects.

## Replays after the merge and after the target rule

All earlier examples were replayed after the requested merge. checker.ts:11225:13's selected function getTypeAliasForTypeLiteral now has no lowering findings. checker.ts:19080:29 proceeds to unchecked casts at 19084:30 and 19085:36. Other examples keep their ordered remaining findings in the attached JSON.

Immediately after the merge, debug.ts:1145:21 and 1148:21 still reproduce assigning to a NonNullExpression (exit 0). After the target rule, neither reproduces (exit 1): each reaches `reading connectors` at its original line after the earlier declaration failed. The string example debug.ts:1199:17's selected writeLane body now has no lowering findings. The field example es2015.ts:3972:21 has no finding at that target, while its enclosing body still hits structural method calls. Uint16 examples reach the named unsupported constructor representation. This preserves observation separately from the inference that represented targets are now supported.

## Tests and mutants

All output is attached as .log.txt, with no test output piped:

- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNonNullAssignment|TestNativeAgreesWithNode/internal/oracle/testdata/assignment_non_null' -count=1 -timeout 15m`: pass 5.234s.
- `python /tmp/non-null-assignment-mutants.py`: duplicate receiver, duplicate index, duplicate current read and RHS-before-current each fail on Node stdout; skipping the presence check fails TestNonNullAssignmentChecksBeforeRight. All restored.
- `python /tmp/non-null-adapter-mutant.py`: wrong optional-number setter unpack emits valid C and fails Node stdout. All six final mutant kills are semantic, with no compile failures counted.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts`: pass 24.519s, exactly eight new rows, no prior row changed.
- `go test ./internal/lower ./internal/ir ./internal/flow -count=1 -timeout 15m`: pass 40.862s, 13.321s, 83.617s respectively.
- `go test ./internal/native -run 'Test.*(Borrow|Reuse|Region|Devirtual|InheritanceMemory)|TestPassThroughsAreNotConsumers' -count=1 -timeout 15m`: pass 0.276s.
- `go test ./cmd/adamic -run 'Test.*NonNull|Test.*Checks' -count=1 -timeout 15m`: pass 1.818s.
- `go vet ./internal/lower ./internal/ir ./internal/flow ./internal/native ./internal/oracle ./cmd/adamic`: exit 0.

The accessor fixture exposed a pre-existing C99-invalid cast from adamic_maybe_number to the same struct type. internal/native/class_accessors.go now assigns that already-unpacked struct directly. The wrong-unpack mutant proves the conversion's value matters. No runtime C files changed. Every file outside the owned assignment function is named in the commit message. Other workers' lower functions are unchanged.
