from pathlib import Path
import json,re
out=Path('/workspace/adamic/stage3/census/for-each-ancestor');summary={}
for name in ['baseline',*'ABCD']:
 rows=[json.loads(s) for s in (out/(name+'-census.jsonl')).read_text().splitlines()];records=[r for r in rows if 'file' in r];assert len(records)==1 and records[0]['file'].endswith('/src/compiler/path.ts')
 source=(out/(name+'-path.ts.txt')).read_text();ranges=[]
 for m in re.finditer(r'export function (forEachAncestorDirectory(?:Path)?)<',source):
  line=source.count('\n',0,m.start())+1;end=source.find('\n}',m.start());signature_end=source.find('\n',m.start());has_body='{' in source[m.start():signature_end]
  ranges.append({'name':m[1],'start':line,'end':source.count('\n',0,end)+2 if has_body else line,'implementation':has_body})
 def contains(where):
  line=int(where.rsplit(':',2)[1]);return any(r['start']<=line<=r['end'] for r in ranges)
 f=[f for f in records[0]['findings'] if contains(f['where'])];u=[u for u in records[0]['units'] if contains(u['where'])]
 unique=[];seen=set()
 for x in f:
  k=(x['kind'],x['where'],x['reason'])
  if k not in seen:seen.add(k);unique.append(x)
 summary[name]={'header':rows[0],'ranges':ranges,'findings':unique,'units':u,'file_findings_raw':len(records[0]['findings'])}
(out/'census-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
for n,s in summary.items():print(n,'diagnostics',len(s['header']['diagnostics']),'findings',[(f['kind'],f['where'].rsplit('/',1)[-1],f['reason']) for f in s['findings']],'units',[(u['name'],u['status'])for u in s['units']])
