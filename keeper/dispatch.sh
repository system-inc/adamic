#!/bin/bash
# dispatch3.sh: keep the test-audit fleet busy. Each clone whose turn finished has its reply saved, then gets the next
# unit of the manifest with the current brief. The clones are read from fanout/clones.txt every sweep, so the ramp adds
# or backs out members without a restart; every state edit goes through state.py's lock. One line per event.
set -uo pipefail
S=$1 brief=$2
state=${S}/fanout/state.json
st() { python3 -I "${S}/fanout/state.py" "${state}" "$@"; }
cd /Users/kirkouimet/Projects/ahra
while true; do
	clones=() && while read -r line; do [ -n "${line}" ] && clones+=("${line}"); done < "${S}/fanout/clones.txt"
	for clone in ${clones[@]+"${clones[@]}"}; do
		status=$(ahra ai status "${clone}" --wait 2> /dev/null | awk '/status/{print $2; exit}')
		case "${status}" in completed | idle) ;; failed | cancelled) echo "clone ${clone:0:13} is ${status}"; continue ;; *) continue ;; esac
		unit=$(st unit "${clone}")
		if [ -n "${unit}" ]; then
			reply=${S}/fanout/replies/${unit}.md
			ahra ai summary "${clone}" --wait > "${reply}" 2>&1
			# A clone whose environment never started audited nothing: its unit goes back on the queue, never into done.
			if grep -q -E 'wait_for_environment failed|environment failed to start|executor configuration' "${reply}"; then
				mv "${reply}" "${reply%.md}.envfail-$(date +%H%M%S).md"
				st requeue "${clone}" > /dev/null
				echo "environment failed for ${unit} on ${clone:0:13}; ${unit} goes back on the queue"
			else
				st finish "${clone}" > /dev/null
				echo "returned ${unit} from ${clone:0:13} ($(head -c 160 "${reply}" | tr '\n' ' '))"
			fi
		fi
		next=$(st take)
		if [ -z "${next}" ]; then echo "queue empty; ${clone:0:13} rests"; continue; fi
		python3 -I "${S}/brief-for.py" "${brief}" "${S}/manifest/units.json" "${next}" > "${S}/fanout/briefs/${next}.md"
		if ahra ai send "${clone}" --message-file "${S}/fanout/briefs/${next}.md" > /dev/null 2>&1; then
			st run "${clone}" "${next}" > /dev/null
			ahra ai label "${clone}" "audit-main-${next}" --fleet test-audit > /dev/null 2>&1
			echo "dispatched ${next} to ${clone:0:13}"
		else
			st put-back "${next}" > /dev/null
			echo "send of ${next} to ${clone:0:13} failed; ${next} goes back on the queue"
		fi
	done
	[ "$(st left)" = 0 ] && { echo "fanout complete"; exit 0; }
	sleep 60
done
