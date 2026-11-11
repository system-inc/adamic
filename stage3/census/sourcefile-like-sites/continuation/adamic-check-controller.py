import pathlib,sys
sys.dont_write_bytecode=True
from variants import invoke
p=pathlib.Path(__file__).resolve().parent
for label,root in [('original','/workspace/scratch/baobab-adapted'),('c','/workspace/scratch/baobab-road2-c'),('c-public','/workspace/scratch/baobab-road2-c-public')]:
 invoke(['/workspace/scratch/baobab-production-checker',str(p/'original/program.json'),root,str(p/label/'whole-src-adamic-checker.json')],p/label/'whole-src-adamic-checker.log',600)
