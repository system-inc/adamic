Built: rebased onto 59451e23e and replaced composed repairs with separate automatic edits through reportNode.
Commit: the pushed branch SHA is reported once with completion; evidence pins the tested source contents.
Checks: registry, gofmt, vet, TestOwnedWitnesses and TestMutants pass; TestRulesAgree stops at a shared non-progress fixer mismatch.
Mutant: object-alias-message-changed is caught on source Node, emitted JavaScript and sanitized native by independent Go bytes.
Not covered: complete captured-corpus parity remains blocked; the full repository gate and timing were not run.

## Individual edits and configured witnesses

Requested base: 59451e23eefc9d83102afd6d1a26050d6c447227 on lint-rules/witness-options,
including fca1616e6's multiple automatic-edit support and configured witnesses.
The port's local Edit class and whole-declaration composition were removed.
It now imports SuggestionEdit and sends every upstream edit separately, in the
original order, to context.reportNode's edits array. No shared file changed.
Messages and the upstream adapter remain unchanged.

Four witnesses cover the default object alias, configured interface heritage,
configured default export, and a declared-global no-fix finding. Sidecars
exercise both the upstream bare string and the captured {"Style":"type"} form.
TestOwnedWitnesses passes: Go, source Node, emitted JavaScript and sanitized native
match on 118,922 complete protocol bytes, including findings, individual fix-edit
rows, fixed source, overlap rejection and the deliberate no-fix finding.
The TestConsistentTypeDefinitions prefix still captures all six real upstream tests.

The valid object-alias-message-changed mutation changes only the interface
message. Shared TestMutants records it caught on Node, emitted JavaScript and
native at case 148, line 2762. These runs must exit cleanly with empty stderr;
only independent Go byte comparison catches the mutation. It is now a proved
mutant, unlike the historical initial report.

## Remaining shared fixer mismatch

Exact source: type Shape = { value: string } followed by a newline, with no
semicolon. Selected rule: @typescript-eslint/consistent-type-definitions.
Options: "interface". Go proposes these edits (all preserved by the port):
replace 0..4 with interface, replace 10..13 with a space, remove 30..30.
The empty trailing removal is a no-op. Go internal/edit/apply.go applyToText
rejects that individual edit with ReasonNoProgress and applies the others.
Its complete wire output includes:

fix-edit	30 30	
rejected @typescript-eslint/consistent-type-definitions 30 30  the fix replaces text with itself
fixed	interface Shape { value: string }\u000a

The shared stage1/cohere/lint/lint.ts fixed() instead executes
panic('nonprogressing fix') when the proposed text equals the original span.
The port cannot remove this edit without changing the oracle's finding bytes.
The fix belongs in the shared engine's refusal handling, which this unit did not
edit or relax. The original multi-edit guard is resolved; this is a distinct
remaining mismatch.

The minimal unchanged-Go and source-Node reproducer is archived under
evidence/multiedit/reproducer, including the raw source, options manifest,
full command arguments and both outputs. Go exits 0, rejects the no-op and
produces the repaired interface; Node exits 70 with nonprogressing fix.
TestRulesAgree captures 2,067 unique source/rule/options cases before failing
at that same panic. Its existing unrelated malformed-input refusal probes
remain intact; no recovery flags or corpus cases were added or removed.

## Reproduction and checks

source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry
go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout 30m > /tmp/consistent-agreement.log 2>&1
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout 30m > /tmp/consistent-owned.log 2>&1
go test ./stage1/cohere/lint -run '^TestMutants$' -count=1 -v -timeout 30m > /tmp/consistent-mutants.log 2>&1

Setup passed in 40 seconds: Go, clang, Node and submodules 0 seconds,
cache warm 40 seconds; nproc 5, CPU quota four cores, memory 17.6 GB.
Registry, gofmt and go vet ./... pass. Logs are under evidence/multiedit.
Previous evidence files describe the superseded single-edit blocker and initial
implementation corrections. The current mutation is caught and owned witness
parity is green; complete upstream certification is not claimed while the
shared no-progress mismatch remains. Only the assigned port branch is pushed.

Full TestMutants passed in 774.744 seconds: 41 mutant subtests passed.
The full raw log is archived alongside the failing agreement log and passing owned-witness log.
