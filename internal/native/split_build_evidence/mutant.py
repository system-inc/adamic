#!/usr/bin/env python3
"""Reverse outlined module initialization; require the output comparison to catch it."""
import json,os,pathlib,subprocess,sys
root=pathlib.Path.cwd();out=pathlib.Path(sys.argv[1]).resolve();out.mkdir(parents=True,exist_ok=True)
source=root/'internal/native/outline.go';text=source.read_text();old='for partIndex, indexes := range modules {'
assert text.count(old)==1
changed=out/'outline.go';changed.write_text(text.replace(old,'for order := range modules {\npartIndex := len(modules)-1-order\nindexes := modules[partIndex]'))
overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(changed)}}))
env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1'
with (out/'mutant.log').open('wb') as log:
 result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/oracle','-run','^TestNativeAgreesWithNode/internal/oracle/testdata/import_cycles/order/main','-count=1','-timeout=10m'],env=env,stdout=log,stderr=subprocess.STDOUT)
text=(out/'mutant.log').read_text()
assert result.returncode!=0 and 'stdout differs' in text and 'clang failed' not in text and 'compiling unit' not in text,text
print('module-order caught on output, after successful compilation and linking')
