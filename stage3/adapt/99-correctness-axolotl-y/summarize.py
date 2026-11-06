"""Compare all exact root paths with the requested reference table."""
from pathlib import Path
import collections,json,re
out=Path(__file__).resolve().parent
reference=(out/'reference-report.txt').read_text().split('## Grouped first stops')[1].split('## Whole-root lowering')[0]
prior={}
for line in reference.splitlines():
 if not re.match(r'\| \d+ \|',line):continue
 cols=[s.strip() for s in line.split('|')[1:-1]]
 for root in re.findall(r'`([^`]+)`',cols[4]):
  prior[root]={'kind':cols[2],'where':cols[1].strip('`'),'reason':cols[3]}
assert len(prior)==79,len(prior)
rows=[]
for i in range(1,5):
 group=json.loads((out/f'group-{i}-rows.json').read_text());assert len(group)==(19 if i==4 else 20)
 rows+=group
assert len(rows)==79
current={}
for r in rows:
 assert len(r['roots'])==1
 root=r['roots'][0].split('/src/compiler/')[1];assert root not in current
 current[root]={'kind':r['kind'],'where':r.get('where','-').split('/src/compiler/')[-1],'reason':r.get('reason','-')}
assert current.keys()==prior.keys()
groups=collections.defaultdict(list)
for root,r in current.items():groups[(r['kind'],r['where'],r['reason'])].append(root)
grouped=[dict(count=len(names),kind=k[0],where=k[1],reason=k[2],roots=names) for k,names in sorted(groups.items(),key=lambda item:(-len(item[1]),item[0]))]
comparison=[dict(root=root,before=prior[root],after=current[root],changed=prior[root]!=current[root]) for root in sorted(current)]
moved=[r for r in comparison if r['before']['where']=='core.ts:966:72']
assert len(moved)==73;assert all(r['after']['where']!='core.ts:966:72' for r in moved)
result={'roots':79,'movedPastCore966':73,'groups':grouped,'rootsBeforeAfter':comparison}
(out/'census-results.json').write_text(json.dumps(result,indent=2)+'\n')
text='''# 99y per-file compiler census

Compiler and apply base: 322f80565edaa978d928ca559803c578ed67c105.
The ordinary census tool uses no overlay, refusal suppression, backend change,
or partial-root fallback. All 79 exact root paths use the completed full
/workspace/y-after tree. Groups contain 20, 20, 20, and 19 roots. Each was
committed and plainly pushed as soon as it finished. Each invocation retains
both raw ordinary rows and counts only the first; their non-timing fields match.

Compared against the grouped table in typescript/next-stops-7e3c0682,
retained as reference-report.txt. All 73 roots at core.ts:966:72 move past it.

| Roots | Location | Kind | Exact reason | Root names |
| ---: | --- | --- | --- | --- |
'''
for g in grouped:
 text+=f"| {g['count']} | `{g['where']}` | {g['kind']} | {g['reason'].replace('|',chr(92)+'|')} | "+', '.join('`'+n+'`' for n in g['roots'])+' |\n'
text+='\n## Comparison with the reference\n\n| Roots | Reference first stop | After first stop | Change |\n| ---: | --- | --- | --- |\n'
transitions=collections.defaultdict(list)
for r in comparison:transitions[(r['before']['where'],r['after']['where'],r['changed'])].append(r['root'])
for (a,b,changed),names in sorted(transitions.items(),key=lambda item:-len(item[1])):
 text+=f"| {len(names)} | `{a}` | `{b}` | {'changed' if changed else 'unchanged'} |\n"
text+='''
These are first lowering stops after checking. Accepted roots indicate ordinary
whole-root Lower acceptance only, with no native-emission or aggregate-program
claim. The reference's three readiness artifact classifications remain incomplete;
no new artifact or whole-root credit is inferred. census-results.json retains
every root's exact before and after row, including duplicate basenames.
'''
(out/'CENSUS.md').write_text(text)
print(json.dumps(grouped,indent=2))
