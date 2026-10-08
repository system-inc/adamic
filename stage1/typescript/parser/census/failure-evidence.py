"""Preserve shortest input and Go tree for each parser failure class."""
import json,pathlib,subprocess,gzip,collections
base=pathlib.Path(__file__).resolve().parent
replay=json.loads((base/'failure-replay.json').read_text());groups={}
for row in replay['failures']:
 for failure in row['failures']:
  error=failure['error'];label=error.split(' in ')[0].split(' at ')[0]
  g=groups.setdefault(label,{'files':0,'shortest_bytes':None});g['files']+=1
  size=pathlib.Path(row['path']).stat().st_size
  if g['shortest_bytes'] is None or size<g['shortest_bytes']:g.update(shortest_bytes=size,path=row['path'],error=error)
folder=base/'failures/examples';folder.mkdir(parents=True,exist_ok=True)
for i,(label,g) in enumerate(groups.items()):
 p=pathlib.Path(g['path']);stem=f'failure-{i}';g['artifact']=stem
 (folder/(stem+p.suffix)).write_bytes(p.read_bytes())
 result=subprocess.run(['/tmp/parser-census-oracle',str(p),'--whole','--recovery'],capture_output=True,text=True,timeout=30)
 with gzip.open(folder/(stem+'.go.tree.gz'),'wt') as out:out.write(result.stdout)
 (folder/(stem+'.port.txt')).write_text(g['error']+'\n')
 g['go_exit']=result.returncode;g['go_stderr']=result.stderr
(base/'failure-classes.json').write_text(json.dumps(groups,indent=2)+'\n')
