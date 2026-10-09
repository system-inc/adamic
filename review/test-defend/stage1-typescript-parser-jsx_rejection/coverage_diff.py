import pathlib,json,urllib.parse
E=pathlib.Path('/tmp/d159/evidence');R=pathlib.Path('/workspace/adamic');names=['TestCompilerExpressionsAgree_Setup','TestPerformance','TestWholePerformance']
def gocov(name):
 p=E/(name+'.cover');d={}
 for l in p.read_text().splitlines()[1:]:
  span,stm,count=l.split();file,ran=span.rsplit(':',1)
  if int(count)>0:d[span]=int(stm)
 return d
out={}
for name,other in [('TestPerformance','TestWholePerformance'),('TestWholePerformance','TestPerformance'),('TestCompilerExpressionsAgree_Setup','Setup-subsumer')]:
 a=gocov(name);b=gocov(other);out[name]=dict(compared_to=other,exclusive_go_blocks=sorted(a.keys()-b.keys()),shared_go_blocks=len(a.keys()&b.keys()))
v8={}
for name in names:
 script={};counts={}
 for p in pathlib.Path('/tmp/d159/v8/'+name).glob('*.json'):
  for s in json.loads(p.read_text())['result']:
   url=s['url'];file=pathlib.Path(urllib.parse.unquote(urllib.parse.urlparse(url).path))
   if not str(file).startswith(str(R/'stage1/typescript/')) or not file.is_file():continue
   text=file.read_text();covered=script.setdefault(str(file.relative_to(R)),set())
   for fn in s['functions']:
    ranges=fn['ranges'];k=str(file.relative_to(R))+':'+fn['functionName'];counts[k]=counts.get(k,0)+ranges[0]['count']
    points=[]
    for ra in ranges:
     points.extend([ra['startOffset'],ra['endOffset']])
    points=sorted(set(points))
    for left,right in zip(points,points[1:]):
     applicable=[ra for ra in ranges if ra['startOffset']<=left and ra['endOffset']>=right]
     if not applicable:continue
     if min(applicable,key=lambda ra:ra['endOffset']-ra['startOffset'])['count']<=0:continue
     start=text[:left].count('\n')+1;end=text[:max(left,right-1)].count('\n')+1;covered.update(range(start,end+1))
 v8[name]={'covered_lines':{file:sorted(lines) for file,lines in script.items()},'function_counts':counts}
for name,other in [('TestPerformance','TestWholePerformance'),('TestWholePerformance','TestPerformance')]:
 out[name]['exclusive_port_lines']={f:sorted(set(ls)-set(v8[other]['covered_lines'].get(f,[]))) for f,ls in v8[name]['covered_lines'].items()};out[name]['exclusive_port_lines']={f:ls for f,ls in out[name]['exclusive_port_lines'].items() if ls}
(E/'coverage-diffs.json').write_text(json.dumps(out,indent=2));(E/'port-v8-coverage.json').write_text(json.dumps(v8,indent=2));print(json.dumps(out,indent=2)[:6000])
