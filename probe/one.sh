#!/usr/bin/env bash
# Seats: the entry census pass at main's tip ef741d4f, base side only (no patch), then every unit's first stop by pair.py's
# rule, then the 26 causes of the 58 unheld firm walls (#bcbp175, firm-walls-plan-1600.md section 4) read from them.
# One pass a job: a unit census no longer fits one 2h job (J-78s0qa8, Nov 25). Its twin is firststops-main-1610-directory/.
set -uo pipefail
pass=entry
here=$(cd "$(dirname "$0")" && pwd)
root="${LOOM_JOB_SCRATCH:-$(mktemp -d "/tmp/seats-firststops-$pass.XXXXXX")}"
mkdir -p "$root/base"
LOOM_JOB_SCRATCH="$root/base" bash "$here/run.sh" base "$pass" || { echo "FAIL base census $pass"; exit 1; }
dir=$(ls -d "$root"/base/jsdoc-census-base.* | head -1)
failed=
# The FS lines go to the log whole, so the kept log is the table and no stream needs keeping.
python3 "$here/firststops.py" "$dir/$pass" "$dir/adapted" "$pass" > "$root/firststops.$pass.txt" || failed=1
cat "$root/firststops.$pass.txt"
echo "== $pass: the 26 causes, by first stop"
python3 "$here/causes.py" "$root/firststops.$pass.txt" || failed=1
echo "VERDICT printed"
[ -z "$failed" ] || exit 1
