Built: territory reservation, ranked dictionary demand, source probes and record reuse evidence; no checked dictionary admission.
Commits: eb3ce2d6 reserves territory on lane 4 9ecdda53; the following checkpoint contains this report and artifacts.
Commands/output: ranking check passes; 3 source Node controls pass; 5 existing record modes match Node; restored compiler and record package pass.
Mutants: skip-check, wrong-shape and dropped-transitive-check were not run; no dictionary production check exists to mutate.
Uncovered: all 409 pairs / 2,256 reads, both-backend source integration, exit-70 pins, alias mutation and shared-flow coverage.

Whole-family estimate: October 16, 2026, 23:00 UTC, conditional on the shared
hooks and reconciled base arriving by October 9, 23:00 UTC. October 14 was a
provisional estimate before reading the representation. This is an estimate,
not a claim that a worker can implement the family without shared dispatch.

The first push, eb3ce2d6, leaves 409 pairs / 2,256 reads remaining. This second
checkpoint also leaves 409 pairs / 2,256 reads remaining. No ledger row is
removed for runtime component tests or source Node controls.

The executable `rank-demand.py` filters the frozen gzip by dictionary family,
checks its 409 unique receiver-type/field keys and 2,256 reads, preserves source
witnesses, ranks by descending reads and emits pair-progress.json and
demand-summary.json. `--check` reproduces both artifacts byte for byte. The
input gzip SHA-256 is pinned in the summary. All receiver provenance remains
Unknown; this ranking supplies no flow solver and erases no guards.

Highest individual pair: ParsedCommandLine.options, CompilerOptions, 114 reads.
Next: Block.statements, NodeArray<Statement>, 110; SourceFile.statements,
NodeArray<Statement>, 72; CallExpression.arguments, NodeArray<Expression>, 63;
VariableDeclarationList.declarations, NodeArray<VariableDeclaration>, 63.
Grouped CompilerOptions accounts for 22 pairs / 201 reads.

Observed inventory detail: 330 pairs / 1,904 reads explicitly display NodeArray.
The inventory tests getIndexInfosOfType after stock isArrayType/isTupleType, so
interfaces carrying numeric index signatures can enter the dictionary family.
The remaining 79 pairs / 352 reads are not necessarily string dictionaries:
JSDocArray, SortedArray and String also occur. Display matching is marked as
such in the ledger, not treated as checker classification or ownership transfer.
Shared array owners must reconcile these rows; all remain assigned and pending.

Observed representation: runtime/record.c already wraps the existing string-key
Map in an ordinary counted object, preserving key ordering, own reads, mutation
and iteration. It must be reused. Neither internal/ir nor lowering provides
record operations on this base. refusals.go still refuses index-signature
syntax; view_objects.go explicitly refuses dictionary views. Runtime record
values use adamic_value without an element logical tag. The table's single
reference_values flag is ownership metadata, not a type certificate. Inferring
number versus boolean from these bits, or from the target contract, is unsound.
Concrete source-probe, registry and emission handoffs are in
../../../../docs/checked-views-plan.md under Dictionary contracts unit.

The new `.a` probes are controls and future regression inputs, not backend
oracles yet. The options probes are small CompilerOptions-style index contracts,
not a claim to instantiate the full pinned CompilerOptions graph. Source Node
prints true/undefined for the valid and absent case and 42 for each malformed
scalar/nested value, all exit 0. Required backend behavior after integration is
exit 70 at holder.options['strict'] (expected boolean | undefined, found number)
and entry.name (expected string, found number). Exact runtime messages have not
been observed or pinned. The nested case must fail at the later field read,
not eagerly when the table entry is returned. check-probes.py preserves those
Node observations and can record compiler frontier results when a compiler builds.

Toolchain: GOPROXY='https://proxy.golang.org|direct'; nproc=5. Setup exit 1.
Timing lines: Go ready 0.113s; Node ready 0.138s; clang ready 0.596s;
markdown dependencies installed, step-duration=0.707s; markdown dependencies
ready 0.981s; submodules ready 20.891s. No build-cache-warm or done line.
Sourced /workspace/adamic-tools/env.sh. Setup fails in its dependency/cache pass:
RootedFilePath/RootedDirectoryPath are missing from the pinned checker; bridge
NewCachedFSCompilerHost takes five arguments where that checker wants six.
Full output is logs/setup.log.

Workaround attempted: temporarily checkout main's cohere 7945d102 and its
TypeScript d92d9bfe, then go build -o /tmp/views-dictionaries-adamic ./cmd/adamic.
Exit 1: this base's loader mixes incompatible APIs, including string versus
RootedFilePath, regexpLibraryFS.ReadFile and sourceFS.AppendFile signatures,
NewParsedCommandLine arguments and missing ComparePathsOptions/ToPath.
Full output is logs/compiler-build-main-pin.log. Both submodules were restored;
no pin update or loader edit is committed. After restoration the compiler build
passed with no diagnostics. The initial failure therefore does not establish a
persistent source incompatibility; its cause remains unproven. The built compiler
was used on every probe in both backends. All six invocations exit 1 and refuse
index-signature syntax, naming the source location and suggesting Map.
probe-observations.json and the .c.log/.js.log files preserve the exact diagnostics.
No successful native/JavaScript dictionary fixture is claimed.

Independent workaround that passed: source the setup environment, then:

```sh
clang -std=c11 -D_POSIX_C_SOURCE=200809L -Wall -Wextra -Werror -pedantic -O2 \
  -I internal/native/runtime internal/native/testdata/records/harness.c \
  internal/native/runtime/*.c -lm -o /tmp/views-dictionaries-records
```

The existing harness and its independent internal/native/testdata/records/oracle.js
match stdout, stderr and exit for semantics, reads, references, iteration and
numeric. stdout sizes are 34,347 / 101 / 46 / 36 / 1,451,860 bytes respectively;
every exit is 0. Results are logs/records-results.json. Full raw outputs remain
in logs/records-*.stdout.log and *.stderr.log locally (ignored to avoid committing
megabytes of unchanged-harness output). This run uses release C; sanitizers,
counts and checked dictionary contracts were not tested by this workaround.

Commands for the new artifacts (each output captured in logs):

```sh
python3 stage3/interface-downcasts/dictionaries/rank-demand.py
python3 stage3/interface-downcasts/dictionaries/rank-demand.py --check
python3 stage3/interface-downcasts/dictionaries/check-probes.py
```

The existing Go record package check was attempted with the restored pins:
`go test ./internal/native -run '^TestRecordsAgainstNode$' -count=1 -timeout 10m`.
It passes: ok internal/native 11.344s. logs/records-package.log preserves its
output. This existing test includes its own sanitizer/count controls; it proves
record representation behavior, not dictionary view admission. The full gate was
not run.
git diff --check passes. No compiler production file was changed.

Dependency observations: lazy-admission 5002bfe0 was fetched and a merge attempted;
it conflicts in cast.go, interface_cast.go, readiness.go, view_objects.go,
native/view_fields.go and the plan. Merge aborted. Callable f13be3e2 was fetched
and a merge attempted; it conflicts in checked-views-plan.md and
checked-views-blockers.md. Merge aborted, rather than overwrite another lane's
content. Array 0b141c26 is already in the base. These are pending merges, not
consumed dependencies. No main or area branch was written and no PR was opened.

Recovered setup rerun: exit 0. Go ready 0.018s; Node ready 0.019s;
submodules ready 0.060s; markdown dependencies skipped with validated bytes,
step-duration=0.006s; markdown ready 0.064s; clang ready 0.258s;
go build ready 24.048s; test binaries deferred 24.149s; build cache warm
24.150s; done 24.177s. nproc=5, cgroup cpu.max=400000 100000.
logs/setup-restored.log preserves every timing line.

The user's coordination change supersedes direct dependency merging. No further
lane branch will be merged. Only codex/views-integration will supply other-lane
code. At the final check, `git ls-remote --heads origin codex/views-integration`
returned exit 0 with no matching ref; the integration branch is not published yet.
The dictionary registry, source probe and indexed read hooks remain blocked.
This checkpoint is ready for the integrator; the worker rests pending that branch
and lazy/read dispatch handoffs, as instructed. It is not a landed/re-greened
integration branch and does not claim a full-family completion.
