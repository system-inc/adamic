Merged the requested compiler area, recorded seven current stops, and admitted proved structural overload results without copying aliases.
Area merge 234467ff1cb968f3107746f7b17c368bc158ddba; stop table bc51bf1b; structural compiler 8c7a0d8466268f2c1b61b235c701947d00fa13a3.
Focused lowering tests, two uncached Node oracle fixtures with sanitizers, counts refresh and final scoped census pass.
Four structural mutants and one measurement-cache mutant fail their intended assertions; no build-warning failure is counted.
Step 16 gains proved structural results and refusal paths; all seven intervals reveal zero bytes and full compiler bodies remain unproved or checker-excluded.

## What the compiler now proves

The existing return proof already checks each return under the overload's
admitted parameter types, using Adamic's own type relation. Structural result
admission now consults that proof before refusing a result that has no runtime
shape check. The result entry keeps the wider implementation representation until
the proof establishes the narrower result. It returns the original object.

The Block witness has a broad Expression kind and a readonly statements array;
a matching kind cannot by itself establish Block. Returning the admitted Block
proves the whole result, preserving the object and its statements. Node and both
backends print true, true, 3.

The EvaluatorResult witness has readonly value and an ordinary writable boolean.
Readonly covariance admits its narrower input as the wider implementation input;
returning that admitted input establishes the narrower result. The boolean keeps
its identical type, and a write through the result reaches the input. Node and
both backends print node, true, missing, false. The result is not copied.

The refusal witnesses run on source Node first. A tagged object without statements
prints undefined; the three wider or writable EvaluatorResult witnesses print 1.
Adamic refuses those contracts. The missing field identifies result.statements,
the wider readonly results identify result.value, and writable parameter widening
retains its parameter refusal. Covariance is never reversed to narrow a wider
result, and checking a readonly result once would not establish a narrow value
against an independently writable wider alias.

## Stops and measurements

[OVERLOAD-STOPS.md](../../../../stage3/hidden-briefs/OVERLOAD-STOPS.md) is the
requested table beside TABLE.md. It records every first boundary, its diagnostic
location, raising function, exact text and whether the owner is lowering or the
checker. The two hidden-01 intervals share a checker-excluded body. Its TS2345
at declarations.ts:1678:110 lies outside the large interval and inside the small
one. There is no observed lowering statement boundary in that excluded body.
The outer callback refusal at 612:90 is separate and outside both intervals.

Both hidden-05 intervals first fail a visitNode call, whose parameter refusal
is diagnosed in visitorPublic.ts:123:5. Hidden-06 fails its TNode parameter
representation. Hidden-13 first refuses the overload at result.kind and also
has an implementation body excluded by TS18048. Hidden-14 refuses at result.value;
its enclosing function separately refuses method destructuring. Dependency and
signature diagnostics are explicitly separated from the boundary inside a region.

The full transformAsyncFunctionBody and evaluate implementations do not meet the
current return proof. Their local result assignments, assertions, branching and
returned callable context remain outside what this proof establishes. These
full implementations are still refused, with the structural component identified.
The accepted fixtures do not establish those whole compiler implementations.
No worker's unlanded branch was merged, including hidden-06's TNode storage change.

The before record uses merge 234467ff. The after record uses compiler 8c7a0d84.
Both attempt the identical seven intervals with the same checker eligibility,
all project declarations registered, and the original no-output guards. All 82
adapted source hashes match pin 388096e6. Exact union/subtraction uses that pin's
hidden.py. The comparison is measured on a checker-rejected program, not a runnable
compiler. Hidden before and after is unchanged: 13,625; 6,899; 6,078; 5,748;
11,417; 7,289; 7,102 bytes. Total 58,158, zero revealed in each interval.

The merged IR added a private emitter cache, argumentFacts. The original census
snapshot copier failed on it before recording useful lowering boundaries. The
measurement-only copier now permits precisely that cache's nil state; a populated
cache or another private field still fails. The repeated before and after records
contain lowering boundaries and checker exclusions rather than cache panics.
No production IR cache or checker option was changed.

## Validation

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower \
  -run 'TestOverloadStructural|TestOverloadCallback|TestOverloadResults|TestCensusOverload|TestPredicateOverload' \
  -count=1 > docs/overload-results/groups/structural/focused.log.txt 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/overload_structural_(block|evaluator).a$' \
  -count=1 -v > /tmp/overload-structural-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts \
  > docs/overload-results/groups/structural/counts.log.txt 2>&1
python3 docs/overload-results/groups/structural/run-mutants.py \
  > docs/overload-results/groups/structural/mutants.log.txt 2>&1
python3 docs/overload-results/groups/structural/validate-cache.py \
  /tmp/overload-structural-after-overlay/overlay.json \
  > docs/overload-results/groups/structural/cache-validation.log.txt 2>&1
```

Both oracle fixtures pass source Node, emitted JavaScript, release native,
sanitized native and the counted build. Logs and updated counts are preserved.
The final negative tests also execute all four Node refusal witnesses. No whole
package tests or full gate ran. Test output was written directly to log files.

The structural mutants are drop-proof (positive fixtures refuse), trust-shape
(a Block liar becomes accepted), reverse-covariance (a wider readonly result
becomes accepted), and mutable-variance (a writable parameter becomes accepted).
The last three fail the refusal assertion with got <nil>. A preliminary shape
mutant changed admission only and was caught by a later unreifiable-array stop;
the final mutation changes the common return proof and demonstrates unchecked
admission instead. The cache mutant removes the nil condition and fails when a
populated private cache is silently discarded. No compiler or clang failure is
counted as a caught rule mutant.

An initial Node witness runner used the wrong relative path and failed with
MODULE_NOT_FOUND. The corrected runner and final witness tests pass. An earlier
negative witness passed string | undefined directly to string-only console.log;
using its explicit missing-value fallback fixed the fixture.

For reproducibility, regenerate the latent overlay with make_overlay.py, apply
c68/census-scope.patch.gz to its generated latent_units file, then build the tool.
The cache support is now in the measurement template and requires no extra patch.
Use LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1 for each run. Compare before.jsonl.gz
and after.jsonl.gz with measure-regions.py --all --compiler 8c7a0d84, the sourcepin
checkout and the exact adapted compiler directory. regions.json and next-stops.json
retain all spans and exact diagnostics. The original before measurement was
completed before production structural edits.

Toolchain setup used GOPROXY=https://proxy.golang.org|direct and cloud/setup.sh,
then /workspace/adamic-tools/env.sh. Timing lines: go ready 0.019s; node ready
0.023s; submodules ready 0.061s; markdown step-duration 0.015s, ready 0.081s;
clang ready 0.343s; go build ready 42.928s; tests deferred 43.111s; cache warm
43.112s; done 43.139s. nproc 5; cgroup cpu.max 400000 100000. Go 1.27.1,
Node 24.19.0, clang 20.1.8. Full setup output is retained in setup.log.txt.
