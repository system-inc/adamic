import pathlib,sys,json
sys.dont_write_bytecode=True
from variants import invoke
p=pathlib.Path(__file__).resolve().parent
root='/workspace/scratch/baobab-road2-c'
for label,mode in [('c','stock'),('c-exact','exact')]:
 invoke(['node',str(p/'source-audit-c.cjs'),root,label,mode],p/label/'source-audit.log',600)
for label,source,strip in [('c','c','on'),('c-unstripped','c','off'),('original-unstripped','adapted','off')]:
 invoke(['node',str(p/'emit-public-c.cjs'),'/workspace/scratch/baobab-road2-c' if source=='c' else '/workspace/scratch/baobab-adapted',label,strip],p/label/'public-emit.log',600)
print('c auxiliary child exit=0',flush=True)
