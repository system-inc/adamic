"""Two first-stop tables, one read: what moved between two census runs.

Usage: python3 -I fsdiff.py <previous log> <current log> [pass]
Each log holds firststops.py's FS lines (FS, pass, file, name, kind, ordinal, finding kind, where, reason). A unit
is (file, name, kind, ordinal). Families are families.py's rules beside it, first fullmatch wins.
Positions after .ts are stripped from reasons first. Prints totals, families by change, then every unit that lowers
now, stops newly, or moved family, and the units whose stop moved site inside one family as a count.
"""
import collections, re, sys
from pathlib import Path

sys.dont_write_bytecode = True
sys.path.insert(0, str(Path(__file__).resolve().parent))
import families

previousPath, currentPath = sys.argv[1:3]
passName = sys.argv[3] if len(sys.argv) > 3 else 'entry'
position = re.compile(r'(\.[cm]?ts|\.a)(?::\d+(?::\d+)?)')

def familyOf(reason):
    return families.familyOf(position.sub(r'\1', reason))

def table(path):
    units = {}
    for line in open(path, encoding='utf-8', errors='replace'):
        parts = line.rstrip('\n').split('\t')
        if len(parts) == 9 and parts[0] == 'FS' and parts[1] == passName:
            units[tuple(parts[2:6])] = (parts[6], parts[7], parts[8], familyOf(parts[8]))
    if len(units) < 100:
        sys.exit(f'{path}: {len(units)} first stops in pass {passName}, under 100; not a census table')
    return units

before, after = table(previousPath), table(currentPath)
count = lambda units: collections.Counter(stop[3] for stop in units.values())
cb, ca = count(before), count(after)
print(f'first stops, {passName} pass: {len(before)} before, {len(after)} after ({len(after) - len(before):+d})')
lowers = sorted(set(before) - set(after))
newly = sorted(set(after) - set(before))
both = set(before) & set(after)
# A stop at the same site whose family changed is a reworded reason, not a move: a rule that no longer matches the
# new words, which families.py should learn. It is listed apart, so it never reads as progress or a regression.
reworded = sorted(u for u in both if before[u][3] != after[u][3] and before[u][1] == after[u][1])
movedFamily = sorted(u for u in both if before[u][3] != after[u][3] and before[u][1] != after[u][1])
# A stop in the same file with the same words, at another line, is the file moving under it (an adaptation that
# grew the file above it), counted apart from a stop that moved to another file or other words.
fileOf = lambda where: where.rsplit(':', 2)[0]
shifted = [u for u in both if before[u][3] == after[u][3] and before[u][1] != after[u][1]
           and fileOf(before[u][1]) == fileOf(after[u][1]) and before[u][2] == after[u][2]]
movedSite = [u for u in both if before[u][3] == after[u][3] and before[u][1] != after[u][1] and u not in shifted]
print(f'units: {len(lowers)} lower now, {len(newly)} stop newly, {len(movedFamily)} moved family, {len(movedSite)} moved site in one family, {len(reworded)} reworded at the same site, {len(shifted)} same stop at a shifted line')
print()
print('families whose count changed (before -> after):')
for family in sorted(set(cb) | set(ca), key=lambda f: (-(abs(ca[f] - cb[f])), f)):
    if ca[family] != cb[family]:
        print(f'  {cb[family]:4} -> {ca[family]:4} ({ca[family] - cb[family]:+d})  {family[:150]}')
def show(label, units, describe):
    print(f'\n{label}: {len(units)}')
    for unit in units:
        print(f'  {unit[0]} {unit[1]}#{unit[3]}: {describe(unit)}')
show('LOWERS NOW', lowers, lambda u: f'was {before[u][1]} {before[u][2][:120]}')
show('NEWLY STOPS', newly, lambda u: f'{after[u][1]} {after[u][2][:120]}')
show('MOVED FAMILY', movedFamily, lambda u: f'{before[u][1]} {before[u][3][:70]} -> {after[u][1]} {after[u][2][:110]}')
show('MOVED SITE IN ONE FAMILY', sorted(movedSite), lambda u: f'{before[u][1]} -> {after[u][1]} {after[u][2][:110]}')
show('REWORDED, SAME SITE (families.py has no rule for the new words)', reworded, lambda u: f'{after[u][1]} {before[u][3][:60]} -> {after[u][2][:110]}')
