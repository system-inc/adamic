#!/usr/bin/env python3
"""Three quiet wall-clock comparisons after the complete byte oracle passes."""
from pathlib import Path
import argparse,json,statistics,subprocess,time
p=argparse.ArgumentParser();p.add_argument('build',type=Path);p.add_argument('directory',type=Path);p.add_argument('--compiler-manifest',required=True);p.add_argument('--compiler-config',required=True);p.add_argument('--repository-manifest',required=True);a=p.parse_args();o=a.directory.resolve();o.mkdir(parents=True,exist_ok=True);r=Path(__file__).resolve().parents[4];result=[]
for name,config,manifest in [('compiler',a.compiler_config,a.compiler_manifest),('repository',r/'tsconfig.json',a.repository_manifest)]:
 values={'go':[],'native':[]}
 for at in range(3):
  output=[]
  for variant,binary in [('go',a.build/'oracle'),('native',a.build/'native-normal')]:
   stem=o/f'{name}-{at}-{variant}';started=time.perf_counter()
   with stem.with_suffix('.stdout').open('wb') as stdout,stem.with_suffix('.stderr').open('wb') as stderr:
    run=subprocess.run([str(binary),str(config),str(manifest),'--count'],stdout=stdout,stderr=stderr)
   elapsed=time.perf_counter()-started
   if run.returncode or variant=='native' and stem.with_suffix('.stderr').read_bytes():raise RuntimeError(str(stem)+' failed')
   values[variant].append(elapsed);output.append(stem.with_suffix('.stdout').read_bytes())
  if output[0]!=output[1]:raise RuntimeError('count outputs differ')
 result.append(dict(corpus=name,findings=output[0].decode().strip(),go_seconds=values['go'],native_seconds=values['native'],go_median_seconds=statistics.median(values['go']),native_median_seconds=statistics.median(values['native'])))
(o/'timings.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
