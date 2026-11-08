"""Independently reconcile grouped reasons, site counts and unit stop ordering."""
import collections,gzip,json
from pathlib import Path
here=Path(__file__).resolve().parent
report=json.loads((here/'record-findings.json').read_text());summary=report['summary'];groups=report['groups']
with gzip.open(here/'record-findings-observations.jsonl.gz','rt') as stream:rows=[json.loads(line) for line in stream]
byreason=collections.defaultdict(list);byunit=collections.defaultdict(list)
for row in rows:byreason[row['reason']].append(row);byunit[row['owner_file'],row['owner_unit']].append(row)
assert len(byunit)==3666
assert len(groups)==len(byreason)==summary['exact_reason_groups']
assert len(rows)==summary['findings_including_bookkeeping']
for group in groups:
 items=byreason[group['reason']];marked=[r for r in items if r['record_shaped']]
 assert group['count']==len(items)
 assert group['locations']==sorted({r['where'] for r in items if r['where']})
 assert group['record_shaped_count']==len(marked)
 assert group['record_shaped']==bool(marked)
 assert group['record_shaped_locations']==sorted({r['where'] for r in marked if r['where']})
 for ex in group['examples']:
  path,line,column=ex['where'].rsplit(':',2);a=Path(path).read_text().splitlines();n=int(line)
  assert ex['source_text']=='\n'.join(a[max(0,n-2):min(len(a),n+3)])
 assert len(group['examples'])==min(2,len(group['record_shaped_locations']))
marked=[r for r in rows if r['record_shaped']]
assert len(marked)==summary['record_shaped_findings']
assert len({r['where'] for r in marked})==summary['record_shaped_distinct_sites']
assert sum(r['kind']!='Boundary' for r in marked)==summary['record_shaped_causal_findings']
first=[];behind=[];causal_first=[];causal_behind=[]
for key,items in byunit.items():
 order=[r for r in items if r['kind']!='Boundary']
 assert order and items[0]['kind']!='Boundary'
 if items[0]['record_shaped']:first.append(key)
 elif any(r['record_shaped'] for r in items):behind.append(key)
 if order[0]['record_shaped']:causal_first.append(key)
 elif any(r['record_shaped'] for r in order):causal_behind.append(key)
assert len(first)==summary['record_shaped_first_stop_units']
assert len(behind)==summary['record_shaped_only_behind_another_stop_units']
assert len(causal_first)==summary['record_shaped_causal_first_stop_units']
assert len(causal_behind)==summary['record_shaped_causal_only_behind_another_stop_units']
assert len(first)+len(behind)==summary['record_shaped_any_units']
assert len(causal_first)+len(causal_behind)==summary['record_shaped_causal_units']
assert {tuple(x.values()) for x in report['units']['first_stop']}==set(first)
assert {tuple(x.values()) for x in report['units']['only_behind_another_stop']}==set(behind)
assert {tuple(x.values()) for x in report['units']['causal_first_stop']}==set(causal_first)
assert {tuple(x.values()) for x in report['units']['causal_only_behind_another_stop']}==set(causal_behind)
assert summary['unresolved_payloads']==0
assert summary['recovered_payloads']==len(report['recovery_validation'])==904
assert all(r['boundary_metadata_and_sites_match'] for r in report['recovery_validation'])
old=json.loads((here/'summary.json').read_text())
assert old['counts']=={'views':59681,'record_literals':0,'dynamic_reads':68,'checked_object_calls':0}
assert old['observed_statuses']=={'lowered':5024,'stopped':3666}
validation={'exact_reason_recount_passed':True,'distinct_sites_passed':True,'source_examples_passed':True,'first_stop_order_passed':True,'all_stopped_payloads_present':True,'original_four_way_counts_unchanged':True,'summary':summary}
(here/'record-findings-validation.json').write_text(json.dumps(validation,indent=2)+'\n');print(json.dumps(validation))
