#!/usr/bin/env bash
# Proves cloud/env-owner.sh's refusal (#zb2bhp9): setup.sh and write-env.sh leave a release-owned env.sh
# byte for byte, and touch no ~/.bashrc, while write-env.sh still rewrites an unmarked one and a forced
# run still proceeds. Run: bash cloud/env-owner-test.sh. The mutant (the refusal call removed from
# write-env.sh) fails "write-env.sh refuses a marked env.sh"; see cloud/env-owner-mutants.md.
set -euo pipefail
repository=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -f -- "$work"/tools/env.sh "$work"/tools/env.sh.writing "$work"/tools/setup.lock "$work"/home/.bashrc; rmdir -- "$work"/tools/bin "$work"/tools "$work"/home "$work" 2>/dev/null || true' EXIT
mkdir -p "$work/tools" "$work/home"
failures=0
check() { if "$@"; then echo "ok: $label"; else echo "FAIL: $label"; failures=$((failures + 1)); fi; }
marked() { printf '%s\nexport PATH="/release/bin:$PATH"\n' '# adamic-env: owned by a Loom release' > "$work/tools/env.sh"; }
writeEnv() { bash "$repository/cloud/write-env.sh" "$work/tools" /tmp/adamic-gate "$work/markdown" "$repository" off ""; }

marked; before=$(cat "$work/tools/env.sh")
set +e; HOME="$work/home" ADAMIC_TOOLS="$work/tools" bash "$repository/cloud/setup.sh" > /dev/null 2>&1; status=$?; set -e
label="setup.sh exits 2 on a marked env.sh"; check test "$status" = 2
label="setup.sh leaves a marked env.sh unchanged"; check test "$(cat "$work/tools/env.sh")" = "$before"
label="setup.sh touches no ~/.bashrc"; check test ! -e "$work/home/.bashrc"
label="setup.sh stops before taking its lock or installing"; check test ! -e "$work/tools/setup.lock"

marked
set +e; writeEnv > /dev/null 2>&1; status=$?; set -e
label="write-env.sh refuses a marked env.sh"; check test "$status" = 2 -a "$(cat "$work/tools/env.sh")" = "$before"

printf 'export PATH="/old/bin:$PATH"\n' > "$work/tools/env.sh"
label="write-env.sh rewrites an unmarked env.sh"; check writeEnv
label="the rewrite is setup's text"; check grep -q 'export GOTOOLCHAIN=auto' "$work/tools/env.sh"

marked
label="a forced write-env.sh proceeds"; check env ADAMIC_SETUP_FORCE=1 bash -c "$(declare -f writeEnv); repository='$repository' work='$work' writeEnv"

[ "$failures" = 0 ] && echo "env-owner: all checks passed" || { echo "env-owner: $failures failed"; exit 1; }
