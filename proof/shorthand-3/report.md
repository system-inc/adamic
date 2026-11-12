# Shorthand nested-function proof — #qk8rztp

HEAD: `f391f3b47f44a3d4db8b6532dff92314dcf1052b` (detached for the proof). Base: `730e7c26e590af0c57c545ba0cb905b09e4bee54`. HEAD contains exactly two commits above base. Date: 2026-11-12 UTC.

## Decisive results

- Both baseline pins passed: 29 lowerer subtests, shorthand identity, and all 64 generator oracle subtests.
- The specified mutant compiled and killed exactly the two expected lowerer shorthand rows. Shorthand identity remained **green** under the mutant; all generator oracle rows also remained green.
- Full lowerer and IR suites passed.
- Full oracle exited 1 with 13 environment failures: twelve TypeScript module-resolution failures (all passed on the corrected isolated follow-up) and missing `WASI_SYSROOT`.
- Admission verdict: **void**, with two native-compilation timeouts. No admission/classification or compiler-stderr delta was observed across 5,302 programs. This is not a passing full admission proof.

## Step 3: mutation sensitivity

The original text occurred exactly once in `internal/lower/object.go`:

```go
if value, handled, err := l.nestedReferenceOf(property, symbol); handled {
```

It was replaced with:

```go
if value, handled, err := l.nestedReferenceOf(property, symbol); handled && false {
```

`go build ./internal/lower`: exit=0 wall_seconds=19.
Both exact step 2 commands were rerun, including `ADAMIC_GATE_UNCACHED=1` on the oracle command. Lowerer pin under mutant: exit=1 wall_seconds=39. Exactly two red subtests; their whole diagnostic lines (including got and want) follow:

```text
generators_nested_test.go:205: got /tmp/adamic-gate/TestNestedGeneratorRoadsAreNotYeta_shorthand_property,_called_t2645321811/001/main.a:4:21: stage 0 can't lower a generator function called other than by its name (#qk8rztp) yet, want a NotYet saying "a generator function declared inside a function, used other than by a call to it by name (#qk8rztp)"
```

```text
generators_nested_test.go:205: got <nil>, want a NotYet saying "a generator function declared inside a function, used other than by a call to it by name (#qk8rztp)"
```

The `called_through` row got a different NotYet diagnostic; the `only_typeof_read` row got `<nil>`, meaning it lowered. The remaining 27 rows passed.
Oracle pins under mutant: exit=0 wall_seconds=105. **TestNestedShorthandIdentityAgreesWithNode is GREEN.** There are no red oracle subtests. Every result is listed below.

```text
--- PASS: TestNestedShorthandIdentityAgreesWithNode (53.77s)
--- PASS: TestGeneratorsAgreeWithNode (0.00s)
--- PASS: TestGeneratorsAgreeWithNode/view_string_items.a (53.66s)
--- PASS: TestGeneratorsAgreeWithNode/view_outer_loop.a (53.67s)
--- PASS: TestGeneratorsAgreeWithNode/array_from.a (53.79s)
--- PASS: TestGeneratorsAgreeWithNode/view_throw.a (53.79s)
--- PASS: TestGeneratorsAgreeWithNode/view_early_return.a (2.45s)
--- PASS: TestGeneratorsAgreeWithNode/view_body_throw.a (2.37s)
--- PASS: TestGeneratorsAgreeWithNode/unused_class.a (2.72s)
--- PASS: TestGeneratorsAgreeWithNode/view_counted_items.a (3.07s)
--- PASS: TestGeneratorsAgreeWithNode/view_mixed.a (3.42s)
--- PASS: TestGeneratorsAgreeWithNode/uncaught_from.a (2.44s)
--- PASS: TestGeneratorsAgreeWithNode/uncaught_spread.a (2.56s)
--- PASS: TestGeneratorsAgreeWithNode/uncaught.a (2.22s)
--- PASS: TestGeneratorsAgreeWithNode/stored_later_throw.a (2.53s)
--- PASS: TestGeneratorsAgreeWithNode/stored_untaken.a (2.88s)
--- PASS: TestGeneratorsAgreeWithNode/stored_exits_finally.a (2.36s)
--- PASS: TestGeneratorsAgreeWithNode/stored_jsx_children.a (2.98s)
--- PASS: TestGeneratorsAgreeWithNode/stored_for_of.a (3.42s)
--- PASS: TestGeneratorsAgreeWithNode/spread.a (2.60s)
--- PASS: TestGeneratorsAgreeWithNode/stored_exits.a (2.99s)
--- PASS: TestGeneratorsAgreeWithNode/split.a (2.02s)
--- PASS: TestGeneratorsAgreeWithNode/return_from_loop.a (2.21s)
--- PASS: TestGeneratorsAgreeWithNode/retained_param.a (2.16s)
--- PASS: TestGeneratorsAgreeWithNode/nullish_view_tests.a (2.05s)
--- PASS: TestGeneratorsAgreeWithNode/new_set_plain.a (2.26s)
--- PASS: TestGeneratorsAgreeWithNode/new_set.a (1.94s)
--- PASS: TestGeneratorsAgreeWithNode/new_map_throw.a (2.43s)
--- PASS: TestGeneratorsAgreeWithNode/nested_stack_parent.a (1.61s)
--- PASS: TestGeneratorsAgreeWithNode/new_map_plain.a (2.52s)
--- PASS: TestGeneratorsAgreeWithNode/new_map_entries.a (2.65s)
--- PASS: TestGeneratorsAgreeWithNode/new_map.a (3.04s)
--- PASS: TestGeneratorsAgreeWithNode/nested_hoisted.a (1.76s)
--- PASS: TestGeneratorsAgreeWithNode/nested_early_break.a (1.90s)
--- PASS: TestGeneratorsAgreeWithNode/nested_dead_zone.a (1.57s)
--- PASS: TestGeneratorsAgreeWithNode/nested_iterable.a (3.26s)
--- PASS: TestGeneratorsAgreeWithNode/nested_carrier_temporary.a (1.95s)
--- PASS: TestGeneratorsAgreeWithNode/nested_checker_shapes.a (3.17s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_overloads.a (2.14s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_view.a (3.32s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_grandchild.a (2.62s)
--- PASS: TestGeneratorsAgreeWithNode/map_iterator_entries.a (2.56s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_generic.a (2.92s)
--- PASS: TestGeneratorsAgreeWithNode/map_iterator.a (2.31s)
--- PASS: TestGeneratorsAgreeWithNode/map_defined_iterator.a (2.65s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_loop.a (2.06s)
--- PASS: TestGeneratorsAgreeWithNode/labeled_exit.a (2.56s)
--- PASS: TestGeneratorsAgreeWithNode/nested_live_capture.a (11.26s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_forms.a (2.39s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_environment.a (2.30s)
--- PASS: TestGeneratorsAgreeWithNode/generic_throw.a (1.65s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_capture.a (2.67s)
--- PASS: TestGeneratorsAgreeWithNode/generic.a (2.58s)
--- PASS: TestGeneratorsAgreeWithNode/first_throw.a (1.74s)
--- PASS: TestGeneratorsAgreeWithNode/generic_map.a (3.02s)
--- PASS: TestGeneratorsAgreeWithNode/first_or_undefined.a (1.89s)
--- PASS: TestGeneratorsAgreeWithNode/elementwise_view.a (2.32s)
--- PASS: TestGeneratorsAgreeWithNode/drop_start.a (2.08s)
--- PASS: TestGeneratorsAgreeWithNode/early_return.a (2.28s)
--- PASS: TestGeneratorsAgreeWithNode/consumer_throw.a (2.19s)
--- PASS: TestGeneratorsAgreeWithNode/declared_iterators.a (2.90s)
--- PASS: TestGeneratorsAgreeWithNode/complete.a (2.61s)
--- PASS: TestGeneratorsAgreeWithNode/break_map.a (2.31s)
--- PASS: TestGeneratorsAgreeWithNode/builder_state_strings.a (2.61s)
--- PASS: TestGeneratorsAgreeWithNode/body_throw.a (2.01s)
--- PASS: TestGeneratorsAgreeWithNode/array_from_view.a (1.77s)
```

The original file was restored byte for byte from its external backup. `git diff --exit-code` exited 0 and `git status --short` printed nothing before the full suites; the diff was checked again before admission. No mutant was committed.

## Step 5: admission delta

Run alone, with no other heavy command or suite running. The manifest was generated using the repository’s own `cmd/adamic-admission-delta/README.md` instructions:

```sh
python3 cloud/admission-corpus/manifest.py --sha HEAD > /tmp/shorthand-proof/manifest.json
go run ./cmd/adamic-admission-delta \
  --base 730e7c26e590af0c57c545ba0cb905b09e4bee54 --head HEAD \
  --manifest /tmp/shorthand-proof/manifest.json \
  --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json
```

Execution: exit=1 wall_seconds=1976. The `go run` wrapper exited **1** and printed `exit status 3`; the admission binary’s own exit was **3** (void).

| Field | Value |
|---|---|
| verdict | `void` |
| void | `True` |
| void_reason | `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout; internal/oracle/testdata/reflect_get/stop_literal_getter.a: native compile timeout` |
| workers | `2` |
| diff_count | `1` |
| admitted | `0` |
| complete_corpus | `True` |
| sampling_size | `115` |
| omitted | `0` |
| budget_seconds | `0` |
| timed_out_runs | `2` |
| retried_runs | `2` |
| still_timed_out_runs | `2` |
| generator_revision | `f391f3b47f44a3d4db8b6532dff92314dcf1052b` |
| generator_blob | `09d061a03b1e016387e975d25ee0cd5b75ac6c1f` |
| manifest_blob | `cedf23c94f57124a40f5266d680489e34b5b1509` |
| newly refused | `0` |

Classification totals: accepted-by-both: 3320, refused-by-both: 1982.
The sole diff program is `internal/oracle/testdata/nested_shorthand/identity.a` (blob `c25d7731ded1569214273124726a398cf0ebf39c`). It is accepted by both revisions. `diff_count` counts changed source programs, not observed behavior differences.
**Every program whose classification or compiler stderr differs:** none. All base/head classification exit statuses, errors, stdout and stderr were compared. There are no differing texts to reproduce. In particular, both stderr texts for each of the two timeout programs are the empty string at the classification stage.
Gate stderr, verbatim:

```text
negative witnesses: 50
ruled run-time stops: 51
ruled backend divergences: 0
runtime timeouts: 2 runs timed out, 2 retried alone, 2 timed out again
changed programs: 1
admission delta is void: internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout; internal/oracle/testdata/reflect_get/stop_literal_getter.a: native compile timeout
exit status 3
```

### Timeout reruns, one command at a time, both SHAs

The gate itself retried both native builds alone and still reached its 10-second limit. After it exited, both programs were rerun serially with the exact cached admission compilers. Cache `.inputs` records tied the binaries to the base and HEAD SHAs. The unchanged HEAD corpus files were used for both compilers, matching the gate’s head-corpus contract. Each run repeated lowering, native build and execution; supplemental JavaScript compilation/execution and source Node execution used the gate’s command forms.

| Revision | Program | Native build at 10s | Extended build at 60s | Native exit |
|---|---|---|---|---|
| base | `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts` | exit 0, 2.434s | not needed | 70 |
| base | `internal/oracle/testdata/reflect_get/stop_literal_getter.a` | exit 124, 10.007s | exit 0, 26.972s | 70 |
| head | `internal/oracle/testdata/intersection_views_stops/nested_boolean.ts` | exit 0, 4.871s | not needed | 70 |
| head | `internal/oracle/testdata/reflect_get/stop_literal_getter.a` | exit 0, 2.941s | not needed | 70 |

Exit 124 is the external `timeout` limit. These timing failures are marked **environment/performance timeouts**, rather than revision-specific compiler failures: both revisions ultimately build and exhibit identical expected stop behavior. The original complete gate remains void; the isolated reruns do not relabel it pass.

`internal/oracle/testdata/intersection_views_stops/nested_boolean.ts` — whole observed texts for both revisions:

```json
{
  "revision": "base",
  "lower": {
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native": {
    "exit": 70,
    "stdout": "other\n",
    "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
  },
  "javascript_compile": {
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "javascript": {
    "exit": 70,
    "stdout": "other\n",
    "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
  },
  "node": {
    "exit": 0,
    "stdout": "other\n4 false\n",
    "stderr": ""
  }
}
```

```json
{
  "revision": "head",
  "lower": {
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native": {
    "exit": 70,
    "stdout": "other\n",
    "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
  },
  "javascript_compile": {
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "javascript": {
    "exit": 70,
    "stdout": "other\n",
    "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n"
  },
  "node": {
    "exit": 0,
    "stdout": "other\n4 false\n",
    "stderr": ""
  }
}
```

`internal/oracle/testdata/reflect_get/stop_literal_getter.a` — whole observed texts for both revisions:

```json
{
  "revision": "base",
  "lower": {
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native": {
    "exit": 70,
    "stdout": "got\n",
    "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
  },
  "javascript_compile": {
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "javascript": {
    "exit": 70,
    "stdout": "got\n",
    "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
  },
  "node": {
    "exit": 0,
    "stdout": "got\nstring\n",
    "stderr": ""
  }
}
```

```json
{
  "revision": "head",
  "lower": {
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native": {
    "exit": 70,
    "stdout": "got\n",
    "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
  },
  "javascript_compile": {
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "javascript": {
    "exit": 70,
    "stdout": "got\n",
    "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
  },
  "node": {
    "exit": 0,
    "stdout": "got\nstring\n",
    "stderr": ""
  }
}
```

A supplemental harness initially used unsupported `build --backend js` syntax and direct Node on `.a`; those harness-only errors were corrected to `adamic js <path>` and `node --disable-warning=ExperimentalWarning oracle/node.mjs <path>`. The tables and texts above contain corrected observations. No compiler source changed for this correction.

## Step 1: checkout and static checks

Loaded `/workspace/adamic-tools/env.sh` first. Fetched `origin emit/shorthand-nested-3`, checked out the requested HEAD detached, and completed `git submodule update --init --recursive` (all exit 0). The checkout was initially clean.

- `cohere`: `5360753e19e4417675a90e92a1823bea66b13918`.
- `cohere/TypeScript`: `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- `gofmt -l internal/lower internal/oracle`: exit 0, **no output**, also rechecked after provisioning.
- `go vet ./internal/lower ./internal/oracle`: exit=0 wall_seconds=291, no diagnostics.

## Environment and provenance

Created `/tmp/adamic-gate` outside the repository. Provisioned upstream `https://github.com/microsoft/TypeScript.git` at exact commit `050880ce59e30b356b686bd3144efe24f875ebc8` in `/tmp/shorthand-proof/TypeScript`. Copied `stage3/api/package.json` and its lockfile to `/tmp/shorthand-proof/api` and ran `npm ci --prefix /tmp/shorthand-proof/api` there; repository packages were not installed or modified. `tsc` resolved to `/tmp/shorthand-proof/api/node_modules/.bin/tsc` and reported `Version 6.0.3`.
Every build/test and the admission gate inherited:

```sh
source /workspace/adamic-tools/env.sh
export GOFLAGS=-buildvcs=false
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/shorthand-proof/TypeScript
export PATH=/tmp/shorthand-proof/api/node_modules/.bin:$PATH
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```

Go 1.27.1 linux/amd64; Node v24.19.0; npm 11.9.0; clang 20.1.8. No two heavy commands ran concurrently. Test-internal parallelism and the gate’s requested two workers were left intact. Logs and provisioning were outside the repository.

## Step 2: baseline pins

```sh
go test ./internal/lower -run '^TestNestedGeneratorRoadsAreNotYet$' -count=1 -v -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestNestedShorthandIdentityAgreesWithNode|TestGeneratorsAgreeWithNode)$' -count=1 -v -timeout 40m
```

Lowerer: exit=0 wall_seconds=45. Oracle: exit=0 wall_seconds=510. No baseline failures, so there are no failing got/want pairs. PASS/FAIL for every test and subtest:

```text
--- PASS: TestNestedGeneratorRoadsAreNotYet (0.05s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/spread (0.14s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/inside_a_generic_function (0.14s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_logical_assignment (0.16s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/an_assignment_to_a_captured_local (0.17s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/stored (0.18s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_shorthand_property,_called_through (0.16s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/in_a_block (0.13s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_shorthand_property,_only_typeof_read (0.16s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/inside_a_generator (0.12s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/the_function_as_a_value (0.14s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_nested_generic_generator (0.08s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_class_in_the_body_whose_method_writes_a_captured_local (0.10s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_named_function_in_the_body_writing_a_captured_local (0.16s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/an_object_literal's_getter_in_the_body (0.17s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/inside_an_async_function (0.22s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/an_increment (0.15s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_destructuring_target (0.16s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/destructured (0.11s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/declaring_a_nested_function (0.12s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_compound_assignment (0.10s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_closure_over_its_own_local (0.13s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_write_by_an_arrow_made_inside_the_generator (0.08s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_prefix_update's_helper (0.10s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/new_Set (0.11s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/new_Map (0.06s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/Array.from (0.06s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/returned (0.07s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/consumed_inside_a_nested_generator (0.06s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/passed_as_an_argument (0.06s)
```

```text
--- PASS: TestNestedShorthandIdentityAgreesWithNode (423.05s)
--- PASS: TestGeneratorsAgreeWithNode (0.00s)
--- PASS: TestGeneratorsAgreeWithNode/view_string_items.a (422.99s)
--- PASS: TestGeneratorsAgreeWithNode/array_from.a (423.01s)
--- PASS: TestGeneratorsAgreeWithNode/view_outer_loop.a (423.27s)
--- PASS: TestGeneratorsAgreeWithNode/view_throw.a (423.32s)
--- PASS: TestGeneratorsAgreeWithNode/unused_class.a (1.93s)
--- PASS: TestGeneratorsAgreeWithNode/view_counted_items.a (2.27s)
--- PASS: TestGeneratorsAgreeWithNode/view_early_return.a (2.59s)
--- PASS: TestGeneratorsAgreeWithNode/view_body_throw.a (3.10s)
--- PASS: TestGeneratorsAgreeWithNode/view_mixed.a (3.50s)
--- PASS: TestGeneratorsAgreeWithNode/uncaught_spread.a (1.80s)
--- PASS: TestGeneratorsAgreeWithNode/uncaught_from.a (2.14s)
--- PASS: TestGeneratorsAgreeWithNode/uncaught.a (1.86s)
--- PASS: TestGeneratorsAgreeWithNode/stored_untaken.a (2.47s)
--- PASS: TestGeneratorsAgreeWithNode/stored_later_throw.a (2.80s)
--- PASS: TestGeneratorsAgreeWithNode/stored_jsx_children.a (2.75s)
--- PASS: TestGeneratorsAgreeWithNode/stored_exits_finally.a (2.45s)
--- PASS: TestGeneratorsAgreeWithNode/split.a (1.27s)
--- PASS: TestGeneratorsAgreeWithNode/stored_for_of.a (3.75s)
--- PASS: TestGeneratorsAgreeWithNode/spread.a (2.09s)
--- PASS: TestGeneratorsAgreeWithNode/return_from_loop.a (2.05s)
--- PASS: TestGeneratorsAgreeWithNode/stored_exits.a (3.25s)
--- PASS: TestGeneratorsAgreeWithNode/retained_param.a (1.62s)
--- PASS: TestGeneratorsAgreeWithNode/new_set_plain.a (1.95s)
--- PASS: TestGeneratorsAgreeWithNode/new_set.a (1.77s)
--- PASS: TestGeneratorsAgreeWithNode/nullish_view_tests.a (2.56s)
--- PASS: TestGeneratorsAgreeWithNode/new_map_throw.a (2.18s)
--- PASS: TestGeneratorsAgreeWithNode/new_map_plain.a (2.58s)
--- PASS: TestGeneratorsAgreeWithNode/nested_stack_parent.a (2.18s)
--- PASS: TestGeneratorsAgreeWithNode/new_map.a (2.46s)
--- PASS: TestGeneratorsAgreeWithNode/new_map_entries.a (2.95s)
--- PASS: TestGeneratorsAgreeWithNode/nested_hoisted.a (2.33s)
--- PASS: TestGeneratorsAgreeWithNode/nested_dead_zone.a (2.01s)
--- PASS: TestGeneratorsAgreeWithNode/nested_early_break.a (2.69s)
--- PASS: TestGeneratorsAgreeWithNode/nested_iterable.a (3.61s)
--- PASS: TestGeneratorsAgreeWithNode/nested_carrier_temporary.a (2.37s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_overloads.a (2.67s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_view.a (2.76s)
--- PASS: TestGeneratorsAgreeWithNode/nested_checker_shapes.a (3.60s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_grandchild.a (2.48s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_generic.a (1.90s)
--- PASS: TestGeneratorsAgreeWithNode/map_iterator_entries.a (3.46s)
--- PASS: TestGeneratorsAgreeWithNode/map_iterator.a (3.57s)
--- PASS: TestGeneratorsAgreeWithNode/map_defined_iterator.a (2.77s)
--- PASS: TestGeneratorsAgreeWithNode/labeled_exit.a (2.80s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_environment.a (1.62s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_loop.a (3.19s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_forms.a (3.10s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_capture.a (3.27s)
--- PASS: TestGeneratorsAgreeWithNode/generic_throw.a (2.73s)
--- PASS: TestGeneratorsAgreeWithNode/generic_map.a (2.27s)
--- PASS: TestGeneratorsAgreeWithNode/generic.a (2.85s)
--- PASS: TestGeneratorsAgreeWithNode/first_throw.a (1.86s)
--- PASS: TestGeneratorsAgreeWithNode/first_or_undefined.a (2.22s)
--- PASS: TestGeneratorsAgreeWithNode/elementwise_view.a (2.42s)
--- PASS: TestGeneratorsAgreeWithNode/drop_start.a (2.26s)
--- PASS: TestGeneratorsAgreeWithNode/early_return.a (2.41s)
--- PASS: TestGeneratorsAgreeWithNode/declared_iterators.a (2.81s)
--- PASS: TestGeneratorsAgreeWithNode/consumer_throw.a (2.41s)
--- PASS: TestGeneratorsAgreeWithNode/complete.a (2.22s)
--- PASS: TestGeneratorsAgreeWithNode/builder_state_strings.a (2.81s)
--- PASS: TestGeneratorsAgreeWithNode/break_map.a (2.03s)
--- PASS: TestGeneratorsAgreeWithNode/body_throw.a (2.07s)
--- PASS: TestGeneratorsAgreeWithNode/array_from_view.a (1.98s)
--- PASS: TestGeneratorsAgreeWithNode/nested_live_capture.a (51.16s)
```

### All lowerer mutant results

```text
--- FAIL: TestNestedGeneratorRoadsAreNotYet (0.06s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/spread (0.28s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_nested_generic_generator (0.31s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_class_in_the_body_whose_method_writes_a_captured_local (0.37s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/inside_a_generic_function (0.40s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_logical_assignment (0.44s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_named_function_in_the_body_writing_a_captured_local (0.24s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/an_object_literal's_getter_in_the_body (0.29s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/stored (0.17s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/declaring_a_nested_function (0.10s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_prefix_update's_helper (0.31s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_write_by_an_arrow_made_inside_the_generator (0.29s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_closure_over_its_own_local (0.18s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/consumed_inside_a_nested_generator (0.17s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/destructured (0.21s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/passed_as_an_argument (0.17s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/an_assignment_to_a_captured_local (0.09s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/returned (0.22s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/an_increment (0.09s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/new_Set (0.08s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_compound_assignment (0.16s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/a_destructuring_target (0.27s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/new_Map (0.17s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/Array.from (0.20s)
--- FAIL: TestNestedGeneratorRoadsAreNotYet/a_shorthand_property,_called_through (0.08s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/inside_an_async_function (0.07s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/the_function_as_a_value (0.24s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/inside_a_generator (0.13s)
--- PASS: TestNestedGeneratorRoadsAreNotYet/in_a_block (0.09s)
--- FAIL: TestNestedGeneratorRoadsAreNotYet/a_shorthand_property,_only_typeof_read (0.15s)
```

## Step 4: full lowerer suite

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

exit=0 wall_seconds=2351. Failing tests: **none**. Environment failures: **none**.

```text
ok  	github.com/system-inc/adamic/internal/lower	2344.806s
```

## Step 4b: full IR suite

```sh
go test ./internal/ir -count=1 -timeout 30m
```

exit=0 wall_seconds=203. Failing tests: **none**.

```text
ok  	github.com/system-inc/adamic/internal/ir	158.343s
```

## Step 6: full oracle suite

```sh
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```

exit=1 wall_seconds=3325.
Every failing test and its first error line:

| Test | First error | Environment? |
|---|---|---|
| `TestAssignmentProofWiderTarget` | assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofWiderTargetMutant` | assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofCompoundMutant` | assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofCompound` | assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofChainedMutant` | assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofChained` | assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofReturnDefinedMutant` | assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofReturnDefined` | assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofIfNarrowingMutant` | assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofIfNarrowing` | assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofScannerKeywordMutant` | assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestAssignmentProofScannerKeyword` | assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: TypeScript module lookup |
| `TestStdoutColumnsRefusesWASI` | library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT | Yes: WASI_SYSROOT missing |

Oracle log (failure diagnostics in full; trailing spaces on blank lines removed):

```text
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestAssignmentProofWiderTarget (0.04s)
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTargetMutant (0.13s)
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.05s)
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.09s)
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.05s)
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.07s)
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.08s)
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.07s)
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.06s)
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.05s)
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.07s)
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.07s)
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^

        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }

        Node.js v24.19.0
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestStdoutColumnsRefusesWASI (0.79s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
FAIL
gate cache: native hits=3083 misses=6594
gate cache: node hits=1747 misses=6612
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	3315.749s
FAIL
```

### Environment-only follow-up

All twelve assignment-proof failures contain `Error: Cannot find module 'typescript'`. The initial follow-up with external NODE_PATH also failed (exit=1 wall_seconds=8), because `assignmentProofTypecheck` explicitly overwrites NODE_PATH with the absent repository `stage3/api/node_modules`. An external wrapper at `/tmp/shorthand-proof/node-wrapper/node` then exported `NODE_PATH=/tmp/shorthand-proof/api/node_modules` before execing the real `/workspace/adamic-tools/bin/node`. No repository dependency directory or source was changed. Corrected follow-up: exit=0 wall_seconds=17. Every follow-up result:

```text
--- PASS: TestAssignmentProofWiderTargetMutant (8.68s)
--- PASS: TestAssignmentProofCompoundMutant (8.69s)
--- PASS: TestAssignmentProofWiderTarget (8.78s)
--- PASS: TestAssignmentProofChained (8.84s)
--- PASS: TestAssignmentProofScannerKeyword (9.20s)
--- PASS: TestAssignmentProofIfNarrowing (2.76s)
--- PASS: TestAssignmentProofCompound (3.21s)
--- PASS: TestAssignmentProofScannerKeywordMutant (3.27s)
--- PASS: TestAssignmentProofReturnDefinedMutant (3.53s)
--- PASS: TestAssignmentProofReturnDefined (3.08s)
--- PASS: TestAssignmentProofChainedMutant (1.92s)
--- PASS: TestAssignmentProofIfNarrowingMutant (1.63s)
```

`TestStdoutColumnsRefusesWASI` remains an environment failure: WASI_SYSROOT was not provisioned. The original full-suite result remains FAIL; no full suite rerun is claimed.

## Publication

Only this report is intended for commit, on new branch `proof/shorthand-3-report` created directly from the checked-out proof HEAD. No compiler changes or other proof artifacts are included.
