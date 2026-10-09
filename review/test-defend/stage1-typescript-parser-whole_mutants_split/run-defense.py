import os,json,pathlib,subprocess,time,re
root=pathlib.Path('/workspace/adamic');out=root/'review/test-defend/stage1-typescript-parser-whole_mutants_split';pkg='./stage1/typescript/parser/'
env=os.environ.copy();env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u158/corpus';env['ADAMIC_PARSER_BENCH']='1'
groups={
 'compiler-expression':['TestCompilerExpressionsAgree_Setup','TestCompilerExpressionsAgreeUnion']+[f'TestCompilerExpressionsAgree_{i:03}' for i in range(16)],
 'expressions':['TestExpressionsAgree'],
 'jsx-native':['TestJsxNative'],
 'other-agreements':['TestGeneratedExpressionsAgree','TestEveryTypeNodeKindAgrees','TestObsoleteImportAttributesAgrees','TestJsxNode','TestTypeOnlyImportCycleCompiles'],
 'defense':['TestWholeCompilerAgrees','TestYieldLookaheadAgrees','TestWholeGeneratedAgrees','TestWholePerformance'],
}
runs=json.loads((out/'runs.json').read_text())
def run(tag,rows,cache):
 e=env.copy();e['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend158/cache/'+cache
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','^('+'|'.join(rows)+')$']; start=time.monotonic();log=out/(tag+'.log')
 with log.open('w') as f:r=subprocess.run(cmd,cwd=root,env=e,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for l in log.read_text().splitlines():
  try:events.append(json.loads(l))
  except ValueError:pass
 fails=[x['Test'] for x in events if x.get('Test') and '/' not in x['Test'] and x['Action']=='fail'];passes=[x['Test'] for x in events if x.get('Test') and '/' not in x['Test'] and x['Action']=='pass'];skips=[x['Test'] for x in events if x.get('Test') and x['Action']=='skip'];outputs={name:[x.get('Output','').strip() for x in events if x.get('Test')==name and x['Action']=='output' and '.go:' in x.get('Output','')] for name in fails}
 result=dict(tag=tag,rows=rows,command=cmd,environment={k:e[k] for k in ['ADAMIC_TYPESCRIPT_SOURCE','ADAMIC_PARSER_BENCH','ADAMIC_BUILD_CACHE_DIR']},exit=r.returncode,wall_seconds=time.monotonic()-start,rows_failed=fails,rows_passed=passes,rows_skipped=skips,unknown=sorted(set(rows)-set(fails+passes+skips)),cooked='test timed out' in log.read_text() or r.returncode==124,failing_lines=outputs)
 runs.append(result);(out/'runs.json').write_text(json.dumps(runs,indent=2)+'\n');print(tag,r.returncode,round(result['wall_seconds'],2),fails,flush=True);return result
# Baseline in separate bounded groups after the previous combined run cooked.
for group in ['expressions']:
 r=run('baseline-'+group,groups[group],'clean')
 if r['exit']!=0:raise SystemExit('baseline not green')
mutants=json.loads((out/'menu.json').read_text())
for m in mutants:
 path=root/m['file'];original=path.read_text();assert original.count(m['old'])==1
 try:
  path.write_text(original.replace(m['old'],m['new'],1)); diff=subprocess.check_output(['git','diff','--',m['file']],cwd=root);(out/(m['id']+'.diff')).write_bytes(diff)
  r=run(m['id']+'-build',['TestProduct_WholeMutantsNative_Control'],m['id']);
  if r['exit']!=0:raise SystemExit('mutant did not compile')
  for group,rows in groups.items():run(m['id']+'-'+group,rows,m['id'])
 finally:path.write_text(original)
