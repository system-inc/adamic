from pathlib import Path
import subprocess,json
root=Path(__file__).resolve().parents[3];out=Path(__file__).resolve().parent;base=(out/'base.txt').read_text().strip();old=subprocess.check_output(['git','show',base+':internal/oracle/counts.md'],cwd=root,text=True);parts=old.split('\n## Predicate direction counts',1);result={};expected=set()
for kind,text in zip(['runtime','predicates'],parts):
 rows=[line for line in text.splitlines() if line.startswith('| ') and not line.startswith('| Fixture')];seen=set()
 for line in rows:
  key=line.split(' | ')[0][2:];assert key not in seen;seen.add(key);path=root/'internal/oracle/testdata/counts'/kind/(key+'.counts');assert path.read_bytes()==(line+'\n').encode(),str(path);expected.add(path)
 result[kind]=len(rows);print(f'{kind}: {len(rows)}/{len(rows)} keys and every value identical; row bytes identical')
actual=set((root/'internal/oracle/testdata/counts').rglob('*.counts'));assert actual==expected
print(f'TOTAL: {sum(result.values())} rows; 0 missing, 0 extra, 0 changed values')
(out/'equality-proof.json').write_text(json.dumps(dict(base=base,rows=result,total=sum(result.values()),missing=0,extra=0,changed=0),indent=2)+'\n')
