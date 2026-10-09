#!/bin/bash
# ramp.sh: grow the test-audit fleet toward a target in steps, measuring whether new Codex sessions cost the star's
# pool workers (@system_adamic's ruling, Oct 9 03:08): only while the codex pool is between runs (nothing queued),
# read how many of its workers asked in the last two minutes, start a step of clones (each on its next unit), read
# again three minutes later, and back the step out if the count fell. A star run starting mid-step pauses the ramp.
# One line per event, for a Monitor.
set -uo pipefail
S=$1 brief=$2 target=${3:-25} step=${4:-10}
state=${S}/fanout/state.json
st() { python3 -I "${S}/fanout/state.py" "${state}" "$@"; }
cd /Users/kirkouimet/Projects/ahra
asking() { timeout 60 "${HOME}/.loom/bin/loom" pool status codex 2> /dev/null | awk '/ asked [0-9]+ s ago/ { for (i = 1; i <= NF; i++) if ($i == "asked" && $(i + 1) <= 120) n++ } END { print n + 0 }'; }
queued() { timeout 60 "${HOME}/.loom/bin/loom" pool status codex 2> /dev/null | awk 'NR == 1 { print $3 }'; }
workers() { timeout 60 "${HOME}/.loom/bin/loom" pool status codex 2> /dev/null | awk 'NR == 1 { print $5 }'; }
count() { grep -c . "${S}/fanout/clones.txt"; }
while [ "$(count)" -lt "${target}" ]; do
	# Between runs: nothing queued, read twice a minute apart so a run that's just starting isn't taken for idle.
	until [ "$(queued)" = 0 ] && sleep 60 && [ "$(queued)" = 0 ]; do sleep 120; done
	before=$(asking) total=$(workers)
	size=$(( target - $(count) < step ? target - $(count) : step ))
	echo "step: codex pool idle, ${before} of ${total} workers asked in the last 120 s; starting ${size} clones ($(count) now)"
	started=()
	for _ in $(seq 1 "${size}"); do
		unit=$(st take)
		[ -n "${unit}" ] || break
		python3 -I "${S}/brief-for.py" "${brief}" "${S}/manifest/units.json" "${unit}" > "${S}/fanout/briefs/${unit}.md"
		id=$(ahra ai start codex --environment adamic --label "audit-main-${unit}" --fleet test-audit --prompt-file "${S}/fanout/briefs/${unit}.md" 2>&1 | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | head -1)
		if [ -z "${id}" ]; then
			echo "start for ${unit} failed; back on the queue"
			st put-back "${unit}" > /dev/null
			continue
		fi
		st run "${id}" "${unit}" > /dev/null
		echo "${id}" >> "${S}/fanout/clones.txt"
		started+=("${id}")
	done
	sleep 180
	after=$(asking) runs=$(queued)
	if [ "${runs}" != 0 ]; then
		echo "step: a star run started during the step (${runs} queued); ${after} asked; holding at $(count) clones until it ends"
		continue
	fi
	# A fall of more than a tenth counts; a worker or two between asks is noise.
	if [ $(( after * 10 )) -lt $(( before * 9 )) ]; then
		echo "step: asking fell ${before} to ${after} with ${#started[@]} new clones; backing the step out (cap measured near $(( $(count) - ${#started[@]} )) audit clones)"
		for id in "${started[@]}"; do
			ahra ai stop "${id}" > /dev/null 2>&1
			st requeue "${id}" > /dev/null
			grep -v "^${id}\$" "${S}/fanout/clones.txt" > "${S}/fanout/clones.txt.partial" && mv "${S}/fanout/clones.txt.partial" "${S}/fanout/clones.txt"
		done
		echo "ramp stopped at $(count) clones"
		exit 0
	fi
	echo "step: asking held ${before} to ${after} with ${#started[@]} new clones; $(count) clones now"
done
echo "ramp reached ${target} clones"
