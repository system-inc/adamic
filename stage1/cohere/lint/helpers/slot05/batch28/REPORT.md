Built bareValuePredicate, bareValueTransform and isValidOpacityValue in separate .a files; nine prerequisites across three blocked rules, no new helper-ready rule.
Claim 38b319ab5 published and success collected before source; implementation SHA is in the final delivery report. Current main 71d7e491 and lint area 7076b4eb remain ancestors.
Final owned suite PASS 52.565s over 37575 Go/source Node/emitted JavaScript/sanitized native comparisons; Node probes PASS 1.067s, vet/format clean; setup 25.787s, nproc 5.
All fourteen compiling semantic mutants caught: nine dispatch, three transformation and two fixed-divisor variants; exact substitutions and output witnesses in evidence/mutants.json.
Not covered: full gate, seventeen required external correctness checks, whole-rule findings, independent Adamic numeric parsing or underlying predicate implementations.

## Contracts and dependency boundary

bareValuePredicate selects the actual supplied predicate function for every named BareValueKind. PositiveInteger, Opacity, SpacingMultiplier and FontStretchPercentage each have their own dependency. StrictPositiveInteger and GridRepeat share strict. Fraction, empty and unknown kinds yield no function. The original Go function-returning-function behavior is represented by a presence object because stage 0 explicitly refuses BarePredicate | null returns. When present is false, its placeholder function must not be invoked. This API representation does not reinterpret nil as a false predicate. The initial refused attempt and its failed log are retained. No compiler or shared harness was edited.

bareValueTransform rewrites only spacing and grid repeat kinds and preserves every other value verbatim, including empty values, Unicode, NUL and newlines. It performs no validation or caller suffix handling. isValidOpacityValue delegates the original string and exact divisor 0.25 to the shared isMultipleOf dependency. Numeric parsing, nonnegative float modulo and canonical rendering remain outside this helper, as in Go. No regex or numeric parser is introduced.

Thin Go overlays export original private helpers without changing their bodies. Pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db is asserted before each run. Predicate selection identity is observed using reflect on the actual function returned by the original Go dispatch. Supplied callbacks use the corresponding unchanged Go predicate answer on the original value. Observing identity distinguishes opacity from spacing even where their present boolean answers coincide. This tests dispatch, not a new implementation of those predicates. Opacity's 0.5 and 1 mutation results also come from actual Go isMultipleOf, not simulated responses.

## Consumer evidence and mutants

Each helper has the same three frozen blocked consumers: better-tailwindcss/enforce-consistent-class-order, better-tailwindcss/enforce-shorthand-classes and better-tailwindcss/no-unknown-classes. Every inventory consumer, including already-ported consumers, is included by the generator. All test files named in the inventory are parsed as Go; their string literals are captured. There are 406 distinct captured strings, 1503 total values after candidate substring projections and controls. Twelve named/empty/unknown kind spellings yield 18036 predicate rows and 18036 transform rows; opacity has 1503 rows. These are helper-input projections and dependency observations, not whole ASTs or whole-rule diagnostics.

Controls cover every ASCII scalar alone and between digits, Unicode supplementary characters, embedded NUL/newlines, empty strings, signs, leading zeroes, quarter increments -2 through 202, malformed numeric spellings, percentage boundaries and all kind spellings. Exact per-consumer files, counts, complete input rows and raw Go outputs are committed in evidence. Every baseline must agree byte-for-byte with actual Go on source Node, emitted Node and ASan/UBSan native.

Nine dispatch mutants omit the two strict-kind branches separately, change each selected predicate, and wrongly provide predicates for Fraction and unknown kinds. Three transformation mutants omit the closing spacing delimiter, change the grid fraction and erase the default identity. Two opacity mutants change the divisor to 0.5 and 1; both fail on an actual valid quarter value. All fourteen variants compile, execute with exit zero and empty stderr, and differ in output. The run wrapper rejects compiler, process or sanitizer failures before granting mutation credit. Complete witness lines and substitutions are retained in mutants.json; the lossless final log is helpers.log.gz. A preceding twelve-mutant green run is also retained; only the final complete fourteen-mutant run is reported as final.

## Landing and readiness

Reservation 38b319ab5de1cd27016a55c10931eab6a9bebf52 was pushed to codex/lint-helpers-05 and its success collected before creating any batch28 source. All twenty origin helper branch claim trees were fetched and read; comments retain the larger shared bundle's ownership in HELPERS.md. These three tie the highest remaining unclaimed concrete fan-out at three. Refresh during validation confirms no competing reservation or changed main/lint base.

The preceding 77 helpers are complete and published. Compiler, runtime, oracle, cohere and library Git objects remain identical to that published state, recorded in retained-input-identity.json. Their earlier proof is retained, not relabeled as freshly rerun here. The new trio removes nine prerequisites across three rules, no final blocker. Cumulative slot05: 80 helpers, 401 dependency occurrences across 76 consumers; 51 helper-ready candidates including the shared base's 46, under the frozen common-adapter assumption. No rule is marked ported. readiness.json records the exact rule lists.

## Commands and limits

Every test output goes directly to a log, never a pipe. Build shells source /workspace/adamic-tools/env.sh and set the requested GOPROXY fallback.

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/lint05-batch28-setup.log 2>&1
ADAMIC_SLOT05_BATCH28_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch28/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch28 -count=1 -v -timeout=20m > /tmp/lint05-batch28-helpers.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch28-node.log 2>&1
go vet ./... > /tmp/lint05-batch28-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05 > /tmp/lint05-batch28-format.log
```

Setup timing lines: Node 0.024s, Go 0.026s, markdown 0.075s, submodules 0.082s, clang 0.196s, Go build 25.631s, test binaries deferred 25.759s, build cache warm 25.760s, done 25.787s. Five processors, quota four cores, 17.6 GB; Go 1.27.1, clang 20.1.8, Node 24.19.0. The initial explicit compiler refusal failed the suite in 3.262s and was corrected through the owned API representation. Final suite has no skipped test. Seven external Node probes were uncached, zero hits and seven misses. Vet and format logs are empty.

Full repository gate and seventeen stage1 external-input checks were not run, skipped to green, relaxed or deleted. No shared compiler, runtime, rule registry, finding adapter or harness was changed. No main or area branch is pushed and no PR is opened.
