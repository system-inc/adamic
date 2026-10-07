#!/usr/bin/env bash
# Merges main into every area branch after main moves, so each area keeps building on what landed.
# It works with git merge-tree and commit-tree, so no worktree or checkout is touched and it can run
# from any clone. A conflict is reported, never resolved here: the area's owner merges main by hand,
# keeping both sides' intent.
#
# usage: cloud/integration/merge-back.sh
set -euo pipefail

directory=$(cd "$(dirname "$0")" && pwd)
git fetch -q origin
main=$(git rev-parse origin/main)

grep -v '^#' "$directory/areas.tsv" | while IFS=$'\t' read -r area owner oracle; do
	[ -n "$area" ] || continue
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
