import json, subprocess
from pathlib import Path
out=Path(__file__).parent
base=subprocess.check_output(['git','show','dbdf977f5b:internal/oracle/counts.md'],text=True)
current=Path('internal/oracle/counts.md').read_text()
def rows(text):
 result={}
 for line in text.splitlines():
  cells=[s.strip() for s in line.strip('|').split('|')]
  if len(cells) in (5,7) and all(s.isdecimal() for s in cells[1:]): result[cells[0]]=line
 return result
a,b=rows(base),rows(current)
changed=[dict(key=k,before=a.get(k),after=b.get(k)) for k in sorted(a.keys()|b.keys()) if a.get(k)!=b.get(k)]
(out/'counts-changes.json').write_text(json.dumps(changed,indent=2)+'\n')
print(json.dumps(dict(added=sum(x['before'] is None for x in changed),removed=sum(x['after'] is None for x in changed),existing=sum(x['before'] is not None and x['after'] is not None for x in changed))))
