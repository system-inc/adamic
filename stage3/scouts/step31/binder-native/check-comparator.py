#!/usr/bin/env python3
"""Reuse the pinned comparator and prove stdout, stderr and exit checks independently."""
import json,subprocess,sys
from pathlib import Path
scout=Path('/workspace/cache/step31-binder-area/stage3/scouts/step31');root=Path('/workspace/cache/step31-binder-compare-mutants');root.mkdir(exist_ok=False)
golden=Path('/tmp/step31-binder-golden.jsonl');data=golden.read_bytes();rows=[]
for name,body,err,code,failures in [('control',data,'',0,[]),('stdout-byte',bytes([data[0]^1])+data[1:],'',0,['stdout']),('stderr-byte',data,'x',0,['stderr']),('exit-byte',data,'',1,['exit'])]:
 source=root/f'{name}.stdout';source.write_bytes(body)
 script='import pathlib,sys;sys.stdout.buffer.write(pathlib.Path(sys.argv[1]).read_bytes());sys.stderr.write(sys.argv[2]);sys.exit(int(sys.argv[3]))'
 with (root/f'{name}.log').open('wb') as log:
  r=subprocess.run([sys.executable,str(scout/'compare.py'),'--output',str(root/name),str(golden),'--',sys.executable,'-c',script,str(source),err,str(code)],stdout=log,stderr=log)
 report=json.loads((root/name/'report.json').read_text());assert r.returncode==bool(failures) and sorted(report['differences'])==failures,(name,report)
 rows.append({'name':name,'exit':r.returncode,'report':report});print(name,report)
Path(__file__).with_name('comparator-mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
