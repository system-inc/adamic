#!/usr/bin/env bash
# Correctness for #e24qsqx: how many computed reads in tsc key a fixed object by a type parameter, the shape at
# moduleNameResolver.ts:368. One full-mode latent census of moduleNameResolver.ts (every finding, not only first
# stops), its ElementAccessExpression findings listed, and a typed scan of the compiler project with stock tsc
# that lists every element access whose key's type is a type parameter.
set -u -o pipefail
probe=$(cd -- "$(dirname -- "$0")" && pwd -P)
repository=$(git -C "$probe" rev-parse --show-toplevel)
cd "$repository"
began=$SECONDS
phase() { echo "PHASE $1 $((SECONDS - began))"; }
stop() { echo "FAIL $1"; echo "VERDICT FAIL at $((SECONDS - began)) s"; exit 1; }

source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh \
	&& sandboxed_npm stage3/api stage3/typescript && sandboxed_typescript_source || stop "sandboxed setup"
work="${LOOM_JOB_SCRATCH:-$TMPDIR}/e24"
mkdir -p "$work/home" "$work/streams" "$work/tools"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 NODE_DISABLE_COMPILE_CACHE=1 PYTHONDONTWRITEBYTECODE=1
export TSC_ADAPT_TYPESCRIPT="$repository/stage3/api/node_modules/typescript"
pin=050880ce59e30b356b686bd3144efe24f875ebc8
echo "census at adamic $(git rev-parse HEAD^ 2>/dev/null) (the probe commit's parent)"
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)" = "$pin" ] || stop "TypeScript source is not at the pin"
git clone -q --depth 1 --no-checkout "file://$ADAMIC_TYPESCRIPT_SOURCE" "$work/tree" && git -C "$work/tree" remote remove origin \
	&& git -C "$work/tree" sparse-checkout set --no-cone '/*' '!/tests/baselines/reference/' '/tests/baselines/reference/api/' \
	&& git -C "$work/tree" -c advice.detachedHead=false checkout -q --detach "$pin" || stop "checkout of the pin"
(HOME="$work/home" XDG_CONFIG_HOME="$work/home" python3 -E -S -B stage3/apply.py --adapt-checkout "$work/tree") > "$work/apply.log" 2>&1 \
	|| { tail -n 60 "$work/apply.log"; stop "apply"; }
phase adapted-tree

# The typed scan first: it needs only the adapted tree and stock tsc.
NODE_PATH="$repository/stage3/api/node_modules" node "$probe/scan.cjs" "$work/tree" || stop "typed scan"
phase scan

latent=stage3/census/latent
tools="$work/tools"
python3 "$latent/make_overlay.py" "$repository" "$tools/latent-overlay" > "$tools/latent-overlay.log" 2>&1 \
	&& gofmt -w "$tools/latent-overlay/"*.go || { tail -n 40 "$tools/latent-overlay.log"; stop "overlay"; }
go build -buildvcs=false -overlay="$tools/latent-overlay/overlay.json" -o "$tools/directory-census" ./$latent/tool > "$tools/build.log" 2>&1 \
	|| { tail -n 60 "$tools/build.log"; stop "directory tool"; }
phase tool
LATENT_ONLY_FILE="$work/tree/src/compiler/moduleNameResolver.ts" LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
	CENSUS_MODULE_FORMAT=CommonJSBundle "$tools/directory-census" "$work/tree/src/compiler" "$work/streams/full.jsonl" \
	> "$work/streams/full.log" 2>&1 || { tail -n 30 "$work/streams/full.log"; stop "full census"; }
phase full-census
python3 -I - "$work/streams/full.jsonl" "$work/tree/" << 'READ'
import collections, json, sys
path, root = sys.argv[1], sys.argv[2]
units, findings, reasons = 0, 0, collections.Counter()
rows = []
for line in open(path, encoding='utf-8'):
    try:
        row = json.loads(line)
    except json.JSONDecodeError:
        break
    units += 1
    for f in row.get('findings') or []:
        findings += 1
        reason = (f.get('reason') or '').replace(root, '')
        reasons[reason] += 1
        if 'ElementAccessExpression' in reason or 'computed field' in reason:
            rows.append(((f.get('where') or '').replace(root, ''), f.get('phase'), reason, row.get('unit') or row.get('name') or ''))
print(f'FULL moduleNameResolver.ts: {units} units, {findings} findings, {len(reasons)} reasons')
seen = set()
for where, phase, reason, unit in rows:
    if (where, reason) in seen:
        continue
    seen.add((where, reason))
    print('FULL-EA', where, phase, reason, sep='\t')
print(f'FULL-EA distinct sites: {len(seen)}')
for reason, count in reasons.most_common(12):
    print('FULL-REASON', count, reason[:160], sep='\t')
READ
phase done
echo "VERDICT PASS"
