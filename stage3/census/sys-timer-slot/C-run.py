"""C measurements; never alter or regenerate the accepted A/B evidence."""
import sys,pathlib,json,hashlib,shutil,os
sys.dont_write_bytecode=True
sys.path.insert(0,str(pathlib.Path(__file__).parent))
from run import run,scratch,repo,evidence
mode=sys.argv[1]
labels=['C-noinfer-a','C-noinfer-b'] if mode.endswith('-alternative') else ['C-a','C-b']
mode=mode.removesuffix('-alternative')
if mode=='C-checks':
 for label in labels:
  run(label+'-stock',['node','--max-old-space-size=8192',str(evidence/'C-check.cjs'),str(scratch/label),str(repo),str(scratch/(label+'-stock.json'))])
elif mode=='C-builds':
 baseline=json.loads((scratch/'main-compiler-js.json').read_text())
 for label in labels:
  run(label+'-build-compiler',['npm','run','build:compiler','--','--lkg'],scratch/label)
  partial={name:hashlib.sha256((scratch/label/'built/local'/name).read_bytes()).hexdigest() for name in baseline if (scratch/label/'built/local'/name).exists()}
  (scratch/(label+'-partial-compiler-js.json')).write_text(json.dumps(partial,indent=2)+'\n')
  run(label+'-build-tests',['npm','run','build:tests','--','--lkg'],scratch/label)
  run(label+'-finish-shim',['node',str(evidence/'finish-shim.mjs'),str(scratch/label)],scratch/label)
  run(label+'-bundle-declarations',['node','scripts/dtsBundler.mjs','--entrypoint','./built/local/typescript/typescript.d.ts','--output','./built/local/typescript.d.ts'],scratch/label)
  hashes={name:hashlib.sha256((scratch/label/'built/local'/name).read_bytes()).hexdigest() for name in baseline}
  (scratch/(label+'-compiler-js.json')).write_text(json.dumps(hashes,indent=2)+'\n')
elif mode=='C-latent':
 for label in labels:
  run(label+'-latent',[str(scratch/'census'),str(scratch/label/'src/compiler'),str(scratch/(label+'-latent.jsonl'))],extra={'LATENT_FULL':'1','LATENT_ONLY_FILE':str(scratch/label/'src/compiler/sys.ts'),'LATENT_ASSERT_NO_OUTPUT':'1','LATENT_TIMES':str(scratch/(label+'-latent-times.jsonl'))})
elif mode=='C-witness':
 run('C-node-witness',['node',str(evidence/'C-node-witness.cjs'),str(scratch/'C-node-witness.json')])
 probe=scratch/'C-clear-probe';probe.mkdir();shutil.copy2(evidence/'C-clear-probe.ts.txt',probe/'probe.ts')
 run('C-clear-probe-latent',[str(scratch/'census'),str(probe),str(scratch/'C-clear-probe-latent.jsonl')],extra={'LATENT_FULL':'1','LATENT_ONLY_FILE':str(probe/'probe.ts'),'LATENT_ASSERT_NO_OUTPUT':'1'})
