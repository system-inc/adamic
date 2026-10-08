"""Generic callable identities preserve actual values, rest and opaque inputs."""
import json,pathlib,subprocess,sys
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-generic-callable-controls');scratch.mkdir(exist_ok=True)
expected={'declarationRead':'conforms and ready (free)','arrowRead':'conforms and ready (free)','joinedRead':'conforms-if','restRead':'unknown','spreadRead':'unknown','hostRead':'unknown'}
results=[]
for name in ('clean','host'):
 work=scratch/name;(work/'src/compiler').mkdir(parents=True,exist_ok=True)
 entry=work/'src/compiler/main.a';entry.write_bytes((root/'stage3/shape-conformance/generic-callables'/(name+'.a')).read_bytes())
 sites=work/'sites.json';output=work/'result.json'
 assert subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(sites)]).returncode==0
 assert subprocess.run([sys.argv[1],str(work),str(sites),str(output)]).returncode==0
 result=json.loads(output.read_text());assert result['checker_rejected'] and result['analysis_only']
 for row in result['sites']:
  owner=row['adapted']['owner']
  if owner not in expected:continue
  assert row['outcome']==expected[owner],(owner,row)
  if owner=='joinedRead':assert any(f['name']=='ready' and f['declared']=='number' and f['expected']=='boolean' for f in row['fields']),row
  if owner=='restRead':assert any('callback rest arguments require an array allocation' in d for d in row['detail']),row
  if owner=='hostRead':assert row['reason']=='host metadata' and 'JSON.parse' in row['host_values'],row
  results.append({'owner':owner,'outcome':row['outcome'],'reason':row['reason'],'detail':row['detail']})
assert {r['owner'] for r in results}==set(expected)
print(json.dumps({'status':'PASS','controls':results},indent=2))
