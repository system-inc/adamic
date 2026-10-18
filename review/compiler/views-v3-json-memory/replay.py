from pathlib import Path
import os,subprocess,sys
out=Path('/workspace/v3-json-artifacts/replay'); base=Path('/workspace/v3-json-artifacts')
source=base/'v3-killed-native/003/main.ts'
for index in range(5):
 cases=base/f'v3-killed-native/006/chunk-{index}.txt'
 with (out/f'node-{index}.stdout').open('wb') as stdout,(out/f'node-{index}.stderr').open('wb') as stderr:
  result=subprocess.run(['/tmp/v3-json-measure/rss',str(out/f'node-{index}.rss'),'node','--disable-warning=ExperimentalWarning','/workspace/adamic/oracle/node.mjs',str(source),'--cases',str(cases)],stdout=stdout,stderr=stderr)
 assert result.returncode==0 and (out/f'node-{index}.stderr').stat().st_size==0
 for label in ['main','v3']:
  binary=base/f'{label}-killed-native/005/sanitized'
  with (out/f'{label}-{index}.stdout').open('wb') as stdout,(out/f'{label}-{index}.stderr').open('wb') as stderr:
   result=subprocess.run(['/tmp/v3-json-measure/rss',str(out/f'{label}-{index}.rss'),str(binary),'--cases',str(cases)],env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1'),stdout=stdout,stderr=stderr)
  assert result.returncode==0 and (out/f'{label}-{index}.stderr').stat().st_size==0,(label,index,result.returncode)
  same=subprocess.run(['cmp',str(out/f'node-{index}.stdout'),str(out/f'{label}-{index}.stdout')],stdout=subprocess.DEVNULL).returncode==0
  assert same,(label,index,'stdout differs')
  print(label,index,(out/f'{label}-{index}.rss').read_text(),flush=True)
