#!/usr/bin/env bash
# Builds the speculative queue for main: with several areas ready, it pushes one candidate per prefix,
# cloud/speculate-<label>-1 = main + the first area, -2 = main + the first two, and so on, so every
# prefix can be gated at once. Integration pushes the longest green prefix and bisects a red one.
# Each candidate is a chain of merge commits built with git merge-tree, so no checkout is touched.
# The chain stops at the first area that conflicts with the ones before it.
#
# An argument without a slash names an area; one with a slash names any branch on origin, so an
# integration branch can join the queue the same way.
#
# usage: cloud/integration/speculate.sh <label> <area or branch> [<area or branch>...]
#   e.g. cloud/integration/speculate.sh 0607a compiler runtime cloud/integrate-16
set -euo pipefail

if [ "$#" -lt 2 ]; then
	echo "usage: $0 <label> <area or branch> [<area or branch>...]" >&2
	exit 2
fi
label=$1
shift

git fetch -q origin
main=$(git rev-parse origin/main)
echo "main ${main}"
candidate=$main
position=0
for name in "$@"; do
	position=$((position + 1))
	case "$name" in
	*/*) area=$name ;;
	*) area=area/$name ;;
	esac
	tip=$(git rev-parse --verify "refs/remotes/origin/$area")
	if git merge-base --is-ancestor "$tip" "$candidate"; then
		echo "${position} $area adds nothing; skipped"
		continue
	fi
	if ! tree=$(git merge-tree --write-tree "$candidate" "$tip" 2>/dev/null); then
		echo "${position} $area ${tip:0:8} conflicts with the prefix before it; stopping here"
		break
	fi
	candidate=$(git commit-tree "$tree" -p "$candidate" -p "$tip" -m "Merge $area at ${tip:0:8} into the speculative main ${label}-${position}")
	git push -q origin "${candidate}:refs/heads/cloud/speculate-${label}-${position}"
	echo "${position} cloud/speculate-${label}-${position} ${candidate} (+$area ${tip:0:8})"
done
