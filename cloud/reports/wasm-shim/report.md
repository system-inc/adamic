Corrected the WASI shim oracle to use the JavaScript backend for checked fixtures, with a wrong-source control.
Commits: prior report 01deb14f8ae47edae318ac0e5a66919dec717dd1; witness correction 6f8090e5790bbd219fba07ccf140c94045977f5f.
Observed: corrected strict suite passes in 104.388s, with 305 three-way agreements and 3 file-access skips; package vet passes.
Controls: writes_past_end.a source exits 0 while backend/shim exit 70; all three iovec, exit, and preopen mutants remain caught.
Not covered: actual browser/Workers/Deno deployment, filesystem support, clocks/randomness, or another complete repository gate.

# W2 report

## Corrected strict rerun

The user clarified the expected witness: fixtures with `checked=true` must use
`onJavaScriptBackend`, exactly as `internal/oracle/wasi_test.go` does. Ordinary
fixtures continue to use source Node. The strict test still compares stdout,
stderr and exit code byte for byte under the shim, Node WASI, and that expected
witness. It now records actual agreements separately from fixtures run.

The corrected strict suite was run against witness-correction commit
`6f8090e5790bbd219fba07ccf140c94045977f5f`:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIShim' \
  -count=1 -v -timeout 30m > /tmp/wasm-shim-corrected-oracle.log 2>&1
# exit 0, ok internal/oracle 104.388s
# SHIM COUNTS run=305 agreed=305 skipped=3

go vet ./internal/oracle > /tmp/wasm-shim-corrected-vet.log 2>&1
# exit 0, no diagnostics

git diff --check
# exit 0, no diagnostics
```

All 305 eligible fixtures agree three ways: 295 use source Node and 10 use the
checked backend. The three skipped file fixtures remain `write_stdout_order.a`,
`write_stderr_order.a`, and `prompt_then_read.a`; their exact imports and the
unchanged skip rule appear below.

`TestWASIShimCheckedWitnessControl` passed in 0.24s. It verifies the existing
`writes_past_end.a` entry is still checked, requires backend/shim agreement, and
requires a source/shim comparison to report `exit codes differ`, with source exit
0 and shim exit 70. This keeps the wrong-source witness detectable.
`TestWASIShimContractsAndMutants` passed in 0.20s, catching each of the three
requested mutants at its named assertion. `TestWASIShimRequest` passed in 0.27s.

The original measurements below are retained as history. The source-only
limitation is resolved by the clarified witness; there is no remaining red
shim gate. No new toolchain setup or full native/repository gate was needed for
this witness-only change. The original full oracle and existing WASI suite
results are recorded below, and the complete shim suite was rerun uncached.

Branch: `codex/wasm-shim`, starting from `origin/wasm/integrate` at
`6f7dce3dc1eace606fe081c8f4ab12034ae11bb4`. The first pushed commit contained
only `cloud/reports/wasm-shim/claim.md`. No main or area branch was pushed.

`git fetch origin` initially fetched only main because this clone's fetch
refspec is narrow. An explicit fetch of
`refs/heads/wasm/integrate:refs/remotes/origin/wasm/integrate` established the
required base. `git fetch origin && git merge origin/main` subsequently said
`Already up to date.` A separate `git ls-remote --heads origin main` confirmed
main remained `e8ba3d5d81de4d3773c723914fccd4c76248b965`, already an ancestor.
No rebase or history rewrite occurred.

## Setup and scope

Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, docs/wasm.md, and the
existing WASI oracle and runner before editing. Changes are confined to the
claimed files; docs/wasm.md only gains its appended Imports section.

`bash cloud/setup.sh --wasi-sdk > /tmp/wasm-shim-setup.log 2>&1` exited 0.
Its timing lines were:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: wasi sdk ready (/workspace/adamic-tools/wasi-sdk) (4s)
setup: submodules ready (4s)
setup: build cache warm (123s)
setup: done in 123s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` = 5. Go 1.27.1, Node 24.19.0, native clang 20.1.8, WASI SDK 27.
Every build/test shell sourced `/workspace/adamic-tools/env.sh`, which exports
`WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot`.
The first existing native `TestWASI` invocation skipped: native clang on PATH
could not open its WASI `libclang_rt.builtins.a`. Rerunning with SDK bin first on
PATH passed. `native.Build` itself selects SDK clang beside the sysroot.

## Initial commands and observed output

All test output went directly to log files, without a pipeline.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_WASI=1 ADAMIC_SHIM_INVENTORY_ONLY=1 go test ./internal/oracle \
  -run '^TestWASIShimAgreesWithNode$' -count=1 -v -timeout 30m \
  > /tmp/wasm-shim-inventory.log 2>&1
# exit 0, ok internal/oracle 95.157s; run=305 skipped=3 (inventory only)

ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIShim' \
  -count=1 -v -timeout 30m > /tmp/wasm-shim-oracle.log 2>&1
# exit 1, 118.188s; 305 Node WASI comparisons agree, 10 source assertions fail

ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIShimRequest$' \
  -count=1 -v -timeout 15m > /tmp/wasm-shim-request-parity.log 2>&1
# exit 0, ok internal/oracle 0.290s

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m \
  > /tmp/wasm-shim-native-oracle.log 2>&1
# exit 0, ok internal/oracle 95.139s; WASI opt-in tests skip in this run

go vet ./... > /tmp/wasm-shim-vet.log 2>&1
# exit 0, no diagnostics

gofmt -l cmd internal > /tmp/wasm-shim-format.log
# exit 0, empty output

git diff --check
# exit 0, empty output
```

The existing runtime test was run with SDK clang first:

```sh
source /workspace/adamic-tools/env.sh
export PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH
ADAMIC_TEST_WASI=1 ADAMIC_WASI_ARTIFACT=/tmp/wasm-shim-request.wasm \
  go test ./internal/native -run '^TestWASI$' -count=1 -v -timeout 15m \
  > /tmp/wasm-shim-request.log 2>&1
# exit 0, 48 strict runtime translation units; requests passes; total 54.699s
```

Contract and mutant controls passed in 0.25s. The first reactor lifetime control
incorrectly expected zero live values after each request; both hosts retained
the module's two initialized globals. The corrected check records the initialized
baseline and requires every request to return to it. Six varied requests then
matched both hosts and the source, with no additional live values.

## Initial source-only run, superseded by the clarified brief

The 10 initially failing fixtures all have `checked=true` in the existing oracle.
For each, source Node exits 0 while the compiled command exits 70 with its
inserted panic. Node WASI and the shim agree exactly, including that panic.
The existing `wasi_test.go` deliberately compares these fixtures to the
JavaScript backend, which carries the same checks. The initial brief instead
required Node running source for every file-free fixture. The initial test stayed
red and reported those mismatches. The user subsequently clarified that checked
fixtures must use the JavaScript backend, matching the existing oracle.

- `internal/oracle/testdata/writes_past_end.a`
- `internal/oracle/testdata/cast_fails.a`
- `internal/oracle/testdata/map_shrinks.a`
- `internal/oracle/testdata/find_shrinks.a`
- `internal/oracle/testdata/find_index_shrinks.a`
- `internal/oracle/testdata/narrowed_numbers.a`
- `internal/oracle/testdata/write_after_shrink.a`
- `internal/oracle/testdata/e4eec87_f1_field_narrowed.a`
- `internal/oracle/testdata/e4eec87_f1_class_narrowed.a`
- `internal/oracle/testdata/e4eec87_f1_alias_narrowed.a`

Initial strict outcomes: 305 commands run, 3 skipped, 295 full three-way passes,
10 source mismatches, zero shim vs Node WASI mismatches. Request reactor parity
passes separately. There are no extra behavior-based skips. The file rule is
conservative: any `path_*`, `fd_read`, `fd_readdir`, or `fd_filestat_get` import.
The skipped names are `write_stdout_order.a`, `write_stderr_order.a`, and
`prompt_then_read.a`, each under `internal/oracle/testdata/`. Their exact imports
are recorded below. File access cannot be honestly provided by this host.

The complete repository test gate was not run. The full touched oracle package,
existing WASI suite, and focused runtime request test are the chosen gate scope.
The original source-only run was not green. The corrected witness policy and
its rerun are documented in the follow-up results.

## Earlier established WASI gate

After the `origin/main` merge check, the existing oracle convention and all
shim contracts/request checks passed:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle \
  -run '^TestWASI(AgreesWithNode|OracleCatchesMutants|RunnerCatchesMutants|Emission|ShimContractsAndMutants|ShimRequest)$' \
  -count=1 -v -timeout 30m > /tmp/wasm-shim-existing-wasi-gate.log 2>&1
# exit 0, ok internal/oracle 165.886s
```

This suite uses the repository's existing JavaScript-backend witness for the
10 checked fixtures. This earlier suite did not include
`TestWASIShimAgreesWithNode`, which still used the initial source-only brief
at that time. Its log includes
passing `TestWASIShimRequest` and all four contract subtests, with all three
requested mutants caught by their named AssertionErrors. The earlier
`/tmp/wasm-shim-contracts.log` also preserves the corrected request-baseline
control failure, and is not claimed as a complete-suite pass.

## Import inventory and per-import policy

The following table and policies are the same appended Imports section saved
in docs/wasm.md. These measurements precede shim implementation.

### Imports

W2 inventory, October 7, 2026, on `wasm/integrate` at
`6f7dce3dc1eace606fe081c8f4ab12034ae11bb4`, WASI SDK 27 and Node 24.19.0.
The inventory uses `WebAssembly.Module.imports` on each actual linked module,
filters `wasi_snapshot_preview1`, and sorts names. No imports are inferred from
C sources. Commands use `native.Build` with `wasm32-wasi`; the compiler request
reactor uses `native.WASI` and `Request: true, Count: true`. The earlier request
prototype comes from `TestWASI/requests` with `ADAMIC_WASI_ARTIFACT`.

These are exact sets, not subsets. The fixture table assigns one set per module.

| Set | Exact `wasi_snapshot_preview1` imports |
| --- | --- |
| A | `args_get, args_sizes_get, fd_close, fd_fdstat_get, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, proc_exit` |
| B | `args_get, args_sizes_get, fd_close, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, proc_exit` |
| R | `fd_close, fd_fdstat_get, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, proc_exit` |
| P | `fd_close, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, proc_exit` |

`internal/native/wasm/shim.mjs` is 38 physical lines including comments and
blank lines, has no Node imports, and exports `wasiShim({ stdout, stderr })`.
Instantiate with `shim.imports`, call `shim.attach(instance)`, then call `_start`
for a command or `_initialize` once for a reactor before using request exports.
The caller must discard an instance after `AdamicExit`; the shim does not unwind
or reset libc state. Callbacks synchronously consume owned `Uint8Array` copies.
They receive raw bytes, preserving NULs, malformed UTF-8 and iovec boundaries.
Do not decode each callback separately if UTF-8 may span iovecs.

| Import | Choice and reason |
| --- | --- |
| `fd_write` | Decode all wasm32 little-endian pointer/length iovecs, copy their bytes to stdout for fd 1 or stderr for fd 2, write the total byte count, return 0. Other fds return EBADF (8). Reacquire memory views for every call and after callbacks. |
| `proc_exit` | Throw exported `AdamicExit`, with `.name = 'AdamicExit'` and `.code` set to the WASI exit code, so termination cannot silently resume. |
| `fd_prestat_get` | EBADF (8), so libc discovers no preopened directories. |
| `fd_prestat_dir_name`, `fd_close`, `fd_seek`, `fd_fdstat_get` | EBADF (8). No directory, seekable file or terminal descriptor metadata is available. Output still works via `fd_write`. |
| Every other `fd_*` and `path_*`, including `fd_read`, `fd_filestat_get`, `fd_readdir`, `path_open` | EBADF (8), without touching output memory or granting host file access. |
| `args_sizes_get`, `environ_sizes_get` | Write zero count and zero buffer size, return 0: this host has an empty argument vector and environment. |
| `args_get`, `environ_get` | Return 0 without writes: zero entries require zero bytes. |
| `clock_time_get`, `random_get` | Throw `AdamicUnsupportedWASI` naming the import. Fake time and predictable random bytes would not satisfy those services. Neither occurs in the inventoried modules. |
| Other imports | Throw `AdamicUnsupportedWASI` naming the import; no success stub. |

No inventoried file-free fixture requires an import that the shim cannot
satisfy. File fixtures cannot run honestly without file access. The test skips
modules importing any `path_*`, `fd_read`, `fd_readdir` or `fd_filestat_get`.
That conservative artifact rule skips exactly these three fixtures:

| Skipped fixture | Exact imports |
| --- | --- |
| `internal/oracle/testdata/write_stdout_order.a` | `args_get, args_sizes_get, fd_close, fd_fdstat_get, fd_filestat_get, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, path_open, proc_exit` |
| `internal/oracle/testdata/write_stderr_order.a` | `args_get, args_sizes_get, fd_close, fd_fdstat_get, fd_filestat_get, fd_prestat_dir_name, fd_prestat_get, fd_seek, fd_write, path_open, proc_exit` |
| `internal/oracle/testdata/prompt_then_read.a` | `args_get, args_sizes_get, fd_close, fd_fdstat_get, fd_filestat_get, fd_prestat_dir_name, fd_prestat_get, fd_read, fd_seek, fd_write, path_open, proc_exit` |

The differential command test uses the same expected witness as
`internal/oracle/wasi_test.go`: Node running source for ordinary fixtures, and
`onJavaScriptBackend` for the 10 fixtures with `checked=true`. The backend
carries Adamic's inserted checks; source Node continues where those checks panic.
The test reports both commands run and actual three-way agreements, plus skips.
The corrected strict run passed with `run=305 agreed=305 skipped=3`: 295
source witnesses and 10 checked backend witnesses, with no mismatches.

`TestWASIShimCheckedWitnessControl` holds `writes_past_end.a` to the checked
backend and separately compares it to source Node. The backend and shim must
agree with exit 70; source must exit 0 and produce `exit codes differ`. Choosing
the source witness for a checked fixture therefore remains a detected error.

`TestWASIShimRequest` compares the compiler reactor under the shim, Node WASI,
and the source handler: empty input, Unicode, echo/reassignment, 1 KiB input,
NULs, dependency/module initialization, and stable counted live values relative
to the module's initialized globals. It passes. The earlier prototype's
100,000-request test also passes under its existing Node WASI host.

| Mutant | Named check that fails |
| --- | --- |
| `fd_write` omits the last iovec | `TestWASIShimContractsAndMutants/iovecs`, assertion `all iovecs, exact bytes` |
| `proc_exit` returns instead of throwing | `TestWASIShimContractsAndMutants/exit`, assertion `exit must terminate` |
| `fd_prestat_get` returns success and a fake directory | `TestWASIShimContractsAndMutants/preopens`, assertion `no preopened directory` |

Every control passed; every isolated mutant process exited 1 with the named
AssertionError. The test also checks empty environment/arguments, named traps,
stdout and stderr independently, byte counts, and memory growth.

Run the command comparison and checked-witness, contract, and request checks with:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIShim' \
  -count=1 -v -timeout 30m > /tmp/wasm-shim-oracle.log 2>&1
```

`ADAMIC_SHIM_INVENTORY_ONLY=1` collects the command import table without running
comparisons. Do not use that option as evidence of execution parity.

| File-free module | Import set |
| --- | --- |
| `internal/native/wasm/request.a` prototype reactor | P |
| `cmd/adamic/testdata/wasi/request.a` compiler reactor | R |
| `internal/oracle/testdata/call_targets_element.a` command | A |
| `internal/oracle/testdata/call_targets_region.a` command | A |
| `internal/oracle/testdata/call_targets_reuse.a` command | A |
| `internal/oracle/testdata/call_targets_closure.a` command | A |
| `internal/oracle/testdata/call_targets_sort.a` command | A |
| `internal/oracle/testdata/library_object_keys.a` command | A |
| `internal/oracle/testdata/library_object_is.a` command | A |
| `internal/oracle/testdata/library_object_has_own.a` command | A |
| `internal/oracle/testdata/library_object_assign.a` command | A |
| `internal/oracle/testdata/library_object_freeze.a` command | A |
| `internal/oracle/testdata/library_object_freeze_write.a` command | A |
| `internal/oracle/testdata/library_object_order.a` command | A |
| `internal/oracle/testdata/library_object_assign_fields.a` command | A |
| `internal/oracle/testdata/library_object_freeze_alias.a` command | A |
| `internal/oracle/testdata/library_object_freeze_assign.a` command | A |
| `internal/oracle/testdata/library_object_same.a` command | A |
| `internal/oracle/testdata/library_object_own.a` command | A |
| `dedication/dedication.a` command | A |
| `cmd/adamic/testdata/wasi/request.a` command | A |
| `internal/oracle/testdata/class_oct6_deep.a` command | A |
| `internal/oracle/testdata/class_oct6_parameters.a` command | A |
| `internal/oracle/testdata/class_oct6_release.a` command | A |
| `internal/oracle/testdata/class_oct6_subclass_holder.a` command | A |
| `internal/oracle/testdata/library_array_join.a` command | A |
| `internal/oracle/testdata/library_array_iterators.a` command | A |
| `internal/oracle/testdata/library_array_metadata.a` command | A |
| `internal/oracle/testdata/library_array_with.a` command | A |
| `internal/oracle/testdata/library_array_flat_map.a` command | A |
| `internal/oracle/testdata/library_array_flat.a` command | A |
| `internal/oracle/testdata/library_array_spliced.a` command | A |
| `internal/oracle/testdata/library_array_copy_within.a` command | A |
| `internal/oracle/testdata/library_array_search.a` command | A |
| `internal/oracle/testdata/library_array_copy.a` command | A |
| `internal/oracle/testdata/library_array_find_last.a` command | A |
| `internal/oracle/testdata/json_stringify_scalars.a` command | A |
| `internal/oracle/testdata/json_stringify_values.a` command | A |
| `internal/oracle/testdata/json_stringify_options.a` command | A |
| `internal/oracle/testdata/json_stringify_escapes.a` command | A |
| `internal/oracle/testdata/json_stringify_numbers.a` command | A |
| `internal/oracle/testdata/json_stringify_undefined.a` command | A |
| `internal/oracle/testdata/json_stringify_indent.a` command | A |
| `internal/oracle/testdata/json_stringify_keys.a` command | A |
| `internal/oracle/testdata/json_stringify_replacer.a` command | A |
| `internal/oracle/testdata/library_function_expressions.a` command | A |
| `internal/oracle/testdata/library_fnexpr_recurse.a` command | A |
| `internal/oracle/testdata/library_fnexpr_store.a` command | A |
| `internal/oracle/testdata/library_fnexpr_loops.a` command | A |
| `internal/oracle/testdata/library_for_in.a` command | A |
| `internal/oracle/testdata/library_for_in_keys.a` command | A |
| `internal/oracle/testdata/library_for_in_live.a` command | A |
| `internal/oracle/testdata/library_globals.a` command | A |
| `internal/oracle/testdata/library_globals_typeof.a` command | A |
| `internal/load/testdata/0.1/compile/01_hello.ts` command | A |
| `internal/load/testdata/0.1/compile/02_fizzbuzz.ts` command | A |
| `internal/load/testdata/0.1/compile/03_shapes.ts` command | A |
| `internal/load/testdata/0.1/compile/04_closures.ts` command | A |
| `internal/load/testdata/0.1/compile/05_wordcount.ts` command | A |
| `internal/load/testdata/0.1/compile/06_stack.ts` command | A |
| `internal/load/testdata/0.1/compile/07_modules/main.ts` command | A |
| `internal/load/testdata/0.1/compile/08_results.ts` command | A |
| `internal/load/testdata/0.1/compile/09_tree.ts` command | A |
| `internal/load/testdata/0.1/compile/10_unicode.ts` command | A |
| `internal/oracle/testdata/strings.a` command | A |
| `internal/oracle/testdata/numbers.a` command | A |
| `internal/oracle/testdata/bitwise_sweep.a` command | A |
| `internal/oracle/testdata/loops.a` command | A |
| `internal/oracle/testdata/booleans.a` command | A |
| `internal/oracle/testdata/shadowing.a` command | A |
| `internal/oracle/testdata/functions.a` command | A |
| `internal/oracle/testdata/effects.a` command | A |
| `internal/oracle/testdata/dead_zone.a` command | A |
| `internal/oracle/testdata/objects.a` command | A |
| `internal/oracle/testdata/modules/main.a` command | A |
| `internal/oracle/testdata/panic.a` command | A |
| `internal/oracle/testdata/maps_and_text.a` command | A |
| `internal/oracle/testdata/lone_surrogates.a` command | A |
| `internal/oracle/testdata/sorting.a` command | A |
| `internal/oracle/testdata/classes.a` command | A |
| `internal/oracle/testdata/closures.a` command | A |
| `internal/oracle/testdata/local_console.a` command | B |
| `internal/oracle/testdata/indexing.a` command | A |
| `internal/oracle/testdata/writes.a` command | A |
| `internal/oracle/testdata/writes_past_end.a` command | A |
| `internal/oracle/testdata/casts.a` command | A |
| `internal/oracle/testdata/cast_fails.a` command | A |
| `internal/oracle/testdata/updates.a` command | A |
| `internal/oracle/testdata/read_order.a` command | A |
| `internal/oracle/testdata/string_index.a` command | A |
| `internal/oracle/testdata/visits.a` command | A |
| `internal/oracle/testdata/searches.a` command | A |
| `internal/oracle/testdata/spreads.a` command | A |
| `internal/oracle/testdata/maybe_numbers.a` command | A |
| `internal/oracle/testdata/defaults.a` command | A |
| `internal/oracle/testdata/search_halves.a` command | A |
| `internal/oracle/testdata/strings_more.a` command | A |
| `internal/oracle/testdata/number_parsing.a` command | A |
| `internal/oracle/testdata/library_math_number_math.a` command | A |
| `internal/oracle/testdata/library_math_number_convert.a` command | A |
| `internal/oracle/testdata/library_math_number_prototype.a` command | A |
| `internal/oracle/testdata/navigation.a` command | A |
| `internal/oracle/testdata/number_formats.a` command | A |
| `internal/oracle/testdata/precision_range.a` command | A |
| `internal/oracle/testdata/radixes.a` command | A |
| `internal/oracle/testdata/radix_range.a` command | A |
| `internal/oracle/testdata/optional_numbers.a` command | A |
| `internal/oracle/testdata/map_iteration.a` command | A |
| `internal/oracle/testdata/sorts.a` command | A |
| `internal/oracle/testdata/sort_releases.a` command | A |
| `internal/oracle/testdata/timsort.a` command | A |
| `internal/oracle/testdata/unused_parameters.a` command | A |
| `internal/oracle/testdata/map_shrinks.a` command | A |
| `internal/oracle/testdata/find_shrinks.a` command | A |
| `internal/oracle/testdata/find_index_shrinks.a` command | A |
| `internal/oracle/testdata/self_assignments.a` command | A |
| `internal/oracle/testdata/method_closures.a` command | A |
| `internal/oracle/testdata/splice_empty.a` command | A |
| `internal/oracle/testdata/narrowed_reads.a` command | A |
| `internal/oracle/testdata/narrowed_writes.a` command | A |
| `internal/oracle/testdata/narrowed_methods.a` command | A |
| `internal/oracle/testdata/narrowed_fields.a` command | A |
| `internal/oracle/testdata/narrowed_numbers.a` command | A |
| `internal/oracle/testdata/narrowed_compared.a` command | A |
| `internal/oracle/testdata/sort_top_level.a` command | A |
| `internal/oracle/testdata/splices.a` command | A |
| `internal/oracle/testdata/fills.a` command | A |
| `internal/oracle/testdata/fill_length.a` command | A |
| `internal/oracle/testdata/array_from.a` command | A |
| `internal/oracle/testdata/array_from_length.a` command | A |
| `internal/oracle/testdata/array_from_undefined.a` command | A |
| `internal/oracle/testdata/weak_parent.a` command | A |
| `internal/oracle/testdata/doubly_linked.a` command | A |
| `internal/oracle/testdata/fresh_parser.a` command | A |
| `internal/oracle/testdata/fresh_writes.a` command | A |
| `internal/oracle/testdata/fresh_calls.a` command | A |
| `internal/oracle/testdata/weak_narrowed.a` command | A |
| `internal/oracle/testdata/exceptions.a` command | A |
| `internal/oracle/testdata/exceptions_uncaught.a` command | A |
| `internal/oracle/testdata/exceptions_empty.a` command | A |
| `internal/oracle/testdata/closures_throw.a` command | A |
| `internal/oracle/testdata/closures_throw_uncaught.a` command | A |
| `internal/oracle/testdata/finally_leaves.a` command | A |
| `internal/oracle/testdata/reuse_foreach_global.a` command | A |
| `internal/oracle/testdata/param_assigned_in_try.a` command | A |
| `internal/oracle/testdata/named_function_values.a` command | A |
| `internal/oracle/testdata/panic_in_try.a` command | A |
| `internal/oracle/testdata/invariance_readonly.a` command | A |
| `internal/oracle/testdata/tuples_kept.a` command | A |
| `internal/oracle/testdata/undefined_keys.a` command | A |
| `internal/oracle/testdata/undefined_strings.a` command | A |
| `internal/oracle/testdata/maybe_booleans.a` command | A |
| `internal/oracle/testdata/maybe_boolean_panic.a` command | A |
| `internal/oracle/testdata/unions.a` command | A |
| `internal/oracle/testdata/maybe_number_slots.a` command | A |
| `internal/oracle/testdata/case_mapping.a` command | A |
| `internal/oracle/testdata/undefined_elements.a` command | A |
| `internal/oracle/testdata/map_zero_keys.a` command | A |
| `internal/oracle/testdata/string_limits.a` command | A |
| `internal/oracle/testdata/string_too_long.a` command | A |
| `internal/oracle/testdata/pad_too_long.a` command | A |
| `internal/oracle/testdata/stack_overflow.a` command | A |
| `internal/oracle/testdata/adversarial_order.a` command | A |
| `internal/oracle/testdata/adversarial_exits.a` command | A |
| `internal/oracle/testdata/adversarial_iteration.a` command | A |
| `internal/oracle/testdata/number_edges.a` command | A |
| `internal/oracle/testdata/long_chain.a` command | A |
| `internal/oracle/testdata/write_after_shrink.a` command | A |
| `internal/oracle/testdata/normalize.a` command | A |
| `internal/oracle/testdata/normalize_form.a` command | A |
| `internal/oracle/testdata/string_positions.a` command | A |
| `internal/oracle/testdata/long_literals.a` command | A |
| `internal/oracle/testdata/class_layouts.a` command | A |
| `internal/oracle/testdata/ascii_scan.a` command | A |
| `internal/oracle/testdata/size_class_churn.a` command | A |
| `internal/oracle/testdata/borrow_reassigned.a` command | A |
| `internal/oracle/testdata/borrow_defined_lent.a` command | A |
| `internal/oracle/testdata/borrow_defined_lent_field.a` command | A |
| `internal/oracle/testdata/writes_in_try.a` command | A |
| `internal/oracle/testdata/class_as_interface.a` command | A |
| `internal/oracle/testdata/optional_class_method.a` command | A |
| `internal/oracle/testdata/set_undefined.a` command | A |
| `internal/oracle/testdata/borrow_map_overwrite.a` command | A |
| `internal/oracle/testdata/spread_snapshot.a` command | A |
| `internal/oracle/testdata/reuse.a` command | A |
| `internal/oracle/testdata/reuse_arrays.a` command | A |
| `internal/oracle/testdata/reuse_forward.a` command | A |
| `internal/oracle/testdata/reuse_global_sibling.a` command | A |
| `internal/oracle/testdata/reuse_weak_during_spread.a` command | A |
| `internal/oracle/testdata/reuse_weak_after_reuse.a` command | A |
| `internal/oracle/testdata/regions.a` command | A |
| `internal/oracle/testdata/regions_throw.a` command | A |
| `internal/oracle/testdata/regions_constructor_capture.a` command | A |
| `internal/oracle/testdata/borrow_element.a` command | A |
| `internal/oracle/testdata/borrow_element_throw.a` command | A |
| `internal/oracle/testdata/borrow_element_virtual_store.a` command | A |
| `internal/oracle/testdata/borrow_element_virtual_move.a` command | A |
| `internal/oracle/testdata/borrow_element_super_move.a` command | A |
| `internal/oracle/testdata/move_throw.a` command | A |
| `internal/oracle/testdata/throw_keeps_old_value.a` command | A |
| `internal/oracle/testdata/throw_keeps_old_value_variants.a` command | A |
| `internal/oracle/testdata/throw_in_writes.a` command | A |
| `internal/oracle/testdata/throw_global_move.a` command | A |
| `internal/oracle/testdata/spread_undefined.a` command | A |
| `internal/oracle/testdata/lent_reads.a` command | A |
| `internal/oracle/testdata/large_output.a` command | A |
| `internal/oracle/testdata/output_then_panic.a` command | A |
| `internal/oracle/testdata/interleaved.a` command | A |
| `internal/oracle/testdata/trig_reduction.a` command | A |
| `internal/oracle/testdata/sets.a` command | A |
| `internal/oracle/testdata/set_maybe_numbers.a` command | A |
| `internal/oracle/testdata/library_map_set.a` command | A |
| `internal/oracle/testdata/library_map_set_keys.a` command | A |
| `internal/oracle/testdata/library_map_set_iterators.a` command | A |
| `internal/oracle/testdata/library_map_set_construct.a` command | A |
| `internal/oracle/testdata/library_map_set_group_by.a` command | A |
| `internal/oracle/testdata/maybe_collections.a` command | A |
| `internal/oracle/testdata/stack_tail_call.a` command | A |
| `internal/oracle/testdata/stack_forever.a` command | A |
| `internal/oracle/testdata/optional_strings.a` command | A |
| `internal/oracle/testdata/concat_too_long.a` command | A |
| `internal/oracle/testdata/replace_all_large.a` command | A |
| `internal/oracle/testdata/collections.a` command | A |
| `internal/oracle/testdata/gaps.a` command | A |
| `internal/oracle/testdata/normalize_long_marks.a` command | A |
| `internal/oracle/testdata/power_of_two_string.a` command | A |
| `internal/oracle/testdata/declared_later.a` command | A |
| `internal/oracle/testdata/return_panic.a` command | A |
| `internal/oracle/testdata/return_panic_fires.a` command | A |
| `internal/oracle/testdata/from_codes.a` command | A |
| `internal/oracle/testdata/from_code_point_fails.a` command | A |
| `internal/oracle/testdata/bitwise.a` command | A |
| `internal/oracle/testdata/tuple_values.a` command | A |
| `internal/oracle/testdata/generic_functions.a` command | A |
| `internal/oracle/testdata/generic_method_return.a` command | A |
| `internal/oracle/testdata/generic_values.a` command | A |
| `internal/oracle/testdata/undefined_references.a` command | A |
| `internal/oracle/testdata/spread_calls.a` command | A |
| `internal/oracle/testdata/utf8_view.a` command | A |
| `internal/oracle/testdata/utf8_view_fails.a` command | A |
| `internal/oracle/testdata/search_from.a` command | A |
| `internal/oracle/testdata/shared_slices.a` command | A |
| `internal/oracle/testdata/string_append.a` command | A |
| `internal/oracle/testdata/shared_slice_append.a` command | A |
| `internal/oracle/testdata/search_from_sweep.a` command | A |
| `internal/oracle/testdata/integer_format.a` command | A |
| `internal/oracle/testdata/reuse_throw.a` command | A |
| `internal/oracle/testdata/reuse_narrowed.a` command | A |
| `internal/oracle/testdata/reuse_lent_global.a` command | A |
| `internal/oracle/testdata/reuse_spread_method.a` command | A |
| `internal/oracle/testdata/reuse_spread_method_alias.a` command | A |
| `internal/oracle/testdata/try_assignments.a` command | A |
| `internal/oracle/testdata/library_string_conversion.a` command | A |
| `internal/oracle/testdata/library_string_prototype.a` command | A |
| `internal/oracle/testdata/library_string_indices.a` command | A |
| `internal/oracle/testdata/library_string_raw.a` command | A |
| `internal/oracle/testdata/library_string_existing.a` command | A |
| `internal/oracle/testdata/library_map_set_visit.a` command | A |
| `internal/oracle/testdata/library_map_set_setops.a` command | A |
| `internal/oracle/testdata/library_map_set_zeros.a` command | A |
| `internal/oracle/testdata/library_map_set_groupby_keys.a` command | A |
| `internal/oracle/testdata/library_map_set_next.a` command | A |
| `internal/oracle/testdata/has_own.a` command | A |
| `internal/oracle/testdata/regexp.a` command | A |
| `internal/oracle/testdata/sweeps/regexp_methods.a` command | A |
| `internal/oracle/testdata/regexp_matchall_nonglobal.a` command | A |
| `internal/oracle/testdata/regexp_replaceall_nonglobal.a` command | A |
| `internal/oracle/testdata/regexp_null_narrowed.a` command | A |
| `internal/oracle/testdata/regexp_replace.a` command | A |
| `internal/oracle/testdata/regexp_split.a` command | A |
| `internal/oracle/testdata/regexp_exec.a` command | A |
| `internal/oracle/testdata/regexp_match.a` command | A |
| `internal/oracle/testdata/regexp_search.a` command | A |
| `internal/oracle/testdata/regexp_unicode.a` command | A |
| `internal/oracle/testdata/regexp_split_pair_pattern.a` command | A |
| `internal/oracle/testdata/class_features_static.a` command | A |
| `internal/oracle/testdata/class_features_static_private.a` command | A |
| `internal/oracle/testdata/class_features_private.a` command | A |
| `internal/oracle/testdata/class_features_accessors.a` command | A |
| `internal/oracle/testdata/class_features_twice.a` command | A |
| `internal/oracle/testdata/class_features_retained.a` command | A |
| `internal/oracle/testdata/class_features_distinct.a` command | A |
| `internal/oracle/testdata/class_inheritance.a` command | A |
| `internal/oracle/testdata/class_inheritance_exceptions.a` command | A |
| `internal/oracle/testdata/class_inheritance_order.a` command | A |
| `internal/oracle/testdata/class_identity.a` command | A |
| `internal/oracle/testdata/class_inheritance_memory.a` command | A |
| `internal/oracle/testdata/class_inheritance_generic.a` command | A |
| `internal/oracle/testdata/class_inheritance_interface.a` command | A |
| `internal/oracle/testdata/class_inheritance_conditional.a` command | A |
| `internal/oracle/testdata/devirtualize.a` command | A |
| `internal/oracle/testdata/user_iterators.a` command | A |
| `internal/oracle/testdata/user_iterators_rest_tdz.a` command | A |
| `internal/oracle/testdata/e4eec87_u02_optional_absent.a` command | A |
| `internal/oracle/testdata/literal_optional_shapes.a` command | A |
| `internal/oracle/testdata/e4eec87_u03_discriminated_undefined.a` command | A |
| `internal/oracle/testdata/e4eec87_u01_undefined_field_widened.a` command | A |
| `internal/oracle/testdata/e4eec87_f1_field_narrowed.a` command | A |
| `internal/oracle/testdata/e4eec87_f1_class_narrowed.a` command | A |
| `internal/oracle/testdata/e4eec87_f1_alias_narrowed.a` command | A |
| `internal/oracle/testdata/e4eec87_f1_field_present.a` command | A |
| `internal/oracle/testdata/object_prototype.a` command | A |
| `internal/fresh/testdata/regexp_tree.ts` command | A |
| `internal/oracle/testdata/regexp_cycle_fields.a` command | A |
| `internal/oracle/testdata/regexp_cycle_collections.a` command | A |
| `internal/oracle/testdata/regexp_cycle_closures.a` command | A |
| `internal/oracle/testdata/regexp_cycle_weak.a` command | A |
