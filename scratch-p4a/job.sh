#!/usr/bin/env bash
# scratch for #6c4swvt piece 4's adaptation, never cut. Whole tree against the tree without
# 99-library-readonly-option-workers: apply, idempotency, sources, tsc, JavaScript, census.
set -u
root=$PWD
export PYTHONDONTWRITEBYTECODE=1
head=99-library-readonly-option-workers
modules="$root/stage3/api/node_modules"
out=$(mktemp -d "$TMPDIR/p4a-out.XXXXXX") || exit 64
work=$(mktemp -d "$TMPDIR/p4a-work.XXXXXX") || exit 64
say() { echo "$*" | tee -a "$out/verdicts.log"; }
copy() { # copy <name> <dropped head>
	local dest="$work/$1/stage3"
	mkdir -p "$dest/adapt" || exit 64
	for input in apply.sh apply.py caches.py source.json api; do cp -a "$root/stage3/$input" "$dest/" || exit 64; done
	for entry in "$root/stage3/adapt"/*; do
		[ "$(basename "$entry")" = "${2:-}" ] || cp -a "$entry" "$dest/adapt/" || exit 64
	done
}
copy whole
copy without "$head"
(python3 -B stage3/census/latent/make_overlay.py "$root" "$work/overlay" > "$out/overlay.log" 2>&1 &&
	gofmt -w "$work/overlay"/*.go &&
	go build -buildvcs=false -overlay="$work/overlay/overlay.json" -o "$work/census" ./stage3/census/latent/tool > "$out/census-build.log" 2>&1; echo $? > "$out/census-build.exit") &
for variant in whole without; do
	(STAGE3_PRODUCT_STORE="$work/products-$variant" STAGE3_CACHE="$work/cache-$variant" bash "$work/$variant/stage3/apply.sh" "$work/$variant/tree" > "$out/apply-$variant.log" 2>&1; echo $? > "$out/apply-$variant.exit") &
done
wait
for variant in whole without; do
	code=$(cat "$out/apply-$variant.exit")
	if [ "$code" != 0 ]; then say "p4a: verdict apply FAIL: $variant exited $code"; grep -E 'Error|expected|want' "$out/apply-$variant.log" | grep -v '^\s*at ' | tail -n 8 | cut -c1-400; exit 1; fi
done
say "p4a: verdict apply PASS: $(grep -F "$head" "$work/whole/tree/patch-set.md")"
grep -A8 '"edits"' "$out/apply-whole.log" | head -n 20
file="$work/whole/tree/src/compiler/commandLineParser.ts"
before=$(sha256sum "$file" | cut -c1-64)
NODE_PATH="$modules" node "$root/stage3/adapt/$head/adapt.cjs" "$work/whole/tree" > "$out/again.log" 2>&1 || { cat "$out/again.log"; say "p4a: verdict idempotent FAIL: the second run exited non-zero"; exit 1; }
after=$(sha256sum "$file" | cut -c1-64)
[ "$before" = "$after" ] && say "p4a: verdict idempotent PASS: a second run leaves commandLineParser.ts's bytes ($before)" || say "p4a: verdict idempotent FAIL: $before then $after"
python3 -I - "$work/without/tree" "$work/whole/tree" << 'PYTHON' | tee -a "$out/verdicts.log"
import hashlib, sys
from pathlib import Path
before, after = Path(sys.argv[1]), Path(sys.argv[2])
def manifest(tree):
    root = tree / 'src'
    return {p.relative_to(root).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
old, new = manifest(before), manifest(after)
differing = sorted(n for n in old.keys() | new.keys() if old.get(n) != new.get(n))
print(f"p4a: verdict sources {'PASS' if differing == ['compiler/commandLineParser.ts'] else 'FAIL'}: {differing}")
for name in differing:
    a = (before / 'src' / name).read_text().split('\n'); b = (after / 'src' / name).read_text().split('\n')
    for i, (x, y) in enumerate(zip(a, b)):
        if x != y: print(f"p4a:   {name}:{i + 1}: {x.strip()[:150]}  ->  {y.strip()[:170]}")
PYTHON
NODE_PATH="$modules" node - "$work/without/tree/src/compiler/commandLineParser.ts" "$work/whole/tree/src/compiler/commandLineParser.ts" << 'NODE' | tee -a "$out/verdicts.log"
const ts = require('typescript'); const fs = require('fs');
const emit = (file, removeComments) => ts.transpileModule(fs.readFileSync(file, 'utf8'), { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS, removeComments } }).outputText;
const same = [false, true].every(r => emit(process.argv[2], r) === emit(process.argv[3], r));
console.log(`p4a: verdict javascript ${same ? 'PASS' : 'FAIL'}: commandLineParser.ts's JavaScript ${same ? 'is byte-identical' : 'differs'}, comments kept and removed`);
NODE
for variant in whole without; do
	for project in compiler services server; do
		dir="$work/$variant/tree/src/$project"
		[ -f "$dir/tsconfig.json" ] || continue
		printf '{ "extends": "./tsconfig.json", "compilerOptions": { "noEmit": true, "composite": false, "declaration": true, "declarationMap": false, "emitDeclarationOnly": false, "incremental": false, "tsBuildInfoFile": null, "types": [] } }\n' > "$dir/tsconfig.p4a.json"
		(cd "$dir" && node "$modules/typescript/lib/tsc.js" -p tsconfig.p4a.json --pretty false) 2>&1 | grep 'error TS' | sed "s#^#$project/#" | sort > "$work/tsc-$variant-$project.txt"
		rm -f "$dir/tsconfig.p4a.json"
	done
done
newErrors=0
for project in compiler services server; do
	[ -f "$work/tsc-without-$project.txt" ] || continue
	count=$(comm -13 "$work/tsc-without-$project.txt" "$work/tsc-whole-$project.txt" | wc -l)
	newErrors=$((newErrors + count))
	echo "p4a: tsc $project: $(wc -l < "$work/tsc-without-$project.txt") without it, $count new with it"
	comm -13 "$work/tsc-without-$project.txt" "$work/tsc-whole-$project.txt" | head -n 10 | cut -c1-300 | sed 's/^/p4a:   new: /'
done
[ "$newErrors" = 0 ] && say "p4a: verdict tsc PASS: 0 new errors over compiler, services and server" || say "p4a: verdict tsc FAIL: $newErrors new errors"
[ "$(cat "$out/census-build.exit" 2> /dev/null)" = 0 ] || { tail -n 20 "$out/overlay.log" "$out/census-build.log"; say "p4a: verdict census FAIL: didn't build"; exit 1; }
for variant in whole without; do
	mkdir -p "$work/c-$variant"
	(LATENT_TIMES="$work/c-$variant/t.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 "$work/census" "$work/$variant/tree/src/compiler" "$work/c-$variant/latent.jsonl" > "$out/census-$variant.log" 2>&1; echo $? > "$out/census-$variant.exit") &
done
wait
echo "p4a: census exits: whole $(cat "$out/census-whole.exit"), without $(cat "$out/census-without.exit")"
python3 -I - "$root/stage3/meter" "$work" << 'PYTHON' | tee -a "$out/verdicts.log"
import json, re, sys
from pathlib import Path
sys.path.insert(0, sys.argv[1])
import report
work = Path(sys.argv[2])
def findings(variant):
    found = set()
    for line in (work / f'c-{variant}' / 'latent.jsonl').read_text().splitlines():
        for f in json.loads(line).get('findings') or []:
            if f.get('phase') == 'tail' or f['kind'] not in ('NotYet', 'Refused'):
                continue
            found.add((f['kind'], report.site_where(f['where']), re.sub(r'/\S*/tree/src/', 'src/', f['reason'])))
    return found
def total(variant):
    s = report.latent_summary(work / variant / 'tree', work / f'c-{variant}')
    return s['totals']['NotYet'] + s['totals']['Refused']
whole, without = findings('whole'), findings('without')
print(f"p4a: census latent total without it {total('without')}, with it {total('whole')}")
gone, came = sorted(without - whole), sorted(whole - without)
print(f"p4a: census {len(gone)} gone, {len(came)} new")
for f in gone: print(f"p4a:   - {f[0]} {f[1]} | {f[2][:400]}")
for f in came: print(f"p4a:   + {f[0]} {f[1]} | {f[2][:1200]}")
for site in ('commandLineParser.ts:2792', 'commandLineParser.ts:3275', 'commandLineParser.ts:2712', 'commandLineParser.ts:3076', 'commandLineParser.ts:3080'):
    print(f"p4a: site {site}: without {sorted(f[0] + ' ' + f[1] + ' ' + f[2][:200] for f in without if site in f[1])}")
    print(f"p4a: site {site}: with    {sorted(f[0] + ' ' + f[1] + ' ' + f[2][:200] for f in whole if site in f[1])}")
PYTHON
echo "p4a: summary"; grep '^p4a: verdict' "$out/verdicts.log"
grep -q '^p4a: verdict .* FAIL' "$out/verdicts.log" && exit 1
exit 0
