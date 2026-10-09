The opt-in `STAGE3_METER_COMPILER=per-ref` mode builds ordinary and latent
census binaries in detached worktrees at the two fetched source pins, including
each pin's recorded submodules. The default continues to build one compiler
from the invoking checkout. Compiler mode and SHAs are recorded before building;
per-tree report SHAs and the Markdown provenance line describe the selected
compilers. Root JSON keeps the area's compiler SHA for existing consumers.

Validation used Node v24.19.0 and Go 1.27.1. Commands:

```sh
source /workspace/adamic-tools/env.sh
bash -n stage3/meter/twice-daily.sh
CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.0qVD1D/census LATENT_CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.0qVD1D/latent-census python3 -m unittest discover -s stage3/meter -p '*_test.py' > /tmp/stage3-meter-per-ref-tests-final.log 2>&1
```

All 23 tests passed, with no skips. [tests.log](per-ref-evidence/tests.log)
retains the output. The three new integration tests use real local Git refs,
detached worktrees and real Go compilation. Main, area and worker compiler
sources have distinct behavior. The default produces the worker's checker and
latent results twice; per-ref mode produces the main and area results separately.
Tests check both actual results and SHA metadata, area compatibility fields,
unchanged headline formatting, per-ref build logs, rejection of mismatched
compiler pins, and early rejection of an invalid option. Existing real imported
type-error and planted-NotYet probes also pass with the p2 binaries.

The wrong-compiler mutant changes exactly this build invocation in a scratch
copy of twice-daily.sh:

```diff
-        build_census "$scratch/$label-source" "$binaries" "$run/$label"
+        build_census "$repository" "$binaries" "$run/$label"
```

Run with:

```sh
METER_SCRIPT_UNDER_TEST=/tmp/stage3-meter-wrong-compiler-mutant.sh python3 -m unittest discover -s stage3/meter -p 'compiler_test.py' > /tmp/stage3-meter-wrong-compiler-mutant.log 2>&1
```

The mutant compiles and finishes the meter, but the per-ref test fails with
`NotYet: compiled with checkout` where `NotYet: compiled with main` is required.
The command exits 1; [wrong-compiler-mutant.log](per-ref-evidence/wrong-compiler-mutant.log)
retains the assertion. Correct-looking claimed SHAs cannot hide the wrong binary.

The production p2 corpus run uses the unchanged single-compiler meter from
84d5eadf. The new per-ref mode was exercised on the local compiler witnesses,
not on a second full TypeScript corpus run. No compiler sources, adaptations,
native backend or Node differential oracle were changed. The full repository
gate was not run. Merged-checkout setup failed at the unrelated adaptation 75
coverage package's string versus RootedFilePath mismatch; targeted census builds
and all meter tests succeeded.
