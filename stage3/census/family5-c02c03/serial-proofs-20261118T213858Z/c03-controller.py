from common import *
API='unittests:: Public APIs for typescript.d.ts should be acknowledged when they change'
def lane(label,sha):
 subprocess.run(['git','switch','--detach',sha],check=True)
 folder=Path('/workspace/family5-final-lane-'+label)
 rc=run('lanes/'+label,['bash','stage3/lane/run.sh',folder],allow=True)
 report=json.loads((folder/'report.json').read_text());oracle=json.loads((folder/'oracle/report.json').read_text())
 record={'sha':sha,'exit_code':rc,'counts':report.get('counts'),'failures':oracle.get('failures'),'execution':report.get('execution'),'phases':oracle.get('phases')}
 state.setdefault('lanes',{})[label]=record;save()
 dest=W/'lanes'/label/'lane'
 for p in folder.rglob('*'):
  if not p.is_file():continue
  rel=p.relative_to(folder)
  if rel.parts[0]=='adapted-tree' or 'node_modules' in rel.parts:continue
  target=dest/rel;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,target)
 tree=folder/'adapted-tree'
 if tree.exists():
  manifest=[]
  for root in ['built/local','lib']:
   for p in sorted((tree/root).rglob('*')):
    if p.is_file():manifest.append({'file':str(p.relative_to(tree)),'bytes':p.stat().st_size,'sha256':digest(p)})
  (W/'lanes'/label/'artifact-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
  shutil.rmtree(tree)
 if report.get('execution',{}).get('apply_exit')!=0 or any(oracle.get('phases',{}).get(k,{}).get('exit')!=0 for k in ['install','build']):raise Stop(label+' infrastructure phase failed')
 if oracle.get('phases',{}).get('tests',{}).get('exit')==124:raise Stop(label+' standard lane test timeout')
 if rc==0:
  assert record['counts']=={'passing':106366,'failing':1,'pending':0},record['counts']
  assert len(record['failures'])==1 and record['failures'][0]['title']==API
 return rc
try:
 sha=state['c03'];assert state.get('c02_pushed');verify_landing(sha,'99-correctness-type-constraint',C02);assert git('rev-parse','HEAD')==sha
 tree=Path('/workspace/family5-c02-tree');tree3=Path('/workspace/family5-c03-tree')
 run('c03/adapter-contract',['node','/tmp/family5-contract.cjs',R/'stage3/adapt/99-correctness-type-constraint',tree/'src/compiler/types.ts','/tmp/family5-final-contract-c03'])
 run('c03/whole-apply',['bash','stage3/apply.sh',tree3])
 stock_and_public('c03',tree,tree3)
 rc=lane('c03-initial',sha)
 if rc:
  lane('main-control-for-c03',MAIN)
  for label,target in [('c03-alt-1',sha),('main-alt-1',MAIN),('c03-alt-2',sha),('main-alt-2',MAIN)]:lane(label,target)
  if not any(v['exit_code']==0 for k,v in state['lanes'].items() if k.startswith('c03-')):raise Stop('C03 has no green lane after main control and alternation')
 subprocess.run(['git','switch','--detach',sha],check=True)
 run('c03/product',['go','test','./stage3/lane/adaptedtree','-run','^TestProduct_Stage3AdaptedTree$','-count=1','-timeout','45m','-v'])
 paired('c03',tree,tree3,['checker.ts'],['G03','G12'],reuse=W/'c02-after')
 mutant('c03',R/'stage3/adapt/99-correctness-type-constraint',tree,tree3,['checker.ts'],{'G03','G12'})
 state['c03_remaining_proofs_green']=True;save();print('READY C03 proofs green; landing awaits review',flush=True)
except Exception as error:
 state['stop']=type(error).__name__+': '+str(error);save();(W/'STOP.txt').write_text(state['stop']+'\n'+traceback.format_exc());print('STOP '+state['stop'],flush=True)
finally:
 (W/'c03-controller-finished.json').write_text(json.dumps(state,indent=2)+'\n')
