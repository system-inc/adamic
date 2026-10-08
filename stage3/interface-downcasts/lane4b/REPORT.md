Built: territory, ranked 73-pair/267-read ledger, eight source frontier probes; no compiler admission.
Commits: cf48926b territory; 6f5b25d3 ranking and source probes; evidence/report commit follows.
Commands: setup passed 94.868s; frontier probe passed; inherited scoped oracle passed 19.332s.
Mutants: inherited helper guard removal caught in release C and JS; this family's three mutants not run.
Uncovered: all 73 pairs and 267 reads, source runtime checks, transitive integration and full gate.

Whole-family working target is October 14, 2026, 23:00 UTC, conditional on the
shared read probes, member adapters and lazy dispatch arriving by October 9.
October 9 was an initial estimate before inspecting missing hooks. This is a
planning estimate, not a claim that a background worker will finish unattended.

The branch starts at lane 4 9ecdda53, a descendant of 583d19b7. Step 1 was pushed
as cf48926b; its plan addition reserves only new lane 4b files and requests owner
handoffs. Step 2 preparation was pushed as 6f5b25d3. Neither push removes demand.
The ranking retains checker identities and source witnesses, with a SHA-256 of
the compressed input. Pair rank and shape rank differ: the leading shape is
string | NodeArray<JSDocComment> | undefined (24 pairs, 63 reads); the leading
individual pair is CommandLineOption.type (33 reads). DiagnosticMessageChain
accounts for four pairs and 24 reads, and true | Node | undefined for one pair
and 20 reads. Together these four shapes account for 30 pairs and 140 reads,
all still pending.

The fixtures are reduced shape witnesses, not the original TypeScript compiler
interfaces. The JSDoc probe uses a readonly Comment array; its NodeArray extras
and the recursive DiagnosticMessageChain are not yet covered. Good controls
exercise primitive and reference branches, plus allowed undefined cases. Wrong
controls put numbers in object fields or a number in the option union. Source
Node prints the malformed nested number in three controls; the malformed option
calls get on a number and throws TypeError. The probe pins that source behavior.
No source expectation was inferred from Adamic output.

Both compiler commands refuse all eight probes before runtime execution, at the
helper's field declaration. For example diagnostic-good.a:5:17 reports:
`stage 0 can't lower a field of type string | Chain yet`.
The native and JavaScript refusals are recorded separately in
frontier-observations.json. An exit-1 compile refusal is not the required exit-70
runtime field check. Therefore none of skip-check, accept-wrong-shape, or
drop-nested-check can currently be tested as a runnable family mutant. No mutant
killed by a compiler refusal is counted.

Commands, with the generated environment sourced in every compiler/test shell:

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/views-object-primitive-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go build -o /tmp/adamic-object-primitive ./cmd/adamic
python3 stage3/interface-downcasts/lane4b/rank-demand.py
python3 stage3/interface-downcasts/lane4b/probe-frontier.py /tmp/adamic-object-primitive
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCheckedViewLane4HelperReads|TestCheckedViewMixedSelection' -count=1 -v -timeout 10m
```

All actual invocations wrote to logs. The scoped oracle reports package pass in
19.332s. Its inherited scalar-helper mutant removes only Property.View, runs
valid release code, and changes C to exit 0 with true/false and JS to exit 0 with
true/42, so both fail the pinned exit-70/message comparison. This evidence is for
the inherited boolean helper check, not object-plus-primitive source admission.
The selector controls pass, including the existing string|Identifier component;
component success never reduces this lane's demand. Full gate not run because
this checkpoint changes no production compiler files.

Setup reports Go ready 0.129s, clang ready 1.059s, Node ready 0.124s, markdown
ready 1.719s, submodules ready 37.924s, build ready 94.630s, cache warm 94.765s,
done 94.868s. nproc is 5 and cpu.max is 400000 100000 (four CPUs).
Go 1.27.1, clang 20.1.8, Node 24.19.0. No setup failure occurred.

Dependency merges were attempted and aborted without resolving owner-only files:

- Lazy admission 5002bfe0: plan, lower/cast.go, interface_cast.go, readiness.go,
  view_objects.go and native/view_fields.go conflict. This tip is a takeover plan,
  with implementation scheduled for October 8, not delivered lazy plumbing.
- Callables be6e4f33: shared plan and blockers conflicts.
- Shape conformance b8316acd: ir/ir.go, javascript/javascript.go, lower/cast.go,
  native/emit_expressions.go conflicts.
- Native arrays bf544d5c: shared plan and blockers conflicts.

The unit prohibits edits to another lane's files. No conflict was resolved by
dropping either side's compiler changes. Mixed-union origin is still 9ecdda53.
The plan names the required normalized presence/readiness slot probe, pure
non-panicking member matcher, read dispatch and selected-member propagation.
These are the concrete owner handoffs needed to proceed soundly. Reimplementing
readiness or guessing object membership from a heap tag would violate the unit.
Repository is left outside merge state, with no uncommitted compiler changes.
