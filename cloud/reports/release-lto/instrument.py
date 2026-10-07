#!/usr/bin/env python3
"""Pin core 3, alternate independent hyperfine runs, retain wall/user/system and load."""
import json,os,subprocess,sys,shlex
from pathlib import Path
def verify_output(text, expected):
 body=text.split('\n',1)[1].split('  Time (',1)[0]
 if body != expected:raise RuntimeError('MISCOMPILE: pinned stdout/stderr differs')
s=Path(sys.argv[1]).resolve(); hf=s/'tools/usr/bin/hyperfine'
commands={'native': [s/'baseline/parse','--manifest',s/'compiler.txt','--count'], 'go':[s/'go-parse','--manifest',s/'compiler.txt','--count'], 'thin':[s/'thin/parse','--manifest',s/'compiler.txt','--count'], 'service-base':[s/'baseline/service','/tmp/wasm-requests-profile/requests.jsonl','run'], 'service-thin':[s/'thin/service','/tmp/wasm-requests-profile/requests.jsonl','run']}
os.environ['GOMAXPROCS']='1'; results={}
groups = [('instrument',['native','go'],10),('lto-parse',['native','thin'],5),('lto-service',['service-base','service-thin'],5)]
if len(sys.argv) > 2 and sys.argv[2] == '--lto-only':
 groups = groups[1:]
for group,names,count in groups:
 for name in names:
  with (s / ('warm-' + group + '-' + name + '.stdout')).open('wb') as out, (s / ('warm-' + group + '-' + name + '.stderr')).open('wb') as err:
   subprocess.run(['taskset','-c','3',*map(str,commands[name])],stdout=out,stderr=err,check=True)
  expected_out = b'7394547\n' if name.startswith('service-') else b'0\n'
  expected_err = b'serve:start\nserve:stop\n' if name.startswith('service-') else b''
  if (s / ('warm-' + group + '-' + name + '.stdout')).read_bytes() != expected_out or (s / ('warm-' + group + '-' + name + '.stderr')).read_bytes() != expected_err:
   raise RuntimeError('MISCOMPILE: pinned warmup output differs')
 rows=[]
 for i in range(count):
  row={'round':i+1,'order':names if i%2==0 else names[::-1]}
  for name in row['order']:
   stem=s/f'pinned-{group}-{name}-{i+1}'; before=os.getloadavg()
   command=shlex.join(['taskset','-c','3',*map(str,commands[name])])
   with Path(str(stem)+'.log').open('wb') as log:
    subprocess.run([str(hf),'--runs','1','--warmup','0','--shell','none','--show-output','--export-json',str(stem)+'.json',command],stdout=log,stderr=subprocess.STDOUT,check=True)
   verify_output(Path(str(stem)+'.log').read_text(), 'serve:start\nserve:stop\n7394547\n' if name.startswith('service-') else '0\n')
   r=json.loads(Path(str(stem)+'.json').read_text())['results'][0]
   row[name]={'wall':r['times'][0],'user':r['user'],'system':r['system'],'load_before':before,'load_after':os.getloadavg()}
  rows.append(row);print(group,json.dumps(row),flush=True)
 results[group]={'rounds':rows,'best':{name:{metric:min(r[name][metric] for r in rows) for metric in ['wall','user','system']} for name in names}}
 (s/'pinned-results.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps({k:v['best'] for k,v in results.items()}),flush=True)
