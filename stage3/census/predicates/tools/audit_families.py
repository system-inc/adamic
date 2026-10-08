"""Record exact predicate diagnostic families and before/after eligibility."""
import json,collections,gzip,pathlib
root=pathlib.Path(__file__).resolve().parent.parent;report=json.loads((root/'REPORT.json').read_text());import subprocess
original=json.loads(subprocess.check_output(['git','show','70456b7:stage3/census/latent/REPORT.json'],cwd=root,text=True))
LABEL=report['measurement']
def body(f):
 prefix=f['where']+': '
 return f['text'][len(prefix):] if f['where'] and f['text'].startswith(prefix) else f['text']
families={'predicate_declarations':set(),'predicate_arguments':set(),'checked_view_fields':set()}
for r in report['runs']:
 for f in r['findings']:
  if f['kind']=='Refused' and '(adamic/no-type-predicate)' in f['text']:
   key='predicate_arguments' if f['reason'].startswith('an unproven predicate argument for parameter ') else 'predicate_declarations'
   families[key].add(body(f))
  elif f['kind']=='NotYet' and f['reason'].startswith('checked view field '):families['checked_view_fields'].add(body(f))
# Recount by exact diagnostic message text, including the complete repair text.
counts={r['name']:{family:sum(body(f) in messages for f in r['findings']) for family,messages in families.items()} for r in report['runs']}
assert counts['cumulative']['predicate_declarations']==581
assert counts['predicates']==dict(predicate_declarations=518,predicate_arguments=451,checked_view_fields=246)
baseline_match={}
for run in report['runs'][:6]:
 old=next(r for r in original['runs'] if r['name']==run['name'])
 key=lambda f:(f['kind'],f['where'],f['reason'],f['text'])
 assert {key(f) for f in run['findings']}=={key(f) for f in old['findings']},run['name']
 baseline_match[run['name']]=True
before,after=[next(r for r in report['runs'] if r['name']==n) for n in ['cumulative','predicates']]
raw={}
for name in ['cumulative','predicates']:
 with gzip.open(root/'data'/(name+'.jsonl.gz'),'rt') as f:raw[name]=[json.loads(l) for l in f]
eligible=lambda records:{u['where'] for r in records[1:] for u in r['units'] if u['status']!='skipped_checker_body'}
be,ae=eligible(raw['cumulative']),eligible(raw['predicates']);common=be&ae
common_counts={}
for name in ['cumulative','predicates']:
 fs={(f['kind'],f['where'],f['reason'],f['text']) for r in raw[name][1:] for f in r['findings'] if f['unit'] in common}
 common_counts[name]={family:sum(body(dict(kind=k,where=w,reason=r,text=t)) in messages for k,w,r,t in fs) for family,messages in families.items()}
result=dict(measurement=LABEL,definition=report['count_definition'],families={k:sorted(v) for k,v in families.items()},counts=counts,baseline_findings_match_original=baseline_match,before_configuration='cumulative',after_configuration='predicates',common_eligibility=dict(common_units=len(common),newly_eligible=len(ae-be),no_longer_eligible=len(be-ae),family_counts=common_counts),scope_limits=['Checker-rejected programs only; no native compilation claim.','The after configuration includes the feature branch and its main/dependency integrations.','The after refusal scans panic in builder.ts and transformers/declarations/diagnostics.ts; their later refusal syntax is not exhaustively scanned.','First lowering error per unit; generic declarations receive no invented specialization.'],after_panics=[f for f in after['findings'] if f['kind']=='panic'])
(root/'FAMILIES.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(dict(counts=counts,common_eligibility=result['common_eligibility'],original_baselines_match=True),indent=2))
