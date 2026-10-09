import pathlib,json,subprocess,bisect,gzip
p=pathlib.Path('/tmp/defend-mdast');profiles={};sources={}
for name in ['TestMdastIdentifierWitnesses','TestNativeMdastConstruction']:
 picked=[]
 for f in p.joinpath('v8',name).glob('*.json'):
  for r in json.loads(f.read_text())['result']:
   if r['url'].startswith('file:///workspace/adamic/stage1/cohere/markdownblocks/') and r['url'].endswith('.ts'):
    rel=r['url'][len('file:///workspace/adamic/'):];picked.append(r)
    if rel not in sources:sources[rel]=subprocess.check_output(['git','show','origin/main:'+rel],text=True)
 profiles[name]=picked
 with gzip.open(p/(name+'-port-v8.json.gz'),'wt') as f:json.dump(picked,f)
p.joinpath('port-sources.json').write_text(json.dumps(sources))
