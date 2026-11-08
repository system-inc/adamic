import sys
sys.dont_write_bytecode=True
from run import *
assert run('DP-stock',['node','--max-old-space-size=8192',S/'check.cjs',S/'DP',ROOT,S/'DP-stock.json'])==0
r=json.loads((S/'DP-stock.json').read_text());print('stock diagnostics: adamic='+str(len(r['adamic']))+' upstream='+str(len(r['solution']))+' solutionExit='+str(r['solutionExit']),flush=True);print('provider types '+json.dumps(r['providerTypes']),flush=True)
assert run('DP-latent',[S/'census',S/'DP/src/compiler',S/'DP-latent.jsonl'],extra={'LATENT_FULL':'1','LATENT_ONLY_FILE':str(S/'DP/src/compiler/sys.ts'),'LATENT_ASSERT_NO_OUTPUT':'1','LATENT_TIMES':str(S/'DP-times.jsonl')})==0
