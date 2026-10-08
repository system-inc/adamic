#!/usr/bin/env bash
# The cohere submodule (with its nested TypeScript) as a cache in R2, keyed by its gitlink, so a fresh
# worker box restores it in seconds instead of cloning 81,500 files (about 287 s on a Codex box).
#
#   cloud/submodule-cache.sh publish <commit>   # on the gate box: build and upload, if not there yet
#   cloud/submodule-cache.sh restore [<repo>]    # in setup: restore it, or leave the clone to setup
#
# Only the gate box writes (its token is ~/.adamic-build-cache-token; no worker holds one). The payload
# is exactly what setup.sh's own submodule step makes (shallow, blob-filtered): cohere's worktree and
# .git/modules/cohere, gzipped, uploaded in parts under the Worker's request limit, with a manifest
# naming every part's sha256 and the whole payload's. Restore verifies every hash before extracting,
# so a worker never trusts a partial or altered payload, and falls back to the network on any doubt.
set -euo pipefail

url=${ADAMIC_BUILD_CACHE_URL:-https://adamic-build-cache.kirk-ouimet.workers.dev}
format=adamic-submodules-v1

key() { printf '%s cohere %s' "${format}" "$1" | sha256sum | cut -d' ' -f1; }
digest() { sha256sum "$1" | cut -d' ' -f1; }

publish() {
  local commit=$1 scratch gitlink name token part
  token=$(cat ~/.adamic-build-cache-token)
  scratch=$(mktemp -d)
  git clone -q --no-checkout https://github.com/system-inc/adamic.git "${scratch}/repository"
  git -C "${scratch}/repository" checkout -q --detach "${commit}"
  gitlink=$(git -C "${scratch}/repository" ls-tree HEAD cohere | awk '{print $3}')
  name=$(key "${gitlink}")
  if curl -fsS -o /dev/null "${url}/${name}.manifest"; then
    echo "cohere ${gitlink}: already cached as ${name}"
    return 0
  fi
  # The same submodule step as cloud/setup.sh, so the payload is what a worker would have made.
  git -C "${scratch}/repository" config submodule.cohere.url https://github.com/system-inc/cohere.git
  git -C "${scratch}/repository" submodule update -q --init --recursive --depth 1 --filter=blob:none
  tar -C "${scratch}/repository" -cf - cohere .git/modules | gzip -1 > "${scratch}/payload.tgz"
  split -b 90m -a 2 "${scratch}/payload.tgz" "${scratch}/part."
  {
    printf '{"format": "%s", "gitlink": "%s", "commit": "%s", "sha256": "%s", "size": %s, "parts": [' "${format}" "${gitlink}" "${commit}" "$(digest "${scratch}/payload.tgz")" "$(stat -c %s "${scratch}/payload.tgz")"
    separator=""
    for part in "${scratch}"/part.*; do
      printf '%s{"suffix": "p%s", "sha256": "%s", "size": %s}' "${separator}" "${part##*.}" "$(digest "${part}")" "$(stat -c %s "${part}")"
      separator=", "
    done
    printf ']}\n'
  } > "${scratch}/manifest.json"
  for part in "${scratch}"/part.*; do
    curl -fsS -X PUT -H "Authorization: Bearer ${token}" --data-binary "@${part}" "${url}/${name}.p${part##*.}" > /dev/null
  done
  # The manifest goes last: a payload is only visible once every part is in.
  curl -fsS -X PUT -H "Authorization: Bearer ${token}" --data-binary "@${scratch}/manifest.json" "${url}/${name}.manifest" > /dev/null
  echo "cohere ${gitlink}: cached as ${name}, $(stat -c %s "${scratch}/payload.tgz") bytes in $(ls "${scratch}"/part.* | wc -l) parts"
}

restore() {
  local repository=${1:-.} gitlink name scratch manifest suffix expected
  [ -d "${repository}/.git" ] || return 0
  # Already there (a resumed container, or setup ran before): nothing to do.
  [ -e "${repository}/cohere/.git" ] && return 0
  gitlink=$(git -C "${repository}" ls-tree HEAD cohere | awk '{print $3}')
  [ -n "${gitlink}" ] || return 0
  name=$(key "${gitlink}")
  scratch=$(mktemp -d)
  curl -fsS --max-time 30 -o "${scratch}/manifest.json" "${url}/${name}.manifest" 2>/dev/null || { echo "submodule cache: none for cohere ${gitlink}"; return 0; }
  manifest=${scratch}/manifest.json
  python3 -c 'import json,sys; m=json.load(open(sys.argv[1])); assert m["format"]==sys.argv[2] and m["gitlink"]==sys.argv[3]; print("\n".join(p["suffix"]+" "+p["sha256"] for p in m["parts"]))' "${manifest}" "${format}" "${gitlink}" > "${scratch}/parts" || { echo "submodule cache: manifest doesn't match"; return 0; }
  while read -r suffix expected; do
    curl -fsS --max-time 300 -o "${scratch}/${suffix}" "${url}/${name}.${suffix}" &
  done < "${scratch}/parts"
  wait
  while read -r suffix expected; do
    [ "$(digest "${scratch}/${suffix}")" = "${expected}" ] || { echo "submodule cache: part ${suffix} doesn't match its hash; cloning instead"; return 0; }
    cat "${scratch}/${suffix}" >> "${scratch}/payload.tgz"
  done < "${scratch}/parts"
  expected=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["sha256"])' "${manifest}")
  [ "$(digest "${scratch}/payload.tgz")" = "${expected}" ] || { echo "submodule cache: payload doesn't match its hash; cloning instead"; return 0; }
  tar -C "${repository}" -xzf "${scratch}/payload.tgz"
  echo "submodule cache: restored cohere ${gitlink}"
}

case ${1:-} in
  publish) publish "${2:?usage: submodule-cache.sh publish <commit>}" ;;
  restore) restore "${2:-.}" ;;
  *) echo "usage: submodule-cache.sh publish <commit> | restore [<repo>]" >&2; exit 2 ;;
esac
