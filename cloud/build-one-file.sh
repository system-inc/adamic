#!/usr/bin/env bash
# The pinned one-file-change build benchmark on every new main, on the gate box, appended to
# documentation/velocity/build-one-file.csv (timestamp_utc,main_sha,box,file,build_seconds) on the
# branch records/build-one-file. Run it from a checkout with a push credential:
#
#   cloud/build-one-file.sh             # loop: each new main, once
#   cloud/build-one-file.sh <full sha>  # one run of that commit, then exit
#
# It takes a fast gate slot (the slot's lock and CPUs) for its minute, so its numbers are measured
# like a landing check's, not beside one; the edits themselves are cloud/build-one-file-box.sh.
set -euo pipefail

box=${ADAMIC_FAST_GATE_BOX:-threadripper}
here=$(cd "$(dirname "$0")/.." && pwd)
state=${ADAMIC_BUILD_ONE_FILE_STATE:-${HOME}/.adamic-build-one-file}
branch=records/build-one-file
csv=documentation/velocity/build-one-file.csv
mkdir -p "${state}"

measure() {
  local sha=$1 rows records
  rows=$(ssh "${box}" bash -s -- "${sha}" <<'BOX'
set -euo pipefail
sha=$1
exec 9> ~/fast-gate/lock-2
flock 9
cpus=$(nproc --all)
taskset -c "$((cpus / 2))-$((cpus - 1))" bash ~/fast-gate/tools-2/cloud/build-one-file-box.sh "${sha}"
BOX
)
  records=${state}/records
  if [ ! -d "${records}" ]; then
    git -C "${here}" fetch -q origin
    if git -C "${here}" ls-remote --exit-code origin "refs/heads/${branch}" > /dev/null; then
      git -C "${here}" worktree add -q --detach "${records}" "origin/${branch}"
    else
      git -C "${here}" worktree add -q --detach "${records}" "$(git -C "${here}" commit-tree "$(git -C "${here}" hash-object -t tree /dev/null)" -m "Start the build-one-file record")"
    fi
  fi
  git -C "${records}" fetch -q origin
  git -C "${records}" ls-remote --exit-code origin "refs/heads/${branch}" > /dev/null && git -C "${records}" switch -q --detach "origin/${branch}"
  mkdir -p "$(dirname "${records}/${csv}")"
  [ -f "${records}/${csv}" ] || echo "timestamp_utc,main_sha,box,file,build_seconds" > "${records}/${csv}"
  echo "${rows}" >> "${records}/${csv}"
  git -C "${records}" add "${csv}"
  git -C "${records}" -c user.name=kirkouimet -c user.email=kirk@kirkouimet.com commit -q -m "Build one file on main ${sha:0:12}" -m "$(printf '%s\n\nCo-Authored-By: Ahra <ahra@ahra.ai>' "${rows}")"
  git -C "${records}" push -q origin "HEAD:refs/heads/${branch}"
  echo "$(date -u +%H:%M:%S) ${sha:0:12}"
  echo "${rows}"
}

if [ -n "${1:-}" ]; then
  measure "$1"
  exit
fi
while true; do
  main=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)
  if [ "${main}" != "$(cat "${state}/last" 2>/dev/null)" ]; then
    measure "${main}" && echo "${main}" > "${state}/last"
  fi
  sleep 60
done
