import json,collections
from pathlib import Path
out=Path(__file__).resolve().parent
rows=json.loads((out/'comparison.json').read_text());counts=collections.Counter(r['verdict'] for r in rows)
assert counts=={'further down':29,'identity':6,'unchanged':3},counts
cleared=[r for r in rows if r['category']=='cleared-array'];gates=[]
for r in cleared:
 selected=[]
 for side in ['parent','head']:
  lines=[line for line in r[side] if line.startswith('probe-widened\targument\t/workspace/scratch/fa-adapted/src/'+r['site']+'\t')]
  assert len(lines)==1,(r['site'],side,lines)
  selected.append(lines[0])
 fields=[dict(p.split('=',1) for p in line.split('\t')[3:]) for line in selected]
 assert all(fields[0][k]==fields[1][k] for k in ['declared','given','slot','same'])
 assert fields[0]['same']==fields[1]['same']=='false'
 assert fields[0]['assignable']=='true' and fields[1]['assignable']=='false'
 gates.append(dict(site=r['site'],parent=selected[0],head=selected[1],input_fields_identical=True,same_flag_switched=False))
assert len(gates)==30
same_flips=[e for r in rows for e in r['identity_evidence'] if e['kind']=='same changed'];assert not same_flips
(out/'append-gates.json').write_text(json.dumps(dict(sites=30,all_inputs_identical=True,all_same_false=True,all_assignable_true_to_false=True,lines=gates),indent=2)+'\n')
summary=('Further down holds at 29 cleared array sites. The remaining cleared site, module/system.ts:802:17, meets the identity criterion only inside its nested expression, at :811:51, where an equal-text slot has different IDs. Identity also holds at all five new bivariance sites: es2017.ts:163:26, es2018.ts:168:26 and :499:40, and esDecorators.ts:1931:42 and :1954:32. All three append controls have identical complete probe sequences. Every site has a probe line on at least one side. No paired same flag switches; at all 30 cleared append arguments, declared, given, slot IDs and texts and same=false match exactly, while assignable changes from true to false. The two logged pointer-equality tests therefore do not explain the cleared append verdicts.')
p=out/'REPORT.md';s=p.read_text();s=s.split('## Verdict\n')[0]+'## Verdict\n\n'+summary+'\n'
s=s.replace('## Site comparison\n','## Site comparison\n\nThe requested identity criterion takes precedence when any selected line contains equal-text types with different IDs. Identical controls are marked unchanged because their results also match. [Append gate audit](append-gates.json) quotes the exact first-argument decisions at all 30 cleared sites: every input field and same=false is identical, and assignable is true on parent and false on head. No paired same flag switches anywhere in the 38-site comparison.\n')
s=s.replace('See [carry validation]', 'See [carry validation]')
s=s.replace('## Finding stream validation\n','Evidence: [carry validation](carry-validation.json), [source validation](source-validation.json), [provenance](provenance.json), [site validation](site-validation.json), [parent stderr](parent-stderr.txt.gz), [head stderr](head-stderr.txt.gz), [parent filtered lines](parent-filtered-probes.txt.gz), [head filtered lines](head-filtered-probes.txt.gz), [parent index](parent-filtered-index.json), [head index](head-filtered-index.json), and [audit](audit.json). Preparation and each completed run have separate pushed checkpoint receipts in this directory. The final branch contains one directory-only commit directly on 1060360b.\n\n## Finding stream validation\n')
p.write_text(s);(out/'closing-paragraph.txt').write_text(summary+'\n')
print(summary)
