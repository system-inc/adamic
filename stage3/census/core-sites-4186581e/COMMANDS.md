# Commands and frozen inputs

Working repository: `/workspace/core-sites-main`, a clean worktree of main `4186581ec18f469c40149bf217d22b028e8374fb`, on `typescript/core-sites`.

Setup, each command's stdout and stderr redirected to a separate retained log:

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /workspace/scratch/b5w3ycg/setup.log 2>&1
npm ci --prefix stage3/api > /workspace/scratch/b5w3ycg/npm-ci.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
# HEAD verified as 050880ce59e30b356b686bd3144efe24f875ebc8, tag v6.0.3.
bash stage3/apply.sh --build-product /workspace/core-sites-adapted > /workspace/scratch/b5w3ycg/apply.log 2>&1
```

The setup and apply launchers measured their wall time with `date +%s`; their result files retain exit code followed by seconds. Setup and apply overlapped. No adapter was selected or omitted. All 83 adapters completed. The tool build also overlapped the later adapters, but the census started only after apply exit 0 and the checkpoint push succeeded.

Tool preparation, retained in `prepare.sh`:

```sh
python3 stage3/census/latent/make_overlay.py /workspace/core-sites-main /workspace/scratch/b5w3ycg/overlay > /workspace/scratch/b5w3ycg/overlay-retry.log 2>&1
# Rename each generated replacement .go file to .go.txt and update overlay.json.
gofmt -w /workspace/scratch/b5w3ycg/overlay/*.go.txt > /workspace/scratch/b5w3ycg/gofmt.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/b5w3ycg/overlay/overlay.json -o /workspace/scratch/b5w3ycg/latent ./stage3/census/latent/tool > /workspace/scratch/b5w3ycg/build.log 2>&1
```

The complete census invocation was equivalent to the following. `run-census.py` provides the actual subprocess invocation, 3,600 second limit, elapsed timing and process-group cleanup. It removed other inherited `LATENT_*` variables. `LATENT_TIMES` only adds measurement timing rows.

```sh
LATENT_FULL=1 LATENT_ONLY_FILE=/workspace/core-sites-adapted/src/compiler/core.ts LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/workspace/scratch/b5w3ycg/unit-times.jsonl /workspace/scratch/b5w3ycg/latent /workspace/core-sites-adapted/src/compiler /workspace/scratch/b5w3ycg/raw.jsonl > /workspace/scratch/b5w3ycg/census.log 2>&1
```

No speculative mode, replay, source stub, ordinary compiler overlay, or patched production lowering source was used. `raw.jsonl` contains the complete selected-file census, including the earlier portion of core.ts, not just the requested range.

`ast-map.cjs` reads both completed core.ts files using the pinned API parser. `analyze.py` reads the raw stream and source map. `emitter-map.py` searches production `internal/lower` reason fragments where dynamic strings cannot match the full reason, retaining exact source lines in `emitters.json` and grep results in `reason-searches.json`. `render.py` renders `REPORT.md`. These scripts retain this run's absolute paths.
