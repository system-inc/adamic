import argparse,subprocess,json,re,sys
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('name');p.add_argument('root',type=Path);a=p.parse_args();out=Path('/workspace/scratch/destructure-price-f2989762/r2');stock=Path('/tmp/destructure-price-scratch/work/stage3/api/node_modules/typescript/lib/tsc.js')
cmd=['node',str(stock),'--project','src/compiler/tsconfig.json','--noEmit','--strictFunctionTypes','true','--incremental','false','--composite','false','--pretty','false','--noErrorTruncation']
meta=out/('strict-'+a.name+'-memory');subprocess.run([sys.executable,str(out/'measure.py'),'--out',str(meta),'--cwd',str(a.root),'--',*cmd],check=False)
text=meta.with_suffix('.log').read_text();diagnostics=[]
for line in text.splitlines():
 m=re.match(r'^(.*?)\((\d+),(\d+)\): error TS(\d+): (.*)$',line)
 global_m=re.match(r'^error TS(\d+): (.*)$',line)
 if m:diagnostics.append({'file':m[1],'line':int(m[2]),'column':int(m[3]),'code':int(m[4]),'message':m[5]})
 elif global_m:diagnostics.append({'file':None,'line':None,'column':None,'code':int(global_m[1]),'message':global_m[2]})
 elif diagnostics:diagnostics[-1]['message']+='\n'+line
measurement=json.loads(meta.with_suffix('.json').read_text());assert measurement['exit'] in [0,1,2],measurement
assert bool(diagnostics)==bool(measurement['exit']),measurement
result={'typescript':'6.0.3','command':cmd,'rootCount':79,'rootManifest':'root-manifest.json','diagnosticCount':len(diagnostics),'diagnostics':diagnostics,'measurement':measurement}
(out/('strict-'+a.name+'.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'road':a.name,'diagnosticCount':len(diagnostics),'peak_process_tree_rss_kib':measurement['peak_process_tree_rss_kib']}))
