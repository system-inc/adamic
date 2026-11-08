set -u
. ~/adamic-tools/env.sh
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /tmp/rg-sub.log 2>&1; echo "submodule exit=$?"
echo "sha $(git rev-parse HEAD)"; nproc; uptime
if ! command -v npm > /dev/null 2>&1; then
  bootstrap=cloud/markdown-width/npm-bootstrap.json
  v=$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1]))["version"])' $bootstrap)
  H=$HOME/npm-$v
  if [ ! -x "$H/bin/npm" ]; then
    mkdir -p "$H.partial/bin"
    curl -fsSL "$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1]))["url"])' $bootstrap)" -o "$H.partial/npm.tgz"
    python3 -I - $bootstrap "$H.partial/npm.tgz" <<'PY'
import base64, hashlib, json, sys
a, e = json.load(open(sys.argv[1]))['integrity'].split('-', 1)
if a != 'sha512' or base64.b64encode(hashlib.sha512(open(sys.argv[2], 'rb').read()).digest()).decode() != e: sys.exit('npm bootstrap integrity mismatch')
PY
    tar --no-same-owner -xzf "$H.partial/npm.tgz" -C "$H.partial"
    printf '#!/bin/sh
exec node %q "$@"
' "$H/package/bin/npm-cli.js" > "$H.partial/bin/npm"; chmod 755 "$H.partial/bin/npm"
    mv "$H.partial" "$H"
  fi
  export PATH="$H/bin:$PATH"
fi
echo "npm $(npm --version)"
O=/tmp/rule-g-delta-$(date +%s); mkdir -p $O
python3 cloud/admission-corpus/manifest.py --sha HEAD > $O/manifest.json 2> $O/manifest.log; echo "manifest exit=$?"
timeout 1560 go run ./cmd/adamic-admission-delta --base e3aa0bb7642163ebd004a71ba40a4657d27eab71 --head HEAD --manifest $O/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json > $O/delta.json 2> $O/delta.log; echo "delta exit=$?"
tail -30 $O/delta.log
python3 - "$O/delta.json" <<'PY'
import json,sys
try: d=json.load(open(sys.argv[1]))
except Exception as e: print("delta.json unreadable:", e); sys.exit()
def walk(x,path=""):
    if isinstance(x,dict):
        for k,v in x.items(): walk(v,path+"/"+k)
    elif isinstance(x,list) and path.split("/")[-1] in ("new_refusals","newly_refused","newly_admitted","admitted","refused","new_admissions"):
        print(path, len(x))
        for e in x[:60]: print("  ", json.dumps(e)[:400])
print("keys:", list(d)[:40] if isinstance(d,dict) else type(d))
walk(d)
PY
