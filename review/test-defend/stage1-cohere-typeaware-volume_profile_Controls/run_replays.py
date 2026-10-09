import pathlib,json,subprocess,os,time,difflib
p=pathlib.Path(__file__).parent
plan=json.loads((p/'plan.json').read_text()); (p/'diffs').mkdir(exist_ok=True)
results=json.loads((p/"matrix.json").read_text())
regex='^(TestVolumeProfileControls(Union|_[0-9]+)|TestVolumeProfileCorpora(Union|_0[0-5][0-9]|_06[0-3])|TestVolumeAgreementAndMutants(Union|_014|_022|_030))$'
for m in plan[:3]:
 f=pathlib.Path(m['file']);original=f.read_text();offset=original.index('func tsgo_go_inspect') if m['id']=='D4' else 0;at=original.index(m['from'],offset);mutated=original[:at]+original[at:].replace(m['from'],m['to'],1)
 (p/'diffs'/f"{m['id']}.diff").write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile='a/'+str(f),tofile='b/'+str(f))))
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-typeaware-volume/cache/'+m['id'];env['ADAMIC_VOLUME_REPOSITORY_MANIFEST']=str((p/'repository.manifest').resolve());f.write_text(mutated)
 try:
  scopes=[('corpora-replay','^TestVolumeProfileCorpora(Union|_0[0-5][0-9]|_06[0-3])$'),('other-corpus','^TestVolumeAgreementRepository_[0-9]+$'),('controls-replay','^TestVolumeProfileControls(Union|_[0-9]+)$'),('agreement-replay','^TestVolumeAgreementAndMutants(Union|_014|_022|_030)$')]
  if m['id']=='D4':
   with (p/'logs'/'D4-vet.log').open('w') as log:vet=subprocess.run(['go','vet','./bridge/tsgo/archive/'],stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  else:vet=None
  for label,scope in scopes:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run',scope];logpath=p/'logs'/f"{m['id']}-{label}.log";start=time.monotonic()
   with logpath.open('w') as log:rc=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env).returncode
   events=[]
   for s in logpath.read_text().splitlines():
    try:events.append(json.loads(s))
    except:pass
   failed=[e['Test'] for e in events if e['Action']=='fail' and e.get('Test') and '/' not in e['Test']];passed=[e['Test'] for e in events if e['Action']=='pass' and e.get('Test') and '/' not in e['Test']];cooked=any('test timed out' in e.get('Output','') for e in events)
   result={'mutant':m['id'],'scope':label,'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'exit':rc,'wall':time.monotonic()-start,'vet':vet,'cooked':cooked,'failed':failed,'passed':passed,'failing_lines':[e['Output'].strip() for e in events if '.go:' in e.get('Output','') and any(s in e['Output'] for s in ['mismatch','released program escaped','escaped','not caught','failed'])]};results.append(result);(p/'matrix.json').write_text(json.dumps(results,indent=2));print(m['id'],label,rc,round(result['wall'],2),failed,flush=True)
 finally:f.write_text(original)
 with (p/'logs'/f"{m['id']}-apply.log").open('w') as log:subprocess.run(['git','apply','--check',str(p/'diffs'/f"{m['id']}.diff")],stdout=log,stderr=subprocess.STDOUT,check=True)
