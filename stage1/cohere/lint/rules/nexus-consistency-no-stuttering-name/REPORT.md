Built: migrated nexus/consistency-no-stuttering-name from batch4 onto the unified handed-node registry, with verbatim messages and typed options.
Commits: branch codex/lint-port-nexus-consistency-no-stuttering-name starts at lint area d3a37422c; see its single owned-rule commit.
Commands/output: lint-registry, registry tests, gofmt, vet, TestRulesAgree, TestOwnedWitnesses and the owned TestMutants subtest PASS.
Mutant: removing receiver/property equality compiles and runs, then the Go comparison catches it on source Node, emitted JavaScript and sanitized native.
Not covered: unrelated mutants, full repository gate or large compiler/repository benchmarks; existing parser-recovery limits remain explicit.

DEDUP_LEDGER.md names batch4:nexus_consistency_no_stuttering_name.ts as
the only batch copy. This port replaces its enabled/kind polling and helper
imports with a concrete Rule/create pair and a node:true descriptor listening
only to PropertyAccessExpression. visit acts on its handed ParseNode; it
looks up semantic children but never refetches the target node or checks its
kind for relevance. No helper is needed outside this rule directory.
messages.a moves the policy text verbatim from batch4_messages.ts, matching
cohere/policy/messages/consistency-no-stuttering-name.json.

The names set is initialized once per file. A nonempty genericNames array
replaces the default vocabulary; missing/null/empty options retain defaults,
matching upstream. The oracle adapter decodes manifest field 5 into
nexus.ConsistencyNoStutteringNameOptions and returns that typed value. It
does not drop configured rows or return nil. No options guard, shared harness,
registry, parser or cohere rule source changed. All new Adamic modules are .a.

Real upstream test names were recorded by go test -list:

- TestConsistencyNoStutteringNameFires
- TestConsistencyNoStutteringNameStaysSilent
- TestConsistencyNoStutteringNameRespectsTheGenericNamesOption
- TestConsistencyNoStutteringNameNamesTheWord

The descriptor prefix TestConsistencyNoStutteringName captures every one,
including the two custom-option observations. Their 18 source observations
cover defaults, optional access, nesting, silence on computed/nonstuttering
access, custom vocabulary replacement and rendered names. The owned witness
adds Unicode/astral source prefixes and a parenthesized receiver, fires on
two default stutters, and contains result.value to expose the equality mutant.
A default witness needs no options sidecar, so it works on the current area
before the optional witness-options extension lands. Configured upstream
rows already exercise the adapter and settings path through field 5.

After sourcing /workspace/adamic-tools/env.sh, exact commands were:

```
gofmt -w stage1/cohere/lint/rules/nexus-consistency-no-stuttering-name/oracle.go
go run ./cmd/lint-registry
go test ./stage1/cohere/lint/registry -count=1 -v
gofmt -l stage1/cohere/lint/rules/nexus-consistency-no-stuttering-name/oracle.go
go vet ./stage1/cohere/lint/registry ./stage1/cohere/lint
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$|^TestRulesAgree$|^TestMutants$/^nexus-stuttering-name-equality$' -count=1 -timeout=30m -v
# From cohere:
go test ./internal/lint/rules/nexus -list '^TestConsistencyNoStutteringName'
```

Every command wrote its own log file, retained under evidence. Registry
tests PASS 0.395s; gofmt and vet output are empty. The unified test package
PASS 199.817s: TestRulesAgree 131.76s and 13058152 identical Go/source-Node/
emitted-JS/sanitized-native bytes over the combined registered/generated/
upstream corpus; TestOwnedWitnesses 34.50s and 92008 identical bytes in
selected and all-rule modes; owned TestMutants 29.64s (parent 33.51s).
This mutation runs cleanly and disagrees at case 145, line 2746: it invents
a stutter for result.value. The oracle stays outside Adamic and the rule has
no automatic fix or suggestion, so unchanged fixed source also compares.
Only this rule's mutant subtest was selected, not all inherited rule mutants.

Existing malformed method-signature recovery cases are reported by the shared
test as explicit parser refusals, independently checked against Go's recovered
answers. They are unrelated to this rule, not hidden successful comparisons.
No new rule case encounters a known blocker. No check was skipped, relaxed
or deleted. No full gate or throughput benchmark is claimed.

Step 1: all previous wave-09 owned implementations remain blocked for unified
certification by its checker boundary and/or dynamic RegExp. Its branch now
records PARKED.md with rule names and reproducer paths. No landing branch
for those rules was produced, per the all-blocked exception. This branch
starts cleanly from the area and includes none of those implementations.
Publication is only to codex/lint-port-nexus-consistency-no-stuttering-name.
