# Main whole-gate skip enforcement

Base: origin/main `48c05d091f0a43c31cbe051b1d6578d99eeedf19`.
Source: area/developer-tools `08e0db20`, limited to this package. The table
was regenerated for main: 70 actual skips, with no area-only inventory rows.
The package and command use only the Go standard library. No adamic-gate or
other area-only implementation is needed. This prefix precedes degraded-input
log detection, which is a separate unit.

After integration's normal whole `go test -json` run, run from the repo root:

```sh
go run ./internal/skipcensus/cmd <log.jsonl>
```

Keep the original test verdict as well. This command checks skip policy, not
whether tests failed, whether a run completed, or whether all tests were selected.
It first validates source against the declarations, then prints every skip's
class, package, test, identity and provider/rationale. Required-input and unknown
skips exit nonzero. Required providers name the variable and input, including
ADAMIC_TYPESCRIPT_SOURCE for TestCompilerAndStage1Agree. Lines do not affect
identity. No cache is added.

## Observed proofs

`go test -v -count=1 ./internal/skipcensus/...` and
`go vet ./internal/skipcensus/...` pass, with output redirected to
`/tmp/census-main-tests.log` and `/tmp/census-main-vet.log`.

The complete real `gate-out/test.jsonl` from
`gate-logs/2e165469ec95/plain` (archive ref ae33bd09) exits 1:
`skips=33 required-input=17 unknown=0`. Output names
TestCompilerAndStage1Agree and ADAMIC_TYPESCRIPT_SOURCE, with its v6.0.3
source pin. `/tmp/census-main-missing-check.log` records the command output.

The complete real log from `gate-logs/47fbaf174d168f14/plain` (archive ref
89eb388f) contains a PASS for TestCompilerAndStage1Agree and no required-input
skip. It passes with `skips=25 required-input=0 unknown=0` when checked against
its matching area source snapshot and declarations:

```sh
go run ./internal/skipcensus/cmd -root /tmp/census-main-historical -table /tmp/census-main-historical-table.json /tmp/census-main-provided-new.jsonl
```

Those scratch inputs were extracted unchanged from `08e0db20`. Five area-only
skips in this historical log have no corresponding site in main; the default
main table correctly refuses them as unknown. They are platform/opt-in/timing
exclusions, not missing required inputs. Proof-only declarations for these five
sites live in provided-historical-rows.json, separate from skips.json. Regression
tests assert both strict refusal with main's table and acceptance with the
historical declarations. The compact log fixture preserves skip output/events
and the compiler-agreement PASS from the real log. Do not add those historical
rows to main's live census to manufacture a green result.

The older `gate-logs/47fbaf174d16/plain` still contains the real WASI builtins
skip and is correctly rejected; it is not the passing witness.

The existing proofs.py script was run on this prefix: an added scratch t.Skip
fails TestCensus naming TestUndeclaredWitness; independently allowing required,
added or removed skips fails the respective regression test. Omitting the
provider in output fails TestRequiredSkipNamesInput. All mutants were caught.
Logs are `/tmp/census-main-mutants.log` and
`/tmp/census-main-input-name-mutant.log`.

Setup exited 0; nproc=5. Its diagnostic timings were Go ready 0.025s, clang ready
0.203s, Node ready 0.020s, submodules ready 3.703s, build ready 172.051s,
cache warm 172.200s and done 172.226s. Exact lines and build flags are in
`/tmp/census-main-setup.log`. Build-flags context: base SHA above, cpu.max
400000 100000, Go 1.27.1, clang 20.1.8, Node v24.19.0, cached setup;
load before 0.05 0.03 0.00 and after 5.79 2.91 1.12. These are setup diagnostics,
not before/after speed measurements. The full live main gate was not rerun.
