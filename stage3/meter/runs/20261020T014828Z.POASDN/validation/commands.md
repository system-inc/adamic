# Validation commands

All test output was sent directly to files and read after execution. No test output was piped.

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 900 bash cloud/setup.sh > /tmp/stage3-meter-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
# 5; cgroup cpu.max is 400000 100000.
GOWORK=off timeout 300 go test -count=1 -timeout 4m ./stage3/census/latent/statecopy ./stage3/census/latent/refusalrewrite > /tmp/stage3-meter-generators.log 2>&1
timeout 300 python3 -m unittest discover -s stage3/meter -p '*_test.py' > /tmp/stage3-meter-tests-final.log 2>&1
timeout 300 python3 -m unittest discover -s stage3/meter -p compiler_test.py > /tmp/stage3-meter-parallel-tests.log 2>&1
timeout 600 python3 stage3/meter/check-daily-mutants.py /tmp/stage3-daily-mutants > /tmp/stage3-meter-mutants-final.log 2>&1
timeout 600 go vet ./... > /tmp/stage3-meter-vet.log 2>&1
timeout 30 gofmt -l cmd internal > /tmp/stage3-meter-gofmt.txt
CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.UhWmVy/main-binaries/census LATENT_CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.UhWmVy/main-binaries/latent-census ENTRY_CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.UhWmVy/main-binaries/entry-census timeout 600 python3 -m unittest discover -s stage3/meter -p '*_test.py' > /tmp/stage3-meter-real-tests.log 2>&1
STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error timeout --kill-after=60s 7200 bash /workspace/adamic/stage3/meter/twice-daily.sh > /tmp/stage3-meter-daily.log 2>&1
```

The default suite passes 41 tests with four optional binary probes skipped. With the actual run binaries, all 41 pass without skips. Both generator packages pass. The nil-hook probe compiles and runs generated Go, accepts nil, rejects a populated hook, and rejects unknown function fields and a changed hook signature before publishing output. It also removes the generated nil guard and proves the populated-hook assertion fails.

Six independent mutants are caught: ignore the requested main pin, allow unequal tree pins, build the checkout compiler despite pinned metadata, omit the previous README row, accept every function hook, and accept every preflight. Their individual witness logs are retained here. The refusal behavior probe records errors from both named preflights and continues through directives and both visitors. Unknown preflights return no generated source.

The Oct 19 run is absent from this branch's base, so the automatic committed baseline is Oct 8 iBLy18. `testdata/oct19-report.json` retains the comparison fields from commit 11a32835995c49e54308bba46b719657dbda68c2 and is tested against that committed Oct 8 report. The new run's ordinary own-file set adds src/compiler/debug.ts versus Oct 19 and loses no files. This observation spans main's changes and does not attribute the improvement to the meter.

No production compiler source was edited. This is a first-error observational census, with diagnosed-body and generic limits; ordinary lowering success is IR success. Native output and the full repository test gate were not run. Repository-wide vet and formatting checks pass.

## Final review checks

After the completed run, final review added two stale-worker protections: a local dependency transport miss retries the gitlinks' recorded origins, and README baseline reads use the measured main revision rather than the worker checkout. Both measurement copies used the successful local transport path. Regenerating the README from pinned main 8c318e45 produces exactly the completed run's original table. `measured-launcher.sh` preserves the exact launcher at execution time; neither measurement was resumed or replaced.

The strengthened suite was rerun with the same real binary command above, writing `/tmp/stage3-meter-real-tests-final2.log`: all 42 tests pass without skips. The mutant command was rerun to `/tmp/stage3-meter-mutants-final2.log`: all eight mutants are caught, adding omission of the recorded-origin retry and ignoring the baseline revision. The generated-copy test separately kills a removed runtime nil guard. Logs are `meter-real-binaries-final.txt`, `mutants-final.txt`, and the individual mutant files.

Both compressed latent streams were decoded and revalidated against their exact 79 implementation sources; every lowering-summary field reproduces report.json. The run records NotYet 930, Refused 4150, SkippedDependency 2, error 0, and panic 2 on each copy. These two retained compiler panics are observations, not infrastructure recovery. The original timeout process's wait status was 0. A monitoring-tool interruption detached the observer; the original launcher and children continued to completion. `completion.json` records the observed process IDs and wait status. No measurement command was resumed, replaced, or rerun.

Elapsed time is 1702 seconds (28m 22s), below the expected 35 to 45 minutes, on nproc 5 with a 4-CPU cgroup quota. This includes cold preparation and the brief optional-probe validation running alongside measurement; setup is separate at 575.544 seconds. Both copies measure 8c318e4528bffb67019116408a819c06e83df9fa, the main SHA fetched at launch. The working branch was cut earlier from b5b2019d6ba92133a76115bacc0416fb652cd069; no main merge was made.

Raw run .log files are retained as .log.gz with unchanged decompressed bytes. Census JSONL streams are also retained compressed.
