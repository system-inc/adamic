#!/usr/bin/env bash
# Preparation executed before this script (on the stable requested branch):
# git fetch origin typescript/stage3-meter-walk-4ac3f04d main
# git fetch origin typescript/stage3-meter-walk-4ac3f04d:refs/remotes/origin/typescript/stage3-meter-walk-4ac3f04d
# Read that branch's commands.sh, K1T2WP9.md and compare.py before setup.
# git switch -c typescript/stage3-meter-time-5e382ce4 5e382ce4baca2f75cc9a452f9b438c8318f375a5
# export GOPROXY='https://proxy.golang.org|direct'
# timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1
# Run below only after setup exits zero. Stop immediately on a failed build/run.
# No whole-directory census is invoked.
set -euo pipefail
repository=/workspace/adamic
cd "$repository"
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
run=$repository/stage3/meter/runs/time-5e382ce4
scratch=/workspace/time-5e382ce4
mkdir -p "$scratch" "$run/timing"
exec > >(tee "$run/commands.log") 2>&1
set -x
cp /tmp/setup.log "$run/setup.log"
{ date -u +%FT%TZ; nproc; cat /sys/fs/cgroup/cpu.max; cat /sys/fs/cgroup/memory.max; free -b; go version; node --version; uname -a; } > "$run/box.txt"
git show -s --format='%H %P %s' HEAD > "$run/pins.txt"
build_census() {
    local source=$1 destination=$2 logs=$3
    mkdir -p "$destination"
    (cd "$source" && go build -o "$destination/census" ./stage3/census/tool) > "$logs/build.log" 2>&1
    python3 "$repository/stage3/census/latent/make_overlay.py" "$source" "$destination/latent-overlay" > "$logs/latent-overlay.log" 2>&1
    gofmt -w "$destination/latent-overlay/"*.go
    python3 "$repository/stage3/meter/entry_overlay.py" "$destination/latent-overlay" "$destination/entry-overlay" > "$logs/entry-overlay.log" 2>&1
    gofmt -w "$destination/entry-overlay/"*.go
    (cd "$source" && go build -buildvcs=false -overlay="$destination/entry-overlay/overlay.json" -o "$destination/entry-census" ./stage3/census/latent/tool) > "$logs/entry-build.log" 2>&1
    (cd "$source" && go build -buildvcs=false -overlay="$destination/latent-overlay/overlay.json" -o "$destination/latent-census" ./stage3/census/latent/tool) > "$logs/latent-build.log" 2>&1
}
for label in parent change; do
    case $label in
      parent) sha=22026c7392f55556e721b6f425dd369ceafdf14f ;;
      change) sha=5e382ce4baca2f75cc9a452f9b438c8318f375a5 ;;
    esac
    source=$scratch/$label-source
    logs=$run/$label
    mkdir -p "$logs"
    git worktree add --detach "$source" "$sha" > "$logs/checkout.log" 2>&1
    git -C "$source" -c protocol.file.allow=always \
        -c "url.file://$repository/cohere.insteadOf=https://github.com/system-inc/cohere.git" \
        -c "url.file://$repository/cohere/TypeScript.insteadOf=https://github.com/system-inc/TypeScript.git" \
        submodule update --init --recursive --depth 1 > "$logs/submodules.log" 2>&1
    git -C "$source" submodule status --recursive > "$logs/submodule-pins.txt"
    build_census "$source" "$scratch/$label-binaries" "$logs"
    sha256sum "$scratch/$label-binaries/census" "$scratch/$label-binaries/entry-census" "$scratch/$label-binaries/latent-census" > "$logs/binaries.sha256"
    (cd "$source/stage3/api" && npm ci --ignore-scripts --no-audit --no-fund) > "$logs/packages.log" 2>&1
    if [[ ! -d $scratch/upstream ]]; then
        git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git "$scratch/upstream" > "$run/upstream.log" 2>&1
    fi
    ADAMIC_TYPESCRIPT_SOURCE=$scratch/upstream bash "$source/stage3/apply.sh" "$scratch/$label-adapted" > "$logs/apply.log" 2>&1
    git -C "$source" status --porcelain --untracked-files=no > "$logs/compiler-status.txt"
    test ! -s "$logs/compiler-status.txt"
done
# Verify byte-identical inputs while ignoring Git metadata and generated evidence.
python3 - "$scratch" > "$run/input-check.txt" <<'PY'
from pathlib import Path
import hashlib, sys
root=Path(sys.argv[1])
def files(label):
    base=root/(label+'-adapted')/'src'
    return {str(p.relative_to(base)):hashlib.sha256(p.read_bytes()).hexdigest() for p in base.rglob('*') if p.is_file()}
a,b=files('parent'),files('change')
assert a==b
print(f'All {len(a)} adapted src files are byte-identical.')
PY
if [[ -x /usr/bin/time ]]; then
    timebin=/usr/bin/time
else
    mkdir -p "$scratch/gnu-time"
    curl -fsSL https://deb.debian.org/debian/pool/main/t/time/time_1.9-0.2_amd64.deb -o "$scratch/gnu-time/time.deb"
    dpkg-deb -x "$scratch/gnu-time/time.deb" "$scratch/gnu-time/extracted"
    sha256sum "$scratch/gnu-time/time.deb" > "$run/timing/package.sha256"
    timebin=$scratch/gnu-time/extracted/usr/bin/time
fi
{ "$timebin" --version; sha256sum "$timebin"; } > "$run/timing/tool.txt"
for index in 1 2 3 4 5; do
    for label in parent change; do
        output=$run/timing/$label-$index
        binary=$scratch/$label-binaries/latent-census
        sourcefile=$scratch/$label-adapted/src/compiler/checker.ts
        sha256sum "$binary" >> "$run/timing/binaries.sha256"
        command=(env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 "$timebin" -v -o "$output.time" "$binary" "$sourcefile" "$output.jsonl")
        { printf '%q ' "${command[@]}"; printf '> %q 2>&1\n' "$output.log"; } >> "$run/timing/commands.log"
        "${command[@]}" > "$output.log" 2>&1
        gzip "$output.jsonl"
        printf '%s sample %s complete\n' "$label" "$index"
    done
done
# Post-run validation/publication commands (executed separately):
# python3 stage3/meter/runs/time-5e382ce4/compare.py
# bash -n stage3/meter/runs/time-5e382ce4/commands.sh
# git diff --check
# git add -f stage3/meter/runs/time-5e382ce4
# git diff --cached --check
# git diff --cached --stat
# git commit -m "Record checker.ts timing for 5e382ce4 against its parent"
# git -c push.followTags=false push -u origin HEAD:refs/heads/typescript/stage3-meter-time-5e382ce4
# git rev-parse HEAD
# git ls-remote --heads origin refs/heads/typescript/stage3-meter-time-5e382ce4
