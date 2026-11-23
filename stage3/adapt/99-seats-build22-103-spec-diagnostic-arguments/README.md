# Family 103 diagnostic arguments

Adamic ruled this adaptation on Nov 23 2026 (#bbssdcy, family 103); Compiler accepts the head after its review and census.

In `src/compiler/commandLineParser.ts`, `validateSpecs` calls its nested
`createDiagnostic` inside the filter callback. The producer
`specToDiagnostic` declares `[DiagnosticMessage, string] | undefined`.
Both explicit returns construct fresh two-element array literals and the
remaining path falls through to undefined. The callback saves the result
in `const diag` and immediately guards the call with `diag !== undefined`.
Slot 0 is the `message: DiagnosticMessage` argument and slot 1 is the
`spec: string` argument. The direct call passes them in the same order
inside the same callback branch.

The family 103 ruling permits the two indexed arguments here under
`docs/rulings.md` line 220, which records the general call-spread refusal.
The adapter pins the complete producer and validator declarations from
the adapted baseline and changes the single spread call.

The baseline spread is at `commandLineParser.ts:4111:42`. The test takes
a tree that `stage3/apply.sh` wrote. It reads this adapter's before and
after snapshots from its `refs/stage3/` refs and runs the adapted call
beside the original.

```sh
adapter=stage3/adapt/99-seats-build22-103-spec-diagnostic-arguments
NODE_PATH="$PWD/stage3/api/node_modules" node "$adapter/test.cjs" \
    "<applied tree>"
```
