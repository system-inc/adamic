#!/bin/bash
set -euo pipefail
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/env-presence-1965/TypeScript
export ADAMIC_BUILD_CACHE_DIR=/tmp/env-presence-1965/cache STAGE3_CACHE=/tmp/env-presence-1965/cache/product
export TMPDIR=/tmp/env-presence-1965/tmp GOTMPDIR=/tmp/env-presence-1965/tmp GOCACHE=/tmp/env-presence-1965/gocache
export npm_config_cache=/tmp/env-presence-1965/npmcache
cp -a --reflink=auto /tmp/env-presence-1965/baseline /tmp/env-presence-1965/counting
python3 stage3/census/env-presence-1965/instrument.py /tmp/env-presence-1965/counting > /tmp/env-presence-1965/instrument.json
export ENV_PRESENCE_COUNTS=/tmp/env-presence-1965/counts.jsonl
export NODE_OPTIONS="--require=/tmp/env-presence-1965/work/stage3/census/env-presence-1965/counter.cjs"
set +e
stage3/oracle/run.sh /tmp/env-presence-1965/counting /tmp/env-presence-1965/count-oracle > /tmp/env-presence-1965/count-oracle.log 2>&1
code=$?
set -e
printf '%s\n' "$code" > /tmp/env-presence-1965/count-oracle.exit
python3 - <<'PY'
import json,pathlib,collections
p=pathlib.Path('/tmp/env-presence-1965')
a=[json.loads(s) for s in p.joinpath('counts.jsonl').read_text().splitlines()]
c=collections.Counter('falsy' if not r['truthy'] else 'truthy_function' if r['function'] else 'truthy_without_function' for r in a)
p.joinpath('counts-summary.json').write_text(json.dumps({'observations':len(a),**{k:c[k] for k in ['truthy_function','truthy_without_function','falsy']},'oracle':json.loads(p.joinpath('count-oracle/report.json').read_text())},indent=2)+'\n')
PY
