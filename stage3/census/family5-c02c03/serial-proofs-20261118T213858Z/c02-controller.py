from common import *
try:
 p=W/'c02/product/status.json'
 while not p.exists():
  supervisor=json.loads((W/'product-supervisor.json').read_text());os.kill(supervisor['pid'],0)
  time.sleep(10)
 result=json.loads(p.read_text());assert result['exit_code']==0, 'Product child exit '+str(result['exit_code'])
 assert git('rev-parse','HEAD')==C02
 base=Path('/tmp/family5-base-tree');tree=Path('/workspace/family5-c02-tree');files=['utilities.ts','parser.ts','program.ts','checker.ts']
 verify_landing(C02,'99-correctness-line-and-character',MAIN)
 provenance=json.loads(Path('/tmp/family5-evidence/setup/tool-provenance.json').read_text())
 assert provenance['source_sha']==MAIN
 for name,h in provenance['tools'].items():assert digest('/tmp/family5-tools/'+name)==h
 paired('c02',base,tree,files,['G02'])
 mutant('c02',R/'stage3/adapt/99-correctness-line-and-character',base,tree,files,{'G02'})
 state['c02_remaining_proofs_green']=True;save();print('READY C02 remaining proofs green; landing awaits review',flush=True)
except Exception as error:
 state['stop']=type(error).__name__+': '+str(error);save();(W/'STOP.txt').write_text(state['stop']+'\n'+traceback.format_exc());print('STOP '+state['stop'],flush=True)
finally:
 (W/'c02-controller-finished.json').write_text(json.dumps(state,indent=2)+'\n')
