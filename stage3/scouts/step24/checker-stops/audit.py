#!/usr/bin/env python3
"""Reject missing, duplicate or misattributed inventory rows; prove each rejection."""
import copy,gzip,json,re
from pathlib import Path
P=Path(__file__).resolve().parent
expected=[(f.split('/src/compiler/')[1],int(l),int(c),int(k)) for f,l,c,k in re.findall(r'^(.+?):(\d+):(\d+): error TS(\d+):',gzip.open(P/'evidence/baseline.log.gz','rt').read(),re.M)]
rows=json.loads((P/'diagnostics.json').read_text())
def check(rs):
 actual=[(r['file'],r['line'],r['column'],r['code']) for r in rs]
 assert len(actual)==320 and actual==expected
 assert len({r['id'] for r in rs})==320
 for r in rs:
  assert r['id']==f"{r['file']}:{r['line']}:{r['column']}:TS{r['code']}"
check(rows)
for name,mutate in [('drop',lambda r:r.pop()),('duplicate',lambda r:r.__setitem__(1,r[0])),('coordinate',lambda r:r[0].__setitem__('line',r[0]['line']+1)),('code',lambda r:r[0].__setitem__('code',2322))]:
 bad=copy.deepcopy(rows);mutate(bad)
 try:check(bad)
 except AssertionError:print(name+' mutant caught')
 else:raise AssertionError(name+' survived')
print('320 ordered identities verified')
# Compare the two measured project logs independently of report.py.
def logkeys(name):
 text=gzip.open(P/'evidence'/name,'rt').read()
 return {(f.split('/src/compiler/')[1],int(l),int(c),int(k)) for f,l,c,k in re.findall(r'^(.+?):(\d+):(\d+): error TS(\d+):',text,re.M)}
comparison=json.loads((P/'comparison.json').read_text()); main=logkeys('main-project.log.gz');topic=logkeys('topic-project.log.gz')
def compare(c):
 assert c['baseline']==320 and c['mainProject']==len(main) and c['topicProject']==len(topic)
 assert c['baselineAbsentInTopic']==len(set(expected)-topic)
 assert c['pairedMainAbsentInTopic']==len(main-topic)
 assert c['newDiagnostics']==len(topic-set(expected))
 assert c['directDriverClearances'] is None
 assert len(c['rows'])==320
 for r in c['rows']:
  key=(r['file'],r['line'],r['column'],r['code'])
  assert r['topicProject']==('remaining' if key in topic else 'absent')
compare(comparison)
for name,mutate in [('clearance-total',lambda c:c.__setitem__('baselineAbsentInTopic',320)),('driver-false-green',lambda c:c.__setitem__('directDriverClearances',320)),('site-status',lambda c:c['rows'][0].__setitem__('topicProject','remaining'))]:
 bad=copy.deepcopy(comparison);mutate(bad)
 try:compare(bad)
 except AssertionError:print(name+' mutant caught')
 else:raise AssertionError(name+' survived')
print('paired 319 -> 67 comparison verified')
