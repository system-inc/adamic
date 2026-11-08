from pathlib import Path
import sys,os,json,subprocess,hashlib,shutil,tarfile,time,signal
sys.dont_write_bytecode=True
from run import run,s
repo=Path('/workspace/adamic');mode=sys.argv[1]
if mode=='build':
 side=sys.argv[2];tree=s/side
 if not (tree/'node_modules').exists():(tree/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
 run(side+'-normal-build',['npm','run','build'],cwd=tree)
elif mode=='types':
 side=sys.argv[2];roots=sorted((s/side/'src/compiler').rglob('*.ts'));assert len(roots)==79
 run(side+'-types79',[s/'adamic','types',*roots])
elif mode=='census':
 side,file=sys.argv[2:4];tree=s/side;env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler'/file),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(s/'raw'/(side+'-'+file+'-times.jsonl')))
 run(side+'-'+file+'-census',['/workspace/scratch/core-cast/latent',tree/'src/compiler',s/'raw'/(side+'-'+file+'-census.jsonl')],env=env)
elif mode=='checkpoint':
 side= sys.argv[2];r=Path(sys.argv[3]);branch=sys.argv[4];directory=r/'stage3/adapt'/('99-correctness-core-getfunctionname-move' if side=='head2' else '99-correctness-core-cast-message');directory.mkdir(exist_ok=True)
 result=json.loads((s/'raw'/(side+'-full-apply-result.json')).read_text());assert result['exit']==0
 manifest={str(p.relative_to(s/side)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((s/side/'src').rglob('*')) if p.is_file()};digest=hashlib.sha256(json.dumps(manifest,sort_keys=True).encode()).hexdigest();entry={'side':side,'apply':result,'sourceFiles':len(manifest),'sourceManifestSha256':digest,'status':'Apply complete, subsequent proof pending'};(directory/(side+'-apply-checkpoint.json')).write_text(json.dumps(entry,indent=2)+'\n');(s/'raw'/(side+'-source-manifest.json')).write_text(json.dumps(manifest,sort_keys=True)+'\n')
 env=dict(os.environ,GIT_INDEX_FILE=str(s/(side+'-checkpoint.index')));head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=r,text=True).strip()
 args=[['git','read-tree',head]]
 allowed=['stage3/adapt/99-correctness-core-cast-message','stage3/upstream/LEDGER.md'] if side!='head2' else ['stage3/adapt/30-indexed-reads','stage3/adapt/41-explicit-any-remaining','stage3/adapt/45-regex-captures','stage3/adapt/99-correctness-core-debug-messages','stage3/adapt/99-correctness-core-getfunctionname-move','stage3/upstream/LEDGER.md']
 args.append(['git','add','--',*allowed])
 for cmd in args:
  child=subprocess.run(cmd,cwd=r,env=env);print('checkpoint child_exit',child.returncode,flush=True);assert child.returncode==0
 tree=subprocess.check_output(['git','write-tree'],cwd=r,env=env,text=True).strip();commit=subprocess.check_output(['git','commit-tree',tree,'-p',head],input='Checkpoint '+side+' full apply; proof continues\n',cwd=r,text=True).strip();run(side+'-checkpoint-push',['git','push','origin',commit+':refs/heads/'+branch],cwd=r);(s/'raw'/(side+'-checkpoint.json')).write_text(json.dumps({'sha':commit,'branch':branch,**entry},indent=2)+'\n');print('PUSHED CHECKPOINT',commit,flush=True)
elif mode=='replay':
 side=sys.argv[2];r=Path(sys.argv[3]);tree=s/('replay-'+side);source=s/side;out=s/('replay-raw-'+side);out.mkdir();shutil.copytree(source,tree,ignore=lambda p,n:[x for x in ['.git','built','node_modules','tests'] if Path(p)==source and x in n]);(tree/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
 journaler=r/'stage3/adapt/99-correctness-core-debug-messages/replay-io.cjs';records=[]
 for adapter in sorted((r/'stage3/adapt').glob('*/adapt.cjs')):
  name=adapter.parent.name;journal=out/(name+'.writes.jsonl');journal.write_text('');env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='--require='+str(journaler),BIND_REPLAY_ROOT=str(tree),BIND_REPLAY_JOURNAL=str(journal));begin=time.monotonic()
  with (out/(name+'.log')).open('w') as log:
   p=subprocess.Popen(['node',str(adapter),str(tree)],cwd=r,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
   try:
    try:code=p.wait(timeout=300)
    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait();code=p.returncode
   finally:
    try:os.killpg(p.pid,signal.SIGKILL)
    except ProcessLookupError:pass
    p.wait()
  text=(out/(name+'.log')).read_text().replace(str(tree),'<tree>').replace(str(r),'<repo>');records.append({'adapter':name,'exit':code,'stdout_stderr':text,'writes':[json.loads(l) for l in journal.read_text().splitlines()]});(out/'records.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps({'side':side,'adapter':name,'child_exit':code,'seconds':time.monotonic()-begin}),flush=True)
 with tarfile.open(s/'raw'/(side+'-replay.tar.gz'),'w:gz') as tf:
  for p in sorted(out.iterdir()):tf.add(p,arcname=p.name)
 print('REPLAY COMPLETE',side,len(records),flush=True)
