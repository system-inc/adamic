Added a separate tsc entry whole-program check and guarded lowering census.
The existing first two lines and compiler-file accounting remain unchanged.

Validation on 2026-10-08:

```sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/stage3-diagnosis-go-tmp
export GOTMPDIR=/workspace/stage3-diagnosis-go-tmp
export GOMAXPROCS=2 GOFLAGS=-p=1
ENTRY_CENSUS_BINARY=/workspace/stage3-entry-main-census \
CENSUS_BINARY=/workspace/stage3-diagnosis-main-census \
LATENT_CENSUS_BINARY=/workspace/stage3-meter-tmp/stage3-meter.nHWE3Q/latent-census \
python3 -m unittest discover -s stage3/meter -p '*_test.py' > tests.log 2>&1
bash -n stage3/meter/twice-daily.sh
git diff --check
```

All 33 tests passed in 29.668s; none skipped. Shell syntax and whitespace checks
passed. Compiler-selection tests execute the meter against actual local Git refs
and distinct compiled Go witnesses in both single and per-ref modes. They require
the independent entry result, diagnostic count and compiler provenance, while
asserting the existing first two lines exactly.

The real entry probe uses the guarded entry driver built from scratch compiler
a609ec184bede2216c81d63230863473d75ca1e3 (main 132a0ed5 plus library 24d980a7).
Its ordinary checker probe is main 132a0ed5. The standalone witness uses identical
checker defaults in both: the clean entry imports an eligible dependency outside
src/compiler, whose debugger statement produces a lowering finding; an unrelated
source with a type error is not loaded. Planting an imported return-type mismatch
changes the entry to checker fail, one diagnostic, and no lowering attempts.
Existing real checker-attribution and planted latent NotYet tests also passed.

Mutant: built the same entry driver with Program.LatentEntryReach changed to an
empty method, retaining only the explicit entry in Program.Files. The binary
built successfully. Running entry_test.py with this binary exited 1, caught by
test_reach_includes_dependency_outside_compiler_and_excludes_unused_file:
`AssertionError: 1 != 2`. The full failure text is in wrapper-only-mutant.txt.
Additional artifact probes reject missing entry roots, missing reachable records,
checker/measurement diagnostic disagreement, lowering on a rejected entry,
a misleading checker-rejected label on clean reach, and a missing overlay hook.

Only scratch overlays extend compiler internals. The entry driver requires
LATENT_ASSERT_NO_OUTPUT; ordinary Load and Lower remain disabled, and the guard
checks both. No backend is imported. The old per-file latent driver is unchanged.
The clean-entry stream relabels its records only after verifying zero diagnostics.

No new twice-daily production run was started. No native execution or full Adamic
Go/oracle gate is claimed for this meter-only change. The lowering census retains
its existing per-unit and best-effort registration limits.
