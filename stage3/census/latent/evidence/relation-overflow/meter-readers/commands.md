# Reproduction

Source the tools before every command:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/8dbaea5/typescript
```

The source checkout must resolve to
`050880ce59e30b356b686bd3144efe24f875ebc8`. Setup and npm logs are retained.
Commands below use file-backed logs. Capture `$?` immediately after each child,
print it, and exit with it; a nonzero mutant exit is expected and asserted by
mutants.py.

```sh
bash cloud/setup.sh > /tmp/8dbaea5-setup.log 2>&1
npm ci --prefix stage3/api > /tmp/8dbaea5-npm.log 2>&1
go test -count=1 -v ./stage3/meter > /workspace/scratch/8dbaea5/meter-tests.log 2>&1
go test -count=1 -json ./stage3/census/... > /workspace/scratch/8dbaea5/census-tests.jsonl 2> /workspace/scratch/8dbaea5/census-tests.stderr
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/8dbaea5/overlay > /workspace/scratch/8dbaea5/overlay.log 2>&1
python3 stage3/meter/entry_overlay.py /workspace/scratch/8dbaea5/overlay /workspace/scratch/8dbaea5/entry-overlay > /workspace/scratch/8dbaea5/entry-overlay.log 2>&1
go test -count=1 -v -overlay=/workspace/scratch/8dbaea5/overlay/overlay.json ./stage3/census/latent/... > /workspace/scratch/8dbaea5/overlay-tests.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/8dbaea5/entry-overlay/overlay.json -o /workspace/scratch/8dbaea5/entry-census ./stage3/census/latent/tool > /workspace/scratch/8dbaea5/entry-build.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/8dbaea5/overlay/overlay.json -o /workspace/scratch/8dbaea5/latent-census ./stage3/census/latent/tool > /workspace/scratch/8dbaea5/latent-build.log 2>&1
go build -o /workspace/scratch/8dbaea5/census ./stage3/census/tool > /workspace/scratch/8dbaea5/census-build.log 2>&1
CENSUS_BINARY=/workspace/scratch/8dbaea5/census LATENT_CENSUS_BINARY=/workspace/scratch/8dbaea5/latent-census ENTRY_CENSUS_BINARY=/workspace/scratch/8dbaea5/entry-census python3 -E -S -B -m unittest discover -s stage3/meter -p '*_test.py' > /workspace/scratch/8dbaea5/meter-real-tests.log 2>&1
python3 stage3/census/latent/evidence/relation-overflow/meter-readers/mutants.py /workspace/scratch/8dbaea5/final-mutants > /workspace/scratch/8dbaea5/final-mutants.log 2>&1
```

The tests use Python unittest discovery through the existing Go wrappers. The
optional real probes run in the separate direct Python invocation with all three
binary paths set. The Go wrappers deliberately use a restricted tool environment
that does not forward those paths.
No new Go test or compiler subprocess test was added. The new regressions are
Python report tests and assertions within the existing real entry probe.
