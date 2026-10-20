#!/usr/bin/env bash
# cloud/meter-daily/run.sh runs the stage 3 meter once on main's tip, the daily recipe in stage3/meter/README.md,
# with no session behind it: meter-daily.service starts it as adamicmeter, a user of its own with network but no
# keys, no push rights and no sudo (Loom's ruling; loomtest stays sealed behind #keydoor). It never commits or
# pushes; submit.sh, as ahra, does that from the finished run directory, and only after this exits 0.
#
#   sudo systemctl start meter-daily.service   (or by hand, as adamicmeter: bash cloud/meter-daily/run.sh)
#
# Everything lives in one state directory ($STATE_DIRECTORY from the unit, else /var/lib/adamic-meter), owned by
# adamicmeter, with an ACL (and a default ACL, so new files inherit it) letting ahra read it:
#
#   work/             the launcher checkout, detached at the main commit this run pins, cohere at its gitlink
#   runs/<stamp>/     the meter's output (STAGE3_METER_RUNS), with toolchain.txt added, so a number that moves
#                     with no sha change can be traced to the tools. Outside work/ on purpose: once a run lands
#                     on main, the next day's checkout would refuse to write a tracked file over its untracked twin
#   logs/<stamp>.log  this script's steps and the meter's whole output, kept on failure
#   status            the last run, key=value: outcome (Running, Succeeded, Failed), the step it reached, the main
#                     sha, the run directory, the log, and the last line the failing step printed
#
# The toolchain is the box's own env.sh (ADAMIC_METER_TOOLS_ENV), which adamicmeter reads through an ACL: traverse
# on /home/ahra, read on ~/adamic-tools. npm is bootstrapped from the pin and integrity in
# cloud/markdown-width/npm-bootstrap.json when the toolchain has none, since twice-daily.sh runs npm ci.
set -euo pipefail

[ "$(uname)" = Linux ] || { echo "meter-daily: Linux build boxes only, never a Mac" >&2; exit 2; }

home=${STATE_DIRECTORY:-/var/lib/adamic-meter}
toolsEnvironment=${ADAMIC_METER_TOOLS_ENV:-/home/ahra/adamic-tools/env.sh}
origin=https://github.com/system-inc/adamic.git
work=$home/work
runs=$home/runs
stamp=$(date -u +%Y%m%dT%H%M%SZ)
mkdir -p "$home/logs" "$runs"
log=$home/logs/$stamp.log
toolchain=$home/logs/$stamp.toolchain.txt
status=$home/status
started=$(date -u +%FT%TZ)
step=start sha="" run="" isLocked=false

say() { printf '%s meter-daily: %s\n' "$(date -u +%FT%TZ)" "$*" | tee -a "$log"; }

# record <outcome> <exit> <detail>: the status file submit.sh reads, written whole and then moved into place.
record() {
	local finished=""
	[ "$1" = Running ] || finished=$(date -u +%FT%TZ)
	printf 'outcome=%s\nstep=%s\nsha=%s\nrun=%s\nlog=%s\nstarted=%s\nfinished=%s\nexit=%s\ndetail=%s\n' \
		"$1" "$step" "$sha" "$run" "$log" "$started" "$finished" "$2" "$(printf '%s' "$3" | tr '\n' ' ')" > "$status.writing"
	mv "$status.writing" "$status"
}

# begin <step>: the status names the step a running run is in, so a run that hangs says where.
begin() {
	step=$1
	record Running 0 ""
}

# A failure anywhere records the step it was in and the last line that step printed; the log stays.
finish() {
	local code=$?
	[ "$code" = 0 ] && return 0
	# A run that never held the lock leaves the status to the run that does.
	"$isLocked" || return 0
	# A red run commits nothing, but the run directory the meter made, if it got that far, keeps what it used.
	[ -n "$run" ] || run=$(sed -n 's/^run: //p' "$log" 2> /dev/null | tail -n 1 || true)
	if [ -n "$run" ] && [ -d "$run" ]; then
		[ ! -f "$toolchain" ] || cp "$toolchain" "$run/toolchain.txt" || true
		chmod -R g+rX "$run" || true
	fi
	local detail
	detail=$(grep -v '^[[:space:]]*$' "$log" 2> /dev/null | tail -n 1 || true)
	[ "$code" = 124 ] && [ "$step" = meter ] && detail="timed out after 2 h; $detail"
	record Failed "$code" "$detail"
	say "failed at $step (exit $code); nothing pushed; log $log"
}
trap finish EXIT

# One run at a time: a hand start while the timer's run is going waits for nothing, it says so and stops.
exec 9> "$home/run.lock"
flock -n 9 || { say "another meter-daily run holds $home/run.lock"; exit 1; }
isLocked=true
record Running 0 ""
say "log $log"

begin toolchain
[ -r "$toolsEnvironment" ] || { say "no readable toolchain env at $toolsEnvironment (ADAMIC_METER_TOOLS_ENV)"; exit 1; }
# shellcheck disable=SC1090
source "$toolsEnvironment"
# A gate box's env.sh names its own gate inputs; the meter clones the tag the measured sha's stage3/source.json
# pins instead (twice-daily.sh does that only when this is unset).
unset ADAMIC_TYPESCRIPT_SOURCE
# The meter keeps its scratch worktrees under TMPDIR. The unit's PrivateTmp gives each run its own /tmp, which
# systemd removes when the run ends, so a day's few GB of adapted trees never pile up; the logs stay in $home.
export TMPDIR=/tmp/meter-daily
mkdir -p "$TMPDIR"
missing=""
for tool in go gofmt node clang python3 git curl; do
	command -v "$tool" > /dev/null 2>&1 || missing="$missing $tool"
done
[ -z "$missing" ] || { say "missing from the toolchain:$missing"; exit 1; }
node --version | grep -q '^v24\.' || { say "node is $(node --version), the meter needs Node 24"; exit 1; }
say "toolchain: $(go version | cut -d' ' -f3) node $(node --version) $(clang --version | head -n 1) $(python3 --version)"

begin clone
if [ ! -d "$work/.git" ]; then
	say "cloning $origin into $work"
	git clone --single-branch --branch main --no-checkout "$origin" "$work" >> "$log" 2>&1
fi

begin fetch
git -C "$work" fetch origin main:refs/remotes/origin/main >> "$log" 2>&1
sha=$(git -C "$work" rev-parse --verify 'origin/main^{commit}')
say "pinned main $sha"

begin checkout
# The launcher must be the pinned commit too: per-ref mode builds each tree's compiler from it, but report.py,
# the overlay generators and the cohere mirror per-ref worktrees clone from all come from this checkout.
# A clone made with --no-checkout has no index until its first switch, and git status calls every file deleted
# there, so only a checkout that has one is judged.
dirty=""
[ ! -f "$work/.git/index" ] || dirty=$(git -C "$work" status --porcelain --untracked-files=no 2> /dev/null || true)
[ -z "$dirty" ] || { say "the launcher checkout has changes no run made: $(printf '%s' "$dirty" | head -n 3 | tr '\n' ' ')"; exit 1; }
git -C "$work" switch --detach "$sha" >> "$log" 2>&1
# Yesterday's scratch worktrees went with its private /tmp; forget their records.
git -C "$work" worktree prune

begin submodules
# Full blobs at depth 1, so the per-ref worktrees can clone cohere from this checkout instead of GitHub.
git -C "$work" submodule update --init --recursive --depth 1 >> "$log" 2>&1

begin npm
if ! command -v npm > /dev/null 2>&1; then
	bootstrap=$work/cloud/markdown-width/npm-bootstrap.json
	npmVersion=$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["version"])' "$bootstrap")
	npmHome=$home/npm-$npmVersion
	if [ ! -x "$npmHome/bin/npm" ]; then
		say "bootstrapping npm $npmVersion from its pin in $bootstrap"
		mkdir -p "$npmHome.partial/bin"
		curl -fsSL "$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["url"])' "$bootstrap")" -o "$npmHome.partial/npm.tgz"
		python3 -I - "$bootstrap" "$npmHome.partial/npm.tgz" << 'PYTHON'
import base64, hashlib, json, sys
algorithm, expected = json.load(open(sys.argv[1]))['integrity'].split('-', 1)
actual = base64.b64encode(hashlib.sha512(open(sys.argv[2], 'rb').read()).digest()).decode()
if algorithm != 'sha512' or actual != expected:
    sys.exit('npm bootstrap integrity mismatch')
PYTHON
		tar --no-same-owner -xzf "$npmHome.partial/npm.tgz" -C "$npmHome.partial"
		printf '#!/bin/sh\nexec node %q "$@"\n' "$npmHome/package/bin/npm-cli.js" > "$npmHome.partial/bin/npm"
		chmod 755 "$npmHome.partial/bin/npm"
		mv "$npmHome.partial" "$npmHome"
	fi
	export PATH="$npmHome/bin:$PATH"
fi
say "npm $(npm --version)"
# What every tool answered for this run, copied into the run directory once the meter has made it.
{
	printf 'node %s\n' "$(node --version)"
	go version
	clang --version | head -n 1
	printf 'npm %s\n' "$(npm --version)"
	python3 --version
	git --version
} > "$toolchain"

begin meter
say "running stage3/meter/twice-daily.sh (per-ref, first-error, 2 h limit) on $sha"
# STAGE3_METER_MAIN_REF is the sha checked out above, so the launcher and both measured trees are one commit even
# when main moves between this fetch and the meter's own.
(
	cd "$work"
	STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error STAGE3_METER_RUNS=$runs STAGE3_METER_MAIN_REF=$sha \
		timeout --kill-after=60s 7200 bash stage3/meter/twice-daily.sh
) >> "$log" 2>&1
run=$(sed -n 's/^run: //p' "$log" | tail -n 1)
[ -n "$run" ] && [ -f "$run/report.md" ] && [ -f "$run/report.json" ] || { say "the meter exited 0 but left no report in '$run'"; exit 1; }
cp "$toolchain" "$run/toolchain.txt"
# twice-daily.sh makes the run with mktemp -d, mode 0700, and under a default ACL those group bits become the mask,
# which would leave ahra's read entry worth nothing. Opening the group bits opens the mask; the group is this user's.
chmod -R g+rX "$run"

step=done
record Succeeded 0 "$(head -n 1 "$run/report.md")"
say "measured $sha in $run"
