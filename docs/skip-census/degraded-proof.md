# Missing input diagnostics cannot pass the gate

Built from `e165f424f45c3d8d71942ef23de98d53b2f9d77a`, as authorized:
area/developer-tools at `6b337090` did not yet contain that census tip.
No stage1 tests were edited. Seven degraded-input source sites are declared,
all required-input: the CSS overlay, CSS numbers, CSS strings, Markdown inline,
selector corpus extraction and both type-aware corpus tests. The JSON inventory
now contains 77 sites, preserving all pre-existing skip identities.

Assumptions: missing-input vocabulary and variable must occur on the same output
line. A test's forwarded overlay diagnostic belongs to that emitting parent test.
Unattributed fmt diagnostics are associated conservatively with active tests;
ambiguous diagnostics fail closed. This is source and log analysis, not a fix to
fixture provisioning or to the stage1 tests' behavior.

## Historical witness and removal mutant

The five CSS tests named in the brief actually ran across shards 1, 2, 5, 8, 12
and 13 of `gate-logs/47fbaf174d16/20261007T153523`. TestCSSPrinterAgreesWithGo
ran multiple times. The checked-in regression fixture is the unchanged CSS
package events from each shard's canonical `gate-out/test.jsonl`, concatenated
in ascending shard order. Its timestamps, test names and output are preserved;
other packages are omitted. All five names and ADAMIC_CSS_FIXTURES are reported,
with seven required-input degraded executions and zero unknown diagnostics.

The complete canonical shard-1 log is also recorded unchanged. Its command
returns 1 and names TestThePortParsesAsGoCohereDoes and ADAMIC_CSS_FIXTURES.
Removing only the output event containing the missing-fixture diagnostic makes
that same complete shard log return 0. Applying the identical removal to the
CSS projection removes seven output events; all terminal PASS events remain,
and the checker returns 0 with `skips=0 required-input=0 unknown=0`.

Commands, each redirected to a log:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 ./internal/skipcensus/...
go vet ./internal/skipcensus/...
go run ./internal/skipcensus/cmd docs/skip-census/degraded-proof/shard-1.jsonl
go run ./internal/skipcensus/cmd internal/skipcensus/testdata/css-degraded-47fbaf17.jsonl
```

Tests and vet exit 0. Both historical checker commands exit 1, as required.
The regression test itself constructs and checks the diagnostic-removal mutant.
Ordinary and ADAMIC_GATE_UNCACHED=1 command output compare byte-identically.
No new cache exists.

## Checks proven to fail

Disabling MissingInputVariables in censusCall in a scratch package makes
TestDegradedSourceAndDrift fail: it receives zero diagnostic rows instead of five.
Disabling CheckLog's pass handling makes TestHistoricalCSSDegradedPasses fail:
the historical degraded passes survive. Neither failure depends on a compiler
warning. Source tests additionally exercise new and removed diagnostic sites,
line moves and forbidden measurement/opt-in-lane classifications. Runtime tests
exercise unknown diagnostics, formatted applicability exemptions, failed-test
verdict separation, and chunked package-level fmt output.

Logs: [tests](degraded-proof/tests.log), [vet](degraded-proof/vet.log),
[source mutant](degraded-proof/source-mutant.log),
[log mutant](degraded-proof/log-mutant.log),
[CSS rejected](degraded-proof/css-rejected.log),
[CSS cleaned](degraded-proof/css-clean.log),
[complete shard rejected](degraded-proof/shard-1-rejected.log),
[complete shard cleaned](degraded-proof/shard-1-clean.log).

## Setup and limits

`bash cloud/setup.sh` exited 0. nproc=5. Setup printed Go ready 0.077s,
submodules ready 0.133s, clang ready 0.395s, Node ready 0.558s, Markdown ready
0.612s, stage3 API ready 0.722s, modules ready 32.295s, build ready 66.957s,
cache warm 67.088s and done 67.245s. These are setup diagnostics, not a speedup
claim. Exact build-flags and before/after load are in [setup.log](degraded-proof/setup.log):
base e165f424, nproc 5, cpu.max 400000 100000, Go 1.27.1, clang 20.1.8,
Node v24.19.0, cached setup, load before 0.00 0.00 0.00 and after 7.05 1.81 0.61.

The full gate and live stage1 suites were not run. Their sources were unchanged;
verification here covers source inventory and gate log policy. Silent omissions
with no matching diagnostic, arbitrary computed messages, indirect helper calls,
and unknown output writers are outside static analysis. Matching unknown
runtime diagnostics still fail closed. The CSS test supplies no corpus commit
pin or corpus count assertion; this unit reports that limitation instead of
inventing a pin for the setup worker.
