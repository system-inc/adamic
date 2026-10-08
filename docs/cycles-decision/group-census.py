import subprocess,re,csv,collections
from pathlib import Path
s=subprocess.check_output(['git','show','fcb7451a:docs/stage3-tsc-cycles.md'],text=True)
rows=[]
for l in s.splitlines():
 m=re.match(r'\| ([DK]\d+) `([^`]+)` \| (.*?) \| (.*?) \| \(([abc])\) (\w+);',l)
 if not m: continue
 id,slot,site,witness,cat,rule=m.groups()
 if rule=='P': group='node.parent'
 elif slot.endswith(('.declarations','.valueDeclaration')) and 'Symbol' in slot: group='symbol declarations'
 elif 'SymbolLinks' in slot or 'SymbolTable' in slot or witness.startswith('SymbolTable') or (id.startswith('K') and re.search(r'(?:, |<)Symbol>',slot)) or slot.startswith(('Symbol.','TransientSymbol.','MappedSymbol.','ReverseMappedSymbol.')): group='symbol tables and links'
 elif rule=='FL' or 'Flow' in slot: group='flow nodes'
 elif slot.startswith('EmitNode.') or slot.endswith('.emitNode'): group='emit nodes'
 elif rule in ('SELF','CK') or (rule=='CACHE' and slot.split('.')[0] not in ('NodeLinks','InferenceContext','InferenceInfo') and not slot.startswith(('CommandLineOption','JSDoc'))): group='type and relation caches'
 else: group='other connecting slots'
 rows.append((id,slot,group,cat,rule,re.sub('<[^>]*>','',site),witness))
assert len(rows)==1685,len(rows)
assert len({r[0] for r in rows}) == 1685
assert collections.Counter(r[3] for r in rows) == {"a":80,"b":2,"c":1603}
p=Path('docs/cycles-decision');p.mkdir(exist_ok=True)
with (p/'census-groups.csv').open('w') as f:
 w=csv.writer(f);w.writerow(['id','slot','group','category','rule','site','witness']);w.writerows(rows)
for group in dict.fromkeys(r[2] for r in rows):
 rs=[r for r in rows if r[2]==group];print(group,len(rs),collections.Counter(r[3] for r in rs));print(rs[:2])
print('rules',collections.Counter(r[4] for r in rows))
