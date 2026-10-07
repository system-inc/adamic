# Date 3 report

Integrated the TS-validity runner from origin/codex/test262-ts-validity (af12899)
with the Date adaptations already on codex/library-date-refusals (4ec07a3).
The claim was the first Date-3 commit, f58c79a. Runner commits 979df68 and af12899
were cherry-picked as 99bd28c and f2ccdb7; the large unrelated integration history
was not merged. New runner fixture sources use .a.

No compiler, Date runtime, JavaScript backend, nullable-string special case or
oracle fixture changed. No new library method family was needed after triage.
52 tests recover their existing Date adaptations in the standalone TS-validity
runner comparison. These are not 52 newly implemented Date behaviors: the
refusals branch's own adapted runner already passed 163 before this unit.
All newly passing programs agree with Node, and no earlier pass was lost.

## Tables

Both runs use compiler/runtime 4ec07a3, --adapt, stock tsc 6.0.3 and test262
3fd3eab12309bd7732f4b5ddeaae19c5d95ad9dd. Directory tables and every reason are in
before.json/after.json; full runner output is in before.log/after.log.

| Runner | Pass | Disagreements | Refused | Not TypeScript | Crashed | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|---:|
| Before: af12899 standalone | 111 | 0 | 90 | 123 | 0 | 270 | 594 |
| After: TS validity plus existing Date adaptations | 163 | 0 | 54 | 107 | 0 | 270 | 594 |

The supplied 201-refusal inventory does not reproduce on the requested branch
and pinned corpus. The 42 Date prototype descriptor tests and UTC descriptor
observations were already implemented on 4ec07a3. The largest standalone runner
refusal was 43 untyped Date result locals; preserving the existing checked,
erasable annotation adapter handles all 43. The remaining 47 originally valid
refusals remain refused; 16 old harness TypeScript rejections move to nine passes
and seven language refusals once existing harness adaptation applies.

The 54 remaining refusals are language work or the intentionally refused clock.
GAPS.md has one-line .a reproducers and category counts. Every actual matching
stock TypeScript rejection is left in not-typescript. No widening of library
argument types, reflective prototype implementation, detached-method implementation
or native exception change was made. Harness includes and other skips remain out
of scope. No new code uses dateNullableString.

## UTC and setup

The reference runner is launched with TZ=UTC; its children inherit it. The
integrated runner's existing runCommandWithLimit also appends TZ=UTC to every
compiler, native and Node subprocess. Thus parsing and local-time methods use
UTC on both sides. The Date oracle's existing package init pins TZ=UTC before
parallel tests. No ambient local timezone is used to judge native behavior.

cloud/setup.sh:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (97s)
setup: done in 97s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

Go 1.27.1, clang 20.1.8, Node v24.19.0. Stock tsc was already installed under
/tmp/date-refusals-tsc/node_modules/.bin and was added to PATH. The isolated runner
checkout uses the same cohere submodule through a scratch symlink; -buildvcs=false
was needed only for that isolated build because its submodule git metadata points
to the primary checkout. Setup itself succeeded.

## Commands and verification

All test output was redirected to logs, never piped.

```
# In an isolated worktree at origin/codex/test262-ts-validity:
go build -buildvcs=false -o /tmp/date3-runner ./cmd/adamic-test262
TZ=UTC /tmp/date3-runner -adapt -json -root /workspace/adamic \
  -test262 /workspace/test262 -work /tmp/date3-before-work built-ins/Date
# In the Date-3 checkout:
go build -o /tmp/date3-runner-integrated ./cmd/adamic-test262
TZ=UTC /tmp/date3-runner-integrated -adapt -json -root /workspace/adamic \
  -test262 /workspace/test262 -work /tmp/date3-after-work built-ins/Date
TZ=UTC go test ./cmd/adamic-test262 -count=1
TZ=UTC ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestDateOracleCatchesMutants$' -count=1 -v
TZ=UTC go test ./cmd/adamic-test262 -run '^TestTypescriptControls$' -count=1 -v
```

Complete runner package: PASS, 46.002s (runner-tests-final.log).
Date Node-only mutants: PASS, 3.128s; 18 mutants, fresh Node executions and fresh
sanitized mutant binaries (mutants.log). Restored TS controls also pass, 3.356s (typescript-restored.log).
The first runner test attempt failed because fixture files were renamed to .a
while that run still read .ts; the completed rename and clean rerun pass. That
failed log is retained as runner-tests-initial.log.

The full compiler gate was not run: only runner integration and fixture extensions
changed. Existing Date fixture/count registrations are unchanged; all 18 family
mutants plus every passing test262 Date program are compared with Node. No new
allocation counts are implied. The audit classifies all 324 attempted Date programs and runs stock tsc on
every compiler-refused source. Seven independent language probes pass stock tsc and Node while
Adamic refuses (gap-proofs.json).

## Every mutant run

The existing TestDateOracleCatchesMutants suite was run unchanged. Every mutant
below remains valid C, exits 0, has empty stderr and no sanitizer error. Only the
comparison with the original source on Node catches it: stdout differs.

| Mutant | Wrong answer |
|---|---|
| number_date | Add 1 to Date-to-Number conversion |
| metadata_name | Date name becomes Dote |
| metadata_length | Constructor length becomes 8 |
| metadata_own | Invert intrinsic hasOwnProperty result |
| nullable_typeof | Null reports undefined |
| nullable_number | Null converts to NaN |
| nullable_stringify | Use nonnullable JSON string conversion |
| date_stringify | Use map stringify for Date |
| toJSON | Invalid Date returns Invalid Date string instead of null |
| dynamic_parse | Add 1 ms to parsing |
| constructor_clip | Add 1 ms before construction |
| UTC | Add 1 ms to Date.UTC result |
| invalid_NaN | Getter NaN becomes 0 |
| getters | Add 1 to getter result |
| setters | Add 1 ms to setter return |
| iso_format | Change penultimate ISO digit |
| format | Shift formatted Date by one day |
| parse | Add 1 ms to parsing |

No nullable-string behavior was added; the nullable mutants are inherited checks,
run only to confirm existing comparison evidence.

An additional TS-validity mutant changed the matching-code check from
`candidate == code` to `candidate != ""`. TestTypescriptControls/disagree failed:
it incorrectly classified Adamic TS2322 versus stock TS2304 as not-typescript.
Exit 1, 3.923s (typescript-mutant.log). This check is caught by the independent
stock-TypeScript oracle, not by Node. The mutation was restored and the same
controls rerun successfully; no mutant remains in the branch.

## Commit sequence

- f58c79a: first claim and triage, before integration.
- 99bd28c: imported TS-validity runner commit.
- f2ccdb7: imported output-overflow crash preservation.
- Final review commit: .a fixture renames, verified gap reproducers, before/after
  artifacts and this report; its SHA is in the final response.
