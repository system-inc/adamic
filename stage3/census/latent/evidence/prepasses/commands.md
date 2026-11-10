# Reproduction commands

Compiler base: `0816e18511b5e1ef4383543ee81c526fc36acfaf`.
The code checkpoint was pushed before either long first-error run.
The baseline checkout is `/workspace/scratch/sygfnmc-main`; the changed checkout
is `/workspace/adamic`. Both share the exact pinned cohere checkout through a
scratch dependency symlink. VCS stamping is disabled in measurement builds.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/sygfnmc-ts-pin
bash cloud/setup.sh > /tmp/sygfnmc-setup.log 2>&1
npm ci --prefix stage3/api > /tmp/sygfnmc-npm.log 2>&1
go test -count=1 -v ./stage3/census/latent/... ./stage3/census/predicates/... > /tmp/sygfnmc-prepush-tests.log 2>&1
python3 -E -S -B stage3/census/predicates/tools/audit_families.py > /tmp/sygfnmc-predicate-families.log 2>&1
python3 -E -S -B -m unittest discover -s stage3/census/pair -p 'test_*.py' > /tmp/sygfnmc-pair-tests.log 2>&1
```

The source checkout was fetched shallowly at the exact TypeScript pin. Main's
`bash stage3/apply.sh --build-product /workspace/scratch/sygfnmc-adapted` supplied
the only adapted tree. The retained apply status reports its child's own exit.
`source-hashes.json` freezes every TypeScript/Adamic source under src; patch-set.md
is copied from that completed main apply. No compiler or adaptation was changed.

For each checkout, generate its own overlay and extend only its entry driver:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/sygfnmc-head-overlay > /tmp/sygfnmc-head-overlay.log 2>&1
python3 stage3/meter/entry_overlay.py /tmp/sygfnmc-head-overlay /tmp/sygfnmc-head-entry > /tmp/sygfnmc-head-entry.log 2>&1
go build -buildvcs=false -overlay=/tmp/sygfnmc-head-entry/overlay.json -o /tmp/sygfnmc-head ./stage3/census/latent/tool > /tmp/sygfnmc-head-build.log 2>&1
```

Use base-named paths and the untouched baseline checkout for the other side.
The first baseline overlay attempt preceded completion of its dependency link;
it failed to find the checker module. The retry and build succeeded. No compiler
source repair was made to obtain a baseline build.

`cost_overlay.py` appends cost-driver.go.txt to the changed lower hook and replaces
normal census execution with eight fresh preparations. Build its overlay as the
cost binary. Invoke once with `src/compiler/checker.ts` and once with
`src/tsc/tsc.ts`, setting `LATENT_ASSERT_NO_OUTPUT=1` and `LATENT_PREPASS_TIMES`
to separate TSV paths. It retains the ordinary output guards and the complete
entry-root checker program. Exact invocations and child resources are in the
cost status records. Both cost samples ran before the long censuses.

Every long run uses run.py, which sends stdout/stderr to a file, records the
child's own return code and getrusage resources, and stops the child's
process group after 1,800 seconds without log or watched-file progress.
`LATENT_TIMES` is a watched stream of completed units. First-error mode leaves
LATENT_FULL, LATENT_SPECULATIVE and LATENT_ONLY_FILE unset. Full core mode sets
`LATENT_FULL=1 LATENT_ONLY_FILE=/workspace/scratch/sygfnmc-adapted/src/compiler/core.ts`
and still loads the complete tsc entry program. The core-only full-mode sides run sequentially and overlap the baseline first-error
run. The changed first-error side starts near the end of the baseline, with a
14 GiB combined VmRSS ceiling on the two census processes. Wall times are shared-host
resource observations and are not a speed benchmark.

Pair both complete streams with main's stage3/census/pair tool and the pinned
stock TypeScript package:

```sh
python3 stage3/census/pair/pair.py BASE.jsonl.gz HEAD.jsonl.gz /workspace/scratch/sygfnmc-adapted --typescript stage3/api/node_modules/typescript --json first-error-paired.json --report first-error-paired.md
```

Repeat for full core. summarize.py renders exact pair-tool counts and groups
new findings by reason and file. It does not mask witness metadata or findings.
The pair tool verifies identical headers, unit identities, attempted eligibility
and source bytes, and performs the stock TypeScript outermost-access-chain
normalization. Raw streams are retained as deterministic gzip archives.

The environment transport reconnected during the baseline. Its outer runner ended,
while the measurement process continued with unchanged arguments and source files.
adopt.py monitors that same PID and records its own kernel wait status from
/proc/PID/stat exit_code after zombie state. Its CPU ticks cover the process lifetime;
its sampled lifetime VmHWM may miss a peak in the final polling interval. The status
record identifies this resource-recovery method explicitly. Later runs use launch.py
to put the resource runner in its own session, so another transport reconnect does
not remove its wait/reaping parent. Launch success never means census success.
