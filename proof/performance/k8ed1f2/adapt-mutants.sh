# #k8ed1f2: adaptation 99-readonly-family3-views' restoring mutants. Each takes one cause back out of adapt.cjs, applies,
# and scans its file's closure; it must bring back exactly that cause's own sites. Run after environment.sh,
# sandboxed-job.sh, sandboxed_typescript_source and sandboxed_npm stage3/api. Never merged.
set -u
bad=0
D=stage3/adapt/99-readonly-family3-views
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; exit 1; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
cp "$D/adapt.cjs" "$TMPDIR/adapt.original"
scan() {
	local name=$1 root=$2
	export SCAN_ALL_SCRATCH="$TMPDIR/scan-$name"; mkdir -p "$SCAN_ALL_SCRATCH"
	bash stage3/apply.sh --build-product "$SCAN_ALL_SCRATCH/adapted" > "$TMPDIR/apply-$name.log" 2>&1 || { echo "FAIL apply $name"; tail -3 "$TMPDIR/apply-$name.log"; bad=1; return 1; }
	"$BIN" "$SCAN_ALL_SCRATCH/adapted/src/compiler/$root" > "$TMPDIR/$name.json" 2>/dev/null
	python3 - "$TMPDIR/$name.json" > "$TMPDIR/$name.sites" <<'PYTHON'
import json, re, sys
for s in json.load(open(sys.argv[1])).get('Stops') or []:
    if 'can write' in s['Words'] or 'becomes writable' in s['Words']:
        print(re.sub(r'^.*/adapted/', '', s['Site']))
PYTHON
}
mutant() {
	local name=$1 root=$2 old=$3 new=$4 want=$5
	cp "$TMPDIR/adapt.original" "$D/adapt.cjs"
	python3 - "$D/adapt.cjs" "$old" "$new" <<'PYTHON'
import sys
p, old, new = sys.argv[1:4]; s = open(p).read()
assert s.count(old) == 1, 'anchor: ' + old
if new == 'DELETE':
    i = s.index(old); j = s.index('},\n', i) + 3
    s = s[:i] + s[j:]
else:
    s = s.replace(old, new)
open(p, 'w').write(s)
PYTHON
	scan "$name" "$root" || return
	cp "$TMPDIR/adapt.original" "$D/adapt.cjs"
	local back=$(sort -u "$TMPDIR/$name.sites" | comm -13 <(sort -u "$TMPDIR/head-$root.sites") - | tr '\n' ' ')
	local missing="" extra=""
	for site in $want; do case " $back" in *" $site "*) ;; *) missing="$missing $site" ;; esac; done
	for site in $back; do case " $want " in *" $site "*) ;; *) extra="$extra $site" ;; esac; done
	[ -z "$missing" ] && echo "PASS mutant $name brings back all $(echo $want | wc -w) of its sites; others back: [${extra# }]" || { echo "FAIL mutant $name misses [${missing# }]; back [$back]"; bad=1; }
}
for root in tsbuildPublic.ts builder.ts; do scan "head-$root" "$root"; done
mutant road2 tsbuildPublic.ts '...statePrunes.map(' '...statePrunes.slice(0, 0).map(' 'src/compiler/tsbuildPublic.ts:676:32 src/compiler/tsbuildPublic.ts:677:32 src/compiler/tsbuildPublic.ts:678:32 src/compiler/tsbuildPublic.ts:679:32 src/compiler/tsbuildPublic.ts:680:32 src/compiler/tsbuildPublic.ts:681:32 src/compiler/tsbuildPublic.ts:682:32 src/compiler/tsbuildPublic.ts:683:32 src/compiler/tsbuildPublic.ts:684:32'
mutant cause15 builder.ts "    { cause: 15, file: builder," DELETE 'src/compiler/builder.ts:1337:16 src/compiler/builder.ts:1344:13 src/compiler/builder.ts:1346:17'
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
