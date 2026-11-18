from pathlib import Path
import json,re,collections,hashlib,sys
W=Path('/workspace');run=W/'family5-run/base-tsc-entry-r1';out=W/'family5-analysis';out.mkdir(exist_ok=True)
status=json.loads((run/'status.json').read_text()) if (run/'status.json').exists() else None
rows=[json.loads(l) for l in (run/'latent.jsonl').read_text().splitlines()];complete=bool(status and status['exit_code']==0 and rows[-1].get('status')=='relation_report')
pattern=re.compile(r'^optional property (\S+) in (.+) absent from structural source (.+), which can hide fields$')
fs=[{'record_file':r.get('file'),'record_status':r.get('status'),'finding':f} for r in rows for f in r.get('findings',[]) if pattern.fullmatch(f['reason'])]
body=[x for x in fs if x['finding'].get('phase')!='tail'];unique={tuple(x['finding'][k] for k in ['kind','where','reason','text']):x['finding'] for x in body};reasons=collections.Counter(f['reason'] for f in unique.values());expected=json.loads((W/'family5-meter-groups.json').read_text());lookup={g['reason']:g for g in expected}
sites=[]
for i,f in enumerate(sorted(unique.values(),key=lambda f:(f['where'],f['reason'])),1):
 p,t,s=pattern.fullmatch(f['reason']).groups();owners=[x['finding'].get('unit') for x in body if all(x['finding'][k]==f[k] for k in ['kind','where','reason','text'])];sites.append({'id':f'S{i:03d}','group':lookup.get(f['reason'],{}).get('id','UNEXPECTED'),'S':s,'T':t,'P':p,**f,'owning_units':sorted(set(u for u in owners if u)),'event_count':len(owners)})
groups=[]
for g in expected:
 group={**g,'regenerated_sites':reasons[g['reason']],'site_ids':[s['id'] for s in sites if s['group']==g['id']]};groups.append(group)
report={'complete':complete,'status':status,'header':rows[0],'raw_family_events':len(fs),'body_family_events':len(body),'tail_family_events':len(fs)-len(body),'meter_unique_sites':len(unique),'distinct_positions':len({f['where'] for f in unique.values()}),'distinct_reasons':len(reasons),'definition':'body findings deduplicated by (kind,where,reason,text), exactly meter/report.py; tail retained separately','expected_unique_sites':131,'expected_reasons':45,'group_count_mismatches':[g for g in groups if g['meter_sites']!=g['regenerated_sites']],'unexpected_reasons':[r for r in reasons if r not in lookup]}
(out/'family-events.jsonl').write_text(''.join(json.dumps(f)+'\n' for f in fs));(out/'sites.json').write_text(json.dumps(sites,indent=2)+'\n');(out/'groups.json').write_text(json.dumps(groups,indent=2)+'\n');(out/'run-summary.json').write_text(json.dumps(report,indent=2)+'\n')
print('complete',complete,'events',len(fs),'meter unique',len(unique),'positions',report['distinct_positions'],'reasons',len(reasons),'mismatched groups',len(report['group_count_mismatches']))
if not complete:sys.exit(2)
