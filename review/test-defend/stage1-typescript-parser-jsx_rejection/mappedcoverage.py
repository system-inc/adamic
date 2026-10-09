import json,pathlib,urllib.parse
p=pathlib.Path('/workspace/adamic/review/test-defend/stage1-typescript-parser-jsx_rejection');alpha='ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/'
def vlq(s):
 out=[];n=shift=0
 for c in s:
  d=alpha.index(c);n|=(d&31)<<shift
  if d&32:shift+=5
  else:out.append(-(n>>1) if n&1 else n>>1);n=shift=0
 return out
cov={}
for name in ['TestJsxMemberNameRejection','TestJsxNode']:
 seen=set()
 for f in pathlib.Path('/workspace/scratch/defend-jsx/v8/'+name).glob('*.json'):
  for script in json.loads(f.read_text())['result']:
   path=urllib.parse.unquote(script['url'].removeprefix('file://'))
   if not path.startswith('/workspace/adamic/stage1/typescript/parser/') or not path.endswith('.ts'):continue
   m=json.loads((p/(pathlib.Path(path).name+'.map.json')).read_text());ranges=[r for fn in script['functions'] for r in fn['ranges']];sourceindex=origline=origcol=nameindex=0;offset=0;lines=m['transformed'].splitlines(True)
   for i,line in enumerate(m['map']['mappings'].split(';')):
    gen=0
    for segment in line.split(','):
     if not segment:continue
     ds=vlq(segment);gen+=ds[0]
     if len(ds)<4:continue
     sourceindex+=ds[1];origline+=ds[2];origcol+=ds[3]
     if len(ds)>4:nameindex+=ds[4]
     active=[r for r in ranges if r['startOffset']<=offset+gen<r['endOffset']]
     if active and min(active,key=lambda r:r['endOffset']-r['startOffset'])['count']>0:seen.add(path.removeprefix('/workspace/adamic/')+':'+str(origline+1))
    if i<len(lines):offset+=len(lines[i].encode('utf-16-le'))//2
 cov[name]=seen
out=dict(member_only=sorted(cov['TestJsxMemberNameRejection']-cov['TestJsxNode']),node_only=sorted(cov['TestJsxNode']-cov['TestJsxMemberNameRejection']),member_covered=sorted(cov['TestJsxMemberNameRejection']),node_covered=sorted(cov['TestJsxNode']),method='Node stripTypeScriptTypes transform source-map mappings, UTF-16 V8 offsets, smallest containing range count, union of original-port runs. Profiles exclude copied built-in mutant paths. Maps are generated from clean HEAD using git show, not the mutated working tree. Native unavailable through V8.')
(p/'mapped-v8-coverage.json').write_text(json.dumps(out,indent=2));print(out['member_only'])
