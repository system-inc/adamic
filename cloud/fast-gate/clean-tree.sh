#!/usr/bin/env bash
# Leaves a gate's checkout exactly its commit: every untracked and ignored file goes, in the tree and in every
# submodule, then the tree is checked clean. A gate box reuses its trees across candidates, so a generated file, a
# node_modules or a leftover from another sha would otherwise be there for a test to read (@system_adamic, Oct 9,
# after TestDotARename's red on 6b2c73f9 raised the question). A fresh checkout on a Codex instance is the same state.
#
#   cloud/fast-gate/clean-tree.sh <tree>
#
# Prints what it removed, and fails naming anything still there.
set -euo pipefail
tree=$1
git -C "${tree}" clean -ffdx | sed 's/^Removing /removed from the gate tree: /'
git -C "${tree}" submodule foreach --recursive --quiet 'git clean -ffdx | sed "s|^Removing |removed from the gate tree: $displaypath/|"'
left=$(git -C "${tree}" status --porcelain --ignored --ignore-submodules=none | grep -E '^(\?\?|!!)' || true)
if [ -n "${left}" ]; then
  echo "the gate tree isn't its commit after cleaning:" >&2
  echo "${left}" >&2
  exit 1
fi
