source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin unit-census-pair || exit 1
# ask: --size F64 --ceiling 120m
# unit-census-pair: tsc's paired census, base main's tip and head, compared by unit by stage3/census/pair/pair.py's rule:
# 0 units newly stopping, or each newly stopping unit named in the commit message with the ruling that allows it
# (cloud/gates/diff-to-gates.json), as a line `allowed: <file> <name> #<task>` alone. Seats' proven form (J-f910xec:
# run.sh, both.sh and units-compare.py, family 141), run at the head's own sha: the base is HEAD^ from the tree's own
# history, never a probe's patch. Each side is laid by
# git archive in the gate's scratch with the laid cohere copied in (so the head may not move cohere's pin), adapted by
# stage3/apply.sh, and measured by the latent census in two passes, tsc's entry and the compiler directory. The base's
# scratch is freed before the head's is built, since a rented box's scratch holds one adapted tree (J-p6913sd), and
# cloud/gates/units-compare.py compares each pass by unit.
# Its ceiling holds its caps (review-r round 3): a proof asks two hours at most (`ahra loom kinds`), so the caps come
# down to fit it rather than the ask going up. Four census passes of at most 22 minutes each are 88, which leaves 32 for
# the npm and tsc installs, two applies and four tool builds; J-f910xec ran both sides whole in 83 and J-ffhyf3v in 69.
# A pass that reaches its cap is killed and fails the gate, so a slow box asks again, never passes.
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_base
[ "$(git rev-parse "$gateBase:cohere")" = "$(git rev-parse HEAD:cohere)" ] ||
	gate_abort cohere "the head moves cohere's pin, and the laid cohere is the head's alone"
gate_preamble npm sandboxed_npm stage3/api
gate_preamble typescript sandboxed_typescript_source
gateRepository=$PWD
git log -1 --format=%B HEAD > "$gateScratch/message"

# census_side <side> <revision>: one side laid, adapted and measured in $gateScratch/<side>, in a subshell so its
# settings stay its own. Its rows are <side>/entry.jsonl and <side>/directory.jsonl, its tree <side>/adapted.
census_side() (
	side=$1 revision=$2
	scratch=$gateScratch/$side
	mkdir -p "$scratch/tmp" "$scratch/source" || exit 1
	export TMPDIR="$scratch/tmp" GOTMPDIR="$scratch/tmp" GOCACHE="$scratch/go-cache" STAGE3_CACHE="$scratch/stage3-cache"
	export ADAMIC_BUILD_CACHE_DIR="$scratch/products"
	unset STAGE3_PRODUCT_STORE ADAMIC_BUILD_CACHE_URL GOCACHEPROG
	git archive "$revision" | tar -xf - -C "$scratch/source" || { echo "  $side: git archive $revision failed"; exit 1; }
	mkdir -p "$scratch/source/cohere"
	tar -C "$gateRepository/cohere" --exclude=.git -cf - . | tar -xf - -C "$scratch/source/cohere" || exit 1
	ln -s "$gateRepository/stage3/api/node_modules" "$scratch/source/stage3/api/node_modules" || exit 1
	cd "$scratch/source" || exit 1
	bash stage3/apply.sh "$scratch/adapted" > "$scratch/apply.log" 2>&1 || { tail -40 "$scratch/apply.log"; echo "  $side: stage3/apply.sh failed"; exit 1; }
	python3 stage3/census/latent/make_overlay.py "$PWD" "$scratch/latent-overlay" > "$scratch/overlay.log" 2>&1 &&
		gofmt -w "$scratch/latent-overlay/"*.go &&
		python3 stage3/meter/entry_overlay.py "$scratch/latent-overlay" "$scratch/entry-overlay" >> "$scratch/overlay.log" 2>&1 &&
		gofmt -w "$scratch/entry-overlay/"*.go &&
		go build -buildvcs=false -overlay="$scratch/entry-overlay/overlay.json" -o "$scratch/entry-census" ./stage3/census/latent/tool &&
		go build -buildvcs=false -overlay="$scratch/latent-overlay/overlay.json" -o "$scratch/directory-census" ./stage3/census/latent/tool ||
		{ tail -40 "$scratch/overlay.log"; echo "  $side: the census tools didn't build"; exit 1; }
	result=0
	for pass in entry directory; do
		target=$scratch/adapted/src/compiler
		[ "$pass" = entry ] && target=$scratch/adapted/src/tsc/tsc.ts
		began=$SECONDS rc=0
		env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 CENSUS_MODULE_FORMAT=CommonJSBundle \
			LATENT_TIMES="$scratch/$pass.times.jsonl" timeout --signal=KILL 22m "$scratch/$pass-census" "$target" \
			"$scratch/$pass.jsonl" > "$scratch/$pass.log" 2>&1 || rc=$?
		printf '  %s %s: exit=%s wall_seconds=%s\n' "$side" "$pass" "$rc" "$((SECONDS - began))"
		# A pass is whole only when it exited 0, opened with its measurement header in first-error mode for the
		# CommonJS bundle, closed with its relation report, and holds a row for every source it measured.
		python3 -I - "$scratch/adapted" "$scratch/$pass.jsonl" "$rc" << 'PYTHON' || result=1
import json, sys
from pathlib import Path
tree, stream, rc = sys.argv[1], Path(sys.argv[2]), int(sys.argv[3])
rows = []
if stream.exists():
    for line in stream.read_text().splitlines():
        try:
            rows.append(json.loads(line))
        except json.JSONDecodeError:
            break
whole = (rc == 0 and len(rows) >= 3 and rows[0].get('status') == 'measurement' and not rows[0].get('checker_rejected')
         and not rows[0].get('diagnostics') and rows[0].get('latent_mode') == 'first-error'
         and rows[0].get('module_format') == 'CommonJSBundle' and rows[-1].get('status') == 'relation_report')
if whole:
    files = {row['file'] for row in rows if 'file' in row}
    expected = set(rows[0].get('sources', [])) or {
        str(p) for p in (Path(tree) / 'src/compiler').rglob('*') if p.is_file() and p.suffix in {'.ts', '.a'}}
    whole = files == expected and len(files) == sum('file' in row for row in rows)
print(f'    {len(rows)} rows, {"whole" if whole else "not whole"}')
sys.exit(0 if whole else 1)
PYTHON
	done
	exit "$result"
)

if census_side base "$gateBase"; then
	gate_pass "census at the base $gateBase"
else
	gate_abort base-census "the census at the base $gateBase"
fi
# Keep the base's rows and its adapted sources' names, then free its scratch for the head's tree.
mkdir -p "$gateScratch/keep"
cp "$gateScratch/base/entry.jsonl" "$gateScratch/base/directory.jsonl" "$gateScratch/keep/" || gate_abort keep "keeping the base's rows"
baseTree=$gateScratch/base/adapted
rm -rf -- "${gateScratch:?}/base"
if census_side head HEAD; then
	gate_pass "census at the head $gateHead"
else
	gate_abort head-census "the census at the head"
fi
for pass in entry directory; do
	printf '== %s: units, base against head\n' "$pass"
	if python3 -I cloud/gates/units-compare.py "$gateScratch/keep/$pass" "$gateScratch/head/$pass" "$baseTree" \
		"$gateScratch/head/adapted" --message "$gateScratch/message"; then
		gate_pass "$pass: 0 units newly stopping that the commit message doesn't allow"
	else
		gate_fail "$pass" "units newly stopping, or a side not read"
	fi
done
gate_end
