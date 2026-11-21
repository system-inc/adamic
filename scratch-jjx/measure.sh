#!/usr/bin/env bash
# measure.sh: a scratch measurement for #jjxtwzs, never to land. Head holds a type as ir.Union where an array may stand
# for it at a checked view (internal/lower/arrays_admitted.go, representationReason's new case); base is 3eb4ba5e, which
# holds every such type as ir.Object. Run from an adamic tree's root, on a sealed job, as
#
#   source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && bash scratch-jjx/measure.sh
#
# It builds ./cmd/adamic twice: head as the tree is laid, and base with go build's -overlay putting the pristine
# 3eb4ba5e copies of the changed lower files (scratch-jjx/base/, so no git history is needed) in place of head's and
# dropping arrays_admitted.go, so the tree itself is never written. Then it compiles (adamic c <file>, compile only)
# every .ts (no .d.ts) and .a under internal/oracle/testdata, stage3/fixtures and stage3/interface-downcasts with both
# binaries, head's with ADAMIC_SCRATCH_ARRAYS_SET naming a file of its own, and prints:
#   - the canary: field_empty_interface.ts's set, which should be exactly { readonly a?: number; };
#   - how many programs have a non-empty set, the distinct types, and the program count per type;
#   - every program whose compile outcome (exit code and stderr's first line) moved from base to head, both on one line;
#   - how many programs' C moved with the outcome unmoved, a count only;
#   - MEASURE DONE.
# Past 200 rows a list prints counts, never rows; the whole lists stay in the evidence directory under TMPDIR.
set -u

root=$PWD
[ -f "$root/go.mod" ] && [ -d "$root/cmd/adamic" ] && [ -f "$root/internal/lower/arrays_admitted.go" ] || {
	echo "jjx: $root is not this scratch tree's root, so nothing ran"
	exit 64
}
for file in expression.go lower.go; do
	[ -f "$root/scratch-jjx/base/internal/lower/$file" ] || {
		echo "jjx: scratch-jjx/base/internal/lower/$file is missing, so nothing ran"
		exit 64
	}
done
missing=""
for tool in go python3 sha256sum xargs timeout; do
	command -v "$tool" > /dev/null 2>&1 || missing="$missing $tool"
done
[ -z "$missing" ] || {
	echo "jjx: missing from the toolchain:$missing"
	exit 65
}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
out="${TMPDIR:-/tmp}/measure-jjx-$stamp"
mkdir -p "$out/sets" "$out/runs" || exit 64
echo "jjx: $(uname -n), $(nproc) threads, $(go version), evidence in $out"

# The base build's overlay: each changed file read from its pristine copy, and the new file gone.
python3 -I - "$root" "$out/overlay.json" << 'PYTHON' || exit 64
import json, sys
root, path = sys.argv[1:3]
replace = {}
for name in ('expression.go', 'lower.go'):
    replace[f'{root}/internal/lower/{name}'] = f'{root}/scratch-jjx/base/internal/lower/{name}'
replace[f'{root}/internal/lower/arrays_admitted.go'] = ''
with open(path, 'w') as file:
    json.dump({'Replace': replace}, file, indent=1)
PYTHON

go build -o "$out/adamic-head" ./cmd/adamic > "$out/build-head.log" 2>&1 || {
	echo "jjx: head didn't build:"
	head -c 4000 "$out/build-head.log"
	exit 66
}
go build -overlay "$out/overlay.json" -o "$out/adamic-base" ./cmd/adamic > "$out/build-base.log" 2>&1 || {
	echo "jjx: base didn't build:"
	head -c 4000 "$out/build-base.log"
	exit 66
}
if cmp -s "$out/adamic-head" "$out/adamic-base"; then
	echo "jjx: head and base built the same binary, so the overlay didn't take; nothing measured"
	exit 66
fi
echo "jjx: built head and base"

find internal/oracle/testdata stage3/fixtures stage3/interface-downcasts -type f \( -name '*.a' -o \( -name '*.ts' ! -name '*.d.ts' \) \) | LC_ALL=C sort > "$out/programs.txt"
echo "jjx: $(wc -l < "$out/programs.txt") programs"

# compileOne <index> <file>: both compiles of one program, as one line of $out/runs/<index>.tsv:
# file, base exit, base stderr's first line, base C sha256, head exit, head stderr's first line, head C sha256, set file.
compileOne() {
	local index=$1 file=$2 side code first sum line=""
	: > "$JJX_OUT/sets/$index.tsv"
	for side in base head; do
		if [ "$side" = head ]; then
			env ADAMIC_SCRATCH_ARRAYS_SET="$JJX_OUT/sets/$index.tsv" timeout 300 "$JJX_OUT/adamic-head" c "$file" > "$JJX_OUT/runs/$index.$side.c" 2> "$JJX_OUT/runs/$index.$side.err"
		else
			env -u ADAMIC_SCRATCH_ARRAYS_SET timeout 300 "$JJX_OUT/adamic-base" c "$file" > "$JJX_OUT/runs/$index.$side.c" 2> "$JJX_OUT/runs/$index.$side.err"
		fi
		code=$?
		first=$(head -n 1 "$JJX_OUT/runs/$index.$side.err" | tr '\t\r' '  ' | cut -c 1-400)
		sum=$(sha256sum < "$JJX_OUT/runs/$index.$side.c" | cut -c 1-16)
		line="$line	$code	$first	$sum"
		rm -f "$JJX_OUT/runs/$index.$side.c"
	done
	printf '%s%s\t%s\n' "$file" "$line" "$JJX_OUT/sets/$index.tsv" > "$JJX_OUT/runs/$index.tsv"
}
export -f compileOne
export JJX_OUT="$out"
awk '{ printf "%d\t%s\n", NR, $0 }' "$out/programs.txt" | tr '\t' '\n' | xargs -d '\n' -n 2 -P "$(nproc)" bash -c 'compileOne "$1" "$2"' _
echo "jjx: compiled"

python3 -I - "$out" << 'PYTHON'
import glob, os, sys
out = sys.argv[1]
limit = 200
rows = []
for path in glob.glob(os.path.join(out, 'runs', '*.tsv')):
    with open(path) as file:
        fields = file.read().rstrip('\n').split('\t')
    if len(fields) != 8:
        print(f'jjx: malformed run {path}: {len(fields)} fields')
        continue
    rows.append(fields)
rows.sort(key=lambda fields: fields[0])
with open(os.path.join(out, 'programs.txt')) as file:
    expected = sum(1 for _ in file)
print(f'jjx: {len(rows)} of {expected} programs recorded')

# The set a program's head compile reported: its last line, one a lowering run; lines that disagree are counted.
sets = {}
disagreeing = []
unreported = 0
for fields in rows:
    lines = []
    if os.path.exists(fields[7]):
        with open(fields[7]) as file:
            lines = [line.rstrip('\n').split('\t') for line in file if line.strip()]
    if not lines:
        unreported += 1
        continue
    if len({tuple(line[1:]) for line in lines}) > 1:
        disagreeing.append(fields[0])
    last = lines[-1]
    names = [name for name in (last[2].split(' | ') if len(last) > 2 and last[2] else [])]
    sets[fields[0]] = (last[1] if len(last) > 1 else '?', names)
print(f'jjx: {unreported} programs reported no set (stopped before the pre-pass, or never lowered)')
print(f'jjx: {len(disagreeing)} programs whose lowering runs reported different sets')
for name in disagreeing[:20]:
    print(f'jjx:   disagreeing {name[-100:]}')

canary = 'internal/oracle/testdata/view_array_object_stops/field_empty_interface.ts'
if canary in sets:
    held = sets[canary][1]
    verdict = 'PASS' if held == ['{ readonly a?: number; }'] else 'FAIL'
    print(f'jjx: canary {verdict}: {canary} holds {sets[canary][0]}: {" | ".join(held)}')
else:
    print(f'jjx: canary FAIL: {canary} reported no set')

nonEmpty = {name: held for name, held in sets.items() if held[1]}
every = sum(1 for held in sets.values() if 'every' in held[0])
print(f'jjx: {len(nonEmpty)} programs have a non-empty set; {every} admitted every type of the shape')
perType = {}
for name, held in nonEmpty.items():
    for admitted in set(held[1]):
        perType.setdefault(admitted, []).append(name)
print(f'jjx: {len(perType)} distinct types (by TypeToString)')
ranked = sorted(perType.items(), key=lambda item: (-len(item[1]), item[0]))
with open(os.path.join(out, 'types.tsv'), 'w') as file:
    for admitted, names in ranked:
        file.write(f'{len(names)}\t{admitted}\t{" ".join(sorted(names))}\n')
for admitted, names in ranked[:limit]:
    example = sorted(names)[0]
    print(f'jjx:   type {len(names)} programs\t{admitted[:120]}\te.g. {example[-80:]}')
if len(ranked) > limit:
    rest = ranked[limit:]
    print(f'jjx:   ... {len(rest)} more types, {sum(len(names) for _, names in rest)} program counts among them (types.tsv)')

moved = [fields for fields in rows if (fields[1], fields[2]) != (fields[4], fields[5])]
cOnly = [fields for fields in rows if (fields[1], fields[2]) == (fields[4], fields[5]) and fields[3] != fields[6]]
with open(os.path.join(out, 'moved.tsv'), 'w') as file:
    for fields in moved:
        file.write('\t'.join(fields[:7]) + '\n')
print(f'jjx: {len(moved)} programs whose compile outcome moved from base to head')
def short(code, first):
    return f'exit {code}: {first[:70]}' if first else f'exit {code}'
if len(moved) <= limit:
    for fields in moved:
        held = sets.get(fields[0], ('?', []))[0]
        print(f'jjx:   moved {fields[0][-80:]} (set {held})\tbase {short(fields[1], fields[2])}\thead {short(fields[4], fields[5])}')
else:
    pairs = {}
    for fields in moved:
        key = (fields[1], fields[4])
        pairs[key] = pairs.get(key, 0) + 1
    for (before, after), count in sorted(pairs.items(), key=lambda item: -item[1]):
        print(f'jjx:   moved exit {before} -> exit {after}: {count} programs')
    heads = {}
    for fields in moved:
        heads[fields[5][:120]] = heads.get(fields[5][:120], 0) + 1
    for first, count in sorted(heads.items(), key=lambda item: -item[1])[:40]:
        print(f'jjx:   moved to head stderr {count}x: {first}')
    print('jjx:   (every moved program in moved.tsv)')
withSet = sum(1 for fields in moved if fields[0] in nonEmpty)
print(f'jjx: {withSet} of the moved programs have a non-empty set')
print(f'jjx: {len(cOnly)} programs whose C moved with the outcome unmoved')
byExit = {}
for fields in rows:
    byExit[(fields[1], fields[4])] = byExit.get((fields[1], fields[4]), 0) + 1
for (before, after), count in sorted(byExit.items()):
    print(f'jjx: exit base {before} head {after}: {count} programs')
PYTHON
echo "MEASURE DONE"
