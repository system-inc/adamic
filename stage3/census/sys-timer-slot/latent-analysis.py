"""Deduplicate full census events and preserve changed checker eligibility."""
import pathlib,json,collections,hashlib
repo=pathlib.Path('/workspace/adamic');scratch=pathlib.Path('/workspace/scratch/sys-timer-slot');out=repo/'stage3/census/sys-timer-slot'
result={};keys={};texts={}
for label in ['main','A','B']:
 rows=[json.loads(l) for l in (scratch/(label+'-latent.jsonl')).read_text().splitlines() if l];header=rows[0];record=rows[1];sites={}
 for f in record['findings']:
  if f['kind'] not in ['Refused','NotYet']:continue
  where=f['where'].replace(str(scratch/label)+'/','');key=(where,f['kind'],f['reason'],f['phase'])
  sites.setdefault(key,{'where':where,'kind':f['kind'],'reason':f['reason'],'phase':f['phase'],'units':[]})['units'].append(f['unit'].replace(str(scratch/label)+'/',''))
 for row in sites.values():row['units']=sorted(set(row['units']))
 keys[label]=set(sites);texts[label]=sites
 result[label]={'checkerRejected':header['checker_rejected'],'diagnostics':header['diagnostics'],'units':len(record['units']),'unitStatuses':dict(collections.Counter(u['status'] for u in record['units'])),'rawEventKinds':dict(collections.Counter(f['kind'] for f in record['findings'])),'distinctSites':len(sites),'sites':list(sites.values()),'splitCheckerBodies':[u for u in record['units'] if u['status']=='split_checker_body']}
for label in ['A','B']:
 result[label]['added']=[texts[label][k] for k in sorted(keys[label]-keys['main'])];result[label]['removed']=[texts['main'][k] for k in sorted(keys['main']-keys[label])];result[label]['common']=len(keys[label]&keys['main'])
emitters=[]
for file,needle in [('internal/lower/any_relations.go','l.refusal.deferred = l.notYet(node, "a member typed any'),('internal/lower/invariance.go','what = "a function taking "'),('internal/lower/invariance.go','return &Refused{Where: l.program.Where(node), What: what, Fix: fix}'),('internal/lower/expression.go','return 0, l.notYet(node, "a value of type "')]:
 source=(repo/file).read_text().splitlines();matches=[{'file':file,'line':i+1,'source':line.strip(),'sourceSha256':hashlib.sha256((repo/file).read_bytes()).hexdigest()} for i,line in enumerate(source) if needle in line];assert matches;(emitters.extend(matches))
result['emitters']=emitters
(out/'latent-sites.json').write_text(json.dumps(result,indent=2)+'\n')
lines=['Full latent site comparison','','| Tree | Checker diagnostics | Units | Distinct NotYet/Refused sites | Common with main | New | Not observed from main |','| --- | --- | --- | --- | --- | --- | --- |']
for label in ['main','A','B']:
 r=result[label];lines.append('| '+label+' | '+str(len(r['diagnostics']))+' | '+str(r['units'])+' | '+str(r['distinctSites'])+' | '+str(r.get('common','N/A'))+' | '+str(len(r.get('added',[])))+' | '+str(len(r.get('removed',[])))+' |')
lines+=['','Deduplication key is (location, kind, exact reason, phase), matching the cited core report. Boundary events are not compiler stops. Raw streams retain every event and unit. Main has zero checker diagnostics despite the preserved legacy measurement label. A has one split_checker_body; B has four. Deferred generic units are not proven clean.','','The timer contract produces one observed main stop, not four separate declaration-token stops: `src/compiler/sys.ts:384:79`, NotYet / refusal_scan, `a member typed any ((0)) related to string at the argument site`. The System declarations at 1447 and 1448 have no independent recorded site. The emission line is:']
for e in emitters:
 if 'any_relations' in e['file']:lines+=['','`'+e['file']+':'+str(e['line'])+'`:', '```go',e['source'],'```']
lines+=['','That callback stop is absent in both candidate streams. This is not a successful-lowering claim: the registering function at sys.ts:383 is checker-split in both candidates. A additionally records these new sites:']
for x in result['A']['added']:lines+=['','`'+x['where']+'`, '+x['kind']+' / '+x['phase']+': `'+x['reason']+'`.']
lines+=['','The two Timeout-to-unknown findings at sys.ts:1502 and 1600 are direct new clearTimeout-contract refusals. Stock tsc accepts these method assignments by bivariance; Adamic deliberately refuses the wider input. The message and raising return are:']
for e in emitters:
 if 'invariance' in e['file']:lines+=['','`'+e['file']+':'+str(e['line'])+'`:', '```go',e['source'],'```']
lines+=['','The new any-related initializer finding at sys.ts:1519:36 is collateral scanner exposure, not a new timer declaration token. A retains 122 main sites. B records no new site, but its getNodeSystem body is checker-split by the two failed generic timer assignments; the lack of A\'s new cancellation refusals in B does not prove them safe. B also splits the two zero-parameter callback registration bodies and the dynamic polling registration body, so its 21 absent observations cannot be credited as fixes.','','The private `pollScheduled: any` finding at sys.ts:470:9 remains under both A and B: `a value of type any`, raised by `internal/lower/expression.go:53`. Retyping the public slot does not retype private storage. All 123/125/102 sites, their recording units, and every added/absent row are in latent-sites.json.']
(out/'LATENT.md').write_text('\n'.join(lines)+'\n')
