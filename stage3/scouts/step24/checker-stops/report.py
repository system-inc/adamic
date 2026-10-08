#!/usr/bin/env python3
"""Package measured logs and compare diagnostic identities, never exit-code absence."""
import collections,csv,gzip,hashlib,json,re
from pathlib import Path
P=Path(__file__).resolve().parent
rows=json.loads((P/'diagnostics.json').read_text())
logs={'main-parser':'/tmp/checker-stops-main-parser.log','topic-driver':'/tmp/checker-stops-topic-parser.log','topic-project-missing-types':'/tmp/checker-stops-project-parser.log','topic-project':'/tmp/checker-stops-project-parser-with-types.log','main-project':'/tmp/checker-stops-project-main-parser.log','setup':'/tmp/step24-checker-stops-setup.log','apply':'/tmp/checker-stops-apply.log','scratch-build':'/tmp/checker-stops-topic-build.log','scratch-merge':'/tmp/checker-stops-merge.log','fixtures':'/tmp/checker-stops-fixtures-final.log','audit':'/tmp/checker-stops-audit.log','inventory':'/tmp/checker-stops-inventory-final.log'}
for label,file in logs.items():
 with gzip.GzipFile(filename=str(P/'evidence'/f'{label}.log.gz'),mode='wb',mtime=0) as f:f.write(Path(file).read_bytes())
def ids(text):return {(f.split('/src/compiler/')[1],int(l),int(c),int(code)) for f,l,c,code in re.findall(r'^(.+?):(\d+):(\d+): error TS(\d+):',text,re.M)}
def key(r):return (r['file'],r['line'],r['column'],r['code'])
base={key(r) for r in rows}; fresh=ids(Path(logs['main-parser']).read_text()); main=ids(Path(logs['main-project']).read_text());topic=ids(Path(logs['topic-project']).read_text())
assert fresh==base and len(base)==320
assert topic<=base and len(topic)==67 and len(main)==319 and main<=base
assert len(base-topic)==253 and len(main-topic)==252
for r in rows:r.update(directDriver='unmeasured: ownership refusal',mainProject='remaining' if key(r) in main else 'absent',topicProject='remaining' if key(r) in topic else 'absent')
(P/'comparison.json').write_text(json.dumps({'baseline':320,'mainProject':319,'topicProject':67,'baselineAbsentInTopic':253,'pairedMainAbsentInTopic':252,'newDiagnostics':0,'directDriverClearances':None,'rows':rows},indent=2)+'\n')
with (P/'groups.csv').open('w',newline='') as f:
 w=csv.writer(f);w.writerow(['group','code','baseline','main_project','topic_project','baseline_absent_in_topic','paired_main_absent_in_topic','examples'])
 for g,code in sorted({(r['group'],r['code']) for r in rows}):
  rs=[r for r in rows if r['group']==g and r['code']==code]; b=sum(key(r) in main for r in rs);t=sum(key(r) in topic for r in rs)
  w.writerow([g,f'TS{code}',len(rs),b,t,len(rs)-t,b-t,'; '.join(r['id'] for r in rs[:3])])
lines=['| Pattern group | Codes (counts) | Baseline | Absent in topic project | Remain | Examples |','|---|---|---:|---:|---:|---|']
for g in sorted({r['group'] for r in rows},key=lambda g:-sum(r['group']==g for r in rows)):
 rs=[r for r in rows if r['group']==g];t=sum(key(r) in topic for r in rs); codes=collections.Counter(r['code'] for r in rs)
 ex='; '.join(f"{r['file']}:{r['line']}:{r['column']}" for r in rs[:3]);
 if len(rs)<3:ex+=f' (only {len(rs)} sites exist)'
 lines.append(f"| {g} | "+', '.join(f'TS{c} ({n})' for c,n in sorted(codes.items()))+f' | {len(rs)} | {len(rs)-t} | {t} | {ex} |')
(P/'TABLE.md').write_text('\n'.join(lines)+'\n')
# All implementation source inputs, not only files with diagnostics.
tree=Path('/workspace/cache/checker-stops-adapted');hashes={str(f.relative_to(tree)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((tree/'src/compiler').rglob('*.ts'))}
for f in [tree/'parser-proof-main.a',tree/'kinds.a']:hashes[f.name]=hashlib.sha256(f.read_bytes()).hexdigest()
(P/'input-hashes.json').write_text(json.dumps(hashes,indent=2)+'\n')
print('320 reproduced; project comparison 319 -> 67; 252 paired disappearances, 1 dependency disappearance; direct driver blocked')
