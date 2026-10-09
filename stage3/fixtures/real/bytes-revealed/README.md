Built: ten more real tsc extracts, selected in descending attributed-byte order.
Pins: ranking 6c4fc1af; exact negative compiler ed6e2975; accepted f9b004ad files preserved.
Checks: ten Node goldens and ten exact diagnostics pass; seven NotYet and three Refused.
Mutants: one stdout byte, one diagnostic byte, and a contradictory a-check header are caught independently.
Limits: CapturedThis and Path extracts stop at earlier casts; reduced dependencies; no compiler edits or native runs.

# Bytes revealed fixtures

The frozen [ranking selection](ranking-selection.json) comes from
`codex/stage3-hidden-ranking` at `6c4fc1af`,
`stage3/census/hidden-ranking/RESULT.json`. These are attributed hidden bytes
under the ranking's outermost-boundary accounting, not measured bytes newly
compiled by a counterfactual fix. Boundary counts are that ranking's
historical deduplicated counts, not a new corpus census.

| Rank | Attributed bytes | Boundaries | Exact reason / extract |
| ---: | ---: | ---: | --- |
| 1 | 273,934 | 377 | [a value of type __String](escaped-string/README.md) |
| 3 | 120,003 | 5 | [a value of type InitializedVariableDeclaration](initialized-variable/README.md) |
| 4 | 113,387 | 4 | [a computed field name](computed-field/README.md) |
| 6 | 97,864 | 3 | [a value of type PrivateIdentifierInExpression](private-identifier-in/README.md) |
| 7 | 93,640 | 2 | [a function returning ImmediatelyInvokedArrowFunction](immediately-invoked-arrow/README.md) |
| 8 | 89,627 | 2,060 | [a function returning any](any-return/README.md) |
| 9 | 79,336 | 10 | [writeTokenText overload void / number result](overload-void-result/README.md) |
| 10 | 63,225 | 4 | [visitBindingElement overload parameter](overload-binding-parameter/README.md) |
| 11 | 61,355 | 3 | [a value of type ParameterPropertyDeclaration](parameter-property/README.md) |
| 13 | 42,328 | 4 | [transformFunctionBody overload Block / ConciseBody result](overload-body-result/README.md) |

Ranking examples can point to an enclosing transformer body. Each fixture
note and [status entry](status.json) also gives the actual extracted function's
`src/compiler/file:line`. A fresh `stage3/apply.sh` from the `ed6e2975`
checkout produced the source at TypeScript commit
`050880ce59e30b356b686bd3144efe24f875ebc8`; source hashes and complete commands
are in [provenance.json](logs/provenance.json). The existing
[fixture NOTICE](../../NOTICE) applies.

# Exclusions in the ranked prefix

Rank 5, structural calls with statics (101,409 bytes), is already covered by
`f9b004ad`. No selected reason was found to have landed on `ed6e2975`: every
selected ordinary program still stops with its exact ranked reason.

Rank 2, `a function returning CapturedThis` (132,345 bytes), could not be
reproduced exactly by the standalone constructor extract. The real
`createCapturedThis` at `transformers/es2015.ts:825` asserts its unique-name
result as CapturedThis. Its source assertion narrows escapedText to the
literal `__this`; the tested unique-name value for `_this` is `_this`.
Even a control that erases __String's brand stops first at
`a cast the runtime can't check`. The original branded alias was also tried;
constructing its supporting value introduces an earlier unchecked cast.
This is an exclusion due to diagnostic precedence, not evidence of landing
or a claim that all possible extracts are impossible.

Rank 12, `a value of type Path` (60,277 bytes), similarly requires a branded
primitive input in the selected examples. The actual `toPath` function at
`path.ts:762`, retained in the exclusion control, returns the canonicalized
string `as Path`; ordinary lowering refuses that assertion first. A direct
Path-input control was masked by its input cast as well. No fake brand or
unchecked cast was added to a selected fixture to force the ranked spelling.
[excluded-controls.json](logs/excluded-controls.json) retains the standalone
source, Node bytes, and actual compiler diagnostics for these two exclusions.
The search therefore continued to rank 13 to obtain ten exact-reason cases.

# Two compiler pins

`ed6e2975` is a pending compiler landing, not an ancestor of the fetched main.
The same topic branch was refreshed by merging `origin/main` at `3ffb1a83`
in merge `73d27a34`; no existing f9b004ad fixture file was edited.
No compiler feature branch was merged and no compiler source was edited.

The first-line a-check headers describe the topic compiler. Its escaped-string
and writeTokenText extracts still refuse the non-null assertion; the other
eight headers are `checked` because the front end accepts a NotYet stop.
The three frozen overload refusals are not yet that compiler's rules.
The manifest separately records `a_check` and the full `ed6e2975` outcome.
Headers must not be read as the frozen negative goldens. Both pins were tested:
exact frozen diagnostics by the shared fixture harness, topic headers by the
unchanged fast gate's `Gate.aCheck` method. The prior unit's ten fixtures and
manifest remain its original historical measurement and were not remeasured
against the pending compiler landing.

# Reproduction

Create a detached checkout at
`ed6e29751ee47d86fad450cd1674139883bc0f70`, initialize its pinned cohere
submodule, and use the configured cloud toolchain:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/real-fixtures-2-setup.log 2>&1
source /workspace/adamic-tools/env.sh
python3 stage3/fixtures/real/bytes-revealed/check.py /tmp/real-fixtures-ed6e /tmp/real-fixtures-2-verification > /tmp/real-fixtures-2-verification.log 2>&1
```

The script verifies the frozen checkout and uses its unchanged shared harness:
`go test ./stage3/fixtures -run '^TestFixtures$/^real$' -count=1 -timeout 10m -v`.
An isolated scratch manifest places this unit under `real/bytes-revealed`,
preserving the harness's exact diagnostic path normalization. It selects only
these ten new programs. `expected.stdout` supplies the authoritative Node
bytes; the committed manifest mirrors those observations. An accidental
checker error or different earlier refusal fails the full-diagnostic check.
All twenty Node/diagnostic leaves passed; native was not reached.

The output mutant changes only the first byte of the scratch escaped-string
`expected.stdout`, from `_` to `X`. The normal manifest preparation reads that
golden and the Node comparison fails; stage0 passes. The diagnostic mutant
changes only the first byte of `__String` in the scratch diagnostic golden;
stage0 fails while Node passes. Each mutant test exits 1 with exactly one
failing leaf. The proof script exits 0 after confirming both catches.
[results.json](logs/results.json) records the actual commands and exits.

For topic headers, obtain `run.py` from `devtools/fast-gate` at
`bebd5966c6747a9ae107c7e23f56f3a90532bde7`, build the topic compiler, then run:

```sh
go build -buildvcs=false -o /tmp/real-fixtures-2-topic-adamic ./cmd/adamic > /tmp/real-fixtures-2-topic-build.log 2>&1
python3 stage3/fixtures/real/bytes-revealed/check_headers.py /tmp/real-fixtures-fastgate.py /tmp/real-fixtures-2-topic-adamic /tmp/real-fixtures-2-verification > /tmp/real-fixtures-2-headers.log 2>&1
```

This focused invocation reuses that built binary for the method's build step;
its header parsing, classification, matching and failure path run unchanged.
All ten headers pass. A scratch computed-field header changed to
`refused nonexistent-mutant` fails only this expectation check.
[headers.json](logs/headers.json) and
[header-mutant.json](logs/header-mutant.json) retain the observations.

Setup timings: Go 0.024s, Node 0.026s, submodules 0.069s, markdown ready
0.080s, clang 0.167s, build 34.505s, tests deferred 34.614s, cache warm
34.616s, done 34.648s. `nproc=5`, cgroup CPU quota four. The detached compiler
build first failed to obtain VCS status for the shared submodule checkout;
`-buildvcs=false` fixed it. Both build logs are retained. The setup succeeded.
The full adaptation output is retained as `logs/apply.log.gz`.

These are negative stage3 fixtures, with trimmed AST views and dependency
helpers described in each note. They do not execute complete tsc or prove
native behavior. No counted oracle fixture was added, so
`internal/oracle/counts.md` does not change. No whole package confirmation,
full gate, or corpus census was run; the shared fixture tests are this unit's
requested measurement. All test output went to retained log files.
