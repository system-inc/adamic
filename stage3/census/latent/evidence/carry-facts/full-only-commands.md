# Commands: current-tip re-cut

Final parent pin: `7e896ce4890d74568585b55973a3c982e7d2629a`. New branch: `typescript/census-carry-facts-2`. Neither prior carry commit is in its ancestry. The original patch was applied as working changes on main, not cherry-picked as two commits. Main advanced from `ae228d0cfe74b319c94020f369ef13da8f63e9df` during validation; both binaries and the complete subset comparison were rebuilt/rerun at the new pin. The previous-tip run is retained separately under `/workspace/scratch/carry2-ae228d-subset` and in `tip-advance-resources.csv`.

## Branch and baseline

```sh
git fetch origin main > /workspace/scratch/carry2-fetch.log 2>&1
git switch -c typescript/census-carry-facts-2 origin/main > /workspace/scratch/carry2-branch.log 2>&1
git diff --binary 9c607b11c6d7a6afad29853db5d86b488baaa08e dad1c20fe9f91fac8f1336fc78e9761c24b2515c -- stage3/census/latent > /workspace/scratch/carry2-existing.patch
git apply --3way /workspace/scratch/carry2-existing.patch > /workspace/scratch/carry2-apply.log 2>&1
# After revising the code, preserve all working changes across main's advance:
git fetch origin main > /workspace/scratch/carry2-refetch.log 2>&1
git stash push --include-untracked -m carry2-before-tip-update -- stage3/census/latent > /workspace/scratch/carry2-stash.log 2>&1
git rev-parse stash@{0} > /workspace/scratch/carry2-stash.sha
git merge --ff-only origin/main > /workspace/scratch/carry2-ff.log 2>&1
git stash apply --index "$(cat /workspace/scratch/carry2-stash.sha)" > /workspace/scratch/carry2-stash-apply.log 2>&1
```

The detached baseline worktree was first created at `ae228d…`, with its empty Cohere directory replaced by a symlink to the already initialized checkout; it was then switched to the final pin. Both pins use Cohere `06af2e399c6c1fb93fd570b774027942747400b4` and its TypeScript submodule `d92d9bfee114c80be2c375d72edae966176e3a4f`. The final baseline build used that worktree's own unmodified generator and templates:

```sh
git -C /workspace/scratch/carry2-main switch --detach origin/main > /workspace/scratch/carry2-main-update.log 2>&1
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
python3 /workspace/scratch/carry2-main/stage3/census/latent/make_overlay.py /workspace/scratch/carry2-main /workspace/scratch/carry2-baseline-overlay > /workspace/scratch/carry2-baseline-overlay.log 2>&1
gofmt -w /workspace/scratch/carry2-baseline-overlay/*.go
(cd /workspace/scratch/carry2-main && go build -buildvcs=false -overlay=/workspace/scratch/carry2-baseline-overlay/overlay.json -o /workspace/scratch/carry2-baseline-census ./stage3/census/latent/tool) > /workspace/scratch/carry2-baseline-build.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/carry2-carried-overlay > /workspace/scratch/carry2-carried-overlay.log 2>&1
gofmt -w /workspace/scratch/carry2-carried-overlay/*.go
go build -buildvcs=false -overlay=/workspace/scratch/carry2-carried-overlay/overlay.json -o /workspace/scratch/carry2-carried-census ./stage3/census/latent/tool > /workspace/scratch/carry2-carried-build.log 2>&1
python3 /workspace/scratch/carry2-measure.py > /workspace/scratch/carry2-measure.log 2>&1
```

[measure-subset.py](measure-subset.py) is an exact copy of that runner, containing measurement only and no profiling code. It checks all 79 input hashes before running. Each binary sequentially attempts utilities, binder, and emitter, retaining the full compiler project. Each invocation uses:

```sh
LATENT_ONLY_FILE=/workspace/scratch/tree/src/compiler/utilities.ts \
LATENT_TIMES=/workspace/scratch/carry2-subset/baseline/utilities.ts/latent-times.jsonl \
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
/workspace/scratch/carry2-baseline-census /workspace/scratch/tree/src/compiler \
/workspace/scratch/carry2-subset/baseline/utilities.ts/latent.jsonl \
> /workspace/scratch/carry2-subset/baseline/utilities.ts/run.log 2>&1
```

`wait4` supplies each child's peak RSS; `/proc/PID/status` is sampled every 0.25 s to enforce the 12 GiB stop. Each changed file is immediately compared byte-for-byte with its matching baseline. A mismatch stops the runner and reports its first differing line. Combined ledgers retain one identical checker header followed by binder/emitter/utilities file records.

## Guards and mutants

All three new Go tests call `t.Parallel()` and use parsing/formatting/generation in-process, without compiler or Go subprocesses. The shell runs the tests; the tests themselves never spawn one.

```sh
guard=TestLatentCarryFactsAuditedFields
git grep -n "func $guard"
guard=TestLatentCarryFactsIgnoresUncarriedFields
git grep -n "func $guard"
guard=TestLatentCarryFactsReplayBodiesMatchMain
git grep -n "func $guard"
python3 /workspace/scratch/carry2-run-mutants.py > /workspace/scratch/carry2-mutant-summary.log 2>&1
go test ./stage3/census/latent/statecopy ./stage3/census/latent/overlaytext -run '^TestLatentCarryFacts' -count=1 > /workspace/scratch/carry2-restored-green.log 2>&1
go test -count=1 ./stage3/census/... > /workspace/scratch/carry2-tests.log 2>&1
go vet ./stage3/census/latent/statecopy ./stage3/census/latent/overlaytext > /workspace/scratch/carry2-vet.log 2>&1
go vet -overlay=/workspace/scratch/carry2-carried-overlay/overlay.json ./stage3/census/latent/tool > /workspace/scratch/carry2-tool-vet.log 2>&1
```

The table mutant removes only the `entriesRecords` entry from the generator table, asserts the named red line, and restores the original source in `finally`. The type mutant changes `entriesProgramRecordsKnown bool` to `int` in `/workspace/scratch/carry2-type-mutant-lower.go`; a scratch test-source overlay redirects the guard's input to that file. The body mutant adds `_ = 0` to the body in `/workspace/scratch/carry2-body-mutant-entries_records.go`; a scratch test-source overlay redirects the fingerprint test to that source. Production `internal/` files are never edited.

Exact mutant test argv:

```sh
go test ./stage3/census/latent/statecopy -run '^TestLatentCarryFactsAuditedFields$' -count=1 > /workspace/scratch/carry2-table-mutant.log 2>&1
go test -overlay=/workspace/scratch/carry2-type-mutant-overlay.json ./stage3/census/latent/statecopy -run '^TestLatentCarryFactsAuditedFields$' -count=1 > /workspace/scratch/carry2-type-mutant.log 2>&1
go test -overlay=/workspace/scratch/carry2-body-mutant-overlay.json ./stage3/census/latent/overlaytext -run '^TestLatentCarryFactsReplayBodiesMatchMain$' -count=1 > /workspace/scratch/carry2-body-mutant.log 2>&1
```

For the non-test overlay lowering source, `go list -overlay=/workspace/scratch/carry2-carried-overlay/overlay.json -json ./internal/lower` provides `Dir` and `GoFiles`; the successful `go vet` argv using those source files is retained in [overlay-source-vet-command.json](overlay-source-vet-command.json). This excludes production tests that reference the intentionally removed ordinary lowering entrypoint.

The earlier CPU profile, literal-answer logs and storage-type counterexample are retained from pin `9c607b…`, explicitly as historical evidence. [prior-commands.md](prior-commands.md) records those exact commands. Profiling code remains scratch-only.

## Unfiltered all-79-files proof on a larger house box

Not run here. Start in the pushed `typescript/census-carry-facts-2` checkout, with initialized Cohere and the identical adapted input at `/workspace/scratch/tree`. This command uses final parent `7e896ce…` for the baseline and removes the file filter. It verifies input hashes before comparing the complete output files. This subset proof must not be described as an all-files proof.

```sh
set -euo pipefail
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
repo=$PWD
house_run=$(mktemp -d /workspace/scratch/carry-facts-house.XXXXXX)
ADAPTED=/workspace/scratch/tree
python3 - "$ADAPTED" "$repo/stage3/census/latent/evidence/carry-facts/adapted-source-shas.csv" > "$house_run/input-check.log" 2>&1 <<'PY'
import csv, hashlib, pathlib, sys
root = pathlib.Path(sys.argv[1])
rows = list(csv.DictReader(open(sys.argv[2])))
assert len(rows) == 79
for row in rows:
    assert hashlib.sha256((root / row['path']).read_bytes()).hexdigest() == row['sha256'], row['path']
actual = {str(p.relative_to(root)) for p in (root/'src/compiler').rglob('*') if p.is_file() and p.suffix in ('.ts', '.a')}
assert actual == {row['path'] for row in rows}
PY
git worktree add --detach "$house_run/base" 7e896ce4890d74568585b55973a3c982e7d2629a > "$house_run/worktree.log" 2>&1
rmdir "$house_run/base/cohere"
ln -s "$repo/cohere" "$house_run/base/cohere"
python3 "$house_run/base/stage3/census/latent/make_overlay.py" "$house_run/base" "$house_run/baseline-overlay" > "$house_run/baseline-overlay.log" 2>&1
gofmt -w "$house_run/baseline-overlay/"*.go
(cd "$house_run/base" && go build -buildvcs=false -overlay="$house_run/baseline-overlay/overlay.json" -o "$house_run/baseline-census" ./stage3/census/latent/tool) > "$house_run/baseline-build.log" 2>&1
python3 stage3/census/latent/make_overlay.py "$repo" "$house_run/carried-overlay" > "$house_run/carried-overlay.log" 2>&1
gofmt -w "$house_run/carried-overlay/"*.go
go build -buildvcs=false -overlay="$house_run/carried-overlay/overlay.json" -o "$house_run/carried-census" ./stage3/census/latent/tool > "$house_run/carried-build.log" 2>&1
env -u LATENT_ONLY_FILE LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$house_run/baseline-times.jsonl" "$house_run/baseline-census" "$ADAPTED/src/compiler" "$house_run/baseline-latent.jsonl" > "$house_run/baseline-run.log" 2>&1
env -u LATENT_ONLY_FILE LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$house_run/carried-times.jsonl" "$house_run/carried-census" "$ADAPTED/src/compiler" "$house_run/carried-latent.jsonl" > "$house_run/carried-run.log" 2>&1
cmp "$house_run/baseline-latent.jsonl" "$house_run/carried-latent.jsonl" > "$house_run/compare.log" 2>&1
sha256sum "$house_run/baseline-latent.jsonl" "$house_run/carried-latent.jsonl" > "$house_run/latent.sha256"
printf '%s\n' "$house_run"
```
