#!/usr/bin/env python3
"""Attempt every source separately without converting upstream faults into a green gate."""
import json,subprocess,time
from pathlib import Path
r=Path(__file__).resolve().parents[5];own=Path(__file__).resolve().parent;out=Path('/workspace/wave04-context/check');compiler=Path('/workspace/typeaware-wave-04-landing-d65/adamic');archives=Path('/workspace/typeaware-wave-04-landing-d65/next')
results={}
def run(label,args,cwd=r):
 start=time.monotonic()
 with (out/(label+'.stdout')).open('wb') as so,(out/(label+'.stderr')).open('wb') as se:code=subprocess.run(list(map(str,args)),cwd=cwd,stdout=so,stderr=se).returncode
 results[label]={'exit':code,'seconds':time.monotonic()-start};(out/'unaffected-results.json').write_text(json.dumps(results,indent=2));return code,(out/(label+'.stdout')).read_bytes(),(out/(label+'.stderr')).read_bytes()
virtual=r/'cohere/adamic_wave04_context_oracle.go';overlay=out/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(own/'oracle.go')}}))
assert run('individual-oracle-build',['go','build','-overlay',overlay,'-o',out/'oracle',virtual],r/'cohere')[0]==0
assert run('individual-native-build',[compiler,'build',own/'source_main.a','-o',out/'native','--tsgo',archives/'checker-normal.a'])[0]==0
config=out/'tsconfig.json';manifest=out/'manifest';records=[]
for at,path in enumerate(manifest.read_text().splitlines()):
 go,truth,error=run(f'individual-{at:03}-go',[out/'oracle',config,manifest,'--select',path]);native,actual,nerror=run(f'individual-{at:03}-native',[out/'native',config,manifest,'default',path]);match=go==native==0 and truth==actual and not nerror
 records.append({'path':path,'go_exit':go,'native_exit':native,'byte_equal':match})
 if not match:print(f'FAILED {Path(path).name}: Go exit {go}, native exit {native}, byte_equal={truth==actual}',flush=True)
(out/'individual-comparisons.json').write_text(json.dumps(records,indent=2))
print(f'Compared every input: {sum(x["byte_equal"] for x in records)}/{len(records)} complete wire matches. Overall gate remains FAILED.',flush=True)
raise SystemExit(1 if any(not x['byte_equal'] for x in records) else 0)
