#!/usr/bin/env bash
# Proves what cloud/block-report/environment.sh writes for the sealed job to source: a TMPDIR the job was handed
# outlives setup.sh's env.sh, a job with none still gets env.sh's /tmp/adamic-gate, and sourcing it under
# set -euo pipefail leaves the shell running with no helper variable behind (Jobs on #system_loom_jobs, Nov 22). The
# writer block is read out of environment.sh by its own first and last lines and run with stand-in pieces in a scratch
# directory, so nothing is laid or fetched. The mutant, environment.sh with the TMPDIR restore deleted, must fail "(a)".
# Run it from the repository root on a Linux box, never on a Mac: bash cloud/block-report/environment-writer-test.sh
set -euo pipefail
repository=$(cd "$(dirname "$0")/../.." && pwd)
writer="$repository/cloud/block-report/environment.sh"
work=$(mktemp -d /tmp/environment-writer-test.XXXXXX)
# Only this run's own mktemp scratch is ever removed whole.
trap 'case "$work" in /tmp/environment-writer-test.?*) rm -rf -- "$work" ;; esac' EXIT
failures=0

# The writer block's own first and last lines, as environment.sh spells them.
first='begin manifest'
last='mv -f "$prefix/environment.sh.writing" "$prefix/environment.sh"'

# writerBlock <environment.sh>: the lines from first to last, and a failure unless both are there.
writerBlock() {
	awk -v first="$first" -v last="$last" '$0 == first, $0 == last { print; if ($0 == last) found = 1 } END { exit found ? 0 : 1 }' "$1"
}

# produce <environment.sh> <directory>: that file's writer block run in a fresh bash under set -euo pipefail, with
# stand-ins for what the steps before it set, leaving what it wrote at <directory>/prefix/environment.sh. The stand-in
# env.sh does what write-env.sh's does to TMPDIR.
produce() {
	local directory=$2
	mkdir -p "$directory/prefix" "$directory/tools" "$directory/go/bin" "$directory/gomodcache" "$directory/npmcache" "$directory/typescript-source" || return 1
	writerBlock "$1" > "$directory/block.sh" || { echo "FAIL: $1 has no writer block from '$first' to '$last'"; return 1; }
	printf '%s\n' 'export TMPDIR=/tmp/adamic-gate' > "$directory/tools/env.sh"
	printf '#!/bin/sh\n' > "$directory/go/bin/go"
	chmod 755 "$directory/go/bin/go"
	env -i PATH="$PATH" commit=0123456789abcdef0123456789abcdef01234567 prefix="$directory/prefix" \
		environmentScript="$directory/tools/env.sh" goRoot="$directory/go" gomodcache="$directory/gomodcache" \
		npmcache="$directory/npmcache" typescriptSource="$directory/typescript-source" installed="" npmBin="" \
		bash -c 'set -euo pipefail; begin() { :; }; source "$0"' "$directory/block.sh" || { echo "FAIL: the writer block in $1 failed"; return 1; }
	[ -r "$directory/prefix/environment.sh" ] || { echo "FAIL: the writer block in $1 wrote no environment.sh"; return 1; }
}

# observe <environment.sh> <script> [<name>=<value> ...]: what a fresh bash holding only PATH and those variables
# prints, stderr too, when it sources the file under set -euo pipefail and then runs the script. A shell the file
# ends prints none of what the script would.
observe() {
	local produced=$1 script=$2
	shift 2
	env -i PATH="$PATH" "$@" bash -c "set -euo pipefail; source \"\$0\"; $script" "$produced" 2>&1
}

# expect <label> <wanted> <got>
expect() {
	if [ "$3" = "$2" ]; then echo "ok: $1"; else echo "FAIL: $1: wanted '$2', got '$3'"; failures=$((failures + 1)); fi
}

real="$work/real"
produce "$writer" "$real" || exit 1
produced="$real/prefix/environment.sh"
tmpdir='printf %s "$TMPDIR"'

expect "(a) a TMPDIR the job was handed stays" /tmp/job-own "$(observe "$produced" "$tmpdir" TMPDIR=/tmp/job-own)"
expect "(b) with no TMPDIR, env.sh's /tmp/adamic-gate stands" /tmp/adamic-gate "$(observe "$produced" "$tmpdir")"
expect "(f) the shell survives sourcing it twice under set -euo pipefail" survived \
	"$(observe "$produced" 'source "$0"; echo survived' TMPDIR=/tmp/job-own)"
expect "(g) no helper variable is left behind" "" "$(observe "$produced" 'compgen -v blockEnvironment || true')"

# The mutant: a copy of environment.sh with the restore line deleted, its block produced in its own directory. An edit
# that changes nothing fails loudly, so a mutant that no longer applies can't pass as red.
mutant="$work/restore-deleted"
sed '/export TMPDIR="\$blockEnvironmentTemporary"/d' "$writer" > "$mutant.sh"
if cmp -s "$writer" "$mutant.sh"; then
	echo "FAIL: mutant restore-deleted: the edit changed nothing in environment.sh"
	failures=$((failures + 1))
elif produce "$mutant.sh" "$mutant"; then
	# A mutant shell that dies is red too, so its status mustn't end this one.
	got=$(observe "$mutant/prefix/environment.sh" "$tmpdir" TMPDIR=/tmp/job-own) || true
	if [ "$got" != /tmp/job-own ]; then
		echo "ok: mutant restore-deleted fails (a) (got '$got')"
	else
		echo "FAIL: mutant restore-deleted passes (a)"
		failures=$((failures + 1))
	fi
else
	failures=$((failures + 1))
fi

[ "$failures" = 0 ] && echo "environment-writer: all checks passed" || { echo "environment-writer: $failures failed"; exit 1; }
