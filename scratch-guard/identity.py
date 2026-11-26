"""Scratch, never landing (#8nj5pn4, Emit's guard): the C identity of every probed program, main's emitter against the
head's. Arguments: main's PROBE lines, the head's, main's C directory, the head's.

A program whose main C reads no field on the uniform road must be byte-identical at the head. A program whose C
changed must be the guard and nothing else: with each of a field's three roads (the bare uniform slot main reads, the
guarded slot, and the by-name read alone a guard with no literal shape becomes) written as one token, every field
cache's numbering erased and its declaration dropped, and blank lines dropped (the first cache a program declares
opens a group of declarations with one), the two sides' lines must be the same multiset (a shape a guard names is
declared earlier, so a declaration may move). An object is a C name or a boxed one, (*name), and a field's name a
string or the char array cString makes of a long one, which spans lines. A uniform read whose name no literal holds
has no shape to compare, so the guard is its by-name read alone: those reads are counted and named, by program."""
import collections
import difflib
import os
import re
import sys

main_lines, head_lines, main_dir, head_dir = sys.argv[1:5]
string = r'(?:"(?:[^"\\]|\\.)*"|\(\(const char\[\]\)\{[0-9,\s]*\}\))'
target = r'(?:\w+|\(\*\w+\))'
guard = re.compile(r'\((?:' + target + r'->shape == &adamic_shape_\w+(?: \|\| )?)+ \? \(&ADAMIC_UNIFORM_OBJECT\((' + target + r'), (' + string + r')\)->slots\[\d+\]\) : adamic_object_data_field\(' + target + r', ' + string + r', &adamic_cache_\w+\)\)')
bare = re.compile(r'\(&ADAMIC_UNIFORM_OBJECT\((' + target + r'), (' + string + r')\)->slots\[\d+\]\)')
byname = re.compile(r'adamic_object_data_field\((' + target + r'), (' + string + r'), &adamic_cache_\w+\)')
cache = re.compile(r'adamic_cache_\w+')
unguarded_read = re.compile(r'(?<!\? )\(&ADAMIC_UNIFORM_OBJECT\(')


def normalized(text):
    text = guard.sub(r'FIELD(\1, \2)', text)
    text = bare.sub(r'FIELD(\1, \2)', text)
    text = byname.sub(r'FIELD(\1, \2)', text)
    text = cache.sub('adamic_cache', text)
    return sorted(line for line in text.split('\n') if line.strip() != '' and line != 'static adamic_slot_cache adamic_cache;')


def names(pattern, text):
    return collections.Counter(re.sub(r'\s+', '', match.group(2))[:40] for match in pattern.finditer(text))


def by_name(main_c, head_c):
    """The uniform reads main took bare that the head takes by name alone, by field name."""
    main_guards, head_guards = names(guard, main_c), names(guard, head_c)
    return (names(bare, main_c) - main_guards) - (head_guards - main_guards)


def read(path):
    rows = {}
    for line in open(path):
        parts = line.split()
        if len(parts) >= 4 and parts[0] == 'PROBE':
            rows[parts[1]] = (parts[2], ' '.join(parts[3:]))
    return rows


main, head = read(main_lines), read(head_lines)
if main.keys() != head.keys():
    print('FAIL identity: the two runs probed different programs')
    sys.exit(1)
same = changed = stops = plain = uniform_same = 0
bad = []
by_name_programs = []
by_name_reads = 0
residues = []
show = os.environ.get('GUARD_SHOW', '').split()
shown = 0
for index in sorted(main, key=int):
    path, main_row = main[index]
    _, head_row = head[index]
    if not main_row.startswith('c ') or not head_row.startswith('c '):
        stops += 1
        if main_row != head_row:
            bad.append(f'{path}: main {main_row[:60]}, head {head_row[:60]}')
        continue
    main_c = open(f'{main_dir}/{index}.c').read()
    head_c = open(f'{head_dir}/{index}.c').read()
    reads = unguarded_read.search(main_c) is not None
    if main_c == head_c:
        same += 1
        if reads:
            uniform_same += 1
        else:
            plain += 1
        continue
    changed += 1
    converted = by_name(main_c, head_c)
    if converted:
        by_name_programs.append(f'{path}: {", ".join(f"{name} x{count}" for name, count in sorted(converted.items()))}')
        by_name_reads += sum(converted.values())
    if not reads:
        bad.append(f'{path}: changed, though main reads no field on the uniform road')
    elif normalized(main_c) != normalized(head_c):
        bad.append(f'{path}: changed by more than the guard')
        # What the normalization leaves on each side, so the residue is read, not guessed.
        main_rest = collections.Counter(normalized(main_c)) - collections.Counter(normalized(head_c))
        head_rest = collections.Counter(normalized(head_c)) - collections.Counter(normalized(main_c))
        residues.append(f'-- {path}: {sum(main_rest.values())} lines main only, {sum(head_rest.values())} head only')
        for line in list(main_rest.elements())[:2]:
            residues.append('   main: ' + line.strip()[:220])
        for line in list(head_rest.elements())[:2]:
            residues.append('   head: ' + line.strip()[:220])
    if any(path.endswith('/' + shown_name) for shown_name in show):
        print(f'-- {path}, diff -u main head, whole:')
        for line in list(difflib.unified_diff(main_c.split('\n'), head_c.split('\n'), lineterm='', n=0))[2:90]:
            print('   ' + line[:500])
    if shown < 2 and not show:
        shown += 1
        print(f'-- {path}, its first changed lines:')
        for line in list(difflib.unified_diff(main_c.split('\n'), head_c.split('\n'), lineterm='', n=0))[2:10]:
            print('   ' + line[:300])
print(f'programs: {len(main)}; lowered to C at both: {same + changed}; stopped at either: {stops}')
print(f'byte-identical: {same} ({plain} reading no field on the uniform road, {uniform_same} reading one); changed: {changed}')
added = 0
for index in head:
    if main[index][1].startswith('c ') and head[index][1].startswith('c '):
        added += len(guard.findall(open(f'{head_dir}/{index}.c').read())) - len(guard.findall(open(f'{main_dir}/{index}.c').read()))
print(f'shape guards added across every program, reads guarded less any a write lost: {added}')
print(f'programs whose uniform reads of a name no literal holds now go by name alone: {len(by_name_programs)}, {by_name_reads} reads')
for line in by_name_programs:
    print('   by name: ' + line[:300])
for line in bad:
    print('FAIL ' + line)
for line in residues[:400]:
    print(line)
print('PASS identity' if not bad else f'FAIL identity: {len(bad)} programs')
sys.exit(1 if bad else 0)
