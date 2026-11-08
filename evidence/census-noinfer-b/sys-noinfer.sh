#!/usr/bin/env bash
# sys.sh: the latent census of the adapted src/compiler/sys.ts, measured with this checkout's compiler.
#
# Runs as `bash -c "$(cat sys.sh)"` from the root of a fresh detached adamic checkout at the compiler under
# test, on a Linux build box whose ~/adamic-tools/env.sh provides go, clang 20 and node. It prepares the way
# cloud/meter-daily/run.sh and stage3/meter/twice-daily.sh do (submodules, the pinned upstream TypeScript from
# stage3/source.json, stage3/api's lockfile packages, stage3/apply.sh), builds census and latent-census the way
# stage3/adapt/99-correctness-core-levenshtein/prepare-census.sh does, then runs review/apple-sys-scout's
# measure.sh commands on sys.ts: census, latent census first-error, latent census full.
#
# Every step is bounded by `timeout` and by one deadline 29 minutes after the start, logs to $OUT/<step>.log
# and writes its exit code to $OUT/<step>.exit. The box's filesystem is gone afterwards, so the end prints a
# summary and the gzipped jsonl base64-encoded between BEGIN and END lines on stdout.
set -uo pipefail

started=$(date +%s)
# 29 minutes for the whole job; the last two are kept for the summary and the payload.
deadline=$((started + 29 * 60))
reserve=120
payloadLimit=4900000
repository=$(pwd)
commit=$(git rev-parse HEAD)
OUT=${OUT:-/tmp/rule-g-census/$(git rev-parse --short HEAD)}
# apply.sh refuses an existing output, and a rerun on the same box must not mix two runs' files.
if [ -e "$OUT" ] && [ -n "$(ls -A "$OUT" 2> /dev/null)" ]; then
	OUT=$OUT.$(date -u +%Y%m%dT%H%M%SZ)
fi
mkdir -p "$OUT/bin"

say() { printf '%s sys.sh: %s\n' "$(date -u +%FT%TZ)" "$*" | tee -a "$OUT/steps.txt"; }

# step <name> <cap seconds> <command...>: runs the command under timeout, capped by the time left before the
# deadline, with its output in $OUT/<name>.log, its exit code in $OUT/<name>.exit and its wall seconds beside.
step() {
	local name=$1 cap=$2
	shift 2
	local remaining=$((deadline - reserve - $(date +%s)))
	if [ "$remaining" -le 0 ]; then
		printf 'skipped: no time left before the deadline\n' > "$OUT/$name.log"
		printf '124\n' > "$OUT/$name.exit"
		say "$name skipped, no time left"
		return 124
	fi
	[ "$cap" -le "$remaining" ] || cap=$remaining
	local stepStarted
	stepStarted=$(date +%s)
	timeout --kill-after=10s "$cap" "$@" > "$OUT/$name.log" 2>&1
	local code=$?
	printf '%s\n' "$code" > "$OUT/$name.exit"
	printf '%s\n' "$(($(date +%s) - stepStarted))" > "$OUT/$name.seconds"
	say "$name exit $code after $(cat "$OUT/$name.seconds")s (cap ${cap}s)"
	return "$code"
}

# skip <name> <why>: a step that cannot run because one it needs failed.
skip() {
	printf 'skipped: %s\n' "$2" > "$OUT/$1.log"
	printf 'skipped\n' > "$OUT/$1.exit"
	say "$1 skipped: $2"
}

say "commit $commit, output $OUT"

# The job's first step, as the brief requires: cohere and its nested TypeScript at their gitlinks.
step submodules 300 git -c protocol.file.allow=always submodule update --init --recursive --depth 1
git submodule status --recursive > "$OUT/submodule-pins.txt" 2>&1

# The toolchain. env.sh is sourced into this shell, so it is logged and its exit kept but not under timeout.
set +u
# shellcheck disable=SC1090
. ~/adamic-tools/env.sh > "$OUT/env.log" 2>&1
printf '%s\n' "$?" > "$OUT/env.exit"
set -u
unset LATENT_SPECULATIVE LATENT_MUTANT_NO_STUBS
{
	printf 'node %s\n' "$(node --version 2>&1)"
	go version 2>&1
	clang --version 2>&1 | head -n 1
	python3 --version 2>&1
	git --version 2>&1
	printf 'npm %s\n' "$(npm --version 2>&1)"
	uname -a
	nproc
} > "$OUT/toolchain.txt" 2>&1
python3 stage3/meter/box.py "$OUT/box.json" start > "$OUT/box.log" 2>&1 || true

# npm, bootstrapped from cloud/markdown-width/npm-bootstrap.json's pin when the toolchain has none, as
# cloud/meter-daily/run.sh does, since stage3/api's packages come from `npm ci`.
if ! command -v npm > /dev/null 2>&1; then
	bootstrap=$repository/cloud/markdown-width/npm-bootstrap.json
	npmHome=$OUT/npm
	mkdir -p "$npmHome/bin"
	step npm-download 120 curl -fsSL "$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["url"])' "$bootstrap")" -o "$npmHome/npm.tgz"
	step npm-verify 30 python3 -I -c '
import base64, hashlib, json, sys
algorithm, expected = json.load(open(sys.argv[1]))["integrity"].split("-", 1)
actual = base64.b64encode(hashlib.sha512(open(sys.argv[2], "rb").read()).digest()).decode()
sys.exit(0 if algorithm == "sha512" and actual == expected else "npm bootstrap integrity mismatch")
' "$bootstrap" "$npmHome/npm.tgz" &&
		step npm-extract 60 tar --no-same-owner -xzf "$npmHome/npm.tgz" -C "$npmHome" &&
		printf '#!/bin/sh\nexec node %q "$@"\n' "$npmHome/package/bin/npm-cli.js" > "$npmHome/bin/npm" &&
		chmod 755 "$npmHome/bin/npm"
	export PATH="$npmHome/bin:$PATH"
	printf 'npm %s (bootstrapped)\n' "$(npm --version 2>&1)" >> "$OUT/toolchain.txt"
fi

# The adapted tree: the pinned upstream, stage3/api's packages, then stage3/apply.sh into $OUT/adapted.
prepareTree() {
	local pin tag url
	pin=$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["commit"])' stage3/source.json)
	tag=$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["tag"])' stage3/source.json)
	url=$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["repository"])' stage3/source.json)
	# A gate box's env.sh may name a checkout of the pin already; apply clones it again into its output, so the
	# adapted tree is a fresh copy either way. Anything else is cloned from the tag, as twice-daily.sh does.
	if [ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ] && [ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD 2> /dev/null)" = "$pin" ]; then
		printf 'using ADAMIC_TYPESCRIPT_SOURCE=%s at %s\n' "$ADAMIC_TYPESCRIPT_SOURCE" "$pin" > "$OUT/upstream.log"
		printf '0\n' > "$OUT/upstream.exit"
	else
		step upstream 300 git clone --depth 1 --branch "$tag" "$url" "$OUT/upstream" || return 1
		export ADAMIC_TYPESCRIPT_SOURCE=$OUT/upstream
	fi
	git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD >> "$OUT/upstream.log" 2>&1
	step packages 300 bash -c 'cd "$1" && npm ci --ignore-scripts --no-audit --no-fund' _ "$repository/stage3/api" || return 1
	# A cache of this run's own, so the tree is built from the pinned upstream here rather than restored from
	# the box's product cache or ADAMIC_BUILD_CACHE_URL (the product key covers only stage3's own inputs).
	export STAGE3_CACHE=$OUT/stage3-cache
	unset STAGE3_PRODUCT_STORE ADAMIC_BUILD_CACHE_URL
	bash stage3/apply.sh --key > "$OUT/apply-key.txt" 2>&1
	step apply 900 bash stage3/apply.sh "$OUT/adapted" || return 1
	[ -f "$OUT/adapted/src/compiler/sys.ts" ] || { say "apply left no src/compiler/sys.ts"; return 1; }
	# Correctness's C-noinfer-b slot (typescript/sys-timer-slot 949b9c3c), applied to the adapted tree after
	# apply, as her scratch experiment was: the System interface's setTimeout and clearTimeout.
	python3 - "$OUT/adapted/src/compiler/sys.ts" > "$OUT/noinfer.log" 2>&1 <<'PY'
import sys
p = sys.argv[1]; s = open(p).read()
pairs = [("declare function setTimeout(handler: () => void, timeout: number): NodeJS.Timeout;\n", ""),
         ("declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;\n", ""),
         ("        pollScheduled: boolean;", "        pollScheduled: unknown;"),
         ("    let pollScheduled: any;", "    let pollScheduled: unknown;"),
         ("setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;", "setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;"),
         ("clearTimeout?(timeoutId: any): void;", "clearTimeout?(timeoutId: unknown): void;")]
for old, new in pairs:
    n = s.count(old)
    print("replaced", n, "of", old)
    if n != 1: sys.exit("expected exactly one")
    s = s.replace(old, new)
open(p, "w").write(s)
PY
	echo "noinfer exit=$?" | tee -a "$OUT/noinfer.log"
	grep -q 'noinfer exit=0' "$OUT/noinfer.log" || { say "the C-noinfer-b slot didn't apply"; return 1; }
}

# census and latent-census from this checkout. latent-census follows prepare-census.sh: make_overlay.py's
# overlay, its .go replacements renamed .go.txt, gofmt, then go build with the overlay.
buildTools() {
	step census-build 600 go build -o "$OUT/bin/census" ./stage3/census/tool
	local censusCode=$?
	step latent-overlay 600 python3 stage3/census/latent/make_overlay.py "$repository" "$OUT/latent-overlay" || return 1
	step latent-overlay-rename 60 python3 -I -c '
from pathlib import Path
import json, sys
p = Path(sys.argv[1]); data = json.loads(p.read_text())
for k, v in data["Replace"].items():
    path = Path(v)
    if path.suffix == ".go":
        new = path.with_suffix(".go.txt"); path.rename(new); data["Replace"][k] = str(new)
p.write_text(json.dumps(data, indent=2) + "\n")
' "$OUT/latent-overlay/overlay.json" || return 1
	step latent-gofmt 120 bash -c 'gofmt -w "$1"/*.go.txt' _ "$OUT/latent-overlay" || return 1
	step latent-build 900 go build -buildvcs=false -overlay="$OUT/latent-overlay/overlay.json" -o "$OUT/bin/latent-census" ./stage3/census/latent/tool || return 1
	return "$censusCode"
}

# The tree and the tools are independent, so they are prepared side by side.
(prepareTree) &
treePid=$!
buildTools
toolsCode=$?
wait "$treePid"
treeCode=$?
say "prepare: tree exit $treeCode, tools exit $toolsCode"

entry=$OUT/adapted/src/compiler/sys.ts
if [ "$treeCode" = 0 ] && [ -x "$OUT/bin/census" ]; then
	step census 300 "$OUT/bin/census" "$entry" "$OUT/census.jsonl"
else
	skip census "no adapted tree or no census binary"
fi
if [ "$treeCode" = 0 ] && [ -x "$OUT/bin/latent-census" ]; then
	step first 480 env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$OUT/first-times.jsonl" "$OUT/bin/latent-census" "$entry" "$OUT/first.jsonl"
	step full 720 env LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$OUT/full-times.jsonl" "$OUT/bin/latent-census" "$entry" "$OUT/full.jsonl"
else
	skip first "no adapted tree or no latent-census binary"
	skip full "no adapted tree or no latent-census binary"
fi
python3 stage3/meter/box.py "$OUT/box.json" end >> "$OUT/box.log" 2>&1 || true

# Counts for every stream, small enough to print whole.
python3 -I - "$OUT/counts.json" census:census:"$OUT/census.jsonl" first:latent:"$OUT/first.jsonl" full:latent:"$OUT/full.jsonl" \
	first-times:times:"$OUT/first-times.jsonl" full-times:times:"$OUT/full-times.jsonl" > "$OUT/counts.log" 2>&1 << 'PYTHON'
import collections, json, sys

def ranked(counter, limit):
    return [[list(key) if isinstance(key, tuple) else key, count] for key, count in counter.most_common(limit)]

summary = {}
for spec in sys.argv[2:]:
    name, kind, path = spec.split(':', 2)
    try:
        lines = open(path, encoding='utf-8').read().splitlines()
    except OSError as error:
        summary[name] = {'missing': str(error)}
        continue
    records, unparsed = [], 0
    for line in lines:
        if line.strip():
            try:
                records.append(json.loads(line))
            except ValueError:
                unparsed += 1
    entry = {'records': len(records), 'unparsed_lines': unparsed}
    if kind == 'census':
        entry['by_kind'] = dict(collections.Counter(str(record.get('kind')) for record in records))
        entry['seconds'] = round(sum(record.get('seconds') or 0 for record in records), 3)
        entry['top_reasons'] = ranked(collections.Counter((str(record.get('kind')), record.get('reason')) for record in records if record.get('reason')), 15)
    elif kind == 'latent':
        header = next((record for record in records if 'file' not in record), None)
        if header is not None:
            entry['header'] = {key: header.get(key) for key in ('status', 'latent_mode', 'checker_rejected', 'measurement')}
            entry['header']['diagnostics'] = len(header.get('diagnostics') or [])
            entry['header']['sources'] = len(header.get('sources') or [])
        files = [record for record in records if 'file' in record]
        entry['file_records'] = len(files)
        entry['units_by_status'] = dict(collections.Counter(str(unit.get('status')) for record in files for unit in record.get('units') or []))
        entry['raw_findings_by_kind'] = dict(collections.Counter(str(finding.get('kind')) for record in files for finding in record.get('findings') or []))
        unique = {(finding.get('kind'), finding.get('where'), finding.get('reason'), finding.get('text'))
                  for record in files for finding in record.get('findings') or []}
        entry['unique_findings_by_kind'] = dict(collections.Counter(str(kind) for kind, _, _, _ in unique))
        entry['top_reasons'] = ranked(collections.Counter((str(kind), reason) for kind, _, reason, _ in unique if reason), 15)
        byFile = collections.Counter()
        for record in files:
            byFile[record['file']] += len({(f.get('kind'), f.get('where'), f.get('reason'), f.get('text')) for f in record.get('findings') or []})
        entry['top_files_by_unique_findings'] = ranked(byFile, 10)
    elif kind == 'times':
        entry['wall_seconds'] = round(sum(record.get('wall_seconds') or 0 for record in records), 3)
        entry['cpu_seconds'] = round(sum(record.get('cpu_seconds') or 0 for record in records), 3)
        slowest = sorted(records, key=lambda record: record.get('wall_seconds') or 0, reverse=True)[:5]
        entry['slowest_units'] = [[record.get('where'), record.get('wall_seconds')] for record in slowest]
    summary[name] = entry
with open(sys.argv[1], 'w') as output:
    json.dump(summary, output, indent=1, sort_keys=True)
    output.write('\n')
PYTHON

# The summary: every step's exit, the last lines of each failed step's log, the counts.
outcome=0
echo
echo "=== rule-g-census sys.sh summary: commit $commit, output $OUT, $(($(date +%s) - started))s ==="
cat "$OUT/toolchain.txt"
for exitFile in "$OUT"/*.exit; do
	name=$(basename "$exitFile" .exit)
	code=$(cat "$exitFile")
	seconds=$(cat "$OUT/$name.seconds" 2> /dev/null || printf '-')
	printf '%-24s exit %-8s %ss\n' "$name" "$code" "$seconds"
done
for name in submodules upstream packages apply census-build latent-overlay latent-overlay-rename latent-gofmt latent-build census first full; do
	[ -f "$OUT/$name.exit" ] || continue
	if [ "$(cat "$OUT/$name.exit")" != 0 ]; then
		outcome=1
		echo "--- last lines of $name.log ---"
		tail -n 8 "$OUT/$name.log" | cut -c 1-400
	fi
done
[ -f "$OUT/first.exit" ] && [ -f "$OUT/full.exit" ] || outcome=1
echo "--- counts.json ---"
cat "$OUT/counts.json" 2> /dev/null || { echo "no counts"; cat "$OUT/counts.log"; }

# The payload: every stream when they fit, else first-error and full alone, as many as fit.
payload=()
total=0
for name in census first first-times full full-times; do
	[ -s "$OUT/$name.jsonl" ] && gzip -9 -n -c "$OUT/$name.jsonl" > "$OUT/$name.jsonl.gz"
done
encoded() { echo $((($(stat -c %s "$1") + 2) / 3 * 4)); }
for name in census first first-times full full-times; do
	[ -f "$OUT/$name.jsonl.gz" ] && total=$((total + $(encoded "$OUT/$name.jsonl.gz")))
done
if [ "$total" -le "$payloadLimit" ]; then
	for name in census first first-times full full-times; do
		[ -f "$OUT/$name.jsonl.gz" ] && payload+=("$name")
	done
else
	echo "all streams together would be $total base64 bytes, over $payloadLimit: printing first-error and full only"
	total=0
	for name in first full; do
		[ -f "$OUT/$name.jsonl.gz" ] || continue
		size=$(encoded "$OUT/$name.jsonl.gz")
		if [ $((total + size)) -le "$payloadLimit" ]; then
			payload+=("$name")
			total=$((total + size))
		else
			echo "omitted $name.jsonl.gz: $size base64 bytes would pass the limit"
		fi
	done
fi
echo "--- payload: ${payload[*]:-none} ---"
# The guarded expansion keeps an empty payload from tripping set -u on bash before 4.4.
for name in ${payload[@]+"${payload[@]}"}; do
	echo "=== BEGIN base64 gzip $name.jsonl.gz commit=$commit ==="
	base64 -w 76 "$OUT/$name.jsonl.gz"
	echo "=== END base64 gzip $name.jsonl.gz ==="
done
echo "=== rule-g-census sys.sh done: outcome $([ "$outcome" = 0 ] && echo ok || echo failed) ==="
exit "$outcome"
