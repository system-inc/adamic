"""A declaration without an initializer is a named missing-value producer."""
import json,pathlib,subprocess,sys
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-missing-initializers-controls');(scratch/'src/compiler').mkdir(parents=True,exist_ok=True)
source=root/'stage3/shape-conformance/missing-initializers/clean.a';entry=scratch/'src/compiler/main.a';entry.write_bytes(source.read_bytes())
node=subprocess.run(['node','stage3/shape-conformance/suspended-callables/node.cjs',str(source)],capture_output=True,text=True,check=True)
assert json.loads(node.stdout)==['missing','missing','boolean','boolean'],node.stdout
mapped=scratch/'map.json';output=scratch/'result.json'
subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(mapped)],check=True)
subprocess.run([sys.argv[1],str(scratch),str(mapped),str(output)],check=True)
r=json.loads(output.read_text());assert r['checker_rejected'] and r['analysis_only']
rows={row['adapted']['owner']:row for row in r['sites']};assert len(rows)==4
for owner in ["assignedRead","absentRead","conditionalRead","initializedRead"]:
 row=rows[owner]
 if owner=='initializedRead':assert row['outcome']=='conforms and ready (free)',(owner,row)
 else:
  assert row['outcome']=='unknown',(owner,row)
  assert any('variable declaration without an initializer at ' in d for d in row['detail']),(owner,row)
  if owner=='assignedRead':assert row['allocation_sites'],(owner,row)
print(json.dumps({'status':'PASS','node':json.loads(node.stdout),'controls':list(rows.values())},indent=2))
