# Reproducing split-build evidence

Use the pinned tools from `cloud/setup.sh` and source `/workspace/adamic-tools/env.sh`. Timing rows include cold runtime compilation and generated C compilation/linking, with loading/lowering/emission outside the timer. Copy `measure.go.txt` into a temporary Go package inside this repository, then build it. The arguments are `label [split [jobs]]`; no split argument exercises the current default. Run three rounds at each revision with no other verification workloads running.

The baseline is area commit `6c891cca313e614ae46de794814bc7c0db4cae01`; the merged policy-free timing revision is `22e8350e`. The final default includes the runtime parallelism changes. JSON records keep all repetitions, exact flags selected in the helper, clang version and load averages.

For a raw oracle recording:

```bash
mkdir -p .split-work/observations
python3 internal/native/split_build_evidence/record.py .split-work/recording
ADAMIC_GATE_UNCACHED=1 ADAMIC_SPLIT_OBSERVATIONS="$PWD/.split-work/observations" XDG_CACHE_HOME=/tmp/unique-native-cache go test -json -overlay=.split-work/recording/overlay.json ./internal/oracle -count=1 -parallel=8 -timeout=30m > .split-work/oracle.jsonl 2>&1
```

Keep Go's existing `GOCACHE` before replacing `XDG_CACHE_HOME`, if desired. The output overlay only observes executions and leaves the normal comparisons intact. The gate's uncached environment disables fixture-result reuse. Comparing the gzip recordings uses the same comparator as directories:

```bash
python3 internal/native/split_build_evidence/compare.py internal/native/split_build_evidence/before-observations.jsonl.gz internal/native/split_build_evidence/after-observations.jsonl.gz
```

It decodes stdout/stderr to bytes and performs no normalization. The main fixture sweep is reported separately from deliberate mutants and timing/resource/crash harnesses. A nonzero comparator status means some raw execution differs, including nondeterministic harness output; see the full differences and integration report rather than discarding them.

Reproduce the required module-order mutant:

```bash
python3 internal/native/split_build_evidence/mutant.py .split-work/module-order-mutant
```

The runner requires successful compilation/linking followed by an oracle stdout disagreement. It changes emitted initializer order, with both program exits zero.

The integrated whole run timed out after 30 minutes without assertion failures. Successful scoped completion logs cover its unfinished tests. `assemble.py` validates and combines these observations; `completion-plan.json` records the scopes, and the compressed test-event logs preserve completion evidence. `parity.json` is the unnormalized comparison result, including all 31 differing harness/mutant records. See ../SPLIT_BUILD.md for their interpretation.
