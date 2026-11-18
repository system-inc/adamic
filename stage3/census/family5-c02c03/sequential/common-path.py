import os,sys,subprocess,json,hashlib,shutil,re
from pathlib import Path
R=Path('/tmp/family5-land');S=Path('/tmp/family5-sequential');tree=Path('/workspace/family5-c02-lane-alone/adapted-tree');adapter=S/'c02-adapter'
shutil.copytree(R/'stage3/adapt/99-correctness-line-and-character',adapter)
os.chdir(R);env0=dict(os.environ,NODE_PATH=str(R/'stage3/api/node_modules'))
sides={'main':'fbe5b72ba115a8f3060d4b5229fd578574f4505f','c02':'28a2e22015d0cf51667fcefa7acb246cab553817'}
results=[]
def proof(name,cmd,cwd=R,env=env0):return subprocess.call([sys.executable,'/tmp/family5-proof.py',str(S/name),*map(str,cmd)],cwd=cwd,env=env)
try:
 for file,runner in [('codeFixClassImplementInterfaceNoTruncation.ts','fourslash'),('templateLiteralTypes1.ts','conformance')]:
  for side,sha in sides.items():
   subprocess.run(['git','switch','--detach',sha],check=True)
   name='common-targeted/'+side+'/'+file.removesuffix('.ts')
   code=proof(name+'/prepare',['node',adapter/('mutant.cjs' if side=='main' else 'adapt.cjs'),tree]);assert code==0
   expected=json.loads((S/('source-'+side+'.json')).read_text());actual={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()};assert actual==expected,(side,'source mismatch')
   built=tree/'built/local';js={str(p.relative_to(built)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(built.rglob('*.js')) if p.is_file()};js['typescript.d.ts']=hashlib.sha256((built/'typescript.d.ts').read_bytes()).hexdigest();assert js==json.loads((S/('built-'+side+'.json')).read_text())
   if (tree/'tests/baselines/local').exists():shutil.rmtree(tree/'tests/baselines/local')
   out=S/name;timing=out/'times';errors=out/'errors';timing.mkdir(parents=True);errors.mkdir()
   env=dict(env0,FAMILY5_TIMING_DIR=str(timing),STAGE3_ERROR_DIR=str(errors));env['NODE_OPTIONS']=' --require='+str(S/'measure-runnables.cjs')+' --require='+str(R/'stage3/oracle/observe-errors.cjs')
   cmd=['npm','test','--','--workers=1','--lint=false','--no-colors','--runners='+runner,'--tests='+re.escape(file)]
   code=proof(name+'/test',cmd,tree,env);events=[]
   for p in timing.glob('*.jsonl'):events.extend(json.loads(l) for l in p.read_text().splitlines())
   wanted=[x for x in events if file in x['title'] and (x['kind']=='test' if runner=='fourslash' else x['kind']=='hook' and 'before all' in x['title'])]
   row={'side':side,'sha':sha,'case':file,'exit_code':code,'elapsed_ms':max((x['elapsed_ms'] for x in wanted),default=None),'source_manifest_matches_side':True,'runtime_manifest_matches_side':True,'shared_path':str(tree),'events':wanted}
   results.append(row);(out/'measurement.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps({k:v for k,v in row.items() if k!='events'}),flush=True)
finally:
 subprocess.run(['git','switch','typescript/line-and-character-1'],check=True)
 code=proof('common-targeted/restore',['node',adapter/'adapt.cjs',tree]);assert code==0
 (S/'common-path-result.json').write_text(json.dumps(results,indent=2)+'\n')
