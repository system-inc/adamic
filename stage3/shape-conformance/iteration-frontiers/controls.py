"""Iterator values and enumeration keys are distinct named allocation frontiers."""
import json,pathlib,subprocess,sys
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-iteration-frontiers-controls');(scratch/'src/compiler').mkdir(parents=True,exist_ok=True)
source=root/'stage3/shape-conformance/iteration-frontiers/clean.a';entry=scratch/'src/compiler/main.a';entry.write_bytes(source.read_bytes())
node=subprocess.run(['node','stage3/shape-conformance/iteration-frontiers/node.cjs',str(source)],capture_output=True,text=True,check=True)
assert json.loads(node.stdout)==['boolean']*4+['undefined','missing','boolean'],node.stdout
mapped=scratch/'map.json';output=scratch/'result.json'
subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(mapped)],check=True)
subprocess.run([sys.argv[1],str(scratch),str(mapped),str(output)],check=True)
r=json.loads(output.read_text());assert r['analysis_only'] and r['checker_rejected'] and len(r['diagnostics'])==1 and 'TS2322' in r['diagnostics'][0]
rows={row['adapted']['owner']:row for row in r['sites']};assert len(rows)==7
for owner,row in rows.items():
 if owner=='initializedRead':assert row['outcome']=='conforms and ready (free)',(owner,row)
 else:
  assert row['outcome']=='unknown',(owner,row)
  cause='for-in binding requires a key producer' if owner=='keyIterationRead' else 'for-of binding requires an iterator value allocation'
  assert any(cause in d for d in row['detail']),(owner,row)
  if owner=='hostIterationRead':assert row['reason']=='host metadata',(owner,row)
print(json.dumps({'status':'PASS','node':json.loads(node.stdout),'controls':list(rows.values())},indent=2))
