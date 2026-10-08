"""Node wrappers are different allocations from the value returned by their body."""
import json,pathlib,subprocess,sys
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-suspended-controls');(scratch/'src/compiler').mkdir(parents=True,exist_ok=True)
source=root/'stage3/shape-conformance/suspended-callables/clean.a'
entry=scratch/'src/compiler/main.a';entry.write_bytes(source.read_bytes())
reference=subprocess.run(['node','stage3/shape-conformance/suspended-callables/node.cjs',str(source)],capture_output=True,text=True,check=True)
observed=json.loads(reference.stdout)
assert len(observed)==12,observed
for owner,value in observed.items():
 if owner=='synchronousAsyncNameRead':
  assert value=={'tag':'[object Object]','readyMissing':False},(owner,value)
  continue
 expected='[object Generator]' if 'generator' in owner.lower() else '[object Promise]'
 assert value=={'tag':expected,'readyMissing':True},(owner,value)
sites=scratch/'map.json';output=scratch/'result.json'
assert subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(sites)]).returncode==0
assert subprocess.run([sys.argv[1],str(scratch),str(sites),str(output)]).returncode==0
result=json.loads(output.read_text());assert result['checker_rejected'] and result['analysis_only']
rows={r['adapted']['owner']:r for r in result['sites']};assert set(rows)==set(observed),rows.keys()
for owner,row in rows.items():
 if owner=='synchronousAsyncNameRead':
  assert row['outcome']=='conforms and ready (free)',(owner,row)
  continue
 assert row['outcome']=='unknown',(owner,row)
 if owner.startswith('diagnosed'):
  assert row['reason']=='diagnosed body' and row['diagnostic_causes'],(owner,row)
 else:
  assert row['reason']=="flow the graph can't see",(owner,row)
  protocol='iterator allocation' if 'generator' in owner.lower() else 'Promise allocation'
  assert any(protocol in d for d in row['detail']),(owner,row)
print(json.dumps({'status':'PASS','node':observed,'controls':[{'owner':owner,'outcome':r['outcome'],'reason':r['reason'],'detail':r['detail']} for owner,r in rows.items()]},indent=2))
