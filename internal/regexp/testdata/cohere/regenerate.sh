#!/usr/bin/env bash
# Run from any directory. Source cloud/setup.sh's printed env.sh first.
set -euo pipefail
corpus=$(cd "$(dirname "$0")" && pwd)
repository=$(cd "$corpus/../../../.." && pwd)
dependencies=$(mktemp -d /tmp/adamic-cohere-regex-npm.XXXXXX)
# Deliberately leave the small scratch install for inspection; no repo dependencies.
npm install --prefix "$dependencies" typescript@6.0.3 @eslint-community/regexpp@4.12.2
export NODE_PATH="$dependencies/node_modules"
cd "$repository"
node "$corpus/extract.cjs"
node "$corpus/trace-selector.cjs"
node "$corpus/generate.cjs"
# Regenerating Node observations does not silently bless Adamic failures.
printf '%s\n' 'Regenerated Node observations. Review them, then explicitly record or verify Adamic as described in README.md.'
