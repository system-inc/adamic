#!/usr/bin/env python3
"""Run the same whole-program CLI census as 99a, then compare complete records."""
import collections, difflib, json, pathlib, re, subprocess, sys
binary, before, after, out = sys.argv[1:]
out = pathlib.Path(out)
out.mkdir(parents=True, exist_ok=True)
records = []
for label, tree in [('before',before),('after',after)]:
    roots = sorted(str(p) for p in pathlib.Path(tree,'src/compiler').rglob('*') if p.suffix in ['.ts','.a'])
    with (out/f'types-{label}.stdout').open('w') as stdout, (out/f'types-{label}.stderr').open('w') as stderr:
        run = subprocess.run([binary,'types',*roots], stdout=stdout, stderr=stderr, timeout=300)
    assert run.returncode == 1, (label,run.returncode)
    text = (out/f'types-{label}.stderr').read_text().replace(tree,'TREE')
    chunks = re.split(r'(?=^TREE/[^\n]+:\d+:\d+:)',text,flags=re.M)
    chunks = [s for s in chunks if s.strip()]
    assert all(re.match(r'TREE/[^\n]+:\d+:\d+:',s) for s in chunks), chunks[:1]
    records.append(chunks)
# Align unchanged source lines after the two inserted guard lines.
a=pathlib.Path(before,'src/compiler/transformers/declarations.ts').read_text().splitlines()
b=pathlib.Path(after,'src/compiler/transformers/declarations.ts').read_text().splitlines()
line_map={}
for block in difflib.SequenceMatcher(a=a,b=b,autojunk=False).get_matching_blocks():
    for i in range(block.size): line_map[block.b+i+1]=block.a+i+1
normalized=[]
for diagnostic in records[1]:
    def remap(m):
        line=int(m[2]); assert line in line_map, diagnostic
        return m[1]+str(line_map[line])+':'
    normalized.append(re.sub(r'^(TREE/src/compiler/transformers/declarations.ts:)(\d+):',remap,diagnostic))
old,new=map(collections.Counter,[records[0],normalized])
removed=list((old-new).elements()); added=list((new-old).elements())
result=dict(before=len(records[0]),after=len(records[1]),removed=removed,added=added,roots=len(roots))
(out/'measurement.json').write_text(json.dumps(result,indent=2)+'\n')
assert not added, added
assert len(removed)==1 and 'declarations.ts:1678:110:' in removed[0] and 'TS2345' in removed[0],removed
print(json.dumps(result,indent=2))
