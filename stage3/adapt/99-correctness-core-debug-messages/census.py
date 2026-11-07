"""Unchanged meter full mode, debug.ts alone, guarded against usable output."""
from pathlib import Path
import sys,os,re,json
sys.dont_write_bytecode=True
sys.path.insert(0,'/workspace/scratch/debug-messages')
from work import run,s,repo
mode=sys.argv[1]
if mode=='build':
 source=repo/'stage3/census/latent/make_overlay.py';text=source.read_text()
 text=text.replace("scratch / name.replace('/', '_')", "scratch / (name.replace('/', '_') + '.txt')")
 text=re.sub(r"((?:scratch|output) / '[^']+\.go)'",r"\1.txt'",text)
 sys.argv=[str(source),str(repo),str(s/'overlay')]
 exec(compile(text,str(source),'exec'),{'__file__':str(source),'__name__':'__main__'})
 run('census-build',['go','build','-buildvcs=false','-overlay='+str(s/'overlay/overlay.json'),'-o',s/'latent','./stage3/census/latent/tool'])
else:
 tree=s/mode;env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
 env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/debug.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(s/'raw'/(mode+'-census-times.jsonl')))
 run(mode+'-census',[s/'latent',tree/'src/compiler',s/'raw'/(mode+'-census.jsonl')],env=env)
