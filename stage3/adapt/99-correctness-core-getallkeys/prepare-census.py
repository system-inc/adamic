"""Use the cited census procedure, changing only physical scratch suffixes."""
from pathlib import Path
import re,sys,json
from supervise import run,SCRATCH
repo=Path('/workspace/adamic')
for relative,args in [('stage3/census/latent/make_overlay.py',[str(repo),str(SCRATCH/'latent-overlay')]),('stage3/meter/entry_overlay.py',[str(SCRATCH/'latent-overlay'),str(SCRATCH/'entry-overlay')])]:
    source=repo/relative;text=source.read_text()
    text=text.replace("scratch / name.replace('/', '_')", "scratch / (name.replace('/', '_') + '.txt')")
    text=re.sub(r"((?:scratch|output) / '[^']+\.go)'",r"\1.txt'",text)
    sys.argv=[str(source),*args]
    exec(compile(text,str(source),'exec'),{'__file__':str(source),'__name__':'__main__'})
for name,command in [('adamic-build',['go','build','-o',str(SCRATCH/'adamic'),'./cmd/adamic']),('latent-build',['go','build','-buildvcs=false','-overlay='+str(SCRATCH/'latent-overlay/overlay.json'),'-o',str(SCRATCH/'latent'),'./stage3/census/latent/tool'])]:
    result=run(name,command)
    if result['exit_code']:raise SystemExit(1)
