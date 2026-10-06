The exit and environment gaps are closed by codex/process-exit-and-tty, merged into this branch.
The probes now use the implemented global prelude API rather than local ambient declarations.
Human, verbose finding, NDJSON, and summary renderers now have native parity tests.
The native clock and the full live CLI/corpus contract remain separate coverage limits.
The report below records the earlier blocked baseline; OUTPUT_REPORT.md records the resumed work.

# Historical baseline diagnosis

# Baseline and scope

Branch: `codex/stage1-lint-output`, from
`origin/codex/typescript-scanner` at `0090256e607c3f2de7d5b67cef680ec010f95c1d`.
Pinned cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.

`git fetch origin` updated only main in this checkout. The first branch creation
failed because the scanner remote-tracking ref was absent. `git ls-remote`
confirmed the remote branch, and an explicit fetch of its ref created it.
No compiler, runtime, oracle, or submodule file was changed.

The user allows stopping when a unit cannot be completed. That applies to the
whole requested native output contract here: its exit status cannot be delivered
through the current Adamic runtime API. The deterministic rendering functions
could be ported separately, but that would not complete this unit's exit,
automatic color, and summary requirements. This change records executable gaps
instead of providing a renderer with those requirements silently weakened.

# Observations

| Proving program | Node observation | Stage 0 observation |
| --- | --- | --- |
| `gaps/6_output_exit_status.ts` | `one finding\n`, exit 1, empty stderr | `lower.NotYet`: `a call returning never`, line 6, column 1 |
| `gaps/7_output_environment.ts`, with `NO_COLOR=` | `color disabled\n`, exit 0, empty stderr | Compiles; sanitized native prints nothing and exits 1, with `runtime error: member access within null pointer of type 'const adamic_object'` |
| `gaps/8_output_clock.ts` | `clock available\n`, exit 0, empty stderr | `load.CheckError`: TS2304, `Cannot find name 'performance'`, line 1, column 17 |

The first two declare only the types of Node's existing `process` object.
They do not supply an implementation. The environment probe exposes an unsafe
ambient-global path, not a supported way to access the native environment.
The generated C initializes `adamic_global_0_process` to NULL and reads its
`env` property. Unsanitized native also crashed, with shell exit 139. No claim
is made that these ambient declarations are intended Adamic runtime facilities.

`internal/load/prelude.d.ts` and `oracle/adamic.mjs` expose files, directory
listing, status, arguments, UTF-8 access, and panic. They expose no caller-chosen
exit status, environment access, terminal detection, or monotonic clock.
`docs/0.1.md` specifies completion as exit 0 and panic as exit 70. Substituting
panic for cohere's failing verdict therefore cannot preserve the contract.

Direct Go observations are preserved under `output_evidence/`. On the same
two-file fixture, two uncached JSON lint runs printed identical findings but
summary `seconds` values of `0.078603435` and `0.089552409`, and `graphSeconds`
values of `0.023631006` and `0.030187572`. Both exited 1. Even Go versus itself
does not satisfy byte equality of the independently measured summary.

Human and verbose lint runs also exited 1. `--no-fix --no-format --json`
reported three findings and two files that would change, exiting 1. The default
fixing run with `--no-format --json` removed all three debugger statements,
reported two fixed files, zero remaining findings, and exited 0. `--github`
was rejected with `flag provided but not defined: -github`, exit 2.

The clock program checks availability instead of asserting a particular elapsed
time. It does not pretend that independent runs have identical timing.

# Mutants

`TestOutputWorkaroundMutants` changes scratch copies of the proving programs.
Each mutant must compile and finish with exit 0 and empty stderr on both Node
and native under ASan/UBSan. Both sides must agree on the mutated program.
Then stdout, stderr, and exit are compared with Node running the original.

| Mutant | Native observation | What catches it |
| --- | --- | --- |
| Remove `process.exit(1)` | `one finding\n`, exit 0 | Original Node exits 1; exit comparison |
| Replace the environment test with `true` | `color permitted\n`, exit 0 | Original Node prints `color disabled\n` for an empty, present `NO_COLOR`; stdout comparison |
| Replace `performance.now()` with `0` | `clock absent\n`, exit 0 | Original Node prints `clock available\n`; stdout comparison |

These are three rejected workaround mutants, not mutants of a completed output
renderer. They prove that dropping the unavailable inputs changes observations.
`TestOutputRuntimeGaps` separately pins the current rejection or sanitizer
failure of each original program so a closed gap requires updating this report.

# Go output inventory

Source inspection of the pinned cohere found these contracts:

- `command/cohere/output_mode.go`: default human, verbose, and newline-delimited
  JSON. `main.go` accepts `--verbose`, `--json`, and `--phases`; verbose and JSON
  together are refused. `--phases` is an additional footer presentation.
- `human_output.go`: findings are `path:line:column severity rule message`.
  Error severity is red, other severity and rule are dim. `NO_COLOR` presence,
  including an empty value, disables color; otherwise a character device enables
  it. Changed files sort by path, show at most 20, align paths by rune count, and
  sort fixing rules by descending count then name.
- `main.go:printRuleDiagnostic`: verbose findings are
  `path:line:column - message [rule/messageId]`. Embedded LF becomes a space.
  Positions use `scanner.GetECMALineAndByteOffsetOfPosition`, not the LF-only
  location computation in the old lint corpus driver.
- `json_output.go`: NDJSON includes finding, fixed, formatted, and summary
  records. Fix rule names sort lexically. JSON uses Go's `encoding/json` byte
  contract, including its escaping. Summaries carry elapsed and graph seconds,
  phases, cache facts, coverage, gaps, and Adamic readiness. Crash records are
  rendered by the command's crash machinery.
- `footer.go`: human and verbose summaries have counts, timing, rewritten files,
  optional phase times, readiness, and unchecked markers. A multi-project run
  adds an overall footer. `profile.go` ends the actual process with the supplied
  exit code; findings and crashes take the exit-1 path in `main.go`.
- `internal/types/program/walk.go:sortDiagnostics`: the command's stable order
  is file path, start, end, rule name, message ID, then description. This differs
  from the existing port's start-only ordering at ties.
- `internal/lint/report/report.go`: the older report view groups by file path,
  then start, stably; prints a location line, an indented rule/message line, and
  a blank line; ends in verdict plus coverage. Its positions count LF only.
  Coverage distinguishes graph warm/cold and formats milliseconds below one
  second, two-decimal seconds otherwise. The existing lint oracle strips this
  footer and is not a command-output oracle.

Searches for `::error`, `::warning`, GitHub annotations, SARIF output, and a
GitHub output flag found no such lint renderer in the pinned Go source.
Inventing one would not be a port byte for byte to that pin.

# Inferences and requirements to resume

Completing this unit needs runtime facilities for exit status, environment
presence, and terminal detection. A clock is needed if the native command itself
owns its timing observations. These additions belong to the compiler/runtime
owners, outside this unit's permitted files.

A separate issue is the comparison contract: Go summaries contain measured
wall-clock durations (`time.Since` and phase elapsed times), including raw numeric
seconds in JSON. Independently executing Go and native cannot guarantee those
bytes agree, even when both correctly render their own timings. To prove the
renderer byte for byte, feed both the same captured run summary, then test actual
clock measurement separately. Normalizing or omitting timing without saying so
would weaken the user's requirement. That change was not assumed here.

# Validation and limits

Setup completed successfully with Go 1.27.1, clang 20.1.8, and Node 24.19.0:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (24s)
setup: done in 24s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed `5`. Source `/workspace/adamic-tools/env.sh` for each command.

The initial focused run was:

```
go test -v -count=1 -timeout 10m ./stage1/cohere/lint \
  -run '^TestOutput(RuntimeGaps|WorkaroundMutants)$' \
  > /tmp/stage1-lint-output-gaps.log 2>&1
```

Exit 0, all six subtests passed; package elapsed 15.134s.
Every mutant finished cleanly under the sanitizers and was caught by the exit or
stdout comparison shown above. The environment original was caught by UBSan.

No TypeScript `src/compiler` or whole-repository output comparison was performed.
No fix/no-fix output matrix, formatter output, suppression, severity resolution,
colored terminal parity, renderer implementation, full summary implementation,
or native exit-1 implementation is claimed. The full gate was not run: this
change adds evidence and tests for blocked facilities, not compiler behavior.

Additional completed checks, each with output directed to a file:

```
go vet ./... > /tmp/stage1-lint-output-vet.log 2>&1
go test -v -count=1 -timeout 10m ./stage1/cohere/lint \
  -run '^Test(RulesAgree|OutputRuntimeGaps|OutputWorkaroundMutants)$' \
  > /tmp/stage1-lint-output-check.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 10m ./internal/oracle \
  -run '^TestOneFileHoldsNodesOrder$' \
  > /tmp/stage1-lint-output-oracle.log 2>&1
```

All exited 0. Vet printed nothing. The lint package completed in 162.978s:
1,272 unique upstream cases, explicit malformed-parser limits, and 751,320
identical bytes from Go, Node, and sanitized native for the existing lint
protocol. This is regression validation, not proof of the requested new output
layer. The filtered uncached oracle passed its `interleaved.a` and
`status_of_stdout.a` cases in 4.101s. See `output_evidence/checks.log` and
`output_evidence/oracle.log` for the complete outputs.

Inside the unmodified cohere checkout:

```
go build -o /tmp/stage1-go-cohere ./command/cohere \
  > /tmp/stage1-go-cohere-build.log 2>&1
go test -v -count=1 -timeout 10m ./command/cohere \
  -run '^Test(FindingLine|StyleForHonorsNoColor|StyleForAFileIsPlain|ChangedFileLinesAlignMixedGlyphs|ChangedFileLinesAreCapped|FooterGolden|FooterColor|JSONLines|JSONSummary)$' \
  > /tmp/stage1-output-go-unit-tests.log 2>&1
```

Both exited 0; the selected Go output tests completed in 0.021s. This validates
the inspected upstream contracts; it does not validate a native renderer.
`output_evidence/go-output-tests.log` preserves the results, including colored
footer checks.

The direct reference fixture was `/tmp/stage1-lint-output-reference`, with:

```
tsconfig.json:
{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["*.ts"]}
CohereSettings.json:
{"rules":{"no-debugger":"error"}}
a.ts: debugger;\ndebugger;\n
b.ts: debugger;\n
```

The three initial views used
`/tmp/stage1-go-cohere --lint --single-threaded --no-cache --directory
/tmp/stage1-lint-output-reference`, with no additional flag for human,
`--verbose` for verbose, and `--json` for JSON. JSON was run twice before
mutating any fixture. Each command redirected stdout and stderr to its log.
The no-fix command was
`/tmp/stage1-go-cohere --no-fix --no-format --single-threaded --no-cache
--directory /tmp/stage1-lint-output-reference --json`; the fixing command was
the same with `--no-fix` removed. Full output bytes are in the corresponding
`output_evidence/go-*.log` files. These are tiny Go reference observations,
not the user's large-corpus comparison.

Source formatting and `git diff --check` excluding `output_evidence` were clean.
The complete diff check reports trailing whitespace in literal upstream output
and spaces before tabs in Go's flag usage text. Those logs retain the observed
bytes. One initial `gofmt` invocation without sourcing the toolchain failed with
`command not found`; sourcing `/workspace/adamic-tools/env.sh` resolved it.

The final source reran the focused six subtests with the same command as the
initial focused run, writing `/tmp/stage1-lint-output-final-gaps.log`. Exit 0,
package elapsed 14.938s. `output_evidence/final-gaps.log` preserves that run.
