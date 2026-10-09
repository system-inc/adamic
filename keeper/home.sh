#!/bin/bash
# home.sh: the keeper's local replay of a bounded survivor: the mutant against the mutated file's own package, in the
# scratch worktree at the unit's base. A survivor its home package catches was only a survivor of the slice.
#	home.sh <worktree> <base sha> <package dir> <name>=<diff> ...   (one line per mutant: name, exit, failing tests)
set -uo pipefail
W=$1 base=$2 package=$3
shift 3
git -C "${W}" checkout -q --detach "${base}" || { echo "home: can't check out ${base}"; exit 2; }
for pair in "$@"; do
	name=${pair%%=*} diff=${pair#*=} log=${diff%.diff}.home.log
	git -C "${W}" apply "${diff}" || { echo "${name}	void	the diff does not apply at ${base:0:10}"; continue; }
	(cd "${W}" && timeout 1500 go test -count=1 -timeout 1200s "./${package}/" > "${log}" 2>&1)
	code=$?
	git -C "${W}" apply -R "${diff}"
	failing=$(grep -E '^\s*--- FAIL: Test[^ /]*' "${log}" | sed -E 's/^\s*--- FAIL: (Test[^ /]*).*/\1/' | sort -u | paste -sd, -)
	printf '%s\texit %s\t%s\n' "${name}" "${code}" "${failing}"
done
