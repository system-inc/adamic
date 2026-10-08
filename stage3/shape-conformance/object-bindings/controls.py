"""Object binding values come from the property key, never the bound name."""
import json,pathlib,subprocess,sys
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-object-bindings-controls');(scratch/'src/compiler').mkdir(parents=True,exist_ok=True)
source=root/'stage3/shape-conformance/object-bindings/clean.a';entry=scratch/'src/compiler/main.a';entry.write_bytes(source.read_bytes())
node=subprocess.run(['node','stage3/shape-conformance/object-bindings/node.cjs',str(source),'shorthandRead','renamedRead','stringRead','numericRead','wrongRead','defaultRead','restRead','arrayRead','nestedRead','computedRead'],capture_output=True,text=True,check=True)
assert json.loads(node.stdout)==['boolean']*4+['number']+['boolean']*5,node.stdout
mapped=scratch/'map.json';output=scratch/'result.json'
subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(mapped)],check=True)
subprocess.run([sys.argv[1],str(scratch),str(mapped),str(output)],check=True)
r=json.loads(output.read_text());assert r['checker_rejected'] and r['analysis_only'] and len(r['diagnostics'])==1 and 'TS2322' in r['diagnostics'][0]
rows={row['adapted']['owner']:row for row in r['sites']};assert len(rows)==10
causes={'defaultRead':'default binding selection','restRead':'rest binding allocation','arrayRead':'array binding iteration','nestedRead':'nested binding allocation','computedRead':'computed binding key'}
for owner,row in rows.items():
 if owner in ['shorthandRead','renamedRead','stringRead','numericRead']:assert row['outcome']=='conforms and ready (free)',(owner,row)
 elif owner=='wrongRead':assert row['outcome']=='conforms-if' and any(f['name']=='ready' for f in row['fields']),(owner,row)
 elif owner=='hostRead':assert row['outcome']=='unknown' and row['reason']=='host metadata',(owner,row)
 else:assert row['outcome']=='unknown' and any(causes[owner] in d for d in row['detail']),(owner,row)
clean_rows=list(rows.values())
entry.write_bytes((source.parent/'host.a').read_bytes())
subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(mapped)],check=True)
subprocess.run([sys.argv[1],str(scratch),str(mapped),str(output)],check=True)
host=json.loads(output.read_text());assert len(host['sites'])==6 and len(host['diagnostics'])==1
for host_row in host['sites']:
 if host_row['adapted']['owner']=='opaqueRead':
  assert host_row['outcome']=='unknown' and any('opaque call may mutate projected fields' in d for d in host_row['detail']),host_row
 else:assert host_row['outcome']=='unknown' and host_row['reason']=='host metadata',(host_row['adapted']['owner'],host_row)
reference=subprocess.run(['node','stage3/shape-conformance/object-bindings/node.cjs',str(source.parent/'host.a'),'hostRead','defaultHostRead','restHostRead','arrayHostRead','nestedHostRead','opaqueRead'],capture_output=True,text=True,check=True)
assert json.loads(reference.stdout)==['boolean']*6
print(json.dumps({'status':'PASS','node':json.loads(node.stdout)+json.loads(reference.stdout),'controls':clean_rows+host['sites']},indent=2))
