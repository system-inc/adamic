#!/usr/bin/env python3
import json,os,subprocess,sys,tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[4]
out=Path(__file__).parent
base=json.loads(Path(sys.argv[1]).read_text())
with tempfile.TemporaryDirectory(prefix='latent-cache-') as tmp:
 tmp=Path(tmp)
 test=tmp/'latent_cache_test.go';test.write_text((out/'cache_test.go.txt').read_text())
 base['Replace'][str(root/'internal/lower/latent_cache_test.go')]=str(test)
 for mutant in (False,True):
  overlay=json.loads(json.dumps(base))
  if mutant:
   original=Path(base['Replace'][str(root/'internal/lower/latent_full.go')]).read_text()
   replacement=tmp/'latent_full.go';replacement.write_text(original.replace(' && v.Field(i).IsNil()', ''))
   overlay['Replace'][str(root/'internal/lower/latent_full.go')]=str(replacement)
  path=tmp/'overlay.json';path.write_text(json.dumps(overlay))
  logpath=out/('cache-mutant.log.txt' if mutant else 'cache.log.txt')
  with logpath.open('w') as log:
   result=subprocess.run(['go','test','-overlay',str(path),'./internal/lower','-run','^TestLatentEmptyEmitterCache$','-count=1','-v'],cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
  text=logpath.read_text()
  if mutant: assert result.returncode!=0 and 'populated emitter cache silently discarded' in text and '[build failed]' not in text
  else: assert result.returncode==0
  print('cache mutant caught' if mutant else 'cache snapshot pass')
