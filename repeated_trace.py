import json,pathlib,subprocess,shutil,hashlib
out=pathlib.Path(__file__).parent;project=pathlib.Path('/workspace/scratch/tracing-witness-project'); observations=[];commands=[]
def norm(value):
 if isinstance(value,list):return [norm(v) for v in value]
 if isinstance(value,dict):return {k:('<timestamp>' if k in ['ts','dur'] else '<pid>' if k=='pid' else norm(v)) for k,v in value.items()}
 return value
source=pathlib.Path('/workspace/scratch/tracing-control/src/compiler/tracing.ts').read_text()
assert source.count('writeEvent("X"')==1
assert 'sampleInterval - (time % sampleInterval) <= endTime - time' in source
for label in ['control','control','after-final','after-final']:
 index=sum(r['tree']==label for r in observations)+1;prefix=label+'-repeat'+str(index)
 for name in ['trace','out']:
  target=project/name
  if target.exists():shutil.rmtree(target)
 args=['node','/workspace/scratch/tracing-'+label+'/built/local/tsc.js','--project',str(project/'tsconfig.json'),'--pretty','false','--generateTrace',str(project/'trace')]
 child=subprocess.run(args,cwd=project,capture_output=True);commands.append(dict(name=prefix,argv=args,exit=child.returncode));(out/'repeated-trace-commands.json').write_text(json.dumps(commands,indent=2)+'\n');print(prefix,'child exit='+str(child.returncode),flush=True)
 if child.returncode:raise SystemExit(child.returncode)
 target=out/prefix;target.mkdir();(target/'stdout').write_bytes(child.stdout);(target/'stderr').write_bytes(child.stderr)
 traces={p.name:p.read_bytes() for p in (project/'trace').iterdir() if p.is_file()}
 for name,data in traces.items():(target/name).write_bytes(data)
 assert set(traces)=={'types.json','legend.json','trace.json'},set(traces)
 events=json.loads(traces['trace.json']);filtered=[e for e in events if e['ph']!='X']
 observations.append(dict(tree=label,index=index,traces=traces,events=events,filtered=norm(filtered),stdout=child.stdout,stderr=child.stderr,emitted={p.name:p.read_bytes() for p in (project/'out').iterdir() if p.is_file()}))
a=observations[0];b=observations[1]
control_different=norm(a['events'])!=norm(b['events'])
if not control_different and any(norm(o['events'])!=norm(a['events']) for o in observations[2:]):raise AssertionError('control self-comparison identical but after differs')
for o in observations[1:]:
 for name in ['types.json','legend.json']:assert o['traces'][name]==a['traces'][name],(name,o['tree'],o['index'])
 assert o['filtered']==a['filtered'],('deterministic inventory changed',o['tree'],o['index'])
 for name in ['stdout','stderr','emitted']:assert o[name]==a[name],(name,o['tree'],o['index'])
rows=[dict(tree=o['tree'],index=o['index'],traceEvents=len(o['events']),sampledEvents=sum(e['ph']=='X' for e in o['events']),retainedEvents=len(o['filtered']),sampledNames=sorted({e['name'] for e in o['events'] if e['ph']=='X'}),typesSha256=hashlib.sha256(o['traces']['types.json']).hexdigest(),legendSha256=hashlib.sha256(o['traces']['legend.json']).hexdigest()) for o in observations]
(out/'repeated-trace.json').write_text(json.dumps(dict(controlSelfInventoryDiffers=control_different,samplingRule='writeStackEvent emits ph X complete events only when [time,endTime) crosses a 10 ms sampling point',removedKinds=['X'],deterministicInventoryEqual=True,typesAndLegendByteIdentical=True,runs=rows),indent=2)+'\n')
