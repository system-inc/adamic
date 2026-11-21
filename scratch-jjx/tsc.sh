set -u
# #jjxtwzs: what the arrays-admitted set holds in tsc, and which tsc stops move, head against base (the base by
# copying scratch-jjx/base's two lower files in and arrays_admitted.go out, then putting head back).
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=60m"
sandboxed_npm stage3/api stage3/typescript > "$TMPDIR/npm.log" 2>&1; echo "tsc: npm exit=$?"
sandboxed_typescript_source > "$TMPDIR/ts.log" 2>&1; echo "tsc: typescript source exit=$?"
export SCAN_ALL_SCRATCH="$TMPDIR/scanall"; mkdir -p "$SCAN_ALL_SCRATCH"
bash stage3/apply.sh --build-product "$SCAN_ALL_SCRATCH/adapted" > "$TMPDIR/apply.log" 2>&1; echo "tsc: apply exit=$?"
roots="src/compiler/sys.ts src/compiler/checker.ts src/compiler/program.ts src/compiler/emitter.ts src/compiler/watch.ts src/compiler/builder.ts src/compiler/factory/nodeFactory.ts src/compiler/executeCommandLine.ts"
for variant in head base; do
	if [ "$variant" = base ]; then
		for f in expression.go lower.go; do cp "internal/lower/$f" "$TMPDIR/$f.head"; cp "scratch-jjx/base/internal/lower/$f" "internal/lower/$f"; done
		mv internal/lower/arrays_admitted.go "$TMPDIR/arrays_admitted.go.head"
	fi
	go test -count=1 -v ./stage3/census/scan-all -run '^TestProduct_ScanAllInstrumentedProbe$' > "$TMPDIR/probe-$variant.log" 2>&1; echo "tsc: $variant probe exit=$?"
	probe=$(grep -o '[^ ]*/probe$' "$TMPDIR/probe-$variant.log" | tail -1); cp "$probe" "$TMPDIR/probe-$variant" || { echo "tsc: no probe for $variant"; tail -20 "$TMPDIR/probe-$variant.log"; }
	for root in $roots; do
		set_file="$TMPDIR/set-$variant-$(basename "$root").tsv"; : > "$set_file"
		( cd "$SCAN_ALL_SCRATCH" && ADAMIC_SCRATCH_ARRAYS_SET="$set_file" timeout 1500 "$TMPDIR/probe-$variant" "$SCAN_ALL_SCRATCH/adapted/$root" > "$TMPDIR/raw-$variant-$(basename "$root").json" 2>"$TMPDIR/err-$variant-$(basename "$root").log" ); echo "tsc: $variant $root exit=$?"
	done
done
for f in expression.go lower.go; do cp "$TMPDIR/$f.head" "internal/lower/$f"; done
mv "$TMPDIR/arrays_admitted.go.head" internal/lower/arrays_admitted.go
python3 -I - "$TMPDIR" "$SCAN_ALL_SCRATCH/adapted/" <<'PY'
import glob, json, os, sys, collections
directory, prefix = sys.argv[1], sys.argv[2]
types = collections.Counter()
for path in glob.glob(os.path.join(directory, 'set-head-*.tsv')):
    for line in open(path):
        parts = line.rstrip('\n').split('\t')
        if len(parts) >= 3 and parts[1] not in ('', '0'):
            for name in parts[2].split(' | '):
                types[name] += 1
print('tsc: arrays-admitted types in tsc roots:', len(types))
for name, count in types.most_common(40):
    print('tsc:   %d lowerings\t%s' % (count, name[:160]))
def triples(variant):
    seen = set()
    for path in glob.glob(os.path.join(directory, f'raw-{variant}-*.json')):
        try: report = json.load(open(path))
        except Exception as error: print('tsc: unreadable', path, error); continue
        for stop in report.get('Stops') or []:
            seen.add((stop['Rule'], stop['Site'].replace(prefix, ''), stop['Words'][:200]))
    return seen
head, base = triples('head'), triples('base')
print('tsc: triples head', len(head), 'base', len(base))
for label, rows in (('ONLY-HEAD', head - base), ('ONLY-BASE', base - head)):
    rows = sorted(rows, key=lambda row: row[1])
    print('tsc: %s %d' % (label, len(rows)))
    for rule, site, words in rows[:60]:
        print('tsc: ' + label + '\t' + site + '\t' + rule + '\t' + words[:200])
PY
echo TSC DONE
