"""Classify every emitted JavaScript hunk against the two permitted source edits."""
from pathlib import Path
import difflib
import hashlib
import json
import re
import sys
before,after,out=map(Path,sys.argv[1:])
old_files={p.relative_to(before):p for p in before.rglob('*.js')}
new_files={p.relative_to(after):p for p in after.rglob('*.js')}
assert old_files.keys()==new_files.keys(), 'emitted file inventory changed'
rows=[];diffs=[]
for name in sorted(old_files):
 old=old_files[name].read_text();new=new_files[name].read_text()
 if old==new:continue
 assert old.count('range.split(whitespaceRegExp)')==1,(name,'split count')
 normalized=old.replace('range.split(whitespaceRegExp)','range.split(/\\s+/)')
 pattern=r'(?m)^[ \t]*(?:var )?'+re.escape('whitespaceRegExp = /\\s+/;')+r'\n'
 normalized,count=re.subn(pattern,'',normalized)
 assert count==1,(name,'declaration initialization count')
 normalized=normalized.replace('logicalOrRegExp, whitespaceRegExp, partialRegExp','logicalOrRegExp, partialRegExp')
 # Exact equality after the permitted edits is stronger than identifier normalization.
 assert normalized==new,(name,'additional JavaScript change requires review')
 diff=list(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='before/'+str(name),tofile='after/'+str(name)))
 hunks=[]
 for line in diff:
  if line.startswith('@@'):hunks.append({'header':line.strip(),'classification':None})
  elif hunks and line[:1] in '+-' and not line.startswith(('+++','---')):
   kind='parseRange split' if 'range.split(' in line else 'shared declaration'
   assert 'whitespaceRegExp' in line or 'range.split(/\\s+/)' in line or ('logicalOrRegExp' in line and 'partialRegExp' in line),(name,line)
   previous=hunks[-1]['classification'];assert previous in (None,kind),(name,'mixed hunk')
   hunks[-1]['classification']=kind
 rows.append({'file':str(name),'hunks':hunks,'normalized_sha256':hashlib.sha256(normalized.encode()).hexdigest(),'additional_changes':0})
 diffs.extend(diff)
assert rows,'no emitted change observed'
(out/'emitted.diff').write_text(''.join(diffs))
report={'files_compared':len(old_files),'changed_files':rows,'hunks':sum(len(row['hunks']) for row in rows),'outside_permitted_edits':0,'identifier_renames':0,'normalization':'exact bytes after only the split substitution and shared declaration removal'}
(out/'emitted.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps(report,indent=2))
