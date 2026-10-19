from pathlib import Path
import json,subprocess,time
out=Path(__file__).resolve().parent;root=out.parents[2];file=root/'internal/native/runtime/share.c';source=file.read_text();before='adamic_object *metadata = array->metadata;';assert source.count(before)==1
changed=out/'retain-interior-metadata.c.txt';changed.write_text(source.replace(before,before+' metadata->heap.references = 1;'));overlay=out/'retain-interior-metadata.json';overlay.write_text(json.dumps({'Replace':{str(file):str(changed)}}))
command=['go','test','-overlay='+str(overlay),'./internal/native','-run','^TestParserNodeArrayMetadataSharing$','-count=1','-v','-timeout','90s'];log=out/'logs/retain-interior-metadata.log';start=time.monotonic()
with log.open('w') as f:r=subprocess.run(command,cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=150)
text=log.read_text();caught=r.returncode==1 and 'interior metadata retained' in text and '[build failed]' not in text
(out/'interior-metadata-mutant-results.json').write_text(json.dumps(dict(mutant='retain-interior-metadata',command=command,exit=r.returncode,caught=caught,seconds=time.monotonic()-start,catcher='TestParserNodeArrayMetadataSharing: interior metadata retained'),indent=2)+'\n')
print('retain-interior-metadata',caught)
if not caught:raise SystemExit(1)
