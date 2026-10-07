Evidence only. Do not merge this branch to main.

Compiler merge: e1198a5772ba4b6d7d322043292696ea11989044.
Base: origin/area/stage3 234ab1aa5f728a5221fb6075c35b94f88a2c6437, containing main ce0750f28ef3943057f1f852b3ae5d93e6c5d644.
Merged compiler fix: 06e8507356eee09111459c6d7e8a5c0aa952adbd.

STAGE3_METER_COMPILER was unset; compiler-mode.json confirms single mode.
The ordinary binary embeds the merge SHA and vcs.modified=true. The only
untracked working-tree content at build time was this new run directory,
including compiler-mode.json, which the meter writes before building.
Tracked files and both submodules were clean. No compiler edits were made.

Setup command: export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh.
Setup exit: 0. Source /workspace/adamic-tools/env.sh; Node v24.19.0 first
on PATH; nproc=5. Full setup timing lines are in setup.log.
An optional --reference submodule initialization failed because its local
reference repositories were shallow. Setup then initialized both pinned
submodules normally and succeeded.

Meter command:
STAGE3_METER_RUNS=$PWD/stage3/meter/runs bash stage3/meter/twice-daily.sh > /tmp/stage3-meter-enum-tags.log 2>&1; echo "exit=$?"

Validation command:
CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.ki64ZZ/census LATENT_CENSUS_BINARY=/tmp/adamic-gate/stage3-meter.ki64ZZ/latent-census python3 -m unittest discover -s stage3/meter -p '*test.py' > meter-tests.log 2>&1
Result: 23 tests passed, exit 0. The real dependency type-error mutant
changes whole-program 2 -> 0 and own-file 2 -> 1, affecting only the dependency.
The real latent mutant plants one NotYet in target; its reason alone increases
by one and Refused stays unchanged. No new meter implementation was changed.

Limits: latent findings are measured on a checker-rejected program and do not
establish successful lowering or native output. The compiler's 865/101
classification is not independently tested by this census. Source snapshots
also differ from the baseline; overall total changes cannot be attributed
solely to the enum fix. No full compiler/oracle gate was run for this meter unit.
