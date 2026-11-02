#!/usr/bin/env python3
import json,re
from pathlib import Path
out=Path(__file__).resolve().parent
def load(p):return json.loads((out/p).read_text())
def esc(v):return str(v).replace('|','&#124;').replace('\n',' ')
def outcome(r):
 if r['outcome']=='lowers':return 'lowers'
 if r['outcome']=='skipped_checker_body':
  text=(r.get('checker_diagnostics') or ['no diagnostic text'])[0]
  match=re.match(r'(.+?:\d+:\d+): error (TS\d+):',text)
  return 'skipped checker body; '+(match[2]+' at '+match[1].replace('/workspace/scratch/tree/','') if match else text)
 if r.get('reason'):return f"{r['kind']}: {r['reason']} at {r['where']}"
 return r['outcome']
rows=load('reason-changes.json')
lines=['# Every changed reason row','','Counts are distinct `(kind, where, reason, text)` sites. Examples are changed locations, up to five removed and five added per row.','','| Kind and reason | A | B | Removed examples | Added examples |','|---|---:|---:|---|---|']
for r in rows:lines.append('| '+ ' | '.join(map(esc,[r['kind']+': '+r['reason'],r['A'],r['B'],'; '.join(r['removed_examples']),' ; '.join(r['added_examples'])]))+' |')
flagged=[r for r in rows if re.search(r'view|branded supertype|predicate|\bread\b|\bwrite\b',r['reason'],re.I)]
lines+=['','## View, branded-supertype, predicate, read/write rows','']
for r in flagged:lines.append(f"- {r['kind']}: {r['reason']} — A {r['A']}, B {r['B']}.")
if not flagged:lines.append('None changed.')
(out/'reason-changes.md').write_text('\n'.join(lines)+'\n')
for filename,title,rows in [('selected-unit-outcomes','Direct tryCast/cast callers and nullParenthesizerRules',load('selected-unit-outcomes.json')),('unit-moves','Every unit whose outcome or first stop moves',load('unit-moves.json')),('lowering-regressions','Every unit lowering on A and stopping on B',load('lowering-regressions.json'))]:
 lines=['# '+title,'','| Unit | A | B |','|---|---|---|']
 for r in rows:lines.append('| '+' | '.join(map(esc,[r['name']+' ('+r['unit']+')',outcome(r['A']),outcome(r['B'])]))+' |')
 if not rows:lines+=['','None.']
 if filename=='selected-unit-outcomes':lines+=['','Checker-body skips and deferred generic declarations are reported explicitly; neither establishes successful lowering. See direct-calls.json for every direct call and its source line.']
 (out/(filename+'.md')).write_text('\n'.join(lines)+'\n')
rows=load('removed-lowering-sites.json')
lines=['# A lowering sites absent from B','','These are distinct diagnostic sites observed during A lowering and absent from B in every phase. Unit coverage and A-success/B-stop regressions are checked separately.','']
for r in rows:lines.append(f"- {r['kind']} at {r['where']}: {r['reason']} — {r['text']}")
if not rows:lines.append('None.')
(out/'removed-lowering-sites.md').write_text('\n'.join(lines)+'\n')
