#!/usr/bin/env bash
# The Darwin leg: does this commit's runtime compile on macOS? Builds cmd/adamic and compiles one native
# fixture (which compiles the whole runtime) on a Mac mini, one leg at a time per Mac. Linux gates can't
# see a macOS-only compile break (area-next 517cb633: node_fs_file.c, st_atimespec, Oct 8).
#
#   cloud/darwin-leg.sh <full sha> [<branch>]   # prints one verdict line first; exit 0 green, 1 red, 2 void
#
# For a landing (cloud/land-*), a green compile is followed by the vanish check (the witness's rule,
# @system_adamic Oct 8): stage 3's fixtures run complete on the candidate and on the main it is built on
# (cloud/darwin-feedback.sh, published as darwin-feedback), and any result that passed on main and is
# absent from the candidate turns the landing red unless the candidate declares it moved on purpose:
# a 'Moved-result: <old> -> <new or removed>' trailer on a commit it carries past main, or a line
# '<old> -> <new or removed>' in stage3/fixtures/moved-results.txt. Naming a test covers its subtests.
#
# Hosts in order (ADAMIC_DARWIN_HOSTS, default "sun2 sun1"): an unreachable host passes to the next, and
# only when none answers is the leg void. Each host's own ~/adamic checkout is used (a fresh worktree
# there can't clone the cohere submodule), serialized by a mkdir lock with a stale-pid check.
set -uo pipefail
sha=${1:?usage: cloud/darwin-leg.sh <full sha> [<branch>]}
branch=${2:-}
here=$(cd "$(dirname "$0")/.." && pwd)
hosts=${ADAMIC_DARWIN_HOSTS:-sun2 sun1}

# The vanish check on host $1: prints its verdict line first when it isn't green, else one summary line.
vanishCheck() {
  local host=$1 checked line base undeclared code=0
  checked=$(mktemp -d)
  line=$(ADAMIC_DARWIN_FEEDBACK_STATE="${checked}" ADAMIC_DARWIN_HOSTS="${host}" bash "${here}/cloud/darwin-feedback.sh" "${branch}" "${sha}" 2> /dev/null | tail -1)
  if [ ! -f "${checked}/last/vanished.txt" ]; then
    echo "void: ${sha} darwin leg's vanish check ran nothing on ${host}: ${line}"
    code=2
  else
    base=$(sed -nE 's/.* against its main ([0-9a-f]{40}).*/\1/p' "${checked}/last/status.txt")
    { git -C "${here}" log --format=%B "${base}..${sha}" | sed -n 's/^Moved-result: *//p'
      git -C "${here}" show "${sha}:stage3/fixtures/moved-results.txt" 2> /dev/null; } > "${checked}/declared.txt"
    undeclared=$(python3 "${here}/cloud/darwin-feedback.py" --undeclared "${checked}/last/vanished.txt" "${checked}/declared.txt")
    if [ -n "${undeclared}" ]; then
      echo "red: ${sha} darwin leg: $(echo "${undeclared}" | wc -l | tr -d ' ') results that pass on main ${base:0:12} are absent with no Moved-result declaring them: $(echo "${undeclared}" | head -5 | awk '{print $2}' | tr '\n' ' ')(darwin-feedback: $(cat "${checked}/last/published"))"
      echo "${undeclared}"
      code=1
    else
      echo "vanish check: $(grep -c . "${checked}/last/vanished.txt") passing results absent, every one declared ($(cat "${checked}/last/published"))"
    fi
  fi
  rm -f "${checked}"/last/* "${checked}/declared.txt"
  rmdir "${checked}/last" "${checked}" 2> /dev/null
  return "${code}"
}

for host in ${hosts}; do
  ssh -o ConnectTimeout=10 -o BatchMode=yes "${host}" true 2> /dev/null || { echo "darwin leg: ${host} unreachable" >&2; continue; }
  compiled=$(mktemp)
  ssh -o BatchMode=yes "${host}" bash -s -- "${sha}" > "${compiled}" 2>&1 <<'MAC'
sha=$1
lock=~/darwin-leg.lock
until mkdir "${lock}" 2> /dev/null; do
  holder=$(cat "${lock}/pid" 2> /dev/null)
  if [ -n "${holder}" ] && ! kill -0 "${holder}" 2> /dev/null; then rm -f "${lock}/pid"; rmdir "${lock}" 2> /dev/null; fi
  sleep 2
done
echo $$ > "${lock}/pid"
trap 'rm -f "${lock}/pid"; rmdir "${lock}" 2> /dev/null' EXIT
source ~/sdk/env.sh
cd ~/adamic || { echo "void: no ~/adamic on $(hostname -s)"; exit 2; }
git fetch -q origin "${sha}" && git checkout -q --detach "${sha}" || { echo "void: $(hostname -s) could not check out ${sha}"; exit 2; }
# The cohere pin rarely moves; when it does, a host without credentials can't follow, and that's the host.
git submodule update -q --init --recursive 2> /dev/null || { echo "void: $(hostname -s) could not update submodules at ${sha}"; exit 2; }
mkdir -p ~/darwin-leg
started=$(date +%s)
if ! go build -o ~/darwin-leg/adamic ./cmd/adamic > ~/darwin-leg/build.log 2>&1; then
  echo "red: darwin build of cmd/adamic failed on $(hostname -s):"; tail -20 ~/darwin-leg/build.log; exit 1
fi
fixture=$(ls internal/oracle/testdata/*.a | head -1)
if ! ~/darwin-leg/adamic build "${fixture}" -o ~/darwin-leg/out > ~/darwin-leg/compile.log 2>&1; then
  echo "red: darwin compile of the runtime and ${fixture} failed on $(hostname -s):"; grep -m20 -E "error|Error" ~/darwin-leg/compile.log || tail -20 ~/darwin-leg/compile.log; exit 1
fi
echo "green: darwin build and compile of ${fixture} on $(hostname -s) in $(( $(date +%s) - started )) s"
MAC
  code=$?
  if [ "${code}" = 0 ] && [[ ${branch} == cloud/land-* ]]; then
    vanished=$(vanishCheck "${host}")
    code=$?
    # One verdict line, and first: a red or void vanish check leads, the compile's green line follows as a detail.
    if [ "${code}" != 0 ]; then
      echo "${vanished}"
      sed 's/^green: /compile: green: /' "${compiled}"
    else
      cat "${compiled}"
      echo "${vanished}"
    fi
  else
    cat "${compiled}"
  fi
  rm -f "${compiled}"
  exit "${code}"
done
echo "void: no Darwin host answered (${hosts})"
exit 2
