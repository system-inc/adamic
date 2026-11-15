# Reproduction

`S=/workspace/bind-extra-parameters-scratch`; both measured worktrees remain detached. The report branch is a separate sparse worktree. No production compiler or adaptation file was edited.

```sh
git fetch origin 55c375d9d245b580c010c3e0b5659f2849f24fb6 ba619e28774130e24331c3026a938cd6f3cd6a34
git merge-base 55c375d9 ba619e28
# 55c375d9d245b580c010c3e0b5659f2849f24fb6
git ls-tree 55c375d9 stage3
git ls-tree ba619e28 stage3
# both: a5ba4feea88de4158706d8326090cb7e0677c653
git worktree add --detach "$S/base" 55c375d9d245b580c010c3e0b5659f2849f24fb6
git worktree add --detach "$S/head" ba619e28774130e24331c3026a938cd6f3cd6a34
# Run in each worktree:
git submodule update --init --recursive --depth 1

git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git "$S/typescript"
git -C "$S/typescript" rev-parse HEAD
# 050880ce59e30b356b686bd3144efe24f875ebc8
(cd "$S/base/stage3/api" && npm ci)
source /workspace/adamic-tools/env.sh
export TMPDIR="$S/tmp" GOTMPDIR="$S/tmp" GOCACHE="$S/go-cache" GOMODCACHE="$S/go-mod-cache"
export STAGE3_CACHE="$S/stage3-cache" ADAMIC_TYPESCRIPT_SOURCE="$S/typescript"
export GOPROXY='https://proxy.golang.org|direct' CENSUS_MODULE_FORMAT=CommonJSBundle
(cd "$S/base" && stage3/apply.sh "$S/adapted")
for side in base head; do
  cd "$S/$side"
  python3 stage3/census/latent/make_overlay.py "$PWD" "$S/$side-overlay"
  gofmt -w "$S/$side-overlay/"*.go
  go build -buildvcs=false -overlay="$S/$side-overlay/overlay.json" -o "$S/$side-census" ./stage3/census/latent/tool
done
# One process at a time: sys base/head, publish; core base/head, publish; remaining files.
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE="$S/adapted/src/compiler/sys.ts" \
LATENT_TIMES="$S/results/full-sys.ts/base-times.jsonl" \
LATENT_PROGRESS_FILE="$S/results/full-sys.ts/base-progress.json" \
python3 "$S/timed.py" "$S/base-census" "$S/adapted/src/compiler" "$S/results/full-sys.ts/base.jsonl"
```

Every run's exact command, selection environment, exit code and wall seconds are in its `*-run.json` receipt. `python3 "$S/timed.py"` records CPU and peak RSS in its log. Scratch `measure.py` pairs full findings by `(kind, phase, reason, message with source position stripped)` and checks site/unit movement within each bucket. `full-summary.json` retains every changed before/after witness and adapted source line. Each full process has a 3,600-second deadline with process-group termination.

The first-error overlay is a scratch copy of each commit's generated overlay. Only the existing census file-selection guard is inserted at the start of its first-error `program.Files()` loop:

```go
if !latentFileSelected(program.FileName(file)) { continue }
```

The existing helper accepts newline-separated exact paths. All roots, checker context, registration and sibling bindings stay intact. The first-error binaries are rebuilt from these copied overlays. `first-pieces.json` records disjoint consecutive groups of at most ten sorted compiler files. Each base/head pair plus comparison shares a 3,600-second deadline. Runs use `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`, with the piece's absolute filenames joined by newline in `LATENT_ONLY_FILE`. Compare each completed pair with the repository tool:

```sh
python3 "$S/base/stage3/census/pair/pair.py" \
  "$S/results/first-01/base.jsonl" "$S/results/first-01/head.jsonl" \
  "$S/adapted" --typescript "$S/base/stage3/api/node_modules/typescript" \
  --json "$S/results/first-01/pair.json" --report "$S/results/first-01/pair.md"
```

The runner verifies both sides' exact piece membership, no repeated file or unit identity between pieces, and complete final compiler-file coverage. Raw observations and scratch runners remain outside Git; `receipts.json` gives their SHA-256 and byte sizes.

The initial timing-wrapper launch failed before starting the compiler because `/usr/bin/time` is absent. Scratch `timed.py` instead launches the census, reaps it with `os.wait4`, and appends its wall seconds, user/system CPU seconds, peak RSS in KiB, and exit code as JSON to the log. The outer runner enforces the deadline on the whole process group.
