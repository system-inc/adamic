#!/usr/bin/env bash
# The #1 roadmap step's landing candidate, built the moment its source moves. The top ready step that
# declares 'Branches: <glob>' may also declare 'Source: <branch>' (compiler/area-next-fixtures for 04,
# Oct 8). Each new tip of the source is merged onto main as integration builds its candidates by hand
# ("Merge <source> at <sha> onto main <sha>"), pushed as <the glob's prefix>-auto-<sha8>, and the
# fast-gate watcher gates it first on the step's reserved box. A source tip already on main, or one
# that conflicts with main, builds nothing; a conflict is told to integration once. Landing stays
# integration's: this only builds and pushes the candidate. Run it from Kirk's Mac:
#
#   cloud/landing-candidate.sh           # the loop (launchd: com.adamic.landing-candidate), every 20 s
#   cloud/landing-candidate.sh --once    # one pass
set -uo pipefail
here=$(cd "$(dirname "$0")/.." && pwd)
state=${ADAMIC_LANDING_CANDIDATE_STATE:-${HOME}/.adamic-landing-candidate}
ahraDirectory=${ADAMIC_FAST_GATE_AHRA_DIR:-/Users/kirkouimet/Projects/ahra}
mkdir -p "${state}"
touch "${state}/built"

tell() {
  # ahra refuses all-caps words.
  local text
  text=$(printf '%s' "$2" | awk '{ out = ""; while (match($0, /[A-Z][A-Z][A-Z]+/)) { out = out substr($0, 1, RSTART - 1) tolower(substr($0, RSTART, RLENGTH)); $0 = substr($0, RSTART + RLENGTH) } print out $0 }')
  (cd "${ahraDirectory}" && ahra os send "$1" "${text}" --from system_adamic_developer_tools > /dev/null 2>&1) || echo "$(date -u +%H:%M:%S) could not tell $1"
}

pass() {
  local step glob source tip main merged code tree commit name candidate ref
  read -r step glob source <<< "$(bash "${here}/cloud/first-step-branches.sh" --source 2> /dev/null | head -1)"
  [ -n "${source:-}" ] && [ -n "${glob:-}" ] || return 0
  tip=$(git -C "${here}" ls-remote origin "refs/heads/${source}" | cut -f1)
  main=$(git -C "${here}" ls-remote origin refs/heads/main | cut -f1)
  [ -n "${tip}" ] && [ -n "${main}" ] || return 0
  grep -q "^${tip} " "${state}/built" && return 0
  git -C "${here}" fetch -q origin "${tip}" "${main}" || return 0
  if git -C "${here}" merge-base --is-ancestor "${tip}" "${main}"; then
    echo "${tip} on-main ${main}" >> "${state}/built"
    return 0
  fi
  # Integration built this tip's candidate by hand already: one gate of a merge is enough.
  while read -r candidate ref; do
    [[ ${ref#refs/heads/} == ${glob} ]] || continue
    git -C "${here}" cat-file -e "${candidate}^{commit}" 2> /dev/null || git -C "${here}" fetch -q origin "${candidate}" 2> /dev/null || continue
    if git -C "${here}" log -1 --format=%P "${candidate}" | grep -qw "${tip}"; then
      echo "${tip} by-hand ${ref#refs/heads/}" >> "${state}/built"
      return 0
    fi
  done <<< "$(git -C "${here}" ls-remote origin 'refs/heads/cloud/*')"
  # merge-tree exits 1 on a conflict, still printing a tree; only a clean merge is a candidate.
  merged=$(git -C "${here}" merge-tree --write-tree "${main}" "${tip}" 2> /dev/null)
  code=$?
  tree=$(echo "${merged}" | head -1)
  if [ "${code}" != 0 ] || [ -z "${tree}" ]; then
    echo "${tip} conflict ${main}" >> "${state}/built"
    echo "$(date -u +%H:%M:%S) ${source} ${tip} conflicts with main ${main}"
    tell system_adamic_integration "Landing candidate for #${step} not built: ${source} at ${tip:0:12} conflicts with main ${main:0:12}. Integration's merge, by hand."
    return 0
  fi
  name=${glob%\*}-auto-${tip:0:8}
  commit=$(git -C "${here}" -c user.name=kirkouimet -c user.email=kirk@kirkouimet.com commit-tree "${tree}" -p "${main}" -p "${tip}" \
    -m "Merge ${source} at ${tip:0:8} onto main ${main:0:8} to land it" -m "Built by cloud/landing-candidate.sh for #${step} the moment ${source} moved." -m "Co-Authored-By: Ahra <ahra@ahra.ai>") || return 0
  git -C "${here}" push -q origin "${commit}:refs/heads/${name}" || { echo "$(date -u +%H:%M:%S) could not push ${name}"; return 0; }
  echo "${tip} ${name} ${main} ${commit}" >> "${state}/built"
  echo "$(date -u +%H:%M:%S) built ${name} ${commit} (${source} ${tip} onto main ${main})"
}

if [ "${1:-}" = --once ]; then
  pass
  exit 0
fi
echo "$(date -u +%H:%M:%S) landing candidates (tools $(git -C "${here}" rev-parse --short HEAD))"
while true; do
  pass
  sleep 20
done
