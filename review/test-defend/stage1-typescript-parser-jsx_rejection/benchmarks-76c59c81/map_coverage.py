import pathlib,json,urllib.parse
P=pathlib.Path('/workspace/adamic/review/test-defend/stage1-typescript-parser-jsx_rejection/benchmarks-76c59c81');alpha='ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/'
def vlq(s):
 out=[];n=shift=0
 for c in s:
  d=alpha.index(c);n|=(d&31)<<shift
  if d&32:shift+=5
  else:out.append(-(n>>1) if n&1 else n>>1);n=shift=0
 return out
cov={}
for name in ['TestPerformance','TestWholePerformance']:
 seen={}
 for f in (P/'v8-raw'/name).glob('*.json'):
  for script in json.loads(f.read_text())['result']:
   path=urllib.parse.unquote(urllib.parse.urlparse(script['url']).path)
   if not path.startswith('/workspace/adamic/stage1/typescript/') or not path.endswith('.ts'):continue
   relative=path.removeprefix('/workspace/adamic/');m=json.loads((P/'source-maps'/(relative.replace('/','__')+'.json')).read_text());ranges=[r for fn in script['functions'] for r in fn['ranges']];src=origline=origcol=ni=0;offset=0;lines=m['transformed'].splitlines(True);covered=seen.setdefault(relative,set())
   for i,line in enumerate(m['map']['mappings'].split(';')):
    gen=0
    for segment in line.split(','):
     if not segment:continue
     ds=vlq(segment);gen+=ds[0]
     if len(ds)<4:continue
     src+=ds[1];origline+=ds[2];origcol+=ds[3]
     if len(ds)>4:ni+=ds[4]
     active=[r for r in ranges if r['startOffset']<=offset+gen<r['endOffset']]
     if active and min(active,key=lambda r:r['endOffset']-r['startOffset'])['count']>0:covered.add(origline+1)
    if i<len(lines):offset+=len(lines[i].encode('utf-16-le'))//2
 cov[name]={f:sorted(ls) for f,ls in seen.items()}
v8=json.loads((P/'port-v8-coverage.json').read_text());diff=json.loads((P/'coverage-diffs.json').read_text())
for name,other in [('TestPerformance','TestWholePerformance'),('TestWholePerformance','TestPerformance')]:
 v8[name]['covered_lines']=cov[name];diff[name]['exclusive_port_lines']={f:sorted(set(ls)-set(cov[other].get(f,[]))) for f,ls in cov[name].items()};diff[name]['exclusive_port_lines']={f:ls for f,ls in diff[name]['exclusive_port_lines'].items() if ls};print(name,diff[name]['exclusive_port_lines'])
v8['mapping_method']='Source-map VLQ segments, UTF-16 generated offsets, innermost V8 range; original source line union. Initial unmapped line estimates are superseded.'
(P/'port-v8-coverage.json').write_text(json.dumps(v8,indent=2));(P/'coverage-diffs.json').write_text(json.dumps(diff,indent=2))
s=(P/'REPORT.md').read_text();s+='\nV8 source-offset correction: the runner uses stripTypeScriptTypes(mode=transform), so direct original-source line offsets were only preliminary estimates. Final coverage-diffs.json and port-v8-coverage.json map generated UTF-16 offsets through source maps. maps.mjs, map_coverage.py and source-maps preserve the mapping evidence. The semantic counter defenses and function-call counts are unchanged. This correction cost extra packaging time.\n';(P/'REPORT.md').write_text(s)
