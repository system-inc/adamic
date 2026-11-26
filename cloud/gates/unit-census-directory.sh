source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin unit-census-directory || exit 1
# ask: --size F64 --ceiling 120m --needs History
# unit-census-entry and unit-census-directory: tsc's paired census in one pass each, base main's tip and head, compared
# by unit by stage3/census/pair/pair.py's rule: 0 units newly stopping, or each newly stopping unit named in the commit
# message with the ruling that allows it (cloud/gates/diff-to-gates.json), as a line `allowed: <file> <name> #<task>`
# alone. Seats' proven form (J-f910xec: run.sh, both.sh and units-compare.py, family 141), run at the head's own sha:
# the base is HEAD^ from the tree's own history, never a probe's patch, so it asks a machine that lays the history
# (`--needs History`, #zx98rbc). Each side is laid by git archive in the gate's scratch with the laid cohere copied in
# (so the head may not move cohere's pin), adapted by stage3/apply.sh, and measured by the latent census in the
# recipe's one pass, censusPass below: tsc's entry, or the compiler directory. The base's scratch is freed before the
# head's is built, since a rented box's scratch holds one adapted tree (J-p6913sd), and cloud/gates/units-compare.py
# compares the pass by unit.
# Its ceiling holds its cap (review-r round 3): a proof asks two hours at most (`ahra loom kinds`), so the cap comes
# down to fit it rather than the ask going up. One pass of at most 50 minutes on each side is 100, which leaves 20 for
# the npm and tsc installs, two applies and two tool builds. On the rented boxes at main ef741d4f (Nov 26), the entry
# pass took 1,788 to 1,864 s a side and the directory pass 1,799 to 1,924 s, and a job of one pass on both sides took 72
# to 77 minutes, 12 to 14 of them outside the census (J-r198seq, J-e7wn882, J-nywaah4, J-rrdpad3, J-f79ghd8, J-b9n0qyg,
# J-1jvwbcz). Both passes on both sides would be 120 to 126 minutes of census alone, the whole ceiling, so each pass is
# its own gate. A pass that reaches its cap is killed and fails the gate, so a slow box asks again, never passes.
# The two recipes are one text but for their gate's name and censusPass, as submit_gates_test.py holds them, and each
# is its own file because cloud/submit-gates.py matches a recipe by the sha256 of its whole command: a pass read from
# the job's environment would be one more line in the command, and no recipe's hash.
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
censusPass=directory
gate_base
[ "$(git rev-parse "$gateBase:cohere")" = "$(git rev-parse HEAD:cohere)" ] ||
	gate_abort cohere "the head moves cohere's pin, and the laid cohere is the head's alone"
gate_preamble npm sandboxed_npm stage3/api
gate_preamble typescript sandboxed_typescript_source
gateRepository=$PWD
git log -1 --format=%B HEAD > "$gateScratch/message" || gate_abort message "the head's commit message wasn't read"

# census_side <side> <revision>: one side laid, adapted and measured in $gateScratch/<side>, in a subshell so its
# settings stay its own. Its rows are <side>/<censusPass>.jsonl, its tree <side>/adapted.
census_side() (
	# pipefail, so a git archive or a cohere copy that fails fails its side, not only the tar that reads it
	set -o pipefail
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
		gofmt -w "$scratch/latent-overlay/"*.go ||
		{ tail -40 "$scratch/overlay.log"; echo "  $side: the census overlay didn't build"; exit 1; }
	# The directory pass measures the compiler directory by the latent overlay; the entry pass measures tsc's entry by
	# the overlay stage3/meter/entry_overlay.py makes from it.
	overlay=$scratch/latent-overlay target=$scratch/adapted/src/compiler
	if [ "$censusPass" = entry ]; then
		overlay=$scratch/entry-overlay target=$scratch/adapted/src/tsc/tsc.ts
		python3 stage3/meter/entry_overlay.py "$scratch/latent-overlay" "$overlay" >> "$scratch/overlay.log" 2>&1 &&
			gofmt -w "$overlay/"*.go ||
			{ tail -40 "$scratch/overlay.log"; echo "  $side: the entry overlay didn't build"; exit 1; }
	fi
	go build -buildvcs=false -overlay="$overlay/overlay.json" -o "$scratch/census" ./stage3/census/latent/tool ||
		{ echo "  $side: the census tool didn't build"; exit 1; }
	began=$SECONDS rc=0
	env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 CENSUS_MODULE_FORMAT=CommonJSBundle \
		LATENT_TIMES="$scratch/$censusPass.times.jsonl" timeout --signal=KILL 50m "$scratch/census" "$target" \
		"$scratch/$censusPass.jsonl" > "$scratch/$censusPass.log" 2>&1 || rc=$?
	printf '  %s %s: exit=%s wall_seconds=%s\n' "$side" "$censusPass" "$rc" "$((SECONDS - began))"
	# A pass is whole only when it exited 0, every line of its stream reads, it opened with its measurement header in
	# first-error mode for the CommonJS bundle, closed with its relation report, and holds a row for every source it
	# measured. The header is held to the recipe's own pass (review-w read 9): the entry pass's names the target as its
	# one root and the sources its program loaded, the root among them, as the meter's report reads it; the compiler
	# directory's names neither, and its sources are the directory's files. So a gate never passes on the other pass.
	python3 -I - "$scratch/adapted" "$scratch/$censusPass.jsonl" "$rc" "$censusPass" "$target" << 'PYTHON'
import json, sys
from pathlib import Path
tree, stream, rc, census_pass, target = sys.argv[1], Path(sys.argv[2]), int(sys.argv[3]), sys.argv[4], sys.argv[5]
rows, unread = [], False
if stream.exists():
    for line in stream.read_text().splitlines():
        try:
            rows.append(json.loads(line))
        except json.JSONDecodeError:
            unread = True
            break
whole = (rc == 0 and len(rows) >= 3 and rows[0].get('status') == 'measurement' and not rows[0].get('checker_rejected')
         and not unread and not rows[0].get('diagnostics') and rows[0].get('latent_mode') == 'first-error'
         and rows[0].get('module_format') == 'CommonJSBundle' and rows[-1].get('status') == 'relation_report')
if whole:
    header, files = rows[0], {row['file'] for row in rows if 'file' in row}
    sources, expected = header.get('sources'), None
    if census_pass == 'entry':
        if (header.get('roots') == [target] and isinstance(sources, list) and len(set(sources)) == len(sources)
                and target in set(sources) | {str(Path(name).resolve()) for name in sources}):
            expected = set(sources)
    elif census_pass == 'directory' and 'roots' not in header and sources is None:
        expected = {str(p) for p in (Path(tree) / 'src/compiler').rglob('*')
                    if p.is_file() and p.suffix in {'.ts', '.a'}}
    whole = expected is not None and files == expected and len(files) == sum('file' in row for row in rows)
print(f'    {len(rows)} rows, {"whole" if whole else "not whole"}')
sys.exit(0 if whole else 1)
PYTHON
)

if census_side base "$gateBase"; then
	gate_pass "census at the base $gateBase"
else
	gate_abort base-census "the census at the base $gateBase"
fi
# Keep the base's rows and its adapted sources' names, then free its scratch for the head's tree.
mkdir -p "$gateScratch/keep"
cp "$gateScratch/base/$censusPass.jsonl" "$gateScratch/keep/" || gate_abort keep "keeping the base's rows"
baseTree=$gateScratch/base/adapted
rm -rf -- "${gateScratch:?}/base" || gate_abort free "the base's scratch wasn't freed for the head's tree"
if census_side head HEAD; then
	gate_pass "census at the head $gateHead"
else
	gate_abort head-census "the census at the head"
fi
printf '== %s: units, base against head\n' "$censusPass"
if python3 -I cloud/gates/units-compare.py "$gateScratch/keep/$censusPass" "$gateScratch/head/$censusPass" "$baseTree" \
	"$gateScratch/head/adapted" --message "$gateScratch/message"; then
	gate_pass "$censusPass: 0 units newly stopping that the commit message doesn't allow"
else
	gate_fail "$censusPass" "units newly stopping, or a side not read"
fi
gate_end
