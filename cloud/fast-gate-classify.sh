# A tip's class for scheduling only (coverage is the gate's business): big for an area, a stage3/
# change, or more than two touched packages (directories of changed Go files, a testdata path counting
# as its package); small otherwise. A big gate runs in the area slot (24 CPUs), a small one in either
# small slot (12 each). Sourced by cloud/fast-gate-watch.sh, which classes each tip when it queues it,
# and by cloud/fast-gate.sh, which classes a gate run without --class (integration's direct landing
# gates) the same way, so a two-file fix-forward never waits for the area slot. Needs ${here}.
classify() {
  local branch=$1 sha=$2 count changed base
  [[ ${branch} == area/* ]] && { echo B; return; }
  base=$(gateBase "${branch}" "${sha}")
  haveCommits "${sha}" || git -C "${here}" fetch -q origin "${sha}" 2>/dev/null || { echo B; return; }
  changed=$(git -C "${here}" diff --name-only "${base#* }...${sha}" 2>/dev/null)
  # A stage3/ change runs the stage 3 lane (about 10 minutes): big, whatever else it touches.
  echo "${changed}" | grep -q '^stage3/' && { echo B; return; }
  count=$(echo "${changed}" | awk '/\/testdata\// {sub("/testdata/.*", ""); print; next} /\.go$/ {sub("/[^/]*$", ""); print}' | sort -u | wc -l)
  [ "${count}" -gt 2 ] && echo B || echo S
}

# The ref a tip is gated against, as "<name> <sha>": main, or the area a worker's branch was cut from.
# Against main, a one-file change on a branch off area/compiler tests the whole area's delta (the clock's
# representations briefs: 60 packages and 2,791 s, Oct 8); against its area it tests its own change, and
# the area's own complete gate holds the rest before anything lands. Only codex/* and devtools/* tips:
# areas and landings are always against main. The area is the one whose fork with the tip is furthest
# past the tip's fork with main. One fetch by sha, no refs written, so gates starting together can't race.
gateBase() {
  local branch=$1 sha=$2 refs main areas mainFork best=main bestSha bestCount=0 tip name fork count
  # Main's canary (#k1n98kx): main gated against itself selects no test (Oct 9 10:42Z: tests=0.0s, and that green
  # promoted tools). It is gated against main ten landings back, so it selects what those landings touched: a real
  # selection whose answer main already holds.
  # A canary-shaped gate mutant (gate-mutants/canary-*, #fyvmsy8 hole 5) is gated the same way, so the suite can plant a
  # canary input whose ten landings select nothing.
  if [ "${branch}" = canary/main ] || [[ ${branch} == gate-mutant/canary-* ]]; then
    haveCommits "${sha}" || git -C "${here}" fetch -q --no-write-fetch-head origin "${sha}" 2> /dev/null
    echo "main~10 $(git -C "${here}" rev-parse "${sha}~10")"
    return
  fi
  if [[ ${branch} != codex/* && ${branch} != devtools/* ]]; then
    echo "main $(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)"
    return
  fi
  refs=$(git -C "${here}" ls-remote origin refs/heads/main 'refs/heads/area/*')
  main=$(echo "${refs}" | awk '$2 == "refs/heads/main" {print $1}')
  areas=$(echo "${refs}" | grep 'refs/heads/area/')
  bestSha=${main}
  if [ -n "${main}" ] && [ -n "${areas}" ] &&
     { haveCommits "${sha}" "${main}" $(echo "${areas}" | cut -f1) ||
       git -C "${here}" fetch -q origin "${sha}" "${main}" $(echo "${areas}" | cut -f1) 2> /dev/null; } &&
     mainFork=$(git -C "${here}" merge-base "${sha}" "${main}" 2> /dev/null); then
    while read -r tip name; do
      fork=$(git -C "${here}" merge-base "${sha}" "${tip}" 2> /dev/null) || continue
      # An area that already holds the tip (merged in before a regate) would leave nothing to test.
      [ "${fork}" = "${sha}" ] && continue
      git -C "${here}" merge-base --is-ancestor "${mainFork}" "${fork}" 2> /dev/null || continue
      count=$(git -C "${here}" rev-list --count "${mainFork}..${fork}")
      if [ "${count}" -gt "${bestCount}" ]; then best=${name#refs/heads/} bestSha=${tip} bestCount=${count}; fi
    done <<< "${areas}"
  fi
  echo "${best} ${bestSha}"
}
# Every named commit is already here, so there's nothing to fetch.
haveCommits() {
  local commit
  for commit in "$@"; do git -C "${here}" cat-file -e "${commit}^{commit}" 2> /dev/null || return 1; done
}
# What a tip is gated as (#11ymb02; @system_adamic, Oct 9 08:17Z): the tip merged onto its base's tip at gate start,
# never its fork point, since V3 and floor1 each went red twice on setup budgets main had already fixed. Prints
# "same <sha>" when the tip already holds the base, "merge <sha>" for a clean merge commit (first parent the base,
# second the tip), or "conflict <paths>" when the two don't merge, which is the tip's red with no gate time spent;
# returns 1 if git can't answer. The merge is made from fixed metadata (one gate identity, the base's committer
# date), so one tip on one base is one sha on every attempt and every machine.
gateMerge() {
  local sha=$1 base=$2 merged code date
  haveCommits "${sha}" "${base}" || git -C "${here}" fetch -q --no-write-fetch-head origin "${sha}" "${base}" 2> /dev/null || return 1
  if git -C "${here}" merge-base --is-ancestor "${base}" "${sha}" 2> /dev/null; then
    echo "same ${sha}"
    return
  fi
  # Exit 1 is a conflict: the first line is the partial tree, the rest the conflicted paths. Above 1, git failed.
  merged=$(git -C "${here}" merge-tree --write-tree --name-only --no-messages "${base}" "${sha}") && code=0 || code=$?
  if [ "${code}" = 1 ]; then
    echo "conflict $(echo "${merged}" | sed 1d | sed '/^$/d' | tr '\n' ' ' | sed 's/ $//')"
    return
  fi
  [ "${code}" = 0 ] || return 1
  date=$(git -C "${here}" show -s --format=%cI "${base}")
  merged=$(GIT_AUTHOR_NAME="Adamic gate" GIT_AUTHOR_EMAIL=gate@adamic.invalid GIT_AUTHOR_DATE="${date}" \
    GIT_COMMITTER_NAME="Adamic gate" GIT_COMMITTER_EMAIL=gate@adamic.invalid GIT_COMMITTER_DATE="${date}" \
    git -C "${here}" -c commit.gpgSign=false commit-tree "${merged}" -p "${base}" -p "${sha}" -m "Gate merge of ${sha} onto ${base}") || return 1
  echo "merge ${merged}"
}
