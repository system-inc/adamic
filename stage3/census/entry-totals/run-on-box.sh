#!/usr/bin/env bash
# Scratch, for a box proof only, never for main (#baobab): the daily meter's tsc entry pass alone, in
# first-error mode, over this checkout, printing latent totals per kind and the top reasons. It is
# stage3/meter/twice-daily.sh's entry recipe (build_census's entry half, apply, LATENT_FULL=0 entry
# census) without the per-file compiler passes, so it fits a two-hour gap.
set -euo pipefail
repository=$(cd "$(dirname "$0")/../../.." && pwd)
cd "$repository"
for environment in "${ADAMIC_TOOLS:+$ADAMIC_TOOLS/env.sh}" /home/ahra/adamic-tools/env.sh \
    "$HOME/adamic-tools/env.sh" /workspace/adamic-tools/env.sh /opt/adamic-tools/env.sh; do
    if [[ -r $environment ]]; then echo "entry: sourcing $environment" >&2; source "$environment"; break; fi
done
for tool in go gofmt node python3 git; do command -v "$tool" > /dev/null || { echo "missing tool: $tool" >&2; exit 1; }; done
export GOWORK="$repository/go.work"
[[ -n ${ADAMIC_TYPESCRIPT_SOURCE:-} && -f $ADAMIC_TYPESCRIPT_SOURCE/src/tsc/tsc.ts ]] || { echo 'missing ADAMIC_TYPESCRIPT_SOURCE' >&2; exit 1; }
[[ $(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD) == 050880ce59e30b356b686bd3144efe24f875ebc8 ]] || { echo 'wrong TypeScript pin' >&2; exit 1; }
if [[ ! -f cohere/TypeScript/tsc/go.mod ]]; then
    git -c submodule.cohere.url=https://github.com/system-inc/cohere.git submodule update --init --recursive --depth 1 --filter=blob:none >&2
fi
[[ -d stage3/api/node_modules ]] || (cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund) >&2
work=$(mktemp -d "${TMPDIR:-/tmp}/entry-totals.XXXXXX")
printf 'entry: adamic %s, work %s\n' "$(git rev-parse HEAD)" "$work" >&2
step() { local label=$1; shift; local start=$SECONDS code=0; "$@" > "$work/$label.log" 2>&1 || code=$?; printf '{"step":"%s","exit":%d,"seconds":%d}\n' "$label" "$code" "$((SECONDS-start))"; if (( code )); then tail -n 60 "$work/$label.log" >&2; exit "$code"; fi; }
step apply bash stage3/apply.sh "$work/adapted"
step latent-overlay python3 stage3/census/latent/make_overlay.py "$repository" "$work/latent-overlay"
gofmt -w "$work/latent-overlay/"*.go
step entry-overlay python3 stage3/meter/entry_overlay.py "$work/latent-overlay" "$work/entry-overlay"
gofmt -w "$work/entry-overlay/"*.go
step entry-build go build -buildvcs=false -overlay="$work/entry-overlay/overlay.json" -o "$work/entry-census" ./stage3/census/latent/tool
entry=$work/adapted/src/tsc/tsc.ts
start=$SECONDS
set +e
LATENT_TIMES="$work/latent-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 "$work/entry-census" "$entry" "$work/latent.jsonl" > "$work/latent.log" 2>&1
code=$?
set -e
printf '{"step":"entry-pass","exit":%d,"seconds":%d}\n' "$code" "$((SECONDS-start))"
(( code == 0 )) || { tail -n 60 "$work/latent.log" >&2; exit "$code"; }
python3 - "$work/latent.jsonl" "$work/latent-times.jsonl" <<'PY'
# Counted as stage3/meter/report.py's lowering_summary does: distinct sites by (kind, where, reason, text)
# over every file record and the relation trailer, plus raw finding events.
import collections, json, sys
records = [json.loads(line) for line in open(sys.argv[1]) if line.strip()]
times = sum(1 for line in open(sys.argv[2]) if line.strip())
header = records[0] if records else {}
trailer = records[-1] if records and records[-1].get('status') == 'relation_report' else None
files = [r for r in records[1:] if r.get('status') != 'relation_report']
sites, events = set(), 0
for record in [*files, *([trailer] if trailer else [])]:
    for finding in record.get('findings') or []:
        sites.add(tuple(finding.get(key) for key in ('kind', 'where', 'reason', 'text')))
        events += 1
kinds = collections.Counter(site[0] for site in sites)
reasons = collections.Counter((site[0], site[2]) for site in sites if site[0] in ('NotYet', 'Refused'))
print(json.dumps({'header_status': header.get('status'), 'files': len(files), 'trailer': bool(trailer),
                  'relation_overflow': (trailer or {}).get('relation_overflow'), 'timed_units': times,
                  'finding_events': events, 'distinct_sites': len(sites), 'sites_by_kind': dict(kinds)}, sort_keys=True))
for (kind, reason), count in sorted(reasons.items(), key=lambda item: (-item[1], item[0]))[:30]:
    print(f'reason\t{count}\t{kind}\t{reason[:180]}')
for site in sorted(s for s in sites if s[0] == 'panic')[:20]:
    print('panic\t' + '\t'.join(str(x)[:180] for x in site[1:]))
PY
