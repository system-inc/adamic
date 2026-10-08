"""Independent frontend controls; census output remains analysis-only."""
import json,pathlib,subprocess,sys
root=pathlib.Path.cwd(); scratch=pathlib.Path('/tmp/shape-dynamic-controls');scratch.mkdir(exist_ok=True)
expected={'finiteString':'conforms and ready (free)','finiteNumber':'conforms and ready (free)','wrongString':'conforms-if','wrongNumber':'conforms-if','openKey':'unknown','absentKey':'unknown','hostKey':'unknown','indexedStore':'unknown','directStore':'unknown','diagnosedKey':'unknown'}
results=[]
for name in ('clean','host','mutation','diagnosed'):
 work=scratch/name; (work/'src/compiler').mkdir(parents=True,exist_ok=True)
 source=root/'stage3/shape-conformance/dynamic-keys/fixtures'/(name+'.a')
 entry=work/'src/compiler/main.a';entry.write_bytes(source.read_bytes())
 sites=work/'sites.json';output=work/'result.json'
 assert subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(sites)]).returncode==0
 assert subprocess.run([sys.argv[1],str(work),str(sites),str(output)]).returncode==0
 measured=json.loads(output.read_text());mapped=json.loads(sites.read_text())
 assert measured['checker_rejected'] and measured['analysis_only'] and measured['diagnostics']
 assert len(measured['sites'])==len(mapped)
 for row in measured['sites']:
  owner=row['adapted']['owner']; assert row['outcome']==expected[owner],(owner,row)
  if owner.startswith('wrong'):assert any(f['name']=='ready' and f['declared']=='number' and f['expected']=='boolean' for f in row['fields']),row
  if owner=='openKey':assert 'dynamic key membership not proven' in row['detail'],row
  if owner=='absentKey':assert any('dynamic projected slot absent' in d for d in row['detail']),row
  if owner=='hostKey':assert row['reason']=='host metadata' and 'JSON.parse' in row['host_values'],row
  if owner=='directStore':assert row['reason']=="flow the graph can't see" and any('mutation effects on field certificate' in d for d in row['detail']),row
  if owner=='diagnosedKey':assert row['reason']=='diagnosed body' and row['diagnostic_causes'],row
  if owner=='indexedStore':assert row['reason']=="flow the graph can't see" and 'indexed store effects not certified' in row['detail'],row
  results.append({'owner':owner,'outcome':row['outcome'],'reason':row['reason'],'detail':row['detail']})
assert {r['owner'] for r in results}==set(expected)
print(json.dumps({'status':'PASS','controls':results},indent=2))
