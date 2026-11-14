from pathlib import Path
import json,sys
s=Path('/workspace/scratch/diag-read-view-2');raw=s/'raw'
files=sys.argv[1:] or ['program.ts','utilitiesPublic.ts','utilities.ts','types.ts']
report=json.loads((raw/'pair-candidate.json').read_text());problems=[];summary={}
for file in files:
 row=report[file]
 for side in ['main','candidate']:
  result=json.loads((raw/(side+'-'+file+'-census.result.json')).read_text());stream=[json.loads(x) for x in (raw/(side+'-'+file+'.jsonl')).read_text().splitlines()]
  if result['exit']!=0 or result['seconds']>6000 or stream[0]['latent_mode']!='full' or stream[0]['module_format']!='CommonJSBundle' or stream[-1]['status']!='relation_report':problems.append(side+' '+file+' did not complete required full census')
  if result['extra'].get('LATENT_FULL')!='1' or result['extra'].get('LATENT_ASSERT_NO_OUTPUT')!='1' or result['extra'].get('CENSUS_MODULE_FORMAT')!='CommonJSBundle' or result['extra'].get('LATENT_ONLY_FILE')!=str(s/(side+'-tree')/'src/compiler'/file):problems.append(side+' '+file+' flags differ')
 if row['new']:problems.append(file+' has new findings')
 if file=='utilities.ts':
  sites={tuple(x['mappedSite']) for x in row['changed']}
  expected={('refusal_scan','Refused','src/compiler/utilities.ts:8808:39'),('refusal_scan','Refused','src/compiler/utilities.ts:8808:44')}
  if sites!=expected or len(row['changed'])!=2 or not all(x.get('acceptedByUser') and x.get('classification')=='changed: target type renamed, pre-existing refusal' for x in row['changed']):problems.append(file+' changed findings differ from the exact accepted pair')
 elif row['changed']:problems.append(file+' has changed findings')
 if file=='program.ts':
  targets={x['where'].rsplit('/',1)[-1] for x in row['removed'] if x['phase']=='refusal_scan' and x['kind']=='Refused'}
  if not {'program.ts:2796:16','program.ts:2829:16'}<=targets:problems.append('program targets were not both removed')
 summary[file]={'new':len(row['new']),'changed':len(row['changed']),'removed':len(row['removed']),'acceptedChanged':sum(bool(x.get('acceptedByUser')) for x in row['changed'])}
result={'passed':not problems,'summary':summary,'problems':problems}
(raw/'census-gate.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
sys.exit(bool(problems))
