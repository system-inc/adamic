Built: member-specific tag filtering, selected-contract preservation and structural fallback components; no source admission.
Commits: territory 6c321f8a; lane 4 refresh d15b4206; selector checkpoint recorded in Git history.
Commands: focused lower/oracle passed (0.006s/3.470s); counted absent controls passed (1.299s); touched-package vet passed.
Mutants: six component defects were caught; independent runner and logs are the next checkpoint.
Uncovered: all 84 pairs and 186 reads remain; full compiler interfaces, lazy read dispatch and transitive propagation are pending.

Working date for the whole family: October 12, 2026 UTC, conditional on the lazy
read hooks and normalized probes arriving by October 9. This is an estimate,
not a promise that unsupported callable, array or branded descendants disappear.
The shared-file merge conflicts make an unconditional completion date indefensible.

The deterministic queue in pair-progress.json retains all overlapping families,
receiver identities, declared types and source witnesses. Its leading pairs are:

| Rank | Receiver / field | Reads |
| --- | --- | ---: |
| 1 | NamedDeclaration & { name: DeclarationName }.name | 25 |
| 2 | CommandLineOptionOfListType.element | 22 |
| 3 | ElementWithComputedPropertyName.name | 12 |
| 4 | TypeNode & LiteralTypeNode & { literal: StringLiteral }.literal | 11 |
| 5 | ParameterPropertyDeclaration.name | 8 |

The 12-member declaration-name shape has another one-read pair: 26 reads total.
The inventory also labels narrowed receivers, array alternatives and callable
alternatives as this family. None is silently removed from the 84-pair queue.

UntaggedViewMembers consumes the existing IR registry and preserves member ids.
Each candidate's required finite scalar fields can filter it independently:
there is no requirement that every alternative names the same discriminant.
Optional tags cannot exclude legal absent values. A tag never certifies payloads.
The C and JavaScript selectors try later alternatives after failure, require a
nonzero contract and a structural matcher, and return the selected contract id.
Shared dispatch must preserve that id on every subsequent checked read.

The component oracle uses reduced 12-member and 5-member shapes, ordered by the
most-read demand, plus two interfaces distinguished by fields alone. These are
representative member-count controls, not copies of tsc's full interfaces.
Source Node establishes their ordinary values and exposes malformed payloads.
Normalized C/JavaScript component snapshots are checked against those outputs;
negative cases pin the field, expected union, found object, and exit 70.
Native positives run release and ASan/UBSan/LeakSanitizer. Panics disable leak
checking because exiting at a panic intentionally does not unwind live data.

The supplied C matcher reads normalized sample data, not production object slots.
The supplied JavaScript matcher uses own data descriptors, without getters.
Shared readiness and complete structural membership remain owner-supplied
adapters. The component does not implement a second readiness bitmap or flow
solver. Its nested-check mutant is an adapter mutation, not proof that every
compiler-generated nested field read is checked. Actual helpers, generics,
callbacks and stored-value propagation remain unverified for this family.

TestCheckedViewUntaggedSourceFrontier records the honest integration frontier:
source Node prints true for each viewed-source probe; common lowering refuses
at the cast with `object union checked view without a common finite discriminant`.
Both backends share that lowering refusal. Allowed absent controls compile as
ordinary source and print true on source Node, native, and emitted JavaScript;
this is not evidence for absence through a viewed union. Count-row additions
are supplied for the shared owner rather than modifying counts.md.

Required concrete hooks and the lazy merge's six conflicting files are recorded
in docs/checked-views-plan.md. Lazy admission 5002bfe0 was fetched and the merge
aborted. Lane 4 9ecdda53 was merged, resolving only the plan conflict by preserving
both additions. No other lane's compiler file was edited by this unit.

Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh,
source /workspace/adamic-tools/env.sh. Go ready 0.231s, Node ready 0.292s,
clang ready 0.742s, markdown ready 1.378s, submodules ready 27.998s,
Go build ready 66.995s, cache warm 67.085s, done 67.116s; nproc 5,
cgroup quota 4 CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Commands wrote complete output to logs, never through a pipe:

```sh
python3 stage3/interface-downcasts/untagged/rank-pairs.py
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestUntaggedView|TestCheckedViewUntagged' -count=1 -v -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewUntaggedSourceFrontier$' -count=1 -v -timeout 10m
go vet ./internal/lower ./internal/javascript ./internal/native ./internal/oracle
```

The full repository gate was not run. The full touched-package run is recorded
separately once complete; the inherited TestSharedArrayContractAdapter failure
on readonly (number | string)[] is still present and was already documented
on the baseline lane. No source pair completion or complete-family gate is claimed.

The user's later coordination ruling supersedes direct lane merging. Only
codex/views-integration is now consumed for other lanes' code. At the final
remote check the integration branch had not yet been published (ls-remote
returned no ref). Further direct lane merges were stopped. This unit is blocked
on that shared integration/lazy read dispatch and can rest after pushing its
component checkpoint. The user will wake it when lazy admission lands.
