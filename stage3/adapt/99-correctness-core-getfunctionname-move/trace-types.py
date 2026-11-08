from pathlib import Path
import sys,json,re,os
sys.dont_write_bytecode=True
from run import run,s
roots=sorted((s/'head2/src/compiler').rglob('*.ts'));assert len(roots)==79
trace=s/'raw/head2-types-file-access.txt';result=run('head2-types79-traced',['strace','-f','-s','4096','-e','trace=openat,openat2,newfstatat,statx','-o',trace,s/'adamic','types',*roots]);text=trace.read_text();opened=[str(p) for p in roots if str(p) in text];assert len(opened)==79,(len(opened),'root accesses missing');assert 'deprecatedCompat' not in text,'compiler imported deprecatedCompat: STOP, wall would move';(s/'raw/head2-no-deprecatedCompat.json').write_text(json.dumps({'passed':True,'method':'strace file operations of the actual adamic types process and children','rootCount':79,'allRootsObservedInTrace':True,'deprecatedCompatFileAccesses':0,'adamicExit':result['exit']},indent=2)+'\n');print('actual adamic process accessed all 79 roots and no deprecatedCompat path')
