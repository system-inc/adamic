# cloud/meter-daily/hand-jobs.sh: whether a Loom hand job holds this box, for run.sh, which sources it. The meter's
# timer runs outside any gap, so it could start beside a hand job; Loom's condition for its timer on a line box
# (Nov 9) is that it never does: it waits up to 30 minutes for the box to come free, then skips the day.
#
# There's no public marker of a held box: the keeper's job files under ~ahra/.loom/hand-jobs are ahra's alone. The
# process table is anyone's to read, and a hand job is a child of the keeper, `loom control keep`, run as ahra. So a
# box is held while that keeper has a child; a box with no keeper hands out no jobs, and is free.
#
# POSIX sh, since the gate's tool list declares sh and its Go test (cloud/meter_daily_test.go) runs it that way.

# handJobKeeper prints the pid of the box's hand-job keeper, or nothing when it has none.
handJobKeeper() {
	# pgrep exits 1 when nothing matches, and run.sh runs under set -e and pipefail, where that would end the run.
	{ pgrep -u "${METER_KEEPER_USER:-ahra}" -f 'loom control keep' 2> /dev/null || true; } | head -n 1
}

# handJobHeld succeeds while the keeper whose pid is $1 has a child: a hand job running on this box.
handJobHeld() {
	[ -n "$({ pgrep -P "$1" 2> /dev/null || true; } | head -n 1)" ]
}

# Its variables carry a hand prefix: sh has no locals, and run.sh's own globals (step, above all) must survive it.
# waitForHandJobs <say>: returns 0 once no hand job holds the box, saying what it waited for through the function
# named <say>; returns 75 when one still holds it after METER_HAND_JOB_WAIT seconds (1800), read every
# METER_HAND_JOB_STEP seconds (60), and the caller skips the day rather than run beside it.
waitForHandJobs() {
	handSay=$1
	handKeeper=$(handJobKeeper)
	if [ -z "$handKeeper" ]; then
		"$handSay" "no hand-job keeper ('loom control keep') runs on this box, so nothing holds it: running"
		return 0
	fi
	handWaited=0
	handLimit=${METER_HAND_JOB_WAIT:-1800}
	handStep=${METER_HAND_JOB_STEP:-60}
	while handJobHeld "$handKeeper"; do
		if [ "$handWaited" -ge "$handLimit" ]; then
			"$handSay" "a hand job still holds this box after ${handWaited} s (keeper $handKeeper): skipping today's run, never beside it"
			return 75
		fi
		[ "$handWaited" -gt 0 ] || "$handSay" "a hand job holds this box (keeper $handKeeper): waiting up to ${handLimit} s, reading every ${handStep} s"
		sleep "$handStep"
		handWaited=$((handWaited + handStep))
	done
	"$handSay" "no hand job holds this box (keeper $handKeeper, waited ${handWaited} s): running"
	return 0
}
