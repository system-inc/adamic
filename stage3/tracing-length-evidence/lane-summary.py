from pathlib import Path
from run import raw,root
import json,re,collections
assert json.loads((raw/'whole-lane.json').read_text())['exit']==0
rows=[json.loads(s) for s in (raw/'whole-lane.log').read_text().splitlines() if s.startswith('{')]
expected=set()
for p in (root/'stage3/lane').glob('*.go'):expected.update(re.findall(r'func (TestShard_\w+)\(',p.read_text()))
passed=[x['Test'] for x in rows if x.get('Action')=='pass' and x.get('Test','').startswith('TestShard_')]
assert len(passed)==len(set(passed))==len(expected);assert set(passed)==expected;assert not any(x.get('Action')=='fail' for x in rows)
units=json.loads((root/'stage3/lane/shards.json').read_text())['units'];assert len(units)==len(expected)
records=[json.loads(p.read_text()) for p in sorted((raw/'lane-counts').glob('*.json'))];assert records
coverage={u['name']:sum(any('/'+u['name']+'/' in arg for arg in p['argv']) for p in records) for u in units};assert all(coverage.values())
def total(records):
 counts={s:collections.Counter() for s in ['popCall','popWrite','popAllCall','popAllWrite']}
 for p in records:
  for site,hist in p['counts'].items():counts[site].update({int(k):v for k,v in hist.items()})
 return {site:dict(sorted(hist.items())) for site,hist in counts.items()}
counts=total(records);controls={}
for directory in ['before-control-counts','after-control-counts']:
 rs=[json.loads(p.read_text()) for p in (raw/directory).glob('*.json')];hist=total(rs)
 assert sum(hist['popWrite'].values())>0 and sum(hist['popAllWrite'].values())>0
 assert hist['popCall'].get(0,0)>0 and hist['popWrite'].get(0,0)==0
 controls[directory]=hist
assert controls['before-control-counts']==controls['after-control-counts']
summary={'wholeLaneExit':0,'units':len(passed),'oracleCountsCheckedByEachGoUnit':{k:sum(u['expect'][k] for u in units) for k in ['pass','fail','skip']},'processRecords':len(records),'unitCounterProcessCoverage':coverage,'histograms':counts,'emptyPopWrites':counts['popWrite'].get(0,0),'controls':controls,'instrumentation':'Private emitted bundle copies with VM-aware preload. Measurement-only Go overlay selects a private product and carries preload through toolchain.Env. Production source and normal built trees contain no counters.'}
(raw/'lane-summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps({k:v for k,v in summary.items() if k!='unitCounterProcessCoverage'},indent=2))
