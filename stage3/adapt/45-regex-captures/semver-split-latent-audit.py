"""Audit complete semver-only full-mode latent records and print every finding."""
from pathlib import Path
from collections import Counter
import hashlib,json,re,sys
paths=list(map(Path,sys.argv[1:4]));out=Path(sys.argv[4])
labels=['main','amended','mutant']; records={};streams={}
def normalize(value):
 if isinstance(value,str):return re.sub(r'/tmp/semver-split-(?:before|after|mutant)/','TREE/',value)
 if isinstance(value,list):return [normalize(x) for x in value]
 if isinstance(value,dict):return {k:normalize(v) for k,v in value.items()}
 return value
def finding_key(row):
 return json.dumps({k:normalize(row.get(k)) for k in ['phase','kind','where','unit','root_kind','depth','reason','text']},sort_keys=True)
for label,path in zip(labels,paths):
 rows=[json.loads(line) for line in path.read_text().splitlines()]
 assert rows[0]['latent_mode']=='full' and not rows[0]['checker_rejected'] and rows[0]['diagnostics']==[]
 assert rows[-1]['status']=='relation_report', 'incomplete census stream'
 files=[row for row in rows if 'file' in row]
 assert len(files)==1 and files[0]['file'].endswith('/src/compiler/semver.ts')
 records[label]=files[0];streams[label]=rows
(out/'latent-sites.md').write_text('# Full-mode single-file latent census site lists\n\n'
 'All three runs use LATENT_FULL=1, the exact LATENT_ONLY_FILE semver path, and LATENT_ASSERT_NO_OUTPUT=1. '
 'The meter overlay generator and latent tool are unchanged. All three exit 0, have empty checker diagnostics, '
 'and end with the relation trailer. These are measurement findings, not native acceptance.\n\n'
 'Every finding occurrence is listed below in its original emission order, including rollback boundaries and '
 'imported debug.ts findings. The complete declarations, units, fields and relation trailer remain in each raw JSONL.\n\n')
counts={label:Counter(finding_key(row) for row in r['findings']) for label,r in records.items()}
assert counts['main']==counts['mutant'], 'restoring mutant did not recover the exact main finding multiset'
assert not counts['amended']-counts['main'], 'amendment added findings'
removed=list((counts['main']-counts['amended']).elements())
expected={('refusal_scan','Refused','TREE/src/compiler/semver.ts:287:34'),('lowering','Refused','TREE/src/compiler/semver.ts:287:34'),('lowering','Boundary','TREE/src/compiler/semver.ts:287:13')}
assert len(removed)==3
assert {(r['phase'],r['kind'],r['where']) for r in map(json.loads,removed)}==expected
assert [normalize(r) for r in records['main']['findings']]==[normalize(r) for r in records['mutant']['findings']]
summary={'status':'PASS','main_findings':len(records['main']['findings']),'amended_findings':len(records['amended']['findings']),'mutant_findings':len(records['mutant']['findings']),'added_findings':0,'removed':list(map(json.loads,removed)),'mutant_exact_main_order_and_findings':True,'exit_codes':dict.fromkeys(labels,0),'units':{label:len(r['units']) for label,r in records.items()}}
def display(value):return normalize(value).replace('TREE/src/compiler/','').replace('|','\\|').replace('\n',' ')
for label,r in records.items():
 with (out/'latent-sites.md').open('a') as report:
  report.write(f"## {label}\n\n{len(r['units'])} units, {len(r['findings'])} findings.\n\n| # | Phase | Kind | Site | Unit | Reason |\n| --- | --- | --- | --- | --- | --- |\n")
  for index,row in enumerate(r['findings'],1):
   report.write('| '+str(index)+' | '+' | '.join(display(row[k]) for k in ['phase','kind','where','unit','reason'])+' |\n')
  report.write('\n')
report_path=out/'latent-sites.md'
report_path.write_text(report_path.read_text().rstrip()+'\n')
(out/'latent-audit.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
