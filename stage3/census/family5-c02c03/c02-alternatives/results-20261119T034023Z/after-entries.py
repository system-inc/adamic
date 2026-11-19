import os,sys,time,json,subprocess,shutil,hashlib,difflib,traceback
from pathlib import Path
W=Path('/tmp/family5-c02-alternatives')
while not (W/'entry-finished.json').exists():time.sleep(10)
info=json.loads((W/'entry-controller.json').read_text())
while Path('/proc/'+str(info['pid'])+'/stat').exists() and Path('/proc/'+str(info['pid'])+'/stat').read_text().split()[2]!='Z':time.sleep(1)
from common import *
def findings(p):
 rows=[json.loads(l) for l in p.read_text().splitlines()];assert rows[-1].get('status')=='relation_report';assert rows[0].get('checker_rejected') is False
 return [f for row in rows for f in row.get('findings',[])]
def compare(label,before,after):
 out=W/'comparisons'/label;out.mkdir(parents=True,exist_ok=True)
 run('comparisons/'+label,[sys.executable,W/'compare.py',before,after,out/'comparison.json','G02'],allow=True)
 return json.loads((out/'comparison.json').read_text())
def source_diff(label,tree):
 d=W/'candidates'/label;edits=json.loads((tree/'candidate-edits.json').read_text());(d/'edits.json').write_text(json.dumps(edits,indent=2)+'\n')
 for file in sorted({e['file'] for e in edits}):
  a=(BASE/file).read_text().splitlines(keepends=True);b=(tree/file).read_text().splitlines(keepends=True);(d/(Path(file).stem+'.diff')).write_text(''.join(difflib.unified_diff(a,b,fromfile='base/'+file,tofile=label+'/'+file)))
 (d/'source-manifest.json').write_text(json.dumps(source_manifest(tree),indent=2)+'\n')
def census(label,tree,file):
 out=W/'full'/label/file.removesuffix('.ts');env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TAIL='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(tree/'src/compiler'/file),LATENT_TIMES='/dev/stderr')
 run('full/'+label+'/'+file.removesuffix('.ts'),[W/'tools/base-entry-census',tree/'src/tsc/tsc.ts',out/'latent.jsonl'],env,inputs={'types_sha256':digest(tree/'src/compiler/types.ts'),'file_sha256':digest(tree/'src/compiler'/file),'tool_sha256':digest(W/'tools/base-entry-census')});findings(out/'latent.jsonl')
try:
 assert state.get('entries_complete'),state.get('stop')
 BASE=Path('/tmp/family5-base-tree');files=['utilities.ts','parser.ts','program.ts','checker.ts']
 for side in ['base','c02']:
  folder=W/('entry-pair-'+side)/'tsc';folder.mkdir(parents=True,exist_ok=True);(folder/'latent.jsonl').symlink_to(W/'entry'/side/'latent.jsonl')
 compare('whole-entry',W/'entry-pair-base',W/'entry-pair-c02')
 phrase='a checked view with callable field getPositionOfLineAndCharacter';fs=findings(W/'entry/c02/latent.jsonl');named=[f for f in fs if phrase in f.get('reason','') or phrase in f.get('text','')]
 (W/'entry/c02/callable-sites.json').write_text(json.dumps({'phrase':phrase,'events':len(named),'all_distinct_positions':len({f['where'] for f in named}),'primary_distinct_positions':len({f['where'] for f in named if f['kind']!='Boundary'}),'by_kind_phase':dict(collections.Counter(f['kind']+'/'+f['phase'] for f in named)),'sites':named},indent=2)+'\n')
 run('publication/whole-entry-pair',[sys.executable,W/'publish-pair.py'])
 run('setup/stock-services-baseline',['node',W/'stock.cjs',BASE,W/'setup/stock-services-baseline.json'],dict(os.environ,FAMILY5_STOCK_PROJECT='src/services/tsconfig.json'),allow=True)
 valid=[]
 for label in ['a','b','c']:
  tree=Path('/workspace/family5-alt-'+label);d=W/'candidates'/label
  run('candidates/'+label+'/prepare',['node',W/'make-candidates.cjs',tree,label]);source_diff(label,tree)
  ordinary=run('candidates/'+label+'/stock',['node',W/'stock.cjs',tree,d/'stock.json'],allow=True)
  strict=run('candidates/'+label+'/stock-strict',['node',W/'stock.cjs',tree,d/'stock-strict.json'],dict(os.environ,FAMILY5_STOCK_STRICT='1'),allow=True)
  merge=run('candidates/'+label+'/services-merge',['node',W/'stock.cjs',tree,d/'services-merge.json'],dict(os.environ,FAMILY5_STOCK_PROJECT='src/services/tsconfig.json'),allow=True)
  js=run('candidates/'+label+'/javascript',['node',W/'javascript.cjs',BASE,tree,d/'javascript.json'],allow=True)
  build=run('candidates/'+label+'/public-build',['node',R/'stage3/typescript/node_modules/hereby/bin/hereby.js','local'],cwd=tree,allow=True)
  bundles=run('candidates/'+label+'/emitted-bundles',[sys.executable,W/'emitted-bundles.py',BASE,tree,d/'emitted-bundles.json'],allow=True)
  pub=tree/'built/local/typescript.d.ts';api={'produced':pub.exists(),'build_exit':build,'baseline_sha256':digest(BASE/'built/local/typescript.d.ts')}
  if pub.exists():
   api['candidate_sha256']=digest(pub);api['byte_identical']=pub.read_bytes()==(BASE/'built/local/typescript.d.ts').read_bytes();run('candidates/'+label+'/public-compare',['cmp',BASE/'built/local/typescript.d.ts',pub],allow=True)
   if not api['byte_identical']:(d/'public-typescript.diff').write_text(''.join(difflib.unified_diff((BASE/'built/local/typescript.d.ts').read_text().splitlines(keepends=True),pub.read_text().splitlines(keepends=True),fromfile='base/typescript.d.ts',tofile=label+'/typescript.d.ts')))
  (d/'public.json').write_text(json.dumps(api,indent=2)+'\n')
  baseline_merge=json.loads((W/'setup/stock-services-baseline.json').read_text());candidate_merge=json.loads((d/'services-merge.json').read_text())
  def diagnostic_key(x):return (x['code'],x.get('file'),re.sub(r':\d+:\d+','',x['message']))
  old=collections.Counter(map(diagnostic_key,baseline_merge['diagnostics']));new=collections.Counter(map(diagnostic_key,candidate_merge['diagnostics']));introduced=list((new-old).elements());(d/'services-merge-delta.json').write_text(json.dumps({'baseline_diagnostic_count':baseline_merge['diagnostic_count'],'candidate_diagnostic_count':candidate_merge['diagnostic_count'],'introduced_diagnostics':introduced},indent=2)+'\n')
  result={'stock_exit':ordinary,'stock_strict_exit':strict,'services_merge_exit':merge,'introduced_services_diagnostics':len(introduced),'javascript_exit':js,'emitted_bundles_exit':bundles,'public_build_exit':build,'public':api,'landing_prerequisites_green':ordinary==0 and strict==0 and not introduced and js==0 and bundles==0 and build==0 and api.get('byte_identical') is True,'eligible_for_full_census':ordinary==0 and strict==0 and not (label=='a' and introduced)}
  state['candidates'][label]=result;save()
  if result['eligible_for_full_census']:valid.append(label)
  else:print('CANDIDATE '+label+' has no checker-clean compiler input or is killed by the explicitly disqualifying (a) merge error; full census not run',flush=True)
 for label in valid:
  for file in files:
   if not (W/'full/base'/file.removesuffix('.ts')/'status.json').exists():
    # Reuse only the completed, isolated utilities baseline with identical input and tool hashes.
    old=Path('/tmp/family5-final-proofs/c02-before/utilities')
    if file=='utilities.ts' and json.loads((old/'status.json').read_text())['inputs']['tool_sha256']==digest(W/'tools/base-entry-census') and json.loads((old/'status.json').read_text())['inputs']['types_sha256']==digest(BASE/'src/compiler/types.ts') and json.loads((old/'status.json').read_text())['inputs']['file_sha256']==digest(BASE/'src/compiler/utilities.ts'):
     dest=W/'full/base/utilities';dest.parent.mkdir(parents=True,exist_ok=True);shutil.copytree(old,dest);(dest/'reuse.json').write_text(json.dumps({'source':str(old),'same_source_and_tool':True,'original_run_was_isolated':True},indent=2)+'\n')
    else:census('base',BASE,file)
   census(label,Path('/workspace/family5-alt-'+label),file)
  compare('candidate-'+label,W/'full/base',W/'full'/label)
 state['alternatives_complete']=True;save();print('READY all alternative measurements complete; report and publication pending',flush=True)
except Exception as error:
 state['stop']=type(error).__name__+': '+str(error);save();(W/'ALTERNATIVES-STOP.txt').write_text(state['stop']+'\n'+traceback.format_exc());print('STOP '+state['stop'],flush=True)
finally:(W/'alternatives-finished.json').write_text(json.dumps(state,indent=2)+'\n')
