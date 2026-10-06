Numeric process exit, terminal observation, and read-only environment lookup are implemented and pushed.
The output port now renders human, verbose findings, NDJSON, summaries, crashes, and the older report.
Renderer comparisons use unmodified pinned Go cohere, Node source, and sanitized native output.
Live rules and fixes are compared on generated fixtures and the pinned compiler/repository corpus.
The complete live CLI contract remains open: clocks, automatic character-device color, and verbose accounting.

Compiler work: branch `codex/process-exit-and-tty`, from origin/main
`5d4c8012a0877094134e6c6bac367ff68f9313e8`; commits `fd2ef77` and `3a9d1fc`, pushed.
See [the compiler report](../../../docs/process-report.md) for commands, outputs, five
executable mutants, and limits. This lint branch merges it at `648b7de` on top of the
previous scanner-based diagnosis, without rewriting either branch.

Setup: `bash cloud/setup.sh`, output `/tmp/process-exit-setup.log`. Timing lines:
Go 0s, clang 0s, Node 0s, submodules 0s, build cache warm 83s, done in 83s.
`nproc` printed 5; cgroup CPU quota is 4. Source `/workspace/adamic-tools/env.sh`
before each command. Go 1.27.1, clang 20.1.8, Node 24.19.0.

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (83s)
setup: done in 83s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

## Implemented output contracts

The `output` directory contains a reusable rendering layer, with explicit run facts
as input. It is not a replacement command executable.

- `model.ts`: complete command summary DTO, failure/exit-1 decision, stable finding
  ordering by UTF-8 path, byte start/end, rule, message ID, and raw description.
- `render.ts`: human findings with or without ANSI colors, verbose finding lines,
  NDJSON finding/fixed/formatted/summary/crash records. JSON preserves Go's field
  order, omitted fields, HTML escaping, U+2028/U+2029 escaping, and raw messages
  when the finding DTO itself contains LF. Live rule conversion separately collapses
  LF, exactly where the Go command does it.
- `footer.ts`: human and verbose summary lines, phase timings, formatting time,
  cache/replay facts, coverage, readiness, and unchecked markers. Counts group
  thousands. Exact binary midpoint rounding agrees with Go fmt's ties-to-even.
- `other.ts`: multi-project footer, human crash diagnostics, older grouped lint
  report and coverage footer, and the checked-nothing line. The old report's stable
  path/start sort and LF-only byte positions remain distinct from the command's
  ECMAScript line and byte positions.
- `location.ts`: command positions account for CR, CRLF, LF, U+2028 and U+2029.
- `corpus.ts`: runs the same thirty ported rules and fix convergence, then renders
  original findings under no-fix or remaining findings after fixes. Fixes apply to
  scratch strings, never the repository or TypeScript checkout. `Linter.fixesByRule`
  records actual applied fixes for changed-file output.

The pin is Go cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`. It ships human,
verbose, and NDJSON command modes. It does not ship a GitHub annotation format or
SARIF renderer: `--github` exits 2 as an unknown flag. No invented annotation
format is represented as an upstream port. Historical reference observations and
the complete source inventory remain in [OUTPUT_GAPS.md](OUTPUT_GAPS.md).

## How parity is held

`output_test.go` overlays observation-only tests into the unmodified Go command
package. Actual Go renderers consume run facts that are exported as TypeScript
input. Source Node and native under ASan/UBSan must produce those exact bytes,
empty stderr, and the same chosen exit code. Native finishes normally using
`process.exitCode`, so LeakSanitizer runs even when the verdict is failure.

Twelve synthetic summaries cover every footer gap, cache/readiness branches,
changed-file glyphs/alignment/capping, non-ASCII ordering, and rounding boundaries.
Each is tested in three modes, two color states, and two phase-display states on
Node and native: 288 observations. Auxiliary cases exercise the actual old Go
report, multi-project verdicts, whole-file/rule crashes, missing/invalid source
positions, singular/plural coverage and exact midpoint duration rounding.

The live rule test uses twenty-four generated manifest rows. The whole-corpus
test pins TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8` and
visits its `src/compiler` plus this repository's `.ts` and `.a` sources, excluding
the Go cohere vendor checkout. Both fix states render all twelve mode/color/phase
combinations. Each rules run now renders every combination in one execution,
avoiding twelve repeated scans of the large corpus.

Go's actual measured time is supplied unchanged to both renderers. It is not
rounded away, normalized, replaced with zero, or presented as an independently
measured native time. This proves rendering the same facts; it cannot prove byte
equality of two independently measured executions. The historical evidence
already shows two identical Go commands produce different JSON timing bytes.

## Mutants and remaining gaps

All three `TestOutputRendererMutants` compile with normal warning flags and
sanitizers, run cleanly, and agree between source Node and native. Only comparison
with unmodified Go output catches them:

| Mutant | Comparison that catches it |
| --- | --- |
| Remove JSON `<` escaping | NDJSON bytes differ |
| Compare UTF-16 units instead of Unicode scalar values | Human file order differs for U+E000 and U+10000 |
| Remove ties-to-even correction | Human footer for 10.5 seconds differs |

The three earlier proposed-workaround mutants remain executable: discarding the
chosen exit code, assuming NO_COLOR absent, and replacing a clock with zero.
The first two original proving programs now agree with Node. The clock original
still reports TS2304, `Cannot find name 'performance'`, whereas Node prints
`clock available`. No silent clock fallback was added.

Automatic color is deliberately limited: the helper uses Node's actual
`process.stdout.isTTY` and the presence of NO_COLOR, including an empty value.
Pinned Go cohere instead tests `os.ModeCharDevice`, which includes `/dev/null`.
Those predicates are different. FORCE_COLOR is readable but Go at this pin does
not consult it. The rendering tests supply the same explicit color fact to both
renderers; they do not claim every possible device has matching automatic policy.

The full CLI pipeline has not been ported: configuration resolution, severity and
suppression application, formatter execution, multi-project discovery, cache
acquisition, invocation envelopes, profiling/progress tables, verbose phase and
wall-clock accounting text remain outside the implemented rendering driver.
Synthetic facts cover formatted-file records, type diagnostics and crashes;
the live corpus uses only the thirty already ported syntax rules at error
severity, without a type checker or formatter. Full independent CLI output
parity, particularly raw JSON timing fields, is not claimed.

## Validation

Logs stay in `/tmp`, rather than being added as source files. Final commands:

```
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-lint-typescript \
go test -v -count=1 -timeout 30m ./stage1/cohere/lint -run '^TestOutput' \
  > /tmp/lint-output-final-stable.log 2>&1
go test -v -count=1 -timeout 10m ./stage1/cohere/lint -run '^TestRulesAgree$' \
  > /tmp/lint-output-rules-final.log 2>&1
go vet ./... > /tmp/lint-output-vet-stable.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/output_test.go \
  stage1/cohere/lint/output_gaps_test.go \
  stage1/cohere/lint/testdata/output_oracle_test.go \
  stage1/cohere/lint/testdata/corpus_output_oracle_test.go
git diff --check
```

The rule regression exited 0 in 36.138s: 1,272 unique upstream cases,
749,959 identical bytes from Go, Node and sanitized native, with malformed-parser
recovery explicitly reported rather than hidden. Vet exited 0 with no output;
formatting and diff checks printed nothing.

The final output command exited 0: `PASS`, package elapsed 308.565s. Its results:

| Check | Observed result |
| --- | --- |
| Closed exit/environment probes and remaining clock gap | PASS, 0.70s |
| Three proposed-workaround mutants | All caught, PASS, 1.03s |
| Twenty-four live generated rule inputs, fix and no-fix, twelve views each | PASS, 47.80s |
| TypeScript compiler and repository, 541 files, fix and no-fix, twelve views each | PASS, 208.10s |
| Three executable renderer mutants | All caught only by Go output comparison, PASS, 11.61s |
| Twelve synthetic summaries, 288 Node/native observations | PASS, byte for byte, 32.82s |
| Six auxiliary cases, both color states, Node/native | PASS, 6.51s |

The whole corpus reports 17,684 findings and 36 files that would change under
no-fix. With fixes applied to scratch text it reports 15,413 remaining findings,
36 fixed files, zero formatted files, and zero files that would change. Both
verdicts exit 1. The timing values in those summaries are the preserved Go
observations described above. Node and sanitized native reproduce every supplied
format's output, including the ANSI bytes. The synthetic matrix also exercises
successful exit-0 and failing exit-1 summaries; the whole-corpus cases themselves
both fail because findings remain.

Earlier attempts are retained in scratch logs: `/tmp/lint-output-whole-first.log`
was stopped because it repeated the entire rules scan per output combination.
`/tmp/lint-output-final.log` then exceeded the one-minute driver budget; it also
observed source-location changes when a renderer source was edited during its
corpus scan. The final run uses a three-minute driver budget, prints every output
combination after one scan, and leaves all source inputs unchanged. The auxiliary
Go fixture initially used a relative source filename, which its parser correctly
rejected; `/tmp/lint-output-auxiliary-fix.log` records that rejection. Giving the
fixture an absolute path corrected it. `/tmp/lint-output-auxiliary-fixed.log`
then passed all five initial auxiliary cases in 11.066s; the final run also adds
a negative-duration rounding case.

The compiler's complete gate was attempted before returning to this unit. It
failed only the new process IR node's registration in `internal/fresh`. After
registration, that whole package passed; the uncached process oracle also passed.
The 14-minute full gate was not repeated, and no complete green gate is claimed.
Exact compiler commands, package results, and the five process mutants are in
[docs/process-report.md](../../../docs/process-report.md).
