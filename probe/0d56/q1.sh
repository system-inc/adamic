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
t="$adapted/src/compiler"
show() { local file=$1 from=$2 count=$3; echo "-- ${file#$adapted/}:$from"; sed -n "${from},$((from + count))p" "$file" | awk -v s="$from" '{ printf "  %d: %s\n", s + NR - 1, $0 }' | cut -c1-200; }
echo "== Q1 Node.parent"
line=$(grep -n -E '^\s*export interface Node extends ReadonlyTextRange' "$t/types.ts" | head -1 | cut -d: -f1)
[ -n "$line" ] && echo "PASS interface Node at types.ts:$line" && show "$t/types.ts" "$line" 16 || echo "FAIL no interface Node"
line=$(grep -n -E '^\s*export interface Node extends ReadonlyTextRange' "$pristine/src/compiler/types.ts" | head -1 | cut -d: -f1)
echo "pristine interface Node at types.ts:$line, its parent line:"; sed -n "${line},$((line + 16))p" "$pristine/src/compiler/types.ts" | grep -n parent
line=$(grep -n -E '^\s*export interface SourceFile extends Declaration' "$t/types.ts" | head -1 | cut -d: -f1)
[ -n "$line" ] && echo "PASS interface SourceFile at types.ts:$line" && awk -v s="$line" 'NR >= s && NR <= s + 220 && /parent/ { print "  types.ts:" NR ": " $0 }' "$t/types.ts" || echo "FAIL no interface SourceFile"
echo "readonly parent lines the adaptations added or changed in types.ts:"
diff "$pristine/src/compiler/types.ts" "$t/types.ts" | grep -E '^[<>].*\bparent\b' | head -40 | cut -c1-200
echo "== the placeholders"
show "$t/utilities.ts" 8544 40
show "$t/utilities.ts" 8592 10
show "$t/utilities.ts" 12488 12
show "$t/binder.ts" 590 10
echo "== accessKind"
line=$(grep -n -E '^\s*function accessKind\(' "$t/utilities.ts" | head -1 | cut -d: -f1)
[ -n "$line" ] && echo "PASS accessKind at utilities.ts:$line" && show "$t/utilities.ts" "$line" 40 || echo "FAIL no accessKind"
echo "its callers in src/compiler:"
grep -rn -E '\baccessKind\(' "$t" | grep -v 'function accessKind' | sed -e "s|$adapted/||" | cut -c1-200
grep -rn -E '\bisWriteAccess\(|\bisWriteOnlyAccess\(' "$t" | sed -e "s|$adapted/||" | cut -c1-160 | head -30
echo DONE
