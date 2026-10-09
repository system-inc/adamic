#!/usr/bin/env python3
"""Independent location recount, snapshot hashes, and executing each real throwing replacement."""
import copy,gzip,json,re,subprocess,tempfile
from pathlib import Path
P=Path(__file__).resolve().parent;ROOT=P.parents[3]
rows=json.loads((P/'stops.json').read_text());fixtures=json.loads((P/'fixtures.json').read_text())
def check(rs, manifest=None):
 assert len(rs)==15 and [r['order'] for r in rs]==list(range(1,16))
 for r in rs:
  i=r['order'];text=gzip.open(P/f'evidence/stop-{i:02}.stderr.gz','rt').read();match=re.search(r'(?:adamic: )?(/[^\n]+?):(\d+):(\d+): (.*)',text)
  assert match and match[1].endswith('/'+r['file']) and (int(match[2]),int(match[3]),match[4])==(r['line'],r['column'],r['message'])
  assert r['exit']==1 and r['owner'] in ['compiler','library','adaptation']
  assert fixtures[i-1]['sameStopMessage'] and fixtures[i-1]['node']['exit']==0
  if i in [10,11,12,14,15]:assert r['owner']=='adaptation'
 if manifest is None:manifest=json.loads((P/'manifest.json').read_text())
 assert manifest['nativeExecution'] is False
 assert manifest['sourceNode']['sha256']==manifest['stock']['sha256'] and manifest['projects']==301
check(rows)
for name,mutate in [('missing-stop',lambda r:r.pop()),('wrong-line',lambda r:r[0].__setitem__('line',0)),('wrong-message',lambda r:r[0].__setitem__('message','green')),('false-owner',lambda r:r[11].__setitem__('owner','compiler'))]:
 bad=copy.deepcopy(rows);mutate(bad)
 try:check(bad)
 except AssertionError:print(name+' mutant caught')
 else:raise AssertionError(name+' survived')
bad=json.loads((P/'manifest.json').read_text());bad['nativeExecution']=True
try:check(rows,bad)
except AssertionError:print('false-native-green mutant caught')
else:raise AssertionError('false-native-green survived')
placeholder_results=[]
with tempfile.TemporaryDirectory(prefix='step31-throw-audit-') as temp:
 for i in range(1,16):
  patch=json.loads((P/f'evidence/patch-{i:02}.json').read_text());replacement=patch['replacement'];assert 'throw ' in replacement and patch['marker'] in replacement
  src=Path(temp)/f'{i:02}.a';src.write_text(replacement+'\n')
  command=['node','--disable-warning=ExperimentalWarning',str(ROOT/'oracle/node.mjs'),str(src)]
  expected=subprocess.run(command,capture_output=True,text=True);assert expected.returncode==70 and expected.stdout=='' and expected.stderr=='adamic: panic: '+patch['marker']+'\n',(i,expected)
  src.write_text(replacement.replace('throw ','void ',1)+'\n');mutant=subprocess.run(command,capture_output=True,text=True)
  assert mutant.returncode==0 and mutant.stdout=='' and mutant.stderr=='',(i,mutant)
  placeholder_results.append({'order':i,'nodeExit':70,'stderr':expected.stderr,'throwErasureExit':0,'caught':True})
(P/'placeholder-mutants.json').write_text(json.dumps(placeholder_results,indent=2)+'\n')
print('15 stop identities verified; five ledger/manifest mutants caught; 15 exact placeholders throw and 15 throw erasures caught')
