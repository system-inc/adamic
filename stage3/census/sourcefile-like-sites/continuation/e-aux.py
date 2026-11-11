import pathlib,time,sys,json
sys.dont_write_bytecode=True
from variants import invoke
p=pathlib.Path(__file__).resolve().parent;root='/workspace/scratch/baobab-road2-e'
while not (p/'e/source-diff.json').exists():time.sleep(2)
for label,tree in [('original-eslint','/workspace/scratch/baobab-adapted'),('e',root)]:
 invoke(['node',str(p/'e-eslint.cjs'),tree,label],p/label/'eslint.log',600)
for label,mode in [('e','stock'),('e-exact','exact')]:
 invoke(['node',str(p/'source-audit-e.cjs'),root,label,mode],p/label/'source-audit.log',600)
invoke(['/workspace/scratch/baobab-production-checker',str(p/'original/program.json'),root,str(p/'e/whole-src-adamic-checker.json')],p/'e/whole-src-adamic-checker.log',600)
invoke(['node',str(p/'emit-public-c.cjs'),root,'e','on'],p/'e/public-emit.log',600)
invoke(['node',str(p/'api-c.cjs'),'e','original'],p/'e/api-analysis.log',600)
invoke(['python3',str(p/'lane-api-c.py'),'e'],p/'e/lane-api-check.log',600)
invoke(['python3',str(p/'sdk-cohere-c.py'),'e'],p/'e/sdk-cohere-controller.log',600)
(p/'e/sdk-program.json').write_text(json.dumps({'roots':['public-adamic-consumer.ts']},indent=2)+'\n')
invoke(['/workspace/scratch/baobab-production-checker',str(p/'e/sdk-program.json'),str(p/'e'),str(p/'e/sdk-adamic-checker.json')],p/'e/sdk-adamic-checker.log',600)
invoke(['python3',str(p/'cohere-c.py'),'e'],p/'e/cohere-controller.log',1500)
print('e auxiliary workflow child exit=0',flush=True)
