#!/usr/bin/env bash
# Builds the probe census at b5733921 exactly as the earlier probe was built on main, niced.
set -u
W=/tmp/m/s1c; S=/tmp/m/s1c-scratch; L=$S/logs; mkdir -p $L $S/main-source
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/m/upstream
cd $W
python3 stage3/census/latent/make_overlay.py $W $S/latent-overlay > $L/latent-overlay.log 2>&1; echo overlay $? > $S/status
gofmt -w $S/latent-overlay/*.go
python3 stage3/meter/entry_overlay.py $S/latent-overlay $S/entry-overlay > $L/entry-overlay.log 2>&1; echo entry-overlay $? >> $S/status
gofmt -w $S/entry-overlay/*.go
# probe patches: same edits as the main-tree probe
cp /home/user/adamic/evidence/compiler/meter-memory/overlay/latent_heapprobe.go.txt $S/latent-overlay/internal_lower_latent_heapprobe.go
python3 - <<'PY'
import json,re
d='/tmp/m/s1c-scratch/latent-overlay/'
p=d+'internal_lower_latent_timing.go'; s=open(p).read()
a="""	if timer == nil {
		return func() {}
	}"""; assert a in s
s=s.replace(a,"""	if timer == nil {
		return func() { heapUnit(file, where) }
	}""")
a="""		if err := timer.encoder.Encode(record); err != nil {
			panic(err)
		}
	}"""; assert a in s
s=s.replace(a,"""		if err := timer.encoder.Encode(record); err != nil {
			panic(err)
		}
		heapUnit(file, where)
	}""")
open(p,'w').write(s)
p=d+'internal_lower_latent_hook.go'; s=open(p).read()
a="func Latent(ctx context.Context, program *load.Program, emit func(LatentFile)) {\n"; assert s.count(a)==1
open(p,'w').write(s.replace(a,a+"\tdefer heapFinal()\n"))
p=d+'internal_lower_functions.go'; s=open(p).read()
a="func (l *lowering) boundarySchema(node *ast.Node, proven *checker.Type) *ir.JSONDecodeSchema {\n"; assert s.count(a)==1
open(p,'w').write(s.replace(a,a+"\tmemoCountBoundary(node, proven)\n"))
src=open('/tmp/m/s1c/internal/lower/library_json_decode.go').read()
a="n := ir.JSONDecodeNode{Expected: l.checker.TypeToString(t)}"; assert src.count(a)==1
open(d+'internal_lower_library_json_decode.go','w').write(src.replace(a,"n := ir.JSONDecodeNode{Expected: l.typeString(t)}"))
for dd in ('latent-overlay','entry-overlay'):
    p='/tmp/m/s1c-scratch/%s/overlay.json'%dd; j=json.load(open(p))
    j['Replace']['/tmp/m/s1c/internal/lower/latent_heapprobe.go']=d+'internal_lower_latent_heapprobe.go'
    j['Replace']['/tmp/m/s1c/internal/lower/library_json_decode.go']=d+'internal_lower_library_json_decode.go'
    json.dump(j,open(p,'w'),indent=2)
PY
echo patched $? >> $S/status
gofmt -l $S/latent-overlay >> $L/gofmt.log
go build -buildvcs=false -overlay=$S/entry-overlay/overlay.json -o $S/entry-census-probe ./stage3/census/latent/tool > $L/entry-probe-build.log 2>&1; echo build $? >> $S/status
git -C $W archive HEAD stage3 | tar -x -C $S/main-source
(cd $S/main-source/stage3/api && npm ci --ignore-scripts --no-audit --no-fund) > $L/packages.log 2>&1; echo npm $? >> $S/status
bash $S/main-source/stage3/apply.sh $S/adapted > $L/apply.log 2>&1; echo apply $? >> $S/status
echo done >> $S/status
