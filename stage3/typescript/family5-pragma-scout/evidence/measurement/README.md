# Measurement commands

The scripts use the observed scratch directory `/workspace/scratch/family5-measure`. Its `main` worktree was detached at [the fetched main tip](main-commit.txt), with recursive shallow submodules initialized from [their recorded pins](submodules.txt). Its `report` worktree began at checkpoint `43bf21686c28679c4782c3d1bff7cf89cbba67cc`. The output directory is this evidence directory in the `report` worktree.

Fetch the upstream repository and tag in `main/stage3/source.json` with `git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git "$scratch/upstream"`. Verify the [upstream commit](upstream-commit.txt) with `git rev-parse HEAD` in that checkout. Run `npm ci --ignore-scripts --no-audit --no-fund` in `main/stage3/api` for stock tsc 6.0.3 and in `upstream` for the pinned declaration bundler dependencies.

From `"$scratch/main"` run the following with Go 1.27.1 on PATH.

```sh
ADAMIC_TYPESCRIPT_SOURCE="$scratch/upstream" bash stage3/apply.sh "$scratch/before"
cp -a "$scratch/before" "$scratch/after"
python3 "$scratch/report/stage3/typescript/family5-pragma-scout/evidence/trial.py" "$scratch/after"

python3 stage3/census/latent/make_overlay.py "$scratch/main" "$scratch/overlay"
gofmt -w "$scratch/overlay/"*.go
go build -buildvcs=false -overlay="$scratch/overlay/overlay.json" \
  -o "$scratch/latent-census" ./stage3/census/latent/tool

e="$scratch/report/stage3/typescript/family5-pragma-scout/evidence/measurement"
bash "$e/check.sh"
bash "$e/emit.sh"
node "$e/runtime.cjs" "$scratch/before-js/compiler/_namespaces/ts.js" \
  "$scratch/after-js/compiler/_namespaces/ts.js"
bash "$e/census.sh"
python3 "$e/grade.py" "$scratch" "$e"
```

The raw compressed census streams preserve their original absolute paths. `parser-grades.jsonl` uses zero-based indices into the `findings` array of each stream's parser record. An index appears exactly once per corresponding stream. Rows carry the aligned finding and attempting-unit coordinates.
