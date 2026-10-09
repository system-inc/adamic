#!/usr/bin/env bash
# The compiler dependency census on a landing's tree, for a landing with no gate record of its own (push-main
# --ruled-gate). @system_adamic, Oct 9 19:17Z, after fc7252a6 went red: floor1 landed through its ruled gate, and the
# census then refused internal/buildcache/cmd/productidentity as an undeclared compiler consumer, which a gate record
# would have caught. A ruled gate still runs the change's cheap static checks on the merged tree before it lands;
# push-main runs the lane checks for every landing, and this is the census.
#
# It runs developer tools' own selection (cloud/fast-gate/run.py --list-units, origin/devtools/fast-gate) over the
# landing commit against old main, the same code that raises "compiler dependency census: undeclared compiler
# consumers" in a gate. The tree is checked out in one long-lived worktree with cohere's submodules from the local
# clone (go list needs them), reused across landings; the tools in a second one.
#
# usage: cloud/integration/census-check.sh <old main> <landing commit>
# Prints one line and exits 0 when the census passes, or prints the census's error and exits 1.
set -euo pipefail
old=$1
landing=$2
repository=${ADAMIC_REPOSITORY:-$HOME/Projects/system/adamic}
trees=${ADAMIC_CENSUS_TREES:-$HOME/.adamic-census}
mkdir -p "$trees"
git -C "$repository" fetch -q origin devtools/fast-gate
for name in tree tools; do
	[ -d "$trees/$name" ] || git -C "$repository" worktree add -q --detach "$trees/$name" "$old"
done
git -C "$trees/tools" checkout -q -f --detach origin/devtools/fast-gate
git -C "$trees/tree" checkout -q -f --detach "$landing"
git -C "$trees/tree" -c protocol.file.allow=always -c "submodule.cohere.url=$repository/cohere" submodule update -q --init cohere
git -C "$trees/tree/cohere" -c protocol.file.allow=always -c "submodule.TypeScript.url=$repository/cohere/TypeScript" submodule update -q --init TypeScript
if output=$(timeout 300 python3 "$trees/tools/cloud/fast-gate/run.py" --list-units --tree "$trees/tree" --base "$old" --sha "$landing" --tools "$trees/tools" 2>&1); then
	echo "census: compiler dependencies declared on ${landing:0:8} against main ${old:0:8} ($(printf '%s\n' "$output" | grep -c .) fast units, tools $(git -C "$trees/tools" rev-parse --short HEAD))"
	exit 0
fi
printf '%s\n' "$output" | grep -E 'census|Error' | tail -n 3
exit 1
