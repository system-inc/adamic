from pathlib import Path
import collections,json
out=Path(__file__).resolve().parent
text=['# Control sites','','Distinct sites are exact source locations; these controls are separate from the 79-file count.','']
for case in ['arrays','k17_captured_unknown','k18_entry_then_declared']:
 r=json.loads((out/'evidence/controls'/(case+'-type.json')).read_text());events=r['Events'];text += [f'## {case}', '', f'{len(events)} events; {len({e["Where"]for e in events})} distinct sites. Forms: '+str(dict(collections.Counter(e['Form']for e in events)))+'.','']
 for e in events:
  file,line,col=e['Where'].rsplit(':',2);line=int(line)
  text += [f'### {Path(file).name}:{line}:{col} — {e["Form"]}','','```typescript',Path(file).read_text().splitlines()[line-1],'```','']
(out/'CONTROL-SITES.md').write_text('\n'.join(text))
p=out/'REPORT.md';t=p.read_text().replace('**9 events**, including that exact','**9 events at 9 distinct sites**, including that exact')
if 'CONTROL-SITES.md' not in t:t=t.replace('Raw control observations and per-process resource receipts','Each control event’s source line is quoted in [CONTROL-SITES.md](CONTROL-SITES.md). Raw control observations and per-process resource receipts')
p.write_text(t)
