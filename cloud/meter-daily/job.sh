#!/usr/bin/env bash
# cloud/meter-daily/job.sh runs the stage 3 meter once as a Loom job, at the commit the job was asked at (#6pgqq4m:
# after each landing, on main's newest tip, coalesced by the server's --on-move). It is run.sh without what the door
# now does: no systemd, no lock and no hand-job wait (a machine holds one job), no clone and no fetch of main (the
# keeper lays the asked commit as the job's worktree, and the job starts in it).
#
#   ahra loom job 'bash cloud/meter-daily/job.sh' --kind meter --task 6pgqq4m --at adamic@<sha>
#
# What the door doesn't do yet (release-0.8.1235, Jobs on #6pgqq4m), the job does for itself: it lays cohere and
# cohere/TypeScript from the machine's own stores, sources the box's toolchain, keeps TMPDIR in a directory beside the
# worktree that a trap removes, and keeps the run's record outside the worktree, which the keeper removes at the end.
# The record stays at $HOME/scratch/<asker>/meter-runs/<stamp>, and the job's last lines name it with its sha256, for
# the poster to fetch, until --keep (#z9n1zvs) carries it. The checked line goes to the log until the voice
# ($LOOM_VOICE, #acvkn2y) exists.
#
# Exit: the meter's own (0, or 3 for a whole record whose relation budget isn't ok), 3 is never a pass; any other
# nonzero is a failed run with no record. Preparation failures exit 64 to 66, so they never read as the meter's 3.
set -u

case "$PWD" in "$HOME"/scratch/*/*) ;; *)
	echo "meter-job: it starts in $PWD, not a job worktree under $HOME/scratch, so nothing ran"
	exit 64
	;;
esac
sha=$(git rev-parse --verify 'HEAD^{commit}') || { echo "meter-job: $PWD holds no commit, so nothing ran"; exit 64; }
work="$PWD.meter"
mkdir "$work" || { echo "meter-job: $work is there already or can't be made, so nothing ran"; exit 64; }
# The record outlives the job: the keeper removes the worktree and this trap removes $work, but not the runs.
runs="$(dirname "$PWD")/meter-runs"
mkdir -p "$runs" || exit 64
jobEnd() {
	local code=$?
	trap - EXIT
	case "$work" in "$HOME"/scratch/*/*.meter) chmod -R u+w -- "$work" 2> /dev/null; rm -rf -- "$work" ;; esac
	exit "$code"
}
trap jobEnd EXIT
trap 'exit 143' TERM
trap 'exit 129' HUP
trap 'exit 130' INT
mkdir "$work/tmp" || exit 64
export TMPDIR="$work/tmp"

. "$HOME/adamic-tools/env.sh" || { echo "meter-job: $HOME/adamic-tools/env.sh didn't source, so nothing ran"; exit 65; }
# A gate box's env.sh names its own gate inputs; the meter clones the tag stage3/source.json pins instead.
unset ADAMIC_TYPESCRIPT_SOURCE
missing=""
for tool in go gofmt node clang python3 git curl; do
	command -v "$tool" > /dev/null 2>&1 || missing="$missing $tool"
done
[ -z "$missing" ] || { echo "meter-job: missing from the toolchain:$missing"; exit 65; }
node --version | grep -q '^v24\.' || { echo "meter-job: node is $(node --version), the meter needs Node 24"; exit 65; }

# lay <repository> <submodule> <store name>: the submodule at its gitlink, from a store on this machine. The meter's
# per-ref worktrees clone cohere from this checkout, so both have to be here whole.
lay() {
	local pin store="" candidate
	pin=$(git -C "$1" rev-parse "HEAD:$2") || return 1
	for candidate in "$HOME/loom-check/modules/$3.git" "$HOME/scratch/system_loom/modules/$3.git"; do
		if git --git-dir="$candidate" cat-file -e "$pin^{commit}" 2> /dev/null; then
			store=$candidate
			break
		fi
	done
	[ -n "$store" ] || { echo "meter-job: no store on this machine holds $2 at $pin"; return 1; }
	git -C "$1" config "submodule.$2.url" "$store" || return 1
	git -C "$1" -c protocol.file.allow=always submodule update -q --init "$2" || return 1
	[ "$(git -C "$1/$2" rev-parse HEAD)" = "$pin" ] || { echo "meter-job: $2 isn't at $pin"; return 1; }
	echo "meter-job: $2 at $pin from $store"
}
lay "$PWD" cohere cohere && lay "$PWD/cohere" TypeScript TypeScript || { echo "meter-job: a submodule didn't lay, so nothing ran"; exit 66; }

# twice-daily.sh runs npm ci. When the toolchain has no npm, it comes from its pin and integrity, as in run.sh.
if ! command -v npm > /dev/null 2>&1; then
	bootstrap=cloud/markdown-width/npm-bootstrap.json
	npmVersion=$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["version"])' "$bootstrap") || exit 65
	npmHome="$work/npm-$npmVersion"
	mkdir -p "$npmHome/bin"
	curl -fsSL "$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["url"])' "$bootstrap")" -o "$npmHome/npm.tgz" || exit 65
	python3 -I - "$bootstrap" "$npmHome/npm.tgz" << 'PYTHON' || exit 65
import base64, hashlib, json, sys
algorithm, expected = json.load(open(sys.argv[1]))['integrity'].split('-', 1)
actual = base64.b64encode(hashlib.sha512(open(sys.argv[2], 'rb').read()).digest()).decode()
if algorithm != 'sha512' or actual != expected:
    sys.exit('npm bootstrap integrity mismatch')
PYTHON
	tar --no-same-owner -xzf "$npmHome/npm.tgz" -C "$npmHome" || exit 65
	printf '#!/bin/sh\nexec node %q "$@"\n' "$npmHome/package/bin/npm-cli.js" > "$npmHome/bin/npm"
	chmod 755 "$npmHome/bin/npm"
	export PATH="$npmHome/bin:$PATH"
fi
toolchain="$work/toolchain.txt"
{
	printf 'node %s\n' "$(node --version)"
	go version
	clang --version | head -n 1
	printf 'npm %s\n' "$(npm --version)"
	python3 --version
	git --version
} > "$toolchain"
echo "meter-job: $(uname -n), $(nproc) threads, measuring $sha; $(tr '\n' ';' < "$toolchain")"

# STAGE3_METER_MAIN_REF is this job's commit, so the launcher and the measured tree are one commit. The meter's own
# limit, 2 h 40 m, sits under the meter kind's 3 h ceiling, so a hang ends here with its last lines, not at a kill
# that runs no trap. A run takes 84 to 102 min on Pupa's class (Nov 10 and 11).
meterLog="$work/meter.log"
meterExit=0
STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error STAGE3_METER_RUNS=$runs STAGE3_METER_MAIN_REF=$sha \
	timeout --kill-after=60s 9600 bash stage3/meter/twice-daily.sh > "$meterLog" 2>&1 || meterExit=$?
run=$(sed -n 's/^run: //p' "$meterLog" | tail -n 1)
if [ "$meterExit" != 0 ] && [ "$meterExit" != 3 ]; then
	echo "meter-job: the meter exited $meterExit on $sha; its last lines:"
	tail -n 20 "$meterLog"
	[ -z "$run" ] || [ ! -d "$run" ] || cp "$meterLog" "$run/job-meter.log"
	exit "$meterExit"
fi
[ -n "$run" ] && [ -f "$run/report.json" ] && [ -f "$run/report.md" ] || {
	echo "meter-job: the meter exited $meterExit but left no report in '$run'"
	exit 1
}
cp "$toolchain" "$run/toolchain.txt"
cp "$meterLog" "$run/job-meter.log"

# The checked line is read from the meter's own report, never written by this wrapper: Adamic's rule on #door.
python3 -I - "$run/report.json" "$sha" << 'PYTHON'
import json, sys
report = json.load(open(sys.argv[1]))
totals = report.get('totals') or {}
latent = (report.get('latent_lowering') or {}).get('totals') or {}
budget = report.get('relation_budget') or {}
files = totals.get('source_files', 0)
print(f"::loom checked {files} tsc files at {sys.argv[2][:12]}: whole program {totals.get('checker_whole_program')}/{files}, "
      f"own file {totals.get('checker_own_file')}/{files}, lowering {totals.get('lowering')}/{files}, "
      f"latent Refused {latent.get('Refused')} NotYet {latent.get('NotYet')} panic {latent.get('panic')}, "
      f"relation budget {budget.get('status')} {budget.get('assignable')}/{budget.get('threshold')}")
PYTHON
verdict=$(python3 -I cloud/meter-daily/outcome.py meter "$meterExit" "$run/report.json")
record="$run.tar.gz"
tar -C "$runs" -czf "$record" "$(basename "$run")"
echo "meter-job: outcome ${verdict%%$'\n'*}"
echo "meter-job: record $record sha256 $(sha256sum "$record" | cut -d' ' -f1)"
exit "$meterExit"
