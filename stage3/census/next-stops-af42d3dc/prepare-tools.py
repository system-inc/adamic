from pathlib import Path
import json,os,sys
from supervise import run,scratch
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent
# The reference runner patches only physical scratch filenames. Virtual Go paths stay unchanged.
status=run('overlay',[sys.executable,str(out/'build-entry.py')],limit=1800)
if status['exit_code']:sys.exit(1)
for name,cmd in [
 ('ordinary-build',['go','build','-o',str(scratch/'census'),'./stage3/census/tool']),
 ('latent-build',['go','build','-buildvcs=false','-overlay='+str(scratch/'latent-overlay/overlay.json'),'-o',str(scratch/'latent'),'./stage3/census/latent/tool']),
 ('entry-build',['go','build','-buildvcs=false','-overlay='+str(scratch/'entry-overlay/overlay.json'),'-o',str(scratch/'entry-census'),'./stage3/census/latent/tool'])]:
 status=run(name,cmd,limit=3600)
 if status['exit_code']:sys.exit(1)
