# #k8ed1f2, family 3's sites with positions: scan-all's instrumented probe on main's apply tree over roots that
# cover tsc's compiler, services and server, every "seen as <B>, which can write <X> where <Y> is read" stop with
# its site, source line and words, distinct by site and words. Run after environment.sh, sealed-job.sh,
# sealed_typescript_source and sealed_npm stage3/api. Never merged.
set -u
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; tail -10 "$TMPDIR/product.log"; exit 3; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
A="$SCAN_ALL_SCRATCH/adapted"
bash stage3/apply.sh --build-product "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -8 "$TMPDIR/apply.log"; exit 3; }
roots="src/compiler/sys.ts src/compiler/program.ts src/compiler/checker.ts src/compiler/watchPublic.ts src/compiler/tsbuildPublic.ts src/compiler/builder.ts src/services/services.ts src/server/editorServices.ts src/tsc/tsc.ts"
for root in $roots; do
	name=$(echo "$root" | tr / _)
	if [ -f "$A/$root" ]; then
		start=$(date +%s)
		"$BIN" "$A/$root" > "$TMPDIR/$name.json" 2>/dev/null
		echo "SCAN $root exit $?, $(( $(date +%s) - start )) s"
	else
		echo "SCAN $root absent"
	fi
done
python3 - "$TMPDIR" "$A" $roots <<'PYTHON'
import glob, json, os, re, sys, collections
root_dir, adapted, roots = sys.argv[1], sys.argv[2], sys.argv[3:]
pattern = re.compile(r"a value of type .+ seen as .+, which can write .+ where .+ is read")
others = re.compile(r"whose readonly field|whose method \S+ becomes a writable field|a type parameter whose constraint")
seen = {}
for root in roots:
    p = os.path.join(root_dir, root.replace('/', '_') + '.json')
    if not os.path.exists(p): continue
    try:
        rows = json.load(open(p)).get('Stops') or []
    except Exception as e:
        print(f"unreadable {p}: {e}"); continue
    for s in rows:
        words = s['Words']
        if not pattern.search(words) or others.search(words): continue
        site = re.sub(r'^.*/adapted/', '', s['Site'])
        seen.setdefault((site, words), (s['Rule'], root))
print(f"FAMILY 3 distinct sites: {len(seen)}")
by = collections.Counter(w for (_, w) in seen)
print(f"distinct reasons: {len(by)}")
for (site, words), (rule, root) in sorted(seen.items()):
    file, line, col = site.rsplit(':', 2)
    try:
        text = open(os.path.join(adapted, file), newline='').read().split('\n')[int(line) - 1].strip()
    except Exception:
        text = '?'
    print(f"SITE {site}\t{rule.split('/')[-1]}\t{words}\t{text[:240]}")
PYTHON
echo "VERDICT DONE"
