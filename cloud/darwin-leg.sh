#!/usr/bin/env bash
# The Darwin leg: does this commit's runtime compile on macOS? Builds cmd/adamic and compiles one native
# fixture (which compiles the whole runtime) on a Mac mini, one leg at a time per Mac. Linux gates can't
# see a macOS-only compile break (area-next 517cb633: node_fs_file.c, st_atimespec, Oct 8).
#
#   cloud/darwin-leg.sh <full sha>      # prints one verdict line; exit 0 green, 1 red, 2 void
#
# Hosts in order (ADAMIC_DARWIN_HOSTS, default "sun2 sun1"): an unreachable host passes to the next, and
# only when none answers is the leg void. Each host's own ~/adamic checkout is used (a fresh worktree
# there can't clone the cohere submodule), serialized by a mkdir lock with a stale-pid check.
set -uo pipefail
sha=${1:?usage: cloud/darwin-leg.sh <full sha>}
hosts=${ADAMIC_DARWIN_HOSTS:-sun2 sun1}
for host in ${hosts}; do
  ssh -o ConnectTimeout=10 -o BatchMode=yes "${host}" true 2> /dev/null || { echo "darwin leg: ${host} unreachable" >&2; continue; }
  ssh -o BatchMode=yes "${host}" bash -s -- "${sha}" <<'MAC'
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
  exit $?
done
echo "void: no Darwin host answered (${hosts})"
exit 2
