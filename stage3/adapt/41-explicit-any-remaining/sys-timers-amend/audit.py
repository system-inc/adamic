import pathlib,json,hashlib,os,subprocess
ROOT=pathlib.Path('/workspace/adamic');S=pathlib.Path('/workspace/scratch/sys-timer-stack');out=pathlib.Path(__file__).parent
scope=json.loads((out/'source-scope.json').read_text());assert scope['sourceFiles']==746 and scope['changed']==['src/compiler/sys.ts'] and scope['actualFullApplyEqualsCensusCandidate']
for name,digest in scope['after'].items():assert hashlib.sha256((S/'head1'/name).read_bytes()).hexdigest()==digest,name
artifacts=json.loads((out/'artifacts.json').read_text());assert artifacts['javascriptCount']==9 and artifacts['allJavascriptIdentical'];assert all(x['identical'] for x in artifacts['declarations'].values())
comparison=json.loads((out/'census-comparison.json').read_text());assert comparison['head1']['pass'] and comparison['head1']['mutantRestoresEntireBaseline'];assert len(comparison['head1']['expected995spev'])==1
assert not comparison['head2']['pass'];assert len(comparison['head2']['new'])==8 and len(comparison['head2']['unexpected'])==1;assert comparison['head2']['unexpected'][0]['where']=='src/compiler/sys.ts:1519:36';assert comparison['head2']['mutantRestores384']
assert not (ROOT/'stage3/adapt/99-correctness-core-sys-timer-slot').exists()
assert subprocess.check_output(['git','diff','--name-only','--','stage3/lane','stage3/upstream/LEDGER.md'],cwd=ROOT,text=True)==''
ancestor=set();pid=os.getpid()
while pid>1 and pid not in ancestor:
 ancestor.add(pid)
 try:pid=int(pathlib.Path(f'/proc/{pid}/stat').read_text().split(') ')[1].split()[1])
 except (FileNotFoundError,ValueError):break
live=[]
for p in pathlib.Path('/proc').iterdir():
 if not p.name.isdigit() or int(p.name) in ancestor:continue
 try:state=(p/'stat').read_text().split(') ')[1].split()[0];cmd=(p/'cmdline').read_bytes().replace(b'\0',b' ').decode(errors='replace');cwd=os.readlink(p/'cwd')
 except (FileNotFoundError,PermissionError,ProcessLookupError):continue
 if state!='Z' and (str(S) in cmd or cwd.startswith(str(S))):live.append({'pid':int(p.name),'command':cmd})
assert not live,live
for p in out.rglob('*'):assert p.suffix not in ['.go','.pyc'],str(p)
result={'head1RequiredProofs':True,'sourceScope':True,'nineJavascriptAndDeclarationsIdentical':True,'expected995spevRetained':True,'ambientMutantRestoresAllSites':True,'head2GateFailedAndNotConstructed':True,'head2AnyMutantRestores384':True,'liveTaskProcesses':live}
(out/'audit.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
