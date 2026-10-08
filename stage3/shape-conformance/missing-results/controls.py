"""Normal fallthrough produces undefined even when another path returns a record."""
import json,pathlib,subprocess,sys
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-missing-results-controls');(scratch/'src/compiler').mkdir(parents=True,exist_ok=True)
source=root/'stage3/shape-conformance/missing-results/clean.a'
entry=scratch/'src/compiler/main.a';entry.write_bytes(source.read_bytes())
reference=subprocess.run(['node','stage3/shape-conformance/suspended-callables/node.cjs',str(source)],capture_output=True,text=True,check=True)
observed=json.loads(reference.stdout)
positive={'completeIfRead','completeFinalRead','nestedReturnRead','completeThrowRead'}
assert len(observed)==11,observed
for owner,value in observed.items():assert value==({'missing':False,'ready':'boolean'} if owner in positive else {'missing':True,'ready':'absent'}),(owner,value)
sites=scratch/'map.json';output=scratch/'result.json'
assert subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(sites)]).returncode==0
assert subprocess.run([sys.argv[1],str(scratch),str(sites),str(output)]).returncode==0
result=json.loads(output.read_text());assert result['checker_rejected'] and result['analysis_only']
rows={r['adapted']['owner']:r for r in result['sites']};assert set(rows)==set(observed),rows.keys()
for owner,row in rows.items():
 if owner in positive:
  assert row['outcome']=='conforms and ready (free)',(owner,row)
 else:
  assert row['outcome']=='unknown' and row['reason']=="flow the graph can't see",(owner,row)
  cause='explicit return without a result' if owner=='explicitVoidRead' else 'function may complete without a result'
  assert any(cause in d for d in row['detail']),(owner,row)
print(json.dumps({'status':'PASS','node':observed,'controls':[{'owner':owner,'outcome':r['outcome'],'reason':r['reason'],'detail':r['detail']} for owner,r in rows.items()]},indent=2))
