from common import *
def publish_side(side):
 src=W/'entry'/side;d=R/'stage3/census/family5-c02c03/c02-alternatives/entry'/side
 subprocess.run(['git','switch','typescript/family5-c02c03-evidence'],check=True)
 subprocess.run(['git','pull','--ff-only','origin','typescript/family5-c02c03-evidence'],check=True)
 d.mkdir(parents=True,exist_ok=False)
 for p in src.iterdir():
  if not p.is_file():continue
  q=d/p.name
  if p.suffix in ['.log','.jsonl']:
   with p.open('rb') as a,gzip.GzipFile(filename=str(q)+'.gz',mode='wb',mtime=0) as b:shutil.copyfileobj(a,b)
  else:shutil.copyfile(p,q)
 root=d.parent.parent
 for n in ['common.py','entry-controller.py']:(root/n).write_text((W/n).read_text())
 (root/'METHOD.md').write_text('Base fbe5b72b; C02 12b9e041. Same pinned adapted source trees as the earlier C02 proof; no landing pushes. Each entry tool rebuilt at its own SHA with the meter overlay and -buildvcs=false. LATENT_FULL=0, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1, CommonJSBundle, no LATENT_ONLY_FILE, tsc.ts entry. Runs strictly sequential on the same idle box. Each completed stream is published before the next run starts. Child exits, times, source manifests and binary hashes retained. Raw stream SHA-256 is in stream.json.\n')
 subprocess.run(['git','add',str(root)],check=True);subprocess.run(['git','commit','-m','stage3: publish '+side+' C02 whole-entry latent stream'],check=True)
 subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/family5-c02c03-evidence'],check=True)
 state['entry_pushes'][side]=git('rev-parse','HEAD');save();print('PUBLISHED '+side+' '+state['entry_pushes'][side],flush=True)
try:
 (W/'tools').mkdir(exist_ok=True)
 for side,sha in [('base',MAIN),('c02',C02)]:
  subprocess.run(['git','switch','--detach',sha],check=True)
  run('setup/build-'+side,['go','build','-buildvcs=false','-overlay=/tmp/family5-tools/entry-overlay/overlay.json','-o',W/'tools'/(side+'-entry-census'),'./stage3/census/latent/tool'])
 assert git('diff',MAIN,C02,'--','internal','cohere','stage3/census','stage3/meter')==''
 subprocess.run(['git','switch','typescript/family5-c02c03-evidence'],check=True)
 for side,tree in [('base',Path('/tmp/family5-base-tree')),('c02',Path('/workspace/family5-c02-tree'))]:
  if side in state['entry_pushes']:continue
  out=W/'entry'/side;out.mkdir(parents=True,exist_ok=True)
  (out/'source-manifest.json').write_text(json.dumps(source_manifest(tree),indent=2)+'\n')
  env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
  env.update(LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TAIL='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_TIMES='/dev/stderr')
  rc=run('entry/'+side,[W/'tools'/(side+'-entry-census'),tree/'src/tsc/tsc.ts',out/'latent.jsonl'],env,allow=True,inputs={'source_sha':MAIN if side=='base' else C02,'tool_sha256':digest(W/'tools'/(side+'-entry-census'))})
  if rc==0:
   rows=[json.loads(l) for l in (out/'latent.jsonl').read_text().splitlines()];assert rows[-1].get('status')=='relation_report';assert rows[0].get('checker_rejected') is False
  (out/'stream.json').write_text(json.dumps({'complete':rc==0,'uncompressed_sha256':digest(out/'latent.jsonl'),'bytes':(out/'latent.jsonl').stat().st_size,'exit_code':rc},indent=2)+'\n')
  publish_side(side)
  if rc:raise Stop(side+' entry real child exit '+str(rc)+'; partial evidence published')
 state['entries_complete']=True;save()
except Exception as error:
 state['stop']=type(error).__name__+': '+str(error);save();(W/'STOP.txt').write_text(state['stop']+'\n'+traceback.format_exc());print('STOP '+state['stop'],flush=True)
finally:(W/'entry-finished.json').write_text(json.dumps(state,indent=2)+'\n')
