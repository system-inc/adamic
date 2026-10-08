#!/usr/bin/env bash
set -euo pipefail
[ "$#" = 3 ] || { echo 'usage: recut.sh <area-worktree> <new-adapted-tree> <new-slice>' >&2; exit 2; }
here=$(cd "$(dirname "$0")" && pwd)
area=$(realpath "$1")
tree=$(realpath -m "$2")
slice=$(realpath -m "$3")
: "${SLICE_TYPESCRIPT:?point at stock TypeScript 6.0.3 lib/typescript.js}"
: "${CENSUS_NODE_TYPES:?point at official @types/node 25.3.3}"
export CENSUS_TYPESCRIPT="$SLICE_TYPESCRIPT"
bash "$area/stage3/apply.sh" "$tree" > "$tree.apply.log" 2>&1
# The current area pipeline already carries 60-64. 65 is still absent there.
node "$here/../../adapt/65-temporary-node-process-cast/adapt.cjs" "$tree" > "$tree.temp65.log" 2>&1
bash "$area/stage3/slice/run.sh" "$tree" "$slice" \
    src/compiler/parser.ts:createSourceFile src/compiler/parser.ts:forEachChild \
    src/compiler/program.ts:flattenDiagnosticMessageText \
    src/compiler/utilitiesPublic.ts:unescapeLeadingUnderscores \
    src/compiler/types.ts:ScriptKind src/compiler/types.ts:ScriptTarget \
    src/compiler/types.ts:SyntaxKind > "$slice.gather.log" 2>&1
node "$area/stage3/slice/verify.cjs" "$slice" > "$slice.verify.log" 2>&1
cp "$here/main.a" "$slice/parser-proof-main.a"
cp "$here/kinds.a" "$slice/kinds.a"
