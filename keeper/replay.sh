#!/bin/bash
# replay.sh: the test audit's central mutant replay (#xphstyt, @system_adamic_tests): one standalone mutant diff in, the
# repo-wide set of failing tests out. A clone settles uniqueness inside its package; this settles it across the repo.
# It rides fast.sh's interface rather than a rerun by hash, since loom-inputs-v1 counts every non-test path as shared and
# a code mutant would rerun the whole suite (developer tools' loom, Oct 9 07:34Z):
#
#	the replay writes  ~/.loom/jobs/fast/<mutant sha>.json   {"branch": "test-audit/replay/<base12>-<diff hash>", "sha", "base",
#	                                                          "base_name": "main", "packages": "select" | [...], "priority": 0}
#	fast.sh writes     ~/.loom/jobs/fast/<mutant sha>.verdict and the work directory's reds.txt
#
#	pilots/adamic-gate/replay.sh <base sha> <mutant.diff>...               # the gate's selection picks the packages
#	REPLAY_PACKAGES="<import path> ..." pilots/adamic-gate/replay.sh ...   # a mutant the selection can't see (a .a
#	                                                                         source Go tests read as data), named outright
#
# Each diff becomes a commit on the base without touching any working tree (a scratch index, git apply --cached), pushed
# as test-audit/replay/<base12>-<diff hash>; the same diff on the same base is the same commit, so a replay already decided is
# read back, never run twice. It prints one line per diff once every verdict is in, tab-separated: the diff hash, the
# verdict word (green: nothing failed, the mutant survived the reached packages; red: the failing tests follow; void:
# Loom broke or the diff didn't apply, so the failing set is unknown), the mutant sha and the failing tests as
# "<package> <test>", comma-separated. Priority 0 on codex-side: side work, behind every landing's fast gate.
set -uo pipefail

[ $# -ge 2 ] || { echo "replay: <base sha> <mutant.diff>..."; exit 2; }
base=$1
shift
[[ ${base} =~ ^[0-9a-f]{40}$ ]] || { echo "replay: the base is a full sha"; exit 2; }
jobs=${LOOM_FAST_JOBS:-${HOME}/.loom/jobs/fast}
gate=${HOME}/Projects/system/adamic-gate
git -C "${gate}" cat-file -e "${base}^{commit}" 2> /dev/null || git -C "${gate}" fetch -q origin "${base}" || { echo "replay: can't fetch ${base}"; exit 2; }
# The gate's selection runs developer tools' run.py from this sha, as the watcher's own jobs pin it.
tools=$(git -C "${gate}" ls-remote origin refs/heads/devtools/fast-gate | cut -f1)
[[ ${tools} =~ ^[0-9a-f]{40}$ ]] || { echo "replay: can't read devtools/fast-gate"; exit 2; }
packages=$(python3 -c 'import json, os, sys; names = os.environ.get("REPLAY_PACKAGES", "").split(); print(json.dumps(names or "select"))')

hashes=() shas=()
for diff in "$@"; do
	hash=$(shasum -a 256 "${diff}" | cut -c1-12)
	index=$(mktemp -u)
	sha=$(
		export GIT_INDEX_FILE=${index}
		git -C "${gate}" read-tree "${base}" &&
			git -C "${gate}" apply --cached --whitespace=nowarn "$(cd "$(dirname "${diff}")" && pwd)/$(basename "${diff}")" &&
			tree=$(git -C "${gate}" write-tree) &&
			GIT_AUTHOR_DATE="@0 +0000" GIT_COMMITTER_DATE="@0 +0000" git -C "${gate}" commit-tree "${tree}" -p "${base}" -m "Test audit mutant ${hash} on ${base:0:12}"
	)
	rm -f "${index}"
	hashes+=("${hash}")
	if [[ ! ${sha} =~ ^[0-9a-f]{40}$ ]]; then
		shas+=("")
		continue
	fi
	shas+=("${sha}")
	if [ -f "${jobs}/${sha}.json" ]; then
		# A replay Loom voided (its fault, not the mutant's) runs again; its old files are kept beside it.
		grep -q '^void:' "${jobs}/${sha}.verdict" 2> /dev/null || continue
		stamp=$(date -u +%Y%m%dT%H%M%SZ)
		for old in "${jobs}/${sha}".json "${jobs}/${sha}".verdict "${jobs}/${sha}".work; do mv "${old}" "${old}.void-${stamp}"; done
	fi
	git -C "${gate}" push -q origin "${sha}:refs/heads/test-audit/replay/${base:0:12}-${hash}" || { echo "replay: pushing ${hash} failed"; exit 2; }
	python3 - "${jobs}/${sha}.json" "${sha}" "${base}" "${hash}" "${packages}" "${tools}" <<'PY'
import json, os, sys
path, sha, base, hash, packages, tools = sys.argv[1:]
job = {"branch": "test-audit/replay/" + base[:12] + "-" + hash, "sha": sha, "base": base, "base_name": "main", "tools": tools, "packages": json.loads(packages), "priority": 0, "publish": False}
json.dump(job, open(path + ".partial", "w"))
os.rename(path + ".partial", path)
PY
done

# Every verdict in, then the lines: the fast server takes eight jobs at a time, so a wave of thirty finishes together.
for sha in "${shas[@]}"; do
	[ -z "${sha}" ] && continue
	until [ -f "${jobs}/${sha}.verdict" ]; do sleep 20; done
done
for at in "${!hashes[@]}"; do
	sha=${shas[$at]}
	if [ -z "${sha}" ]; then
		printf '%s\tvoid\t\tthe diff does not apply to %s\n' "${hashes[$at]}" "${base:0:12}"
		continue
	fi
	verdict=$(head -1 "${jobs}/${sha}.verdict" | cut -d: -f1)
	failing=$(grep '^FAIL ' "${jobs}/${sha}.work/reds.txt" 2> /dev/null | cut -d' ' -f2-3 | sed -E 's#^github.com/system-inc/adamic/##; s#^([^ ]+ Test[^/ ]*)/.*#\1#' | sort -u | paste -sd, -)
	# Red with no failing test is a mutant that broke go build or vet: caught, but by the compiler, not a test.
	[ "${verdict}" = red ] && [ -z "${failing}" ] && failing="(go build or vet)"
	printf '%s\t%s\t%s\t%s\n' "${hashes[$at]}" "${verdict}" "${sha}" "${failing}"
done
