# Shorthand nested compiler proof — #qk8rztp

2026-11-12, Linux amd64. Tested HEAD `dc473599cabd5882f4643d5126345d46610af920` from `emit/shorthand-nested-4`, detached, against main `c41e5377e05326b76d69618cb65daa8f65073c92` (exactly two commits). All requested unmutated checks passed. The mutant built and was killed by three red subtests. The full admission delta passed with no classification or compiler-stderr changes.

## Preparation

Loaded `/workspace/adamic-tools/env.sh` first. Created `/tmp/adamic-gate` outside the repository and exported `GOFLAGS=-buildvcs=false`. Fetched `origin emit/shorthand-nested-4`, checked out the exact HEAD detached, and ran `git submodule update --init --recursive`: exit 0. Recursive pins:

- `cohere`: `5360753e19e4417675a90e92a1823bea66b13918`
- `cohere/TypeScript`: `d92d9bfee114c80be2c375d72edae966176e3a4f`

Provisioned `/tmp/adamic-gate/typescript` as a separate checkout, reusing local submodule objects, with origin `https://github.com/microsoft/TypeScript.git`; fetched and checked out upstream `050880ce59e30b356b686bd3144efe24f875ebc8` detached. Exported `ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-gate/typescript`.

Copied `stage3/api/package.json` and `package-lock.json` outside the repository to `/tmp/adamic-gate/api` and ran `npm ci --prefix /tmp/adamic-gate/api`: exit 0, three packages installed. Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`. Prepended the external `node_modules/.bin` to PATH. Versions: Go 1.27.1, Node v24.19.0, `tsc` 6.0.3. One heavy command ran at a time throughout; the admission delta ran alone.

## 1. Formatting and vet

- `gofmt -l internal/lower internal/oracle`: exit 0, **no output**.
- `go vet ./internal/lower ./internal/oracle`: exit 0, no diagnostics.

## 2. Selected lowering pins

Command (stdout and stderr redirected to `/tmp/pins.log`):

```sh
go test ./internal/lower -run '^(TestNestedGeneratorRoadsAreNotYet|TestOverloadLiteralResultIsChecked|TestOverloadLiteralResultLieStopsAtTheCall|TestOverloadLiteralResultFollowsThePromise|TestOverloadTagResultIsChecked|TestOverloadTagResultLieStopsAtTheCall|TestOverloadOpenTagResultIsChecked|TestOverloadOpenTagResultLieStopsAtTheRead|TestOverloadBrandedResultIsChecked|TestOverloadNestedResultLieStopsAtTheCall|TestOverloadNestedSiblingResultLieStopsAtTheCall|TestOverloadNarrowAsAValueStops|TestOverloadDiscardedResultIsErased|TestOverloadDiscardedResultKeepsTheImplementationsRepresentation|TestOverloadNarrowTagStops|TestOverloadNarrowResultStaysRefusedInA|TestOverloadUncheckableResultStaysRefused)$' -count=1 -v -timeout 30m > /tmp/pins.log 2>&1
```

Exit **0**; package **PASS**. All 17 tests and 38 subtests passed; no FAIL, so there are no failing got/want pairs. Every result follows verbatim:

```text
--- PASS: TestOverloadNarrowResultStaysRefusedInA (0.41s)
--- PASS: TestOverloadBrandedResultIsChecked (0.48s)
--- PASS: TestOverloadDiscardedResultIsErased (22.34s)
--- PASS: TestOverloadTagResultLieStopsAtTheCall (22.05s)
--- PASS: TestOverloadOpenTagResultLieStopsAtTheRead (21.67s)
--- PASS: TestOverloadOpenTagResultIsChecked (0.58s)
--- PASS: TestOverloadNestedResultLieStopsAtTheCall (23.09s)
--- PASS: TestOverloadLiteralResultFollowsThePromise (0.70s)
--- PASS: TestOverloadTagResultIsChecked (0.97s)
--- PASS: TestOverloadNarrowTagStops (0.12s)
--- PASS: TestOverloadLiteralResultLieStopsAtTheCall (0.93s)
--- PASS: TestOverloadDiscardedResultKeepsTheImplementationsRepresentation (24.18s)
--- PASS: TestOverloadLiteralResultIsChecked (1.13s)
--- PASS: TestOverloadNestedSiblingResultLieStopsAtTheCall (2.02s)
--- PASS: TestOverloadNarrowAsAValueStops (0.01s)
    --- PASS: TestOverloadNarrowAsAValueStops/const (0.12s)
    --- PASS: TestOverloadNarrowAsAValueStops/nested_in_a_const (0.14s)
    --- PASS: TestOverloadNarrowAsAValueStops/nested_in_a_shorthand (0.17s)
    --- PASS: TestOverloadNarrowAsAValueStops/nested_into_a_narrower_slot (0.16s)
    --- PASS: TestOverloadNarrowAsAValueStops/bind (0.34s)
    --- PASS: TestOverloadNarrowAsAValueStops/call (0.24s)
    --- PASS: TestOverloadNarrowAsAValueStops/union_of_two_sets (0.23s)
--- PASS: TestNestedGeneratorRoadsAreNotYet (0.35s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/spread (0.16s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/consumed_inside_a_nested_generator (0.30s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_closure_over_its_own_local (0.34s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/destructured (0.36s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/declaring_a_nested_function (0.55s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/returned (0.26s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/stored (0.25s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_shorthand_property,_called_through (0.12s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_shorthand_property,_only_typeof_read (0.10s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/inside_a_generator (0.26s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/in_a_block (0.29s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/passed_as_an_argument (0.59s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/inside_an_async_function (0.26s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_nested_generic_generator (0.15s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/inside_a_generic_function (0.15s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/the_function_as_a_value (0.45s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/an_object_literal's_getter_in_the_body (0.16s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_write_by_an_arrow_made_inside_the_generator (0.16s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_class_in_the_body_whose_method_writes_a_captured_local (0.31s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_prefix_update's_helper (0.18s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_destructuring_target (0.15s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_compound_assignment (0.23s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/an_increment (0.27s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/an_assignment_to_a_captured_local (0.13s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_named_function_in_the_body_writing_a_captured_local (0.59s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/a_logical_assignment (0.34s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/new_Set (0.10s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/Array.from (0.10s)
    --- PASS: TestNestedGeneratorRoadsAreNotYet/new_Map (0.14s)
--- PASS: TestOverloadUncheckableResultStaysRefused (0.00s)
    --- PASS: TestOverloadUncheckableResultStaysRefused/array_elements (0.73s)
    --- PASS: TestOverloadUncheckableResultStaysRefused/untagged_subtype (0.10s)
```

## 3. Uncached Node identity oracle

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNestedShorthandIdentityAgreesWithNode$' -count=1 -v -timeout 30m
```

Exit **0**; `TestNestedShorthandIdentityAgreesWithNode` **PASS** (358.79s); package PASS. Wall time 379.746s including build overhead. Native cache hits=0/misses=3; Node hits=0/misses=2; probe hits=0/misses=0.

## 4. Mutation proof

Asserted the following original text occurred exactly once in `internal/lower/object.go`, then replaced it exactly:

```go
// Original:
if value, handled, err := l.nestedReferenceOf(property, symbol); handled {
// Mutant:
if value, handled, err := l.nestedReferenceOf(property, symbol); handled && false {
```

`go build ./internal/lower`: exit **0**. Reran the exact step 2 test command, redirecting both streams to `/tmp/adamic-gate/mutant-pins.log`: exit **1**, package FAIL. Exactly these three subtests were red; their whole diagnostic lines, including got and want, follow:

### `TestNestedGeneratorRoadsAreNotYet/a_shorthand_property,_called_through` — FAIL

```text
    generators_nested_test.go:205: got /tmp/adamic-gate/TestNestedGeneratorRoadsAreNotYeta_shorthand_property,_called_t89939095/001/main.a:4:21: stage 0 can't lower a generator function called other than by its name (#qk8rztp) yet, want a NotYet saying "a generator function declared inside a function, used other than by a call to it by name (#qk8rztp)"
```

### `TestNestedGeneratorRoadsAreNotYet/a_shorthand_property,_only_typeof_read` — FAIL

```text
    generators_nested_test.go:205: got <nil>, want a NotYet saying "a generator function declared inside a function, used other than by a call to it by name (#qk8rztp)"
```

### `TestOverloadNarrowAsAValueStops/nested_in_a_shorthand` — FAIL

```text
    overload_result_check_test.go:284: want the stop "an overloaded function as a value" where the overload set is read as a value, got <nil>
```

Both expected “a shorthand property” generator rows were red. **Yes, the “nested in a shorthand” overload row was also red.** The remaining 35 subtests passed; the two parent tests containing those red rows failed, while the other 15 top-level tests passed.

Restored `object.go` byte-for-byte from the original saved contents. `git diff --exit-code`: exit **0**, no output. The tree was also clean before report creation.

## 5. Full lowering suite

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

Exit **0**; wall time **1490.114s (24m 50.114s)**. Package output:

```text
ok  	github.com/system-inc/adamic/internal/lower	1481.245s
```

Failing tests / first error lines: **none**. Environment failures in the completed run: **none**.

An earlier attempt was interrupted when the managed environment restarted, after roughly 13 minutes. It produced no recorded command exit or package result and is inconclusive, not a test failure. Its empty log was preserved as `/tmp/adamic-gate/lower-interrupted.log`. The complete rerun above supplied the reported result. Previously completed steps survived and were retained.

## 6. Admission delta

Generated the pinned manifest with `python3 cloud/admission-corpus/manifest.py > /tmp/adamic-gate/manifest.json`: exit 0. Then ran alone:

```sh
go run ./cmd/adamic-admission-delta --base c41e5377e05326b76d69618cb65daa8f65073c92 --head HEAD --workers 2 --manifest /tmp/adamic-gate/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json
```

JSON stdout: `/tmp/adamic-gate/delta.json`; stderr: `/tmp/adamic-gate/delta.log`.

| Metric | Result |
| --- | --- |
| Exit | 0 |
| Verdict | pass |
| Void | false |
| Classification mode | lowering-only |
| diff_count | 1 |
| admitted (newly accepted) | 0 |
| Newly refused | 0 |
| Accepted by both | 3320 |
| Refused by both | 1982 |
| Complete corpus | true |
| Programs | 5302 |
| Workers | 2 |
| Wall time | 1470.560s (24m 30.560s) |

Corpora: witnesses=2, fixtures=4354, gaps=105, review=20, fuzz=0, stage3=821. Manifest blob: `a2470fdebd73be89738947886ad8d3bc61fb005d`; generator blob: `09d061a03b1e016387e975d25ee0cd5b75ac6c1f`, at the exact tested HEAD.

`diff_count=1` is changed-program coverage, not a classification change: `internal/oracle/testdata/nested_shorthand/identity.a` (blob `c25d7731ded1569214273124726a398cf0ebf39c`). It was accepted by both revisions with identical empty compiler stderr.

**Programs whose classification or compiler stderr differs: none.** Compared every entry's base/head classification and exact `base_compile.stderr` / `compile.stderr` strings. There are therefore no differing base/head texts to enumerate.

Runtime scope reported by the tool: sampling_size=115, omitted=0, negative witnesses=50, accepted witnesses=50, Node agreements=13, ruled runtime stops=52. Two runtime runs timed out and were retried alone; both retries completed (`still_timed_out_runs=0`). The complete-corpus flag covers admission classification; it does not claim every admitted program was executed against Node.

## Report delivery

Delivery branch: `proof/shorthand-4-report`, created from the exact tested HEAD. The report commit contains only `proof/shorthand-4/report.md`; the push is restricted to that branch on origin. Raw execution logs remain at the `/tmp` paths above.
