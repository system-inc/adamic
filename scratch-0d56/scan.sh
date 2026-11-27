set -u
# #0d56xpd: scan-all at the head and at main dacb2da2 (the base), the lower files this change touched put back to the
# base's for the second, and the triples each side alone has. Verdict lines begin "gate:".
source /opt/adamic-block/environment.sh && echo "gate: PASS environment.sh sourced" || { echo "gate: FAIL environment.sh"; exit 1; }
source cloud/sandboxed-job.sh && echo "gate: PASS sandboxed-job.sh sourced" || { echo "gate: FAIL sandboxed-job.sh"; exit 1; }
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=60m"
work="${LOOM_JOB_SCRATCH:-${TMPDIR:-/tmp}}/scan0d56"
mkdir -p "$work" && echo "gate: PASS scratch $work" || { echo "gate: FAIL scratch $work"; exit 1; }
base=dacb2da27f24bc89bc1ea09cb6adca55cbe77e9f
echo "gate: box $(hostname) at $(git rev-parse --short HEAD), base $base"
sandboxed_npm stage3/api stage3/typescript > "$work/npm.log" 2>&1 && echo "gate: PASS npm" || { echo "gate: FAIL npm"; tail -20 "$work/npm.log"; exit 1; }
sandboxed_typescript_source > "$work/ts.log" 2>&1 && echo "gate: PASS typescript source" || { echo "gate: FAIL typescript source"; tail -20 "$work/ts.log"; exit 1; }
git cat-file -e "$base^{commit}" 2>/dev/null && echo "gate: PASS base present" || { echo "gate: FAIL base $base not in this clone"; exit 1; }
export SCAN_ALL_SCRATCH="$work/scanall"
mkdir -p "$SCAN_ALL_SCRATCH" && echo "gate: PASS scan scratch" || { echo "gate: FAIL scan scratch"; exit 1; }
bash stage3/apply.sh --build-product "$SCAN_ALL_SCRATCH/adapted" > "$work/apply.log" 2>&1 && echo "gate: PASS apply" || { echo "gate: FAIL apply"; tail -20 "$work/apply.log"; exit 1; }
touched="object.go placeholder.go switch_nullish_numeric.go"
for variant in head base; do
	if [ "$variant" = base ]; then
		for f in $touched; do cp internal/lower/$f "$work/$f.head"; git show "$base:internal/lower/$f" > internal/lower/$f; done
		echo "gate: base files laid: $(git diff --stat -- internal/lower | tail -1)"
	fi
	go test -count=1 -v ./stage3/census/scan-all -run '^TestProduct_ScanAllInstrumentedProbe$' > "$work/probe-$variant.log" 2>&1
	echo "gate: $variant probe exit=$?"
	probe=$(grep -o '[^ ]*/probe$' "$work/probe-$variant.log" | tail -1)
	[ -n "$probe" ] && cp "$probe" "$work/probe-$variant" && echo "gate: PASS $variant probe built" || { echo "gate: FAIL $variant probe"; tail -20 "$work/probe-$variant.log"; exit 1; }
	for root in src/compiler/sys.ts src/compiler/checker.ts src/compiler/program.ts src/compiler/emitter.ts src/compiler/watch.ts src/compiler/builder.ts src/compiler/factory/nodeFactory.ts src/compiler/executeCommandLine.ts; do
		( cd "$SCAN_ALL_SCRATCH" && timeout 1200 "$work/probe-$variant" "$SCAN_ALL_SCRATCH/adapted/$root" > "$work/raw-$variant-$(basename $root).json" 2>/dev/null )
		echo "gate: $variant $root exit=$?"
	done
done
for f in $touched; do cp "$work/$f.head" internal/lower/$f; done
echo "gate: head files back: $(git diff --stat -- internal/lower | tail -1)"
python3 -I - "$work" "$SCAN_ALL_SCRATCH/adapted/" <<'PY'
import glob, json, os, sys
directory, prefix = sys.argv[1], sys.argv[2]
def triples(variant):
    seen = set()
    for path in glob.glob(os.path.join(directory, f'raw-{variant}-*.json')):
        try: report = json.load(open(path))
        except Exception as error: print('gate: unreadable', path, error); continue
        for stop in report.get('Stops') or []:
            seen.add((stop['Rule'], stop['Site'].replace(prefix, ''), stop['Words'][:200]))
    return seen
head, base = triples('head'), triples('base')
print('gate: triples head', len(head), 'base', len(base))
family = '?. to a'
print('gate: family head', len({t for t in head if family in t[2]}), 'base', len({t for t in base if family in t[2]}))
for label, rows in (('ONLY-HEAD', head - base), ('ONLY-BASE', base - head)):
    for rule, site, words in sorted(rows, key=lambda row: row[1]):
        print(label + '\t' + site + '\t' + rule + '\t' + words)
PY
echo "gate: DONE"
