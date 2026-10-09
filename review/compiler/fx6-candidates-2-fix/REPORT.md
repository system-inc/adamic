Built receiver-keyed checked-read repairs for all four candidate-2 gate reds; tests are unchanged.
Compiler fix: aaca4be5; main merge: 98f9c270 (origin/main c4c59914).
Required regression commands, full lowering, reader guard and counts are recorded in this directory.
Three independent revert mutants were caught by the original regression tests.
The existing optional-read NotYet boundary remains; the full repository gate was not run.

The branch starts at candidate 2, 3e2166d5. Gate evidence was fetched from
refs/heads/gate-logs/60b7c503cb5b/20261009T135048Z/fast and preserved in reds.txt.
All reported failures reproduced after merging current main.

| Red | Fix SHA | Cause and resulting behavior |
|---|---|---|
| TestCheckedViewOptionalReadBoundary, JavaScript and native | aaca4be5 | Nullable receiver identities now select the present object's registered contract, restoring the pinned located NotYet. |
| TestCheckedViewObjectPrimitiveSource/optional-receiver | aaca4be5 | The same nullable receiver normalization preserves its optional-read boundary. |
| TestCheckedViewUntaggedSourceFlows/generic/wrong | aaca4be5 | Resolve instantiated receivers and remaining generic constraints before recording read metadata. All three backend runs stop at the checked read with exit 70. |
| TestFixturesAssertions/assertions/19_identifier_kind.a/stage0 | aaca4be5 | A common union receiver retains a boundary when one of its member types is registered as checked. The original located NotYet is restored. |

Registration remains keyed by receiver identity and member. Neither nullable
normalization nor union membership uses a global field-name fallback. Unrelated
same-named fields remain covered by the full lowering suite. Optional checked
reads keep their existing boundary rather than escaping it. Source Node controls
and all expected check failures remain exactly as the original tests specify.

Revert mutants are .diff files, generated and run by run-mutants.py:

- nullable-receiver: remove nonnullable receiver normalization. The optional-read
  boundary test reports "optional checked read escaped"; optional-receiver also fails.
- generic-receiver: leave the receiver binder rigid. The generic/wrong fixture
  exits 0 and prints true on native, sanitized native and JavaScript; exit-code
  comparisons catch the missing check.
- union-receiver: remove union-member receiver registration. The stage3 assertion
  compiles and its pinned gap assertion reports "gap changed".

Every mutant was restored before final verification. No tests, fixtures, status
pins or expectations were weakened or changed. No new test leaf was introduced.

Setup used GOPROXY=https://proxy.golang.org|direct and bash cloud/setup.sh.
Timing lines: Node ready 0.020s, Go ready 0.021s, submodules ready 0.060s,
markdown dependencies ready 0.072s, clang ready 0.151s, build cache warm 47.764s,
done 47.792s. nproc=5; cgroup cpu.max=400000 100000. Each test shell sources
/workspace/adamic-tools/env.sh.

Final commands (each writes to its named log, with an outer hard timeout):

```sh
go test ./internal/lower -count=1 -timeout 240s
go test ./internal/oracle -run 'TestCheckedViewOptionalReadBoundary|TestCheckedViewObjectPrimitiveSource|TestCheckedViewUntaggedSourceFlows|TestCountsAreRecorded' -count=1 -timeout 300s -args -update-counts
go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s
go test ./stage3/fixtures -run 'TestFixturesAssertions/assertions/19_identifier_kind.a/stage0' -count=1 -timeout 90s
python3 review/compiler/fx6-candidates-2-fix/run-mutants.py
git fetch -q origin main devtools/fast-gate cloud/merge-tree
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

Counts change only the two generic flow rows: the good fixture now keeps its
checked-read retain/release pair; the wrong fixture stops before allocating the
result it previously returned unchecked. Compiler scope is interface_cast.go
and view_member_read.go. Other gate scheduling and lint failures in reds.txt
are outside this repair unit.

The delivery merge is ad81d9b7, including origin/main 157a4355. The newer main
commits add test-only changes outside lowering and oracle emission. Required
checks were repeated after this merge. The final lowering package passed in
80.922s. The preceding restored-source run also passed: lowering 81.769s,
oracle regressions plus counts 93.533s, reader guard 13.213s, and stage3 0.861s.
Delivery lane output: "lane checks 3.5 s: gofmt and tools on 15 Go files,
t.Parallel on 2 test packages; vet 2 packages".

Delivery oracle regressions plus counts passed in 91.642s; reader guard passed
in 1.079s; stage3 assertion passed in 0.830s. All three mutants exited nonzero
with their expected behavioral catcher; the runner exited 0. The final
counts table still changes only the two generic rows.
