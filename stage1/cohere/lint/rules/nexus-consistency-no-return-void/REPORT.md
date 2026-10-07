Built two .a rule implementations; skipped the existing no-for-in port. Integration is blocked.
Commits: claim 69e2e1a; delay 8e8acca; return-void 865a747; foundation merges 48cd934 and 33b9c9e.
Checks: scratch-overlay corpus/mutant suite PASS in 144.313s, 24,710,680 matching canonical bytes.
Mutants: wrong resolve name and omitted braces both caught only by Go comparison on all three backends.
Not covered: ordinary .a registration, one JSX case, one top-level-await case, cohere CLI source lint, full gate.

## Scope and ownership

Positions 37, 38 and 39 are obtained by concatenating the 30 option-ready rules
and 16 policy-ready rules in `../../HELPERS.md`, referenced by
`../../helpers/REPORT.md`. The report itself does not reproduce an ordered list.

- `nexus/consistency-no-for-in` was already ported on
  `origin/codex/stage1-lint-batch4`, in `nexus_consistency_no_for_in.ts`; skipped.
- `nexus/consistency-no-hand-rolled-delay` implements the upstream executor,
  timer, callback, reject-reference and primitive-definition decisions, with no fix.
- `nexus/consistency-no-return-void` implements findings and guarded fixes,
  including indentation, statement lists and braces around conditional/loop bodies.

The two rule directories own descriptors, `.a` modules, Go overlay adapters,
raw witnesses and semantic mutants. Policy descriptions are copied from the
helper's resolved Go catalog; their complete bytes are compared to the real
Go policy renderer. These rules have no rule options to decode. No compiler,
parser, runtime, submodule or shared dispatch source is changed by implementation.

`docs/parallel-work.md` is absent from the fetched main and both foundations.
The available contract is `docs/lint-registration.md` and the merged CLAUDE.md.

The branch began at fetched main `d090af5`. The checkout's fetch refspec initially
included only main; fetching all branch refs made the foundation refs available.
Registration merged cleanly. Helpers conflicted in README, lint.ts, lint_test.go,
main.ts, settings.ts and testdata/oracle.go. Those entry points retain the
registration side, while incoming helper/inventory artifacts are present.
The ownership claim was pushed before either implementation.

## Evidence and exact commands

Toolchain commands were `bash cloud/setup.sh`, then
`source /workspace/adamic-tools/env.sh`. Both setup attempts exited 1 during
cache warming. The first overlapped the merge and reported missing import
artifacts for regexp and registry plus undefined ownedWitnesses. That overlap
is observed; a build/merge race is an inference, not a proven diagnosis.
The retry's exact blocker is:

```
profile_test.go:32:23: cannot range over portFiles (value of type func(t *testing.T) []string)
```

First timing lines: Go ready 1s; clang ready 1s; Node ready 1s; submodules ready
1s. Retry: all four ready 0s. Neither run emitted a successful build-cache-warm
or setup-done timing line. Both complete logs are in `evidence/`.
`nproc` printed 5; cgroup `cpu.max` is `400000 100000`.
Go 1.27.1, clang 20.1.8, Node 24.19.0; Linux x86_64, AMD EPYC 9V74.

The narrow workaround is a scratch Go overlay, not a shared-file change.
`evidence/required-infrastructure.patch` makes registry discovery/rendering
accept `rule.a`, allows `.a` mutant files, copies `.a` modules in the test
harness and fixes the stale profiling test's portFiles call. `git apply --check`
on that patch exited 0. Registry checks under the overlay passed in 0.022s.
Your territory restriction excludes applying these changes to shared files;
the scope clarification is still unanswered. The patch is reviewable and unapplied.

All test commands wrote directly to log files, never to a pipe:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
  go test -overlay=/tmp/lint-wave1-13/overlay.json ./stage1/cohere/lint \
  -run '^TestWave13' -count=1 -v -timeout=20m \
  > /tmp/lint-wave1-13/full-corpus-retry.log 2>&1

go test -overlay=/tmp/lint-wave1-13/overlay.json ./stage1/cohere/lint \
  -run '^TestWave13JsxGap$' -count=1 -v -timeout=10m \
  > /tmp/lint-wave1-13/jsx-gap.log 2>&1

go test -overlay=/tmp/lint-wave1-13/overlay.json ./stage1/cohere/lint/registry \
  -count=1 -v > /tmp/lint-wave1-13/registry.log 2>&1

go vet -overlay=/tmp/lint-wave1-13/overlay.json ./... \
  > /tmp/lint-wave1-13/vet.log 2>&1

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m \
  > /tmp/lint-wave1-13/oracle.log 2>&1
```

Corpus/mutants: PASS, 144.313s. JSX gap proof: PASS, 22.767s.
Vet: exit 0, empty log. Filtered uncached oracle: PASS, 4.930s,
native and Node cache hits 0, misses 1 each. The owned-witness comparison
also passed: 9,782 canonical bytes, in 12.71s. Its surrounding initial
TestRulesAgree run failed on the JSX replay and is retained as failed evidence.
Gofmt on the two adapters and source diff checks printed nothing.

The external corpus is TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`: all 77 src/compiler files.
All 135 current stage1 `.ts` and `.a` files are included, 212 files total.
The real Go rule tests ran through an assertion-capture overlay, with no Go
rule/body/message/fixer changes. There are 48 distinct delay cases and 23
return-void cases; one of each is an explicit adapter gap. The other 47 delay
and 22 return-void cases compare findings, message IDs, descriptions, byte
ranges, fix categories/replacements and complete fixed sources.

Source Node and ASan/UBSan native matched Go on 12,358,833 delay bytes and
12,351,847 return-void bytes. Emitted JavaScript, run through the external
Node runtime loader, matched those same outputs. Successful baseline and mutant
processes must exit 0 with empty stderr; sanitizer/compile/panic failures are
never credited as caught semantic mutants.

`evidence/reproduce.py --compiler /path/to/typescript-6.0.3` reconstructs the
scratch overlays from the saved Go snapshots and runs the bounded comparisons,
owned witnesses, registry checks and vet. Source the toolchain environment
first. The script's syntax and repository path were checked; the recorded
original commands above produced the evidence. The complete replay script
itself was not rerun after its final zero-test control assertion was added.

## Mutants

| Mutation | What the independent comparison caught |
|---|---|
| Delay resolves(callback, resolve) becomes resolves(callback, 'wrongResolve') | missing handRolledDelay finding |
| Return-void braceless body loses its braces | replacement `stop(); return;` differs from `{ stop(); return; }`, and fixed source differs |

Both temporary variants compile and run successfully on source Node, emitted
JavaScript and sanitized native before comparison. Both are caught on all three
sides, only by the Go-output comparison. The production sources were never
mutated. These controls use the owned raw witnesses.

## Throughput

Best of three interleaved Go/native/Node rounds over each rule's supported
upstream cases plus the 212 source files. Includes process startup, file reads,
scanning, parsing and visitation; excludes build time and output/fix formatting.
Native throughput uses an unsanitized release build; correctness uses sanitizers.
These are mixed-corpus rates with sparse findings, not isolated visitor rates.

| Rule | Findings | Native findings/s | Node findings/s | Go findings/s |
|---|---:|---:|---:|---:|
| hand-rolled-delay | 21 | 17.07 | 25.16 | 99.00 |
| return-void | 18 | 14.35 | 22.49 | 84.29 |

All count outputs matched. Every sample and duration is preserved in
`evidence/corpus-and-mutants.log`.

## Explicit blockers and limits

1. Normal registry generation exits 1 because it insists on `rule.ts`:
   `open .../nexus-consistency-no-hand-rolled-delay/rule.ts: no such file or directory`.
   New authored Adamic modules are `.a`, as instructed. The scratch extension
   patch is necessary for the experimental green runs and remains unapplied.
2. The upstream top-level `await new Promise<void>(...)` fixture is accepted
   by Go and refused identically by source Node and sanitized native, exit 70,
   `parser slice expected semicolon at 6`. It is not counted as parity.
3. The upstream clean JSX callback fixture passes the actual Go cohere test.
   Source Node and sanitized native both refuse it, exit 70,
   `expected GreaterThanToken, got Identifier at 24`. The inherited capture
   harness also replays JSX as plain TS. This fixture is not counted as parity.
   Both gap witnesses are saved under the corresponding owned `gaps/` directories.
4. The pinned cohere CLI dry run against the four `.a` modules exits 1:
   `nothing to check: none of the named paths is in the program`, followed by
   `.a is not a TypeScript or JavaScript file`. This is a failed self-lint
   observation, never a clean lint result. Native/emitted-JS builds did typecheck
   and lower the `.a` modules successfully through Adamic's own loader.
5. No full repository gate or all inherited lint tests were run to green.
   The foundations also bring stale scanner profiling tests. The branch is
   not merge-ready without shared registration/test integration and parser work.
   No arbitrary malformed-source, JSX, general config/suppression, checker-backed
   behavior or overlapping-fix contract beyond the inherited driver is claimed.

Early dependency-download stderr, unsupported two-argument lastIndexOf, JSX
replay failure, top-level-await refusal and missing emitted-JS runtime loader
failures are retained in the early evidence logs. The final working return-void
implementation uses slice-then-lastIndexOf. Final emitted JS runs through
oracle/node.mjs so the adamic runtime import resolves.
