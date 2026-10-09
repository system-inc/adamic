import pathlib,json,urllib.parse
R=pathlib.Path('/workspace/adamic');E=pathlib.Path('/tmp/d159/evidence');out={}
for name in ['TestPerformance','TestWholePerformance']:
 files={};functions={}
 for p in pathlib.Path('/tmp/d159/v8/'+name).glob('*.json'):
  for s in json.loads(p.read_text())['result']:
   file=pathlib.Path(urllib.parse.unquote(urllib.parse.urlparse(s['url']).path))
   if not str(file).startswith(str(R/'stage1/typescript/')) or not file.is_file():continue
   key=str(file.relative_to(R));text=file.read_text();ranges=[]
   for fn in s['functions']:
    ranges.extend(fn['ranges']);k=key+':'+fn['functionName'];functions[k]=functions.get(k,0)+fn['ranges'][0]['count']
   covered=files.setdefault(key,set());offset=0
   for num,l in enumerate(text.splitlines(True),1):
    point=offset+len(l)-len(l.lstrip());offset+=len(l)
    applicable=[ra for ra in ranges if ra['startOffset']<=point<ra['endOffset']]
    if applicable and min(applicable,key=lambda ra:ra['endOffset']-ra['startOffset'])['count']>0:covered.add(num)
 out[name]={'covered_lines':{f:sorted(ls) for f,ls in files.items()},'function_counts':functions}
(E/'port-v8-coverage.json').write_text(json.dumps(out,indent=2));diff=json.loads((E/'coverage-diffs.json').read_text())
for name,other in [('TestPerformance','TestWholePerformance'),('TestWholePerformance','TestPerformance')]:
 diff[name]['exclusive_port_lines']={f:sorted(set(ls)-set(out[other]['covered_lines'].get(f,[]))) for f,ls in out[name]['covered_lines'].items()};diff[name]['exclusive_port_lines']={f:ls for f,ls in diff[name]['exclusive_port_lines'].items() if ls}
 print(name,diff[name]['exclusive_port_lines']);print({f:c for f,c in out[name]['function_counts'].items() if f.endswith(':countTree') or f.endswith(':printTree')})
(E/'coverage-diffs.json').write_text(json.dumps(diff,indent=2))
