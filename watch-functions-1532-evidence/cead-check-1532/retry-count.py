from pathlib import Path
import time,subprocess,os,shutil,json
s=Path('/tmp/watch-file-check-cead');tree=Path('/tmp/watch-functions-1532/main-tree')
while not (s/'APPLY_OK').exists():
 if (s/'apply-after-exit.txt').exists() and (s/'apply-after-exit.txt').read_text()!='0':raise RuntimeError('full apply failed')
 time.sleep(1)
for name in ['count-oracle','count-oracle.log','count-exit.txt']:
 p=s/name
 if p.exists():p.rename(s/(name+'-oom'))
for p in tree.glob('.parallelperf*.json'):shutil.move(p,s/('prior-'+p.name))
(s/'memory-events-before-count.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
with (s/'retry-count.log').open('w') as log:subprocess.run(['python3',str(s/'count-run.py')],stdout=log,stderr=subprocess.STDOUT,check=True)
(s/'memory-events-after-count.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
with (s/'post-count.log').open('w') as log:subprocess.run(['python3',str(s/'post-count.py')],stdout=log,stderr=subprocess.STDOUT,check=True)
