#!/bin/bash
# resume-failed.sh: after an account 429 left clones "failed" (the dispatcher never sends to a failed clone), give each
# one its next unit again, 30 s apart so the account isn't hit at once. One line per clone.
S=$1
st() { python3 -I "${S}/fanout/state.py" "${S}/fanout/state.json" "$@"; }
cd /Users/kirkouimet/Projects/ahra
while read -r clone; do
	[ -n "${clone}" ] || continue
	status=$(ahra ai status "${clone}" 2> /dev/null | awk '/^  status/{print $2; exit}')
	[ "${status}" = failed ] || { echo "${clone:0:13} is ${status}, left to the dispatcher"; continue; }
	[ -z "$(st unit "${clone}")" ] || { echo "${clone:0:13} still holds a unit"; continue; }
	next=$(st take)
	[ -n "${next}" ] || { echo "queue empty"; break; }
	python3 -I "${S}/brief-for.py" "${S}/test-audit-brief-v8.md" "${S}/manifest/units.json" "${next}" > "${S}/fanout/briefs/${next}.md"
	if ahra ai send "${clone}" --message-file "${S}/fanout/briefs/${next}.md" > /dev/null 2>&1; then
		st run "${clone}" "${next}" > /dev/null
		echo "resumed ${clone:0:13} with ${next}"
	else
		st put-back "${next}" > /dev/null
		echo "send to ${clone:0:13} failed; ${next} back on the queue"
	fi
	sleep 30
done < "${S}/fanout/failed-429.txt"
