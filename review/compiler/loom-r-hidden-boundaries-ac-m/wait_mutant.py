import json,pathlib,subprocess,difflib
folder=pathlib.Path('review/compiler/loom-r-hidden-boundaries-ac-m');p=pathlib.Path('cloud/fast-gate/product-order.json');original=p.read_bytes();data=json.loads(original);del data['waits']['cmd/adamic TestProduct_CommandRuntimeCanonicalClosures'];mutated=(json.dumps(data,indent=2)+'\n').encode()
(folder/'removed-canonical-wait.patch').write_text(''.join(difflib.unified_diff(original.decode().splitlines(keepends=True),mutated.decode().splitlines(keepends=True),fromfile=str(p),tofile=str(p))))
command=['timeout','120','go','test','./cmd/adamic','-run','^TestCommandRuntimeProductsWaitForOwners$','-count=1','-timeout','90s','-json']
try:
 p.write_bytes(mutated)
 with (folder/'removed-canonical-wait.log').open('w') as log:run=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=150)
 text=(folder/'removed-canonical-wait.log').read_text();assert run.returncode==1 and 'TestProduct_CommandRuntimeCanonicalClosures must wait for internal/oracle TestProduct_NativeRuntimesSanitized2' in text,text
 (folder/'removed-canonical-wait.json').write_text(json.dumps(dict(command=command,exit=run.returncode,caught=True,catcher='TestCommandRuntimeProductsWaitForOwners',mutation='remove only the new canonical-closures wait'),indent=2)+'\n')
 print('CAUGHT removed canonical-closures wait')
finally:p.write_bytes(original)
