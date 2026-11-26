#!/usr/bin/env bash
# #0d56xpd scout, questions 1 and 3. It sources its own sandboxed-job.sh, since a shell function does not cross into bash.
set -u -o pipefail
echo "box: $(hostname) at $(git rev-parse --short HEAD 2>/dev/null || echo "${LOOM_JOB_COMMIT:-unknown}")"
work="${LOOM_JOB_SCRATCH:-$TMPDIR}/scout0d56"
mkdir -p "$work"
source cloud/sandboxed-job.sh || { echo "no sandboxed-job.sh"; exit 1; }
echo "== setup"
sandboxed_npm stage3/api stage3/typescript > "$work/npm.log" 2>&1; echo "npm exit=$?"
sandboxed_typescript_source > "$work/ts.log" 2>&1; echo "typescript source exit=$?"
export SCAN_ALL_SCRATCH="$work/scanall"; mkdir -p "$SCAN_ALL_SCRATCH"
adapted="$SCAN_ALL_SCRATCH/adapted"
bash stage3/apply.sh --build-product "$adapted" > "$work/apply.log" 2>&1; echo "apply exit=$?"
[ -d "$adapted/src/compiler" ] || { tail -40 "$work/apply.log"; echo "no adapted tree"; exit 1; }
pristine="$ADAMIC_TYPESCRIPT_SOURCE"
echo "pristine at $(git -C "$pristine" rev-parse HEAD 2>/dev/null)"

echo "== Q1 Node.parent, adapted types.ts"
start=$(grep -n "^    export interface Node extends ReadonlyTextRange" "$adapted/src/compiler/types.ts" | head -1 | cut -d: -f1)
echo "interface Node at types.ts:$start"
[ -n "$start" ] && sed -n "${start},$((start + 22))p" "$adapted/src/compiler/types.ts" | grep -n "" | sed -e "s/^/  /"
echo "pristine:"
grep -n "^    export interface Node extends ReadonlyTextRange" -A22 "$pristine/src/compiler/types.ts" | grep -E "parent|interface Node"
echo "SourceFile's own parent lines:"
start=$(grep -n "^    export interface SourceFile extends Declaration" "$adapted/src/compiler/types.ts" | head -1 | cut -d: -f1)
echo "interface SourceFile at types.ts:$start"
[ -n "$start" ] && awk -v s="$start" 'NR >= s && NR <= s + 200 && /parent/ { print "  types.ts:" NR ": " $0 }' "$adapted/src/compiler/types.ts"
echo "readonly parent declarations, pristine vs adapted:"
echo "  pristine $(grep -c 'readonly parent' "$pristine/src/compiler/types.ts") adapted $(grep -c 'readonly parent' "$adapted/src/compiler/types.ts")"
echo "lines the adaptations changed that name parent in a type or a placeholder (src/compiler):"
diff -ru "$pristine/src/compiler" "$adapted/src/compiler" 2>/dev/null | grep -E '^[-+][^-+].*(readonly parent|parent\?:|parent: Node\b|parent = undefined|parent: undefined|\.parent = |parent!)' | head -60 | cut -c1-240
echo "every parent placeholder in adapted src/compiler:"
grep -rn -E "parent = undefined!|parent: undefined!|parent!:" "$adapted/src/compiler" | sed -e "s|$adapted/||" | cut -c1-200
echo "the source file constructor:"
grep -n -E "getSourceFileConstructor|SourceFileConstructor|function Node\(|function Token\(|function Identifier\(" "$adapted/src/compiler/utilities.ts" | head -20
grep -n -E "createBaseSourceFileNode|getSourceFileConstructor" "$adapted/src/compiler/factory/baseNodeFactory.ts" "$adapted/src/compiler/factory/nodeFactory.ts" | sed -e "s|$adapted/||" | head
echo "writes of a SourceFile's parent in parser.ts and binder.ts:"
grep -n -E "sourceFile\.parent|(file|sourceFile)\)\.parent = " "$adapted/src/compiler/parser.ts" "$adapted/src/compiler/binder.ts" | sed -e "s|$adapted/||" | head
echo "accessKind:"
line=$(grep -n "^    function accessKind(node: Node): AccessKind" "$adapted/src/compiler/utilities.ts" | cut -d: -f1)
echo "  at utilities.ts:$line"
[ -n "$line" ] && sed -n "${line},$((line + 40))p" "$adapted/src/compiler/utilities.ts" | grep -n "" | sed -e "s/^/  /"
echo "tracing.ts 236 to 250:"
sed -n 236,250p "$adapted/src/compiler/tracing.ts" | grep -n "" | sed -e "s/^/  /"
echo "adaptations naming accessKind or tracing.ts:243:"
grep -rln -E "accessKind|tracing\.ts:24[0-9]" stage3/adapt | head

echo "== Q3 siblings"
node probe/0d56/siblings.cjs "$PWD/stage3/api/node_modules/typescript" "$adapted" > "$work/siblings.txt" 2> "$work/siblings.err"
echo "siblings exit=$?"
head -5 "$work/siblings.err"
grep -E "^SIB(COUNT)? |^SIB config" "$work/siblings.txt"
grep -E "^SIB	" "$work/siblings.txt" | awk -F'\t' '$3 == "number" || $3 == "boolean"' | head -300
echo "string and other results, first 40:"
grep -E "^SIB	" "$work/siblings.txt" | awk -F'\t' '$3 == "string" || $3 == "other"' | head -40
echo DONE
(cd "$work" && sha256sum siblings.txt)
