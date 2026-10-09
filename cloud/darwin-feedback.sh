#!/usr/bin/env bash
# Darwin feedback for the #1 roadmap step: every new tip of the step's candidate branches (the fast-gate
# watcher's first-step-globs, from the step's 'Branches:' line) runs stage 3's fixtures complete on a Mac
# mini, and so does the main it is built on. The outcomes that moved are published to
# gate-logs/<sha12>/<UTC stamp>/darwin-feedback and sent to integration, in about a minute, while the
# Linux gate of record takes twenty. Feedback, not a verdict: no gate reads it until the darwin base is
# green (#a1n05b6). Run it from Kirk's Mac, where the watcher's state is:
#
#   cloud/darwin-feedback.sh                 # the loop (launchd: com.adamic.darwin-feedback), every 30 s
#   cloud/darwin-feedback.sh <branch> <sha>  # one run, published and printed, nothing sent
#
# Hosts in order (ADAMIC_DARWIN_HOSTS, default "sun2 sun1"), each its own ~/adamic under the Darwin leg's
# lock (cloud/darwin-leg.sh), so the two never check out over each other. A base's outcomes are kept on
# the Mac by sha, so a run of the same main is paid once.
set -uo pipefail
here=$(cd "$(dirname "$0")/.." && pwd)
watchState=${ADAMIC_FAST_GATE_WATCH_STATE:-${HOME}/.adamic-fast-gate-watch}
state=${ADAMIC_DARWIN_FEEDBACK_STATE:-${HOME}/.adamic-darwin-feedback}
hosts=${ADAMIC_DARWIN_HOSTS:-sun2 sun1}
mkdir -p "${state}"

# One candidate: prints the summary line, leaves the published directory in ${state}/last.
feedback() {
  local branch=$1 sha=$2 main base host out started stamp tree commit index gitDirectory logBranch
  # By sha, writing no refs: the watcher fetches main into this checkout too, and a ref lock race would void a landing.
  main=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)
  git -C "${here}" fetch -q origin "${sha}" "${main}" || { echo "void: could not fetch ${sha}"; return 2; }
  base=$(git -C "${here}" merge-base "${sha}" "${main}") || { echo "void: no merge base for ${sha}"; return 2; }
  out=$(mktemp -d)
  started=$(date -u +%s)
  for host in ${hosts}; do
    ssh -o ConnectTimeout=10 -o BatchMode=yes "${host}" true 2> /dev/null || { echo "darwin feedback: ${host} unreachable" >&2; continue; }
    if ssh -o BatchMode=yes "${host}" bash -s -- "${base}" "${sha}" > "${out}/remote.log" 2>&1 <<'MAC'
base=$1 candidate=$2
lock=~/darwin-leg.lock
until mkdir "${lock}" 2> /dev/null; do
  holder=$(cat "${lock}/pid" 2> /dev/null)
  if [ -n "${holder}" ] && ! kill -0 "${holder}" 2> /dev/null; then rm -f "${lock}/pid"; rmdir "${lock}" 2> /dev/null; fi
  sleep 2
done
echo $$ > "${lock}/pid"
trap 'rm -f "${lock}/pid"; rmdir "${lock}" 2> /dev/null' EXIT
source ~/sdk/env.sh
mkdir -p ~/darwin-feedback/npm
cd ~/adamic || { echo "void: no ~/adamic on $(hostname -s)"; exit 2; }
# stage3/api's pinned packages (@types/node for node:* imports), as the gate's npmPackages installs them:
# one install per lockfile hash, parked by hash under ~/darwin-feedback/npm when another sha needs another.
npmPackages() {
  local api=stage3/api want have
  want=$([ -f "${api}/package-lock.json" ] && shasum -a 256 "${api}/package-lock.json" | cut -c1-64)
  have=$(cat "${api}/node_modules/.lockfile-hash" 2> /dev/null)
  [ "${want}" = "${have}" ] && [ -n "${want}" ] && return 0
  [ -d "${api}/node_modules" ] && mv "${api}/node_modules" ~/darwin-feedback/npm/"${have:-unknown-$$}"
  [ -n "${want}" ] || return 0
  if [ -d ~/darwin-feedback/npm/"${want}" ]; then mv ~/darwin-feedback/npm/"${want}" "${api}/node_modules"; return 0; fi
  (cd "${api}" && npm ci --ignore-scripts --no-audit --no-fund --registry=https://registry.npmjs.org > ~/darwin-feedback/npm.log 2>&1) || return 1
  echo "${want}" > "${api}/node_modules/.lockfile-hash"
}
for sha in "${base}" "${candidate}"; do
  [ "${sha}" = "${base}" ] && [ -s ~/darwin-feedback/"${sha}".jsonl ] && continue
  git fetch -q origin "${sha}" && git checkout -q --detach "${sha}" || { echo "void: $(hostname -s) could not check out ${sha}"; exit 2; }
  git submodule update -q --init --recursive 2> /dev/null || { echo "void: $(hostname -s) could not update submodules at ${sha}"; exit 2; }
  npmPackages || { echo "void: npm ci of stage3/api failed at ${sha} on $(hostname -s):"; tail -20 ~/darwin-feedback/npm.log; exit 2; }
  # A failing test is the answer, not a void: only an empty run is.
  go test -count=1 -timeout 15m -json ./stage3/fixtures/... > ~/darwin-feedback/"${sha}".partial 2> ~/darwin-feedback/"${sha}".err
  [ -s ~/darwin-feedback/"${sha}".partial ] || { echo "void: go test printed nothing at ${sha} on $(hostname -s):"; tail -20 ~/darwin-feedback/"${sha}".err; exit 2; }
  mv ~/darwin-feedback/"${sha}".partial ~/darwin-feedback/"${sha}".jsonl
done
echo "ran on $(hostname -s)"
MAC
    then
      scp -q -o BatchMode=yes "${host}:darwin-feedback/${base}.jsonl" "${out}/base.jsonl" &&
        scp -q -o BatchMode=yes "${host}:darwin-feedback/${sha}.jsonl" "${out}/candidate.jsonl" && break
    fi
    echo "darwin feedback: ${host}: $(tail -3 "${out}/remote.log" | tr '\n' ' ')" >&2
  done
  if [ ! -s "${out}/candidate.jsonl" ]; then
    echo "void: no Darwin host ran ${sha} ($(tail -1 "${out}/remote.log" 2> /dev/null))"
    rm -f "${out}"/*; rmdir "${out}"
    return 2
  fi
  python3 "${here}/cloud/darwin-feedback.py" "${out}/base.jsonl" "${out}/candidate.jsonl" "${out}" > /dev/null
  gzip -9 "${out}/base.jsonl" "${out}/candidate.jsonl"
  printf 'feedback, not a verdict: %s on darwin (stage3/fixtures, complete), %s %s against its main %s, %s in %s s\n' \
    "$(cat "${out}/summary.txt")" "${branch}" "${sha}" "${base}" "$(tail -1 "${out}/remote.log" | sed 's/^ran on //')" "$(( $(date -u +%s) - started ))" > "${out}/status.txt"
  stamp=$(date -u +%Y%m%dT%H%M%SZ)
  logBranch=gate-logs/${sha:0:12}/${stamp}/darwin-feedback
  index=$(mktemp -u)
  gitDirectory=$(git -C "${here}" rev-parse --absolute-git-dir)
  rm -f "${out}/remote.log"
  tree=$(cd "${out}" && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" --work-tree=. add -A -f . && GIT_INDEX_FILE=${index} git --git-dir="${gitDirectory}" write-tree)
  commit=$(git -C "${here}" commit-tree "${tree}" -m "Darwin feedback for ${branch} ${sha}: $(head -1 "${out}/status.txt")")
  git -C "${here}" push -q origin "${commit}:refs/heads/${logBranch}" || echo "darwin feedback: could not publish ${logBranch}" >&2
  rm -f "${index}"
  echo "${logBranch}" > "${out}/published"
  rm -f "${state}"/last/*; rmdir "${state}/last" 2> /dev/null
  mv "${out}" "${state}/last"
  head -1 "${state}/last/status.txt"
}

if [ $# -eq 2 ]; then
  feedback "$1" "$2"
  exit $?
fi

# The step's candidate tips: "branch sha" lines matching the watcher's globs.
candidates() {
  local glob
  git -C "${here}" ls-remote origin 'refs/heads/*' | awk '{sub("refs/heads/", "", $2); print $2, $1}' |
    while read -r branch sha; do
      while read -r glob; do
        [ -n "${glob}" ] && [[ ${branch} == ${glob} ]] && { echo "${branch} ${sha}"; break; }
      done < "${watchState}/first-step-globs"
    done
}

# Tips present at the first start are the backlog, left alone.
[ -f "${state}/seen" ] || candidates > "${state}/seen"
echo "$(date -u +%H:%M:%S) darwin feedback watching $(tr '\n' ' ' < "${watchState}/first-step-globs" 2> /dev/null)(tools $(git -C "${here}" rev-parse --short HEAD))"
while true; do
  candidates > "${state}/now" 2> /dev/null
  while read -r branch sha; do
    grep -qxF "${branch} ${sha}" "${state}/seen" && continue
    echo "${branch} ${sha}" >> "${state}/seen"
    # Already landed: nothing to tell.
    main=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)
    git -C "${here}" fetch -q origin "${sha}" "${main}" 2> /dev/null
    git -C "${here}" merge-base --is-ancestor "${sha}" "${main}" 2> /dev/null && continue
    echo "$(date -u +%H:%M:%S) feedback for ${branch} ${sha}"
    line=$(feedback "${branch}" "${sha}")
    code=$?
    echo "$(date -u +%H:%M:%S) ${line}"
    [ "${code}" = 0 ] || continue
    failing=$(head -5 "${state}/last/failing.txt" | tr '\n' ' ')
    message="Darwin feedback, not a verdict, for ${branch} ${sha:0:12}: ${line#feedback, not a verdict: }. Newly failing, first five: ${failing:-none}. Log: $(cat "${state}/last/published") (moved.txt has each failure's output)."
    # ahra refuses all-caps words.
    message=$(printf '%s' "${message}" | awk '{ out = ""; while (match($0, /[A-Z][A-Z][A-Z]+/)) { out = out substr($0, 1, RSTART - 1) tolower(substr($0, RSTART, RLENGTH)); $0 = substr($0, RSTART + RLENGTH) } print out $0 }')
    (cd "${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}" &&
      ahra os send system_adamic_integration "${message}" --from system_adamic_developer_tools > /dev/null 2>&1) || echo "$(date -u +%H:%M:%S) could not tell integration"
  done < "${state}/now"
  sleep 30
done
