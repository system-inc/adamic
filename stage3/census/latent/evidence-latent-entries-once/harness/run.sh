#!/usr/bin/env bash
# run.sh <binary> <adapted tree> <output dir>: the meter's latent step (twice-daily.sh line 113), timed.
set -euo pipefail
binary=$1 adapted=$2 out=$3
mkdir -p "$out"
python3 - "$binary" "$adapted" "$out" <<'PY'
import os, resource, subprocess, sys, time, json
binary, adapted, out = sys.argv[1:]
env = dict(os.environ, LATENT_TIMES=out + '/latent-times.jsonl', LATENT_FULL=os.environ.get('LATENT_FULL', '1'), LATENT_ASSERT_NO_OUTPUT='1')
start = time.monotonic()
with open(out + '/latent.log', 'w') as log:
    code = subprocess.call([binary, adapted + '/src/compiler', out + '/latent.jsonl'], env=env, stdout=log, stderr=subprocess.STDOUT)
wall = time.monotonic() - start
usage = resource.getrusage(resource.RUSAGE_CHILDREN)
result = {'exit': code, 'wall_seconds': round(wall, 1), 'user_seconds': round(usage.ru_utime, 1), 'system_seconds': round(usage.ru_stime, 1), 'max_rss_kb': usage.ru_maxrss}
json.dump(result, open(out + '/wall.json', 'w')); print(json.dumps(result))
PY
sha256sum "$out/latent.jsonl" | tee "$out/latent.sha256"
