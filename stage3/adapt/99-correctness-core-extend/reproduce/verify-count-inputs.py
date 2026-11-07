import os,json,hashlib
from pathlib import Path
s=Path('/workspace/scratch/core-extend');before=s/'before';counted=s/'counted';missing=[];different=[];files=0
for directory,dirs,names in os.walk(before):
 if Path(directory)==before:dirs[:]=[d for d in dirs if d not in ['.git','node_modules','built']]
 for name in names:
  p=Path(directory)/name;relative=p.relative_to(before);q=counted/relative;files+=1
  if not q.is_file():missing.append(str(relative))
  elif relative.as_posix()!='src/compiler/core.ts' and p.read_bytes()!=q.read_bytes():different.append(str(relative))
result={'files':files,'missing':missing,'differentOutsideInstrumentedOwner':different};(s/'complete-count-input-check.json').write_text(json.dumps(result,indent=2)+'\n');assert not missing and not different,result;print('complete non-build inputs verified',files,flush=True)
