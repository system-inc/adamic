#!/usr/bin/env bash
# #0d56xpd scout, questions 1 and 3. It sources its own sandboxed-job.sh, since a shell function does not cross into bash.
set -u -o pipefail
echo "box: $(hostname) at $(git rev-parse --short HEAD 2>/dev/null || echo "${LOOM_JOB_COMMIT:-unknown}")"
work="${LOOM_JOB_SCRATCH:-$TMPDIR}/scout0d56"
# step prints PASS or FAIL for one setup step; a FAIL prints its log's last 30 lines and ends the job.
step() {
	local name=$1 log=$2
	shift 2
	if "$@" > "$log" 2>&1; then
		echo "PASS $name"
	else
		echo "FAIL $name (exit $?), the last 30 lines of $log:"
		tail -n 30 "$log"
		echo "VERDICT FAIL at $name"
		exit 1
	fi
}
mkdir -p "$work" && echo "PASS scratch $work" || { echo "FAIL scratch $work"; exit 1; }
# A shell function does not cross into this bash from the ask's shell, so the helpers are sourced here too.
source cloud/sandboxed-job.sh && echo "PASS sandboxed-job.sh sourced" || { echo "FAIL sandboxed-job.sh"; exit 1; }
for helper in sandboxed_npm sandboxed_typescript_source; do
	declare -F "$helper" > /dev/null && echo "PASS $helper defined" || { echo "FAIL $helper undefined"; exit 1; }
done
[ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ] && echo "PASS ADAMIC_TYPESCRIPT_SOURCE=$ADAMIC_TYPESCRIPT_SOURCE" || { echo "FAIL ADAMIC_TYPESCRIPT_SOURCE unset"; exit 1; }
step npm "$work/npm.log" sandboxed_npm stage3/api stage3/typescript
# sandboxed_typescript_source exports the copy's path, so it runs in this shell, never a subshell.
sandboxed_typescript_source > "$work/ts.log" 2>&1 && echo "PASS typescript source $ADAMIC_TYPESCRIPT_SOURCE" || { echo "FAIL typescript source"; tail -n 30 "$work/ts.log"; exit 1; }
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD 2>/dev/null)" = 050880ce59e30b356b686bd3144efe24f875ebc8 ] && echo "PASS typescript source at the pin" || { echo "FAIL typescript source not at the pin: $(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD 2>&1 | head -1)"; exit 1; }
export SCAN_ALL_SCRATCH="$work/scanall"; mkdir -p "$SCAN_ALL_SCRATCH"
adapted="$SCAN_ALL_SCRATCH/adapted"
step apply "$work/apply.log" bash stage3/apply.sh --build-product "$adapted"
[ -d "$adapted/src/compiler" ] && echo "PASS adapted tree $adapted" || { echo "FAIL no adapted tree"; exit 1; }
[ -d stage3/api/node_modules/typescript ] && echo "PASS typescript API $(node -p "require('./stage3/api/node_modules/typescript').version")" || { echo "FAIL no stage3/api/node_modules/typescript"; exit 1; }
pristine="$ADAMIC_TYPESCRIPT_SOURCE"

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
if node probe/0d56/siblings.cjs "$PWD/stage3/api/node_modules/typescript" "$adapted" > "$work/siblings.txt" 2> "$work/siblings.err"; then
	echo "PASS siblings"
else
	echo "FAIL siblings (exit $?)"
	head -20 "$work/siblings.err"
fi
grep -E "^SIB(COUNT)? |^SIB config" "$work/siblings.txt"
grep -E "^SIB	" "$work/siblings.txt" | awk -F'\t' '$3 == "number" || $3 == "boolean"' | head -300
echo "string and other results, first 40:"
grep -E "^SIB	" "$work/siblings.txt" | awk -F'\t' '$3 == "string" || $3 == "other"' | head -40
echo DONE
(cd "$work" && sha256sum siblings.txt)
