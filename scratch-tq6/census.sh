set -u
# #tq6yzqy scout: family 6's optional-boolean-in-an-intersection sites, read from the latent census's raw streams at
# this commit, as the meter's first-error passes write them (tsc's entry, then src/compiler). Scratch only, never lands.
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
echo "tq6: box $(hostname) at $(git rev-parse HEAD)"
audit="$TMPDIR/tq6"; mkdir -p "$audit"
sandboxed_npm stage3/api > "$audit/npm.log" 2>&1; echo "tq6: npm exit=$?"
sandboxed_typescript_source > "$audit/ts.log" 2>&1; echo "tq6: typescript source exit=$? at ${ADAMIC_TYPESCRIPT_SOURCE:-unset}"
bash stage3/apply.sh "$audit/adapted" > "$audit/apply.log" 2>&1; echo "tq6: apply exit=$?"; tail -2 "$audit/apply.log"
python3 stage3/census/latent/make_overlay.py "$PWD" "$audit/latent-overlay" > "$audit/overlay.log" 2>&1; echo "tq6: latent overlay exit=$?"
gofmt -w "$audit/latent-overlay/"*.go
python3 stage3/meter/entry_overlay.py "$audit/latent-overlay" "$audit/entry-overlay" >> "$audit/overlay.log" 2>&1; echo "tq6: entry overlay exit=$?"
gofmt -w "$audit/entry-overlay/"*.go
go build -buildvcs=false -overlay="$audit/entry-overlay/overlay.json" -o "$audit/entry-census" ./stage3/census/latent/tool > "$audit/build.log" 2>&1; echo "tq6: entry build exit=$?"
go build -buildvcs=false -overlay="$audit/latent-overlay/overlay.json" -o "$audit/latent-census" ./stage3/census/latent/tool >> "$audit/build.log" 2>&1; echo "tq6: latent build exit=$?"
tail -5 "$audit/overlay.log" "$audit/build.log"
export LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 CENSUS_MODULE_FORMAT=CommonJSBundle
LATENT_TIMES="$audit/entry-times.jsonl" timeout 1500 "$audit/entry-census" "$audit/adapted/src/tsc/tsc.ts" "$audit/entry.jsonl" > "$audit/entry.out" 2>&1; echo "tq6: entry census exit=$?"
[ "${TQ6_COMPILER:-0}" = 1 ] && LATENT_TIMES="$audit/compiler-times.jsonl" timeout 1500 "$audit/latent-census" "$audit/adapted/src/compiler" "$audit/compiler.jsonl" > "$audit/compiler.out" 2>&1; echo "tq6: compiler census exit=$?"
tail -3 "$audit/entry.out" "$audit/compiler.out"
python3 -I - "$audit" <<'PY'
import json, os, sys
audit = sys.argv[1]
for name in ("entry.jsonl", "compiler.jsonl"):
    path = os.path.join(audit, name)
    if not os.path.exists(path):
        print("tq6:", name, "missing"); continue
    lines = open(path, encoding="utf-8").read().splitlines()
    events = family = 0
    keys = list()
    seen = set()
    for line in lines:
        try:
            event = json.loads(line)
        except Exception:
            continue
        events += 1
        if not keys and isinstance(event, dict):
            keys = sorted(event.keys())
        text = json.dumps(event)
        if "a value of type" in text:
            family += 1
        sites = os.environ.get("TQ6_SITES", "")
        if sites and isinstance(event, dict) and any(str(event.get("file", "")).endswith(site) for site in sites.split(",")):
            found = []
            def walk(node):
                if isinstance(node, dict):
                    for value in node.values(): walk(value)
                elif isinstance(node, list):
                    for value in node: walk(value)
                elif isinstance(node, str) and ("stage 0" in node or "Adamic 0.1" in node or "compiler bug" in node):
                    found.append(node)
            walk(event)
            print("SITE", str(event.get("file", "")).split("/src/")[-1], "keys", sorted(event.keys()), "stops", json.dumps(found[:3])[:900])
        if os.environ.get("TQ6_FILTER", "optional boolean") not in text:
            continue
        compact = {k: v for k, v in event.items() if not isinstance(v, (dict, list))} if isinstance(event, dict) else event
        identity = json.dumps(compact, sort_keys=True)
        if identity in seen:
            continue
        seen.add(identity)
        print("ROW", name, identity[:900])
    print("tq6:", name, "events", events, "family-6 events", family, "filtered rows", len(seen), "keys", keys)
PY
echo DONE
