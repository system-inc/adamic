from pathlib import Path
import json,subprocess,time
out=Path(__file__).resolve().parent;root=out.parents[2];file=root/'internal/native/runtime/share.c';source=file.read_text();before='if (array->metadata != NULL) {';assert source.count(before)==1
changed=out/'skip-metadata-sharing.c.txt';changed.write_text(source.replace(before,'if (false && array->metadata != NULL) {'));overlay=out/'skip-metadata-sharing.json';overlay.write_text(json.dumps({'Replace':{str(file):str(changed)}}))
command=['go','test','-overlay='+str(overlay),'./internal/native','-run','^TestParserNodeArrayMetadataSharing$','-count=1','-v','-timeout','90s'];log=out/'logs/skip-metadata-sharing.log';start=time.monotonic()
with log.open('w') as f:r=subprocess.run(command,cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=150)
text=log.read_text();caught=r.returncode==1 and 'metadata child not shared' in text and '[build failed]' not in text
(out/'metadata-sharing-mutant-results.json').write_text(json.dumps(dict(mutant='skip-metadata-sharing',command=command,exit=r.returncode,caught=caught,seconds=time.monotonic()-start,catcher='TestParserNodeArrayMetadataSharing: metadata child not shared'),indent=2)+'\n')
print('skip-metadata-sharing',caught)
if not caught:raise SystemExit(1)
