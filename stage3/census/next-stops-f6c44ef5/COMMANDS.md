# Reproduction and evidence

Compiler and adaptations: f6c44ef5f5d3908e134e2f098bf5c9cabdff424f.
Run from that checkout. The ordinary census uses no overlay.

```sh
source /workspace/adamic-tools/env.sh
git submodule update --init --recursive > /tmp/next-stops-f6c44ef5/submodules.log 2>&1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/next-stops-f6c44ef5/upstream > /tmp/next-stops-f6c44ef5/upstream.log 2>&1
npm ci --prefix stage3/api > /tmp/next-stops-f6c44ef5/npm.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/next-stops-f6c44ef5/upstream stage3/apply.sh /tmp/next-stops-f6c44ef5/adapted > /tmp/next-stops-f6c44ef5/apply.log 2>&1
go build -o /tmp/next-stops-f6c44ef5/census ./stage3/census/tool > /tmp/next-stops-f6c44ef5/build.log 2>&1
timeout --kill-after=10s 1790s /tmp/next-stops-f6c44ef5/census /tmp/next-stops-f6c44ef5/adapted/src/compiler /workspace/adamic/stage3/census/next-stops-f6c44ef5/census.jsonl > stage3/census/next-stops-f6c44ef5/census.log 2>&1
```

The complete ordinary run finished in 816.834 recorded seconds. No compiler-root
split was needed. The separate ordinary tsc entry was measured with the same
binary and both its per-file and aggregate observations agree:

```sh
timeout --kill-after=10s 300s /tmp/next-stops-f6c44ef5/census /tmp/next-stops-f6c44ef5/adapted/src/tsc/tsc.ts /workspace/adamic/stage3/census/next-stops-f6c44ef5/tsc/census.jsonl > stage3/census/next-stops-f6c44ef5/tsc/census.log 2>&1
```

## Artifact-label certificate

Build the meter's existing guarded entry tool separately from the ordinary tool:

```sh
python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/next-stops-f6c44ef5/latent-overlay > /tmp/next-stops-f6c44ef5/overlay.log 2>&1
python3 stage3/meter/entry_overlay.py /tmp/next-stops-f6c44ef5/latent-overlay /tmp/next-stops-f6c44ef5/entry-overlay >> /tmp/next-stops-f6c44ef5/overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/next-stops-f6c44ef5/entry-overlay/overlay.json -o /tmp/next-stops-f6c44ef5/entry-census ./stage3/census/latent/tool > /tmp/next-stops-f6c44ef5/entry-build.log 2>&1
```

The certificate evaluates the exact candidate-expression AST from the pinned
meter. Its required attempted-file memberships are witnessed by native full-mode
unit timing records. The two original groups have disjoint selections covering
the 81 resolved source files; selections, headers, timings and stopped-run status
are retained in auxiliary-entry-attempts.tar.gz. Their complete body scans are not
claimed. For each group the actual command was, with PART set to 1 or 2:

```sh
PART=1
LATENT_ONLY_FILE="$(cat stage3/census/next-stops-f6c44ef5/tsc-full-parts/part-$PART/files.txt)" LATENT_TIMES=/workspace/adamic/stage3/census/next-stops-f6c44ef5/tsc-full-parts/part-$PART/times.jsonl LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 timeout --kill-after=10s 1790s /tmp/next-stops-f6c44ef5/entry-census /tmp/next-stops-f6c44ef5/adapted/src/tsc/tsc.ts /workspace/adamic/stage3/census/next-stops-f6c44ef5/tsc-full-parts/part-$PART/latent.jsonl > stage3/census/next-stops-f6c44ef5/tsc-full-parts/part-$PART/latent.log 2>&1
```

The attempted-unit timer is reached only with status attempted or panic, both
accepted by the meter. Readiness findings are impossible in this guarded binary:
the only emitter is namespaceInitialization, whose only production call is
removed by the meter overlay. The binary contains neither its symbol nor either
reason literal. The ordinary binary retains all of them. This proves the empty
readiness set globally, without waiting for unrelated body units.

```sh
go tool nm /tmp/next-stops-f6c44ef5/entry-census > stage3/census/next-stops-f6c44ef5/entry-symbols.txt 2> stage3/census/next-stops-f6c44ef5/entry-symbols.log
go tool nm /tmp/next-stops-f6c44ef5/census > stage3/census/next-stops-f6c44ef5/census-symbols.txt 2> stage3/census/next-stops-f6c44ef5/census-symbols.log
```

The unchanged meter Node helper supplied the probes. Its tsc was selected from
stage3/api/node_modules/.bin, the lockfile-pinned TypeScript 6.0.3 install:

```sh
PATH=/workspace/adamic/stage3/api/node_modules/.bin:$PATH python3 stage3/meter/node_readiness.py /tmp/next-stops-f6c44ef5/adapted stage3/census/next-stops-f6c44ef5/initial-node-probes /tmp/next-stops-f6c44ef5/adapted/src/compiler/debug.ts /tmp/next-stops-f6c44ef5/adapted/src/compiler/commandLineParser.ts /tmp/next-stops-f6c44ef5/adapted/src/compiler/diagnosticInformationMap.generated.ts > stage3/census/next-stops-f6c44ef5/node-probes.log 2>&1
PATH=/workspace/adamic/stage3/api/node_modules/.bin:$PATH python3 stage3/meter/node_readiness.py /tmp/next-stops-f6c44ef5/adapted stage3/census/next-stops-f6c44ef5/types-node-probe /tmp/next-stops-f6c44ef5/adapted/src/compiler/types.ts > stage3/census/next-stops-f6c44ef5/types-node-probe.log 2>&1
python3 stage3/census/next-stops-f6c44ef5/classify-artifacts.py
python3 stage3/census/next-stops-f6c44ef5/summarize.py
```

classify-artifacts.py also reads compressed symbol tables and the archived timing
inputs. It requires the measured binaries and generated overlay to verify their
hashes and active source graph. The ordinary stream is never rewritten. No full
entry body census, entry-program next stop, or promotion of artifact roots to
whole lowering is claimed.

## Supplementary attempts

The archive retains the incomplete legacy entry scan, stopped at its final
105-minute bound, an unsplit full-mode scan that reached its 30-minute bound,
the stopped two-group scans, and independently checkpointed file attempts stopped
after the artifact certificate made further work unnecessary. These are separate
from the completed 79-root ordinary census. entry-attempts.json summarizes them.
Every task-owned census process has exited or been stopped; process-cleanup.json
records the cleanup. The archive preserves original data and attempt helpers for
inspection, not as evidence of complete entry-program lowering.

The native graph-order inspector uses the same guarded loader and entry reach
hook. It only records source paths and never lowers a program:

```sh
python3 - <<'PYTHON'
import json
from pathlib import Path
metadata=json.loads(Path('/tmp/next-stops-f6c44ef5/entry-overlay/overlay.json').read_text())
keys=[k for k in metadata['Replace'] if k.endswith('/stage3/census/latent/tool/main.go')]
assert len(keys)==1
metadata['Replace'][keys[0]]=str(Path('stage3/census/next-stops-f6c44ef5/graph-order.go.txt').resolve())
Path('/tmp/next-stops-f6c44ef5/graph-overlay.json').write_text(json.dumps(metadata,indent=2)+'\n')
PYTHON
go build -buildvcs=false -overlay=/tmp/next-stops-f6c44ef5/graph-overlay.json -o /tmp/next-stops-f6c44ef5/graph-order ./stage3/census/latent/tool > stage3/census/next-stops-f6c44ef5/graph-build.log 2>&1
timeout --kill-after=5s 60s /tmp/next-stops-f6c44ef5/graph-order /tmp/next-stops-f6c44ef5/adapted/src/tsc/tsc.ts > stage3/census/next-stops-f6c44ef5/graph-order.json 2> stage3/census/next-stops-f6c44ef5/graph-order.log
```

To restore the timing-group inputs before rerunning them:

```sh
tar -xzf stage3/census/next-stops-f6c44ef5/auxiliary-entry-attempts.tar.gz -C stage3/census/next-stops-f6c44ef5 tsc-full-parts
```

Plain logs and Go build metadata were gzip-compressed after measurement, preserving their original bytes. The ordinary JSONL remains uncompressed beside REPORT.md.
