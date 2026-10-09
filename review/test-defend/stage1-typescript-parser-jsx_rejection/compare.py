import pathlib,json
p=pathlib.Path('/workspace/adamic/review/test-defend/stage1-typescript-parser-jsx_rejection');names=['TestJsxMemberNameRejection','TestJsxNode','TestJsxNative'];go={};v8={}
for n in names:
 go[n]={l.split()[0] for l in (p/(n+'.cover')).read_text().splitlines()[1:] if int(l.split()[-1])>0}
 ranges={}
 for f in pathlib.Path('/workspace/scratch/defend-jsx/v8/'+n).glob('*.json'):
  d=json.loads(f.read_text())
  for s in d['result']:
   url=s['url']
   if '/stage1/typescript/parser/' not in url or not url.endswith('.ts'):continue
   if '/workspace/adamic/stage1/' not in url:continue
   rs=ranges.setdefault(url,[])
   for fn in s['functions']:rs.extend(fn['ranges'])
 v8[n]=ranges
 (p/(n+'-v8.json')).write_text(json.dumps(ranges,indent=2))
result=[]
for a,b in [(names[0],names[1]),(names[1],names[2]),(names[2],names[1])]:
 result.append(dict(row=a,subsumer=b,go_exclusive_blocks=sorted(go[a]-go[b]),go_shared_blocks=len(go[a]&go[b])))
(p/'coverage-diffs.json').write_text(json.dumps(result,indent=2));print([(r['row'],len(r['go_exclusive_blocks']),r['go_shared_blocks']) for r in result]);print([(n,len(v8[n])) for n in names])
