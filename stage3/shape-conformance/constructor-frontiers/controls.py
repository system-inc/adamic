"""Constructor dependencies never substitute argument shapes for the result."""
import json,pathlib,subprocess,sys
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-constructor-controls');scratch.mkdir(exist_ok=True)
expected={'plainRead':"flow the graph can't see",'argumentRead':"flow the graph can't see",'hostConstructorRead':'host metadata','hostArgumentRead':'host metadata','diagnosedConstructorRead':'diagnosed body','diagnosedArgumentRead':'diagnosed body'}
results=[]
for name in ('clean','host','diagnosed'):
 work=scratch/name;(work/'src/compiler').mkdir(parents=True,exist_ok=True)
 entry=work/'src/compiler/main.a';entry.write_bytes((root/'stage3/shape-conformance/constructor-frontiers'/(name+'.a')).read_bytes())
 sites=work/'sites.json';output=work/'result.json'
 assert subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(sites)]).returncode==0
 assert subprocess.run([sys.argv[1],str(work),str(sites),str(output)]).returncode==0
 result=json.loads(output.read_text());assert result['checker_rejected'] and result['analysis_only']
 for row in result['sites']:
  owner=row['adapted']['owner']
  if owner not in expected:continue
  assert row['outcome']=='unknown' and row['reason']==expected[owner],(owner,row)
  assert any('constructor allocation and initialization body not modeled' in d for d in row['detail']),row
  if owner.startswith('host'):assert 'JSON.parse' in row['host_values'],row
  if owner.startswith('diagnosed'):assert row['diagnostic_causes'] and all(c['scope']=='dependency' for c in row['diagnostic_causes']),row
  results.append({'owner':owner,'outcome':row['outcome'],'reason':row['reason'],'detail':row['detail']})
assert {r['owner'] for r in results}==set(expected)
print(json.dumps({'status':'PASS','controls':results},indent=2))
