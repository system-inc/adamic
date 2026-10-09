#!/usr/bin/env bash
# The 90-second probe (#c89weq5, for Outcome #awn479j: a change re-tests only the strands it touched, each fetching its
# build cache, building and testing in under 90 s). It times one fixed, typical change end to end through the normal
# candidate path, every 30 minutes under launchd (cloud/probe-90/com.adamic.probe-90.plist):
#
#   cloud/probe-90.sh            # the leaf probe
#   cloud/probe-90.sh --all      # the forced-wide mutant (or ADAMIC_PROBE_FORCE_ALL=1): its wall must read larger
#   cloud/probe-90.sh --dry-run  # everything but the push: prints the branch and the commit it would gate
#
# One throwaway commit on main's tip appends one comment line (gofmt-clean) to one fixed file and is pushed as
# devtools/probe-90-<UTC stamp>, a devtools/* tip the watcher (cloud/fast-gate-watch.sh) gates against main like any
# candidate: Loom's pool first, a box racing it when the pool is slow. The probe waits for the first green or red
# record under gate-logs/<sha12>/, reads it as five phases (cloud/probe-90/record.py says how each is mapped), posts
# ONE machine status line on the outcome with `ahra tasks status <task> "<line>" --auto`, and deletes its branch.
#
# The leaf: cmd/adamic-stage1-progress/main_test.go. Nothing in the module imports cmd/adamic-stage1-progress (go list:
# no Imports, TestImports or XTestImports name it), a _test.go edit selects its own package alone (run.py touched()),
# no executors.txt reads rule or compiler path matches it, and its five tests ran in 0.0 to 0.5 s summed across seven
# fast records on Oct 9 with no failure. Kept fixed so every probe's number is comparable. stage3/ leaves are out (a
# stage3/ path runs the stage 3 lane and classes big), and statecopy is the gate-mutant control's.
# The mutant: one comment line in internal/load/load.go, a production file of the package most of the module imports
# (46 of 85 packages; its reverse closure holds 42 of the 60 test packages), and a compiler path, so nothing is deferred
# and the oracle runs whole: the widest selection the gate makes short of a --full run.
set -uo pipefail

here=$(cd "$(dirname "$0")/.." && pwd)
helper=${here}/cloud/probe-90/record.py
state=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}
ahraDirectory=${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}
watchLog=${ADAMIC_PROBE_WATCH_LOG:-/Users/kirkouimet/Projects/system/adamic-gate-logs/fast-gate-watch.log}
jobs=${LOOM_FAST_JOBS:-${HOME}/.loom/jobs/fast}
task=${ADAMIC_PROBE_TASK:-awn479j}
leafFile=cmd/adamic-stage1-progress/main_test.go
wideFile=internal/load/load.go

mode=leaf dryRun=no
[ "${ADAMIC_PROBE_FORCE_ALL:-}" = 1 ] && mode=all
for argument in "$@"; do
  case ${argument} in
    --all) mode=all ;;
    --dry-run) dryRun=yes ;;
    *) echo "usage: cloud/probe-90.sh [--all] [--dry-run]" >&2; exit 2 ;;
  esac
done
# Under launchd the next start is 30 minutes on, so a probe gives up before it: 35 minutes for the leaf. The wide
# mutant runs most of the module's tests and is run by hand.
limit=${ADAMIC_PROBE_SECONDS:-$([ "${mode}" = all ] && echo 3300 || echo 2100)}
file=${leafFile}
[ "${mode}" = all ] && file=${wideFile}

say() {
  echo "$(date -u +%H:%M:%S) $*"
}
post() {
  if [ "${dryRun}" = yes ] || [ "${ADAMIC_PROBE_POST:-1}" = 0 ]; then
    say "would post on #${task}: $1"
    return 0
  fi
  # ahra takes at most 160 characters on one line.
  local line
  line=$(printf '%s' "$1" | tr '\n' ' ' | cut -c1-160)
  (cd "${ahraDirectory}" && ahra tasks status "${task}" "${line}" --auto) && say "posted on #${task}: ${line}"
}

stamp=$(date -u +%Y%m%dT%H%M%SZ)
branch=$(python3 "${helper}" branch "${stamp}" "${mode}") || exit 2
if glob=$(python3 "${helper}" skipped "${branch}" "${state}/skip-globs"); then
  say "not probing: ${branch} is skipped by ${glob} in ${state}/skip-globs"
  exit 2
fi

main=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)
[[ ${main} =~ ^[0-9a-f]{40}$ ]] || { say "not probing: origin's main is unreadable"; exit 1; }
git -C "${here}" cat-file -e "${main}^{commit}" 2> /dev/null || git -C "${here}" fetch -q --no-write-fetch-head origin "${main}" || exit 1

# The commit is made from main's tree in a scratch index, so no checkout moves and the watcher's own tree is untouched.
scratch=$(mktemp -d)
pushed=""
cleanup() {
  if [ -n "${pushed}" ]; then
    git -C "${here}" push -q origin --delete "${branch}" && say "deleted ${branch}"
  fi
  rm -f "${scratch}"/*
  rmdir "${scratch}"
}
trap cleanup EXIT
fileMode=$(git -C "${here}" ls-tree "${main}" -- "${file}" | awk '{print $1}')
[ -n "${fileMode}" ] || { say "not probing: ${file} is not on main ${main:0:12}"; exit 1; }
# A blank line first, as gofmt sets a comment after a closing brace.
{ git -C "${here}" show "${main}:${file}"; echo; echo "// probe-90 ${stamp}: a throwaway line the 90-second probe gates, never merged"; } > "${scratch}/file"
blob=$(git -C "${here}" hash-object -w "${scratch}/file")
export GIT_INDEX_FILE=${scratch}/index
git -C "${here}" read-tree "${main}"
git -C "${here}" update-index --cacheinfo "${fileMode},${blob},${file}"
tree=$(git -C "${here}" write-tree)
unset GIT_INDEX_FILE
sha=$(GIT_AUTHOR_NAME=kirkouimet GIT_AUTHOR_EMAIL=kirk@kirkouimet.com GIT_COMMITTER_NAME=kirkouimet GIT_COMMITTER_EMAIL=kirk@kirkouimet.com \
  git -C "${here}" -c commit.gpgSign=false commit-tree "${tree}" -p "${main}" \
  -m "Probe-90 (${mode}): one comment line in ${file}, gated to time the 90-second verdict" \
  -m "Throwaway: cloud/probe-90.sh deletes this branch once its record publishes. Never merge it." \
  -m "Co-Authored-By: Ahra <ahra@ahra.ai>") || exit 1
say "${branch} ${sha}: ${mode} probe of ${file} on main ${main:0:12}"
if [ "${dryRun}" = yes ]; then
  say "dry run: not pushed"
  exit 0
fi

git -C "${here}" push -q origin "${sha}:refs/heads/${branch}" || { say "push of ${branch} failed"; exit 1; }
pushed=$(date -u +%s)
deadline=$((pushed + limit))

# The first green or red record of the sha. A void (the watcher queues it again) or a lost race's stopped record is no
# verdict, so it is passed over.
record="" checked=" "
while [ -z "${record}" ]; do
  if [ "$(date -u +%s)" -ge "${deadline}" ]; then
    post "probe-90 ${mode}: no green or red record for ${sha:0:9} within ${limit}s of its push (branch ${branch}); see the watcher log"
    exit 1
  fi
  # The watcher said it won't gate it: no record is coming.
  refused=$(grep -E "(skipped by|already on main|superseded|given up).* ${branch} ${sha}|${branch} ${sha}.*given up" <(tail -n 20000 "${watchLog}" 2> /dev/null) | tail -1)
  if [ -n "${refused}" ]; then
    post "probe-90 ${mode}: the watcher will not gate ${sha:0:9}: $(echo "${refused}" | cut -d' ' -f2- | cut -c1-90)"
    exit 1
  fi
  while read -r commit reference; do
    [ -n "${commit}" ] || continue
    [[ ${checked} == *" ${commit} "* ]] && continue
    checked="${checked}${commit} "
    git -C "${here}" fetch -q --no-write-fetch-head origin "${reference}" 2> /dev/null || { checked=${checked/ ${commit} / }; continue; }
    status=$(git -C "${here}" show "${commit}:status.txt" 2> /dev/null | head -1)
    case ${status} in
      green:* | red:*) record=${reference#refs/heads/} recordCommit=${commit}; break ;;
      *) say "passed over ${reference#refs/heads/}: ${status:0:100}" ;;
    esac
  done < <(git -C "${here}" ls-remote origin "refs/heads/gate-logs/${sha:0:12}/*" 2> /dev/null | sort -k2)
  [ -n "${record}" ] || { sleep 20 & wait $!; }
done

for name in fast.json status.txt record.jsonl record.jsonl.gz; do
  git -C "${here}" show "${recordCommit}:${name}" > "${scratch}/${name}" 2> /dev/null || rm -f "${scratch}/${name}"
done
recordStamp=$(echo "${record}" | cut -d/ -f3)
stampEpoch=$(date -u -j -f %Y%m%dT%H%M%SZ "${recordStamp}" +%s)
commitEpoch=$(git -C "${here}" show -s --format=%ct "${recordCommit}")
# The tools the gate ran: the pool job's (cloud/pool-job.sh writes them), else the box record's, else the good tools.
tools=$(python3 -c 'import json, sys; print(json.load(open(sys.argv[1])).get("tools", ""))' "${jobs}/${sha}.json" 2> /dev/null)
[ -n "${tools}" ] || tools=$(python3 -c 'import json, sys; print(json.load(open(sys.argv[1])).get("tools_sha", ""))' "${scratch}/fast.json" 2> /dev/null)
[ -n "${tools}" ] || tools=$(cat "${state}/tools-good" 2> /dev/null)
line=$(python3 "${helper}" line "${scratch}" --mode "${mode}" --sha "${sha}" --ref "${record}" --tools "${tools}" \
  --stamp-epoch "${stampEpoch}" --commit-epoch "${commitEpoch}" --pushed-epoch "${pushed}") || { say "could not read ${record}"; exit 1; }
say "record ${record} for ${branch} ${sha}"
post "${line}" || exit 1
