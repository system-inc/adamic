import json
from pathlib import Path
out=Path(__file__).resolve().parent;tree=Path('/workspace/scratch/next-stops-adapted/src/compiler')
source=out/'grouped-first-stops.json'
groups=json.loads((source if source.exists() else out/'previous-grouped-first-stops.json').read_text())
result=[]
for group in groups:
 if group['where']=='-':continue
 name,line,col=group['where'].rsplit(':',2);line=int(line);lines=(tree/name).read_text().splitlines(keepends=True)
 start=max(1,line-3);end=min(len(lines),line+3);text=''.join(lines[start-1:end])
 destination=out/(name.replace('/','_')+f'-{line}-{col}.txt');destination.write_text(text)
 result.append({'where':group['where'],'start_line':start,'end_line':end,'text':text,'file':destination.name})
(out/'source-excerpts.json').write_text(json.dumps(result,indent=2)+'\n')
