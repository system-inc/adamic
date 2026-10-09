#!/usr/bin/env bash
# Merges main into every area branch after main moves, so each area keeps building on what landed.
# It works with git merge-tree and commit-tree, so no worktree or checkout is touched and it can run
# from any clone. A conflict is reported, never resolved here: the area's owner merges main by hand,
# keeping both sides' intent. An area whose area-merge.sh lock is held, here or as locks/area-<area> on origin, is skipped
# and caught up on the next landing, so merge-back never takes a running merge's push from it.
#
# usage: cloud/integration/merge-back.sh
set -euo pipefail

directory=$(cd "$(dirname "$0")" && pwd)
git fetch -q origin
main=$(git rev-parse origin/main)
remoteLocks=$(git ls-remote origin 'refs/heads/locks/area-*' | sed 's#.*refs/heads/locks/area-##')

grep -v '^#' "$directory/areas.tsv" | while IFS=$'\t' read -r area owner oracle; do
	[ -n "$area" ] || continue
	lock="${ADAMIC_AREA_WORKTREES:-$HOME/.adamic-areas}/$area.lock"
	if [ -d "$lock" ]; then
		# An area-merge.sh run holds the area; pushing over it would cost that run its push.
		# The next landing's merge-back catches the area up.
		echo "area/$area skipped: a merge into it is running ($lock, since $(date -r "$lock" '+%H:%M'))"
		continue
	fi
	if printf '%s\n' "$remoteLocks" | grep -qx "$area"; then
		echo "area/$area skipped: a merge into it is running on another machine (locks/area-$area on origin)"
		continue
	fi
	# An owner can ask integration to hold its area between runs (a seat chaining area-merge.sh runs,
	# where a lock pushed by hand would block its own next run): a file named for the area in the hold
	# directory, holding why and until when, keeps merge-back off it until integration removes it.
	hold="${ADAMIC_MERGE_BACK_HOLDS:-$HOME/.adamic-merge-back-holds}/$area"
	if [ -f "$hold" ]; then
		echo "area/$area skipped: held for its owner ($(head -n 1 "$hold"))"
		continue
	fi
	if ! tip=$(git rev-parse -q --verify "refs/remotes/origin/area/$area"); then
		echo "area/$area missing"
		continue
	fi
	if git merge-base --is-ancestor "$main" "$tip"; then
		echo "area/$area ${tip:0:8} already holds main"
		continue
	fi
	if git merge-base --is-ancestor "$tip" "$main"; then
		# Nothing on the area that main lacks: move it forward without a merge commit.
		if git push -q origin "${main}:refs/heads/area/$area"; then
			echo "area/$area fast-forwarded to ${main:0:8}"
		else
			echo "area/$area moved while merging back; run again"
		fi
		continue
	fi
	if tree=$(git merge-tree --write-tree "$tip" "$main" 2>/dev/null); then
		commit=$(git commit-tree "$tree" -p "$tip" -p "$main" -m "Merge main ${main:0:8} into area/$area so the area builds on what landed")
		if git push -q origin "${commit}:refs/heads/area/$area"; then
			echo "area/$area merged main: ${commit:0:8}"
		else
			echo "area/$area moved while merging back; run again"
		fi
	else
		echo "area/$area conflicts with main ${main:0:8}: @$owner merges main into it by hand"
	fi
done
