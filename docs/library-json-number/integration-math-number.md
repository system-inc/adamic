# Integration's Math and Number probes from 1d3929c

The supplied `math_edges`, `own_names` and `convert_sources` are preserved as
`internal/oracle/testdata/library_number_math_edges.a`,
`library_number_own_names.a` and `library_number_convert_sources.a`.
All three agree with Node on this branch. Their registration is in the slice's
oracle test file; integration's `oracle_test.go` is unchanged. Their allocation
rows are recorded in `internal/oracle/counts.md`.

The final uncached oracle passed in 1.017s: Node source, JavaScript backend,
sanitized native, release native and leak checks. The full counts update passed
in 20.589s and changed only these three rows. `go vet ./internal/oracle` passed
with no diagnostics. The full flow gate passed in 52.662s, including the three
new root fixtures. Gofmt produced no paths. Test output was written to log files,
never piped.

## Four independently caught mutants

Each mutant compiled, both source Node and native exited zero with empty
stderr, release matched sanitized native, the JavaScript backend matched Node,
and there were no leak failures. Only Node's stdout comparison caught each
mutant. The runtime was restored after each run, including in a finally block.
No runtime implementation change is committed.

| Mutant | Fixture | Observed Node result | Observed mutant result | Time |
|---|---|---|---|---|
| Negative fround cutoff moved one ULP toward negative infinity | math_edges | `Math.fround(-3.4028235677973366e38)` is `-Infinity` | `-3.4028234663852886e38` | 8.916s |
| Names table renames `EPSILON` to `Epsilon` | own_names | `EPSILON true false` | `EPSILON false false` | 8.769s |
| Octal and binary prefix recognition removed, hexadecimal retained | convert_sources | The first line ends with `3` for `Number(' 0b11 ')` | `NaN`; three output lines differ | 8.912s |
| `adamic_number_from_string(NULL)` returns zero | convert_sources | Missing array source: `NaN NaN NaN NaN` | `NaN NaN NaN 0`; two output lines differ | 8.790s |

`run-math-number-mutants.py` checks that the negative cutoff moves exactly one
binary64 ULP, requires each selected subtest to fail solely against Node, and
restores the original runtime bytes. `math-number-mutants.json` records the
four outcomes and complete log paths under `/tmp/library-json-number-math-mutants/`.

## Optional-field disagreement

`optional_field_min` does **not** agree with Node. On Linux, Node exits zero
and prints `true\n`. Sanitized and release native agree with each other but
exit 70, printing nothing and reporting:

```text
adamic: panic: compiler bug: a field the checker proved is there is missing
```

The exact source is preserved in
`docs/library-json-number/probes/library_number_optional_field_min.a` for
@system_adamic. It is not registered as a passing oracle fixture, and no
library workaround hides the disagreement. The initial four-probe comparison
and this diagnostic are in `integration-math-number-baseline.txt`.

One-line reproducer:

```text
interface Box { size?: number } const boxes: Box[] = [{}]; for (const box of boxes) console.log(`${box.size === undefined}`);
```

## Commands

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestNativeAgreesWithNode/(internal|docs)/(oracle|library-json-number)/(testdata|probes)/library_number_(math_edges|own_names|convert_sources|optional_field_min).a' \
  > /tmp/library-json-number-integration-math-baseline.log 2>&1
# Exit 1 only for the optional-field disagreement; the other three passed.
python3 docs/library-json-number/run-math-number-mutants.py \
  > /tmp/library-json-number-math-mutants-run.log 2>&1
# Exit 0: all four mutants caught only by Node.
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_number_(math_edges|own_names|convert_sources).a' \
  > /tmp/library-json-number-math-final-oracle.log 2>&1
# Exit 0.
go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded \
  -args -update-counts > /tmp/library-json-number-math-counts.log 2>&1
# Exit 0.
go vet ./internal/oracle > /tmp/library-json-number-math-vet.log 2>&1
# Exit 0.
go test -count=1 -timeout 30m ./internal/flow \
  > /tmp/library-json-number-math-flow.log 2>&1
# Exit 0, 52.662s.
```

The broader gate is not repeated for this fixture-only addition. No lowering
or runtime behavior changed, and no new test262 survey is claimed here.
