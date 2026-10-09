# Spread calls and JSON shorthand

Roadmap step 42, tasks #aqh658t and #kqxkvg0. The supplied witnesses are the source of these reductions; `ahra` is not installed.

## Spread arguments

Array push and unshift use the existing call-argument lowering, which already accepts readonly arrays and packs rest calls. The receiver is evaluated first. An ordinary array literal then expands each spread exactly where it occurs, before later arguments can mutate its source. A generated IR function inserts the saved items and returns the receiver's resulting length. Unshift reverses only the private argument pack and inserts at the front, preserving argument order. The existing single-value push path is retained.

Different represented element types are converted through ordinary IR before the next argument runs. An empty literal spread contributes no items and takes the destination representation. The existing tuple and iterable representation boundaries remain; unsupported representations report their source location, and new insertion diagnostics explain how to convert explicitly. No new backend operation or runtime layout is needed.

`call_spread_array.a` covers the reduced readonly-string push witness, push and unshift lengths, once-only evaluation, empty spreads, multiple spreads, self-spread, source mutation by a later argument, receiver and argument order, readonly rest calls and function-value calls, inline literals, and conversion from number to number | undefined.

The evaluated-twice mutant replaces the actual spread operand with a comma expression containing two evaluations. It compiles, exits zero, and passes ASan/UBSan and leak checks. Source Node catches stdout: the first evaluation counter is 1 rather than the mutant's 2, and the later counter is 2 rather than 3. The JavaScript backend also catches the mutant against source Node.

## Validation

Tools: Node v24.19.0, Go go1.27.1 linux/amd64, clang 20.1.8. `nproc` is 5; cpu.max is 400000 100000, a four-CPU quota. Setup used GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh.

Setup timing lines: node ready 0.022s; go ready 0.022s; submodules ready 0.064s; markdown dependencies skipped step-duration=0.008s; markdown dependencies ready 0.071s; clang ready 0.184s; go build ready 41.931s; test binaries deferred 42.159s; build cache warm 42.161s; done 42.189s.

The source-Node comparisons include JavaScript output and native release, ASan/UBSan, and leak builds. Every added top-level test begins with t.Parallel. The final uncached test command and measurements are recorded below with the shorthand item. Linux counts were regenerated; existing rows did not change.

## JSON shorthand

JSONStringify's complete-literal lowering now accepts shorthand properties. It uses the existing shorthand lowering to read the checker's value binding, rather than the property's own symbol. The JSON schema comes from that binding's declared type, preserving the representation of a number | undefined local even inside a narrowed branch. Evaluation and schema key order continue through the existing literal path.

`json_shorthand.a` covers the reduced `{ status }` argument, the corresponding explicit assignment, parameter and block shadowing, built strings, field evaluation order, booleans, omitted undefined fields, readonly arrays, and a narrowed optional-number binding. The existing structural-reference, toJSON, prototype, and duplicate-key boundaries remain.

The wrong-binding mutant replaces the parameter read in `report` with the outer status binding. It compiles, exits zero, and passes ASan/UBSan and leak checks. Source Node prints `{"status":"innerinner"}` where native prints `{"status":"outerouter"}`. The JavaScript backend also catches the mutant against source Node.

## Commands and measurements

All test output went directly to log files. Final fixture and mutant run:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestCallSpread|TestJSONStringifyShorthand)' -v
```

| Added top-level test | Seconds | Result |
|---|---:|---|
| TestCallSpreadArray | 0.63 | Pass |
| TestCallSpreadEvaluatedTwiceMutant | 0.56 | Node catches wrong stdout in both backends |
| TestJSONStringifyShorthand | 0.50 | Pass |
| TestJSONStringifyShorthandWrongBindingMutant | 0.44 | Node catches wrong stdout in both backends |

An earlier runtime-cache-cold spread run took 21.17s for the fixture and 17.97s for its mutant, also under the 60s leaf limit.

Additional checks:

```
go build -o /tmp/spread-shorthand-adamic ./cmd/adamic
go vet ./internal/lower ./internal/oracle
/tmp/spread-shorthand-adamic types internal/oracle/testdata/call_spread_array.a internal/oracle/testdata/json_shorthand.a
go test ./internal/oracle -run '^TestJSONStringifyRefusals$' -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/arguments_length_(spread|extended).a$' -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts
```

Build, vet, and both .a checks exited 0. JSON refusals passed in 0.22s. Existing rest spread and extended fixtures passed in 0.59s and 9.33s. Linux counts regeneration passed in 50.713s, adding only these two fixture rows. The lane-check output is included in the delivery report after running it on the committed branch.

No protected compiler files were edited, no code was copied from cohere, and no pending tests or acceptance dependencies were added. The full gate, WASI, arbitrary iterables, and existing unrepresented tuple spreads were not covered by this unit.
