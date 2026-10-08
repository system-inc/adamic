"""A non-null syntax wrapper preserves finite identities, not an unknown target proof."""
import json,pathlib,subprocess,sys
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-nonnull-callables-controls');(scratch/'src/compiler').mkdir(parents=True,exist_ok=True)
source=root/'stage3/shape-conformance/nonnull-callables/clean.a';entry=scratch/'src/compiler/main.a';entry.write_bytes(source.read_bytes())
node=subprocess.run(['node','stage3/shape-conformance/nonnull-callables/node.cjs',str(source)],capture_output=True,text=True,check=True)
expected=['boolean']*4+['number']+['boolean']*4;assert json.loads(node.stdout)==[expected,expected],node.stdout
mapped=scratch/'map.json';output=scratch/'result.json'
subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(mapped)],check=True)
subprocess.run([sys.argv[1],str(scratch),str(mapped),str(output)],check=True)
r=json.loads(output.read_text());assert r['analysis_only'] and r['checker_rejected'] and len(r['diagnostics'])==1 and 'TS2322' in r['diagnostics'][0]
rows={row['adapted']['owner']:row for row in r['sites']};assert len(rows)==10
for owner,row in rows.items():
 if owner in ['aliasIdentity','directIdentity','passedIdentity','joinedOne','joinedTwo','nullableIdentity']:assert row['outcome']=='conforms and ready (free)',(owner,row)
 elif owner=='wrongIdentity':assert row['outcome']=='conforms-if' and any(f['name']=='ready' and f['declared']=='number' for f in row['fields']),(owner,row)
 elif owner=='hostIdentity':assert row['outcome']=='unknown' and row['reason']=='host metadata',(owner,row)
 else:assert row['outcome']=='unknown' and row['reason']=="flow the graph can't see",(owner,row)
print(json.dumps({'status':'PASS','node':json.loads(node.stdout),'controls':list(rows.values())},indent=2))
